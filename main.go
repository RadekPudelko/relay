package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"pcfs/particle"
	"pcfs/db"
    "pcfs/middleware"
)

var particleToken string 
var dbConn *sql.DB

type CreateTaskRequest struct {
    SomId  string `json:"som_id"`
    ProductId int `json:"product_id"`
    CloudFunction string `json:"cloud_function"`
    Argument *string `json:"argument,omitempty"`
    DesiredReturnCode *int `json:"desired_return_code,omitempty"`
    ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
}

func (p CreateTaskRequest) String() string {
    str := fmt.Sprintf("som: %s, product %d, function: %s", p.SomId, p.ProductId, p.CloudFunction)
    if p.Argument != nil {
        str += fmt.Sprintf(", argument: %s", *p.Argument)
    }
    if p.DesiredReturnCode != nil {
        str += fmt.Sprintf(", desired return code: %d", *p.DesiredReturnCode)
    }
    return str
}

func backgroundTask() {
    lastTaskId := 0
    for true {
        // Get ready tasks, starting from the lastTaskId, limited 1 per som 
        // This implementation does not care about the order of tasks
        // To take into account order, would first need to get list of soms with ready tasks, then query the min for each
        taskIds, err := db.SelectTaskIds(dbConn, db.TaskReady, lastTaskId, time.Now(), 10)
        if err != nil {
            log.Fatal("backgroundTask: ", err)
        }

        log.Printf("Loading %d ready tasks ids from the dbConn\n", len(taskIds))
        if len(taskIds) == 0 {
            lastTaskId = 0
            time.Sleep(2*time.Second)
            continue
        }
        nTasks := len(taskIds)
        lastTaskId = taskIds[nTasks-1]

        for _, taskId := range taskIds {
            task, err := db.SelectTask(dbConn, taskId)
            if err != nil {
                log.Println("backgroundTask:", err)
                continue
            }
            // Consider pinging a som if its been more than n seconds since last check
            if !task.Som.LastOnline.Valid || time.Since(task.Som.LastOnline.Time) > 5 * time.Minute {
                // Only ping a som if we have not pinged in n seconds
                if task.Som.LastPing.Valid && time.Since(task.Som.LastPing.Time) < 5 * time.Minute {
                    log.Printf("backgroundTask: skipping task %d, som %s is not online\n", taskId, task.Som.SomId)
                    continue
                }
                log.Printf("backgroundTask: pinging som %s\n", task.Som.SomId)
                online, err := particle.Ping(task.Som.SomId, task.Som.ProductId, particleToken)
                now := sql.NullTime{Time: time.Now(), Valid: true}
                if err != nil {
                    log.Println("backgroundTask:", err)
                    err = db.UpdateSom(dbConn, task.Som.Id, task.Som.ProductId, task.Som.LastOnline, now)
                    if err != nil {
                        log.Println("backgroundTask: ", err)
                    }
                    continue
                }
                if !online {
                    log.Printf("backgroundTask: som %s is offline\n", task.Som.SomId)
                    err = db.UpdateSom(dbConn, task.Som.Id, task.Som.ProductId, task.Som.LastOnline, now)
                    if err != nil {
                        log.Println("backgroundTask: ", err)
                    }
                    continue
                }
                err = db.UpdateSom(dbConn, task.Som.Id, task.Som.ProductId, now, now)
                if err != nil {
                    log.Println("backgroundTask:", err)
                }
            }

            log.Printf("backgroundTask: runnning task %d\n", taskId)
            log.Printf("backgroundTask: som %s is online\n", task.Som.SomId)
            // TODO: may want to get return value from function
            // TODO: may want to add some way to store error history in the database
            success, err := particle.CloudFunction(task.Som.SomId, task.Som.ProductId, task.CloudFunction, task.Argument, particleToken, task.DesiredReturnCode)
            fiveMinLater := time.Now().Add(5 * time.Minute)
            if err != nil {
                log.Println("backgroundTask:", err)
                if task.Tries == 2 { // Task is considered failed on third attempt
                    log.Printf("backgroundTask task %d has failed due to exceeding max tries, err %v:\n", taskId, err)
                    err = db.UpdateTask(dbConn, taskId, task.ScheduledTime, db.TaskFailed, task.Tries+1)
                } else {
                    err = db.UpdateTask(dbConn, taskId, fiveMinLater, db.TaskReady, task.Tries+1)
                }
                if err != nil {
                    log.Println("backgroundTask:", err)
                }
                continue
            }
            if !success {
                log.Printf("backgroundTask task %d has failed due to mismatch in returned code\n", taskId)
                err = db.UpdateTask(dbConn, taskId, task.ScheduledTime, db.TaskFailed, task.Tries+1)
            } else {
                log.Printf("backgroundTask: task %d, success\n", taskId)
                err = db.UpdateTask(dbConn, taskId, task.ScheduledTime, db.TaskComplete, task.Tries+1)
            }
            if err != nil {
                log.Println("backgroundTask:", err)
                continue
            }
        }
        time.Sleep(2*time.Second)
    }
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
    taskIdStr := r.PathValue("id")

	if taskIdStr == "" {
        log.Println("getTaskHandler: missing task id in url: ", r.URL.Path)
        http.Error(w, "Missing task id", http.StatusBadRequest)
        return
    }

    taskId, err := strconv.Atoi(taskIdStr)
    if err != nil {
        log.Println("getTaskHandler: invalid task id: ", taskIdStr)
        http.Error(w, "Invalid task id", http.StatusBadRequest)
        return
    }

    log.Printf("getTaskHandler: request for task %d\n", taskId)

    task, err := db.SelectTask(dbConn, taskId)
    if err != nil {
        log.Println("getTaskHandler: ", err)
        http.Error(w, "Error in getting task", http.StatusInternalServerError)
        return
    }

    if task == nil {
        msg := fmt.Sprintf("getTaskHandler: task %d does not exist", taskId)
        log.Println(msg)
        http.Error(w, msg, http.StatusBadRequest)
        return
    }

    jsonData, err := json.Marshal(task)
    if err != nil {
        log.Println("getTaskHandler: json.Marshal: ", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(jsonData)
}

// TODO: Want to add some sort of id to these logs so that I can know whats going on if there are multiple requests at once
func createTaskHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        log.Println("createTaskHandler: io.ReadAll:", err)
        http.Error(w, "Error reading request body", http.StatusBadRequest)
        return
    }

    var req CreateTaskRequest
    err = json.Unmarshal(body, &req)
    if err != nil {
        log.Println("createTaskHandler: json.Unmarshal:", err)
        log.Println("request body:", string(body))
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("createTaskHandler: received request body: %s\n", req)
    if req.SomId == "" || req.ProductId == 0 || req.CloudFunction == "" {
        log.Println("createTaskHandler: Atleast one field in the post payload was blank or invalid")
        http.Error(w, "som_id, product_id and cloud_function are required fields", 
        http.StatusUnprocessableEntity)
        return
    }

    somKey, err := db.InsertOrUpdateSom(dbConn, req.SomId, req.ProductId)
    if err != nil {
        log.Println("createTaskHandler:", err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // TODO: validate the scheduled time
    scheduledTime := time.Now()
    if req.ScheduledTime != nil {
        scheduledTime = *req.ScheduledTime
    }
    argument := ""
    if req.Argument != nil {
        argument = *req.Argument
    }

    taskId, err := db.InsertTask(dbConn, somKey, req.CloudFunction, argument, req.DesiredReturnCode, scheduledTime)
    if err != nil {
        log.Println("createTaskHandler:", err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    log.Printf("createTaskHandler: new task created, id: %d scheduled for %s\n", taskId, scheduledTime.String())

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    io.WriteString(w, fmt.Sprintf("%d", taskId))
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
    fmt.Printf("Hello\n")
    var err error

	err = godotenv.Load(".env")
	if err != nil {
        log.Fatalf("main: Error loading .env file: %v", err)
	}

	particleToken = os.Getenv("PARTICLE_TOKEN")
    if particleToken == "" {
        log.Fatalf("main: missing PARTICLE_TOKEN in .env file")
    }
    // TODO: Test the token

    // TODO: What is a database pool?
    dbConn, err = db.Connect("my.db3")
 	if err != nil {
        log.Fatal("main: %w", err)
	}   
    defer dbConn.Close()

    err = db.SetupTables(dbConn)
 	if err != nil {
        log.Fatal("main: %w", err)
	}   

    go backgroundTask()

    router := http.NewServeMux()
	router.HandleFunc("GET /{$}", getRoot)
    router.HandleFunc("POST /api/tasks", createTaskHandler)
    router.HandleFunc("GET /api/tasks/{id}", getTaskHandler)

	err = http.ListenAndServe(":8080", middleware.Logging(router))
    if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed\n")
	} else if err != nil {
		log.Printf("error starting server: %s\n", err)
		os.Exit(1)
    }
}

