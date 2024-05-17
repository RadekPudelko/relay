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

	"pcfs/db"
	"pcfs/middleware"
	"pcfs/particle"
)

var particleToken string
var myDB *sql.DB
var taskLimit = 10

type CreateTaskRequest struct {
	SomId             string     `json:"som_id"`
	ProductId         int        `json:"product_id"`
	CloudFunction     string     `json:"cloud_function"`
	Argument          *string    `json:"argument,omitempty"`
	DesiredReturnCode *int       `json:"desired_return_code,omitempty"`
	ScheduledTime     *time.Time `json:"scheduled_time,omitempty"`
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

func CreateTask(dbConn *sql.DB, somId string, productId int, cloudFunction string, argument string, desiredReturnCode sql.NullInt32, scheduledTime time.Time) (int, error) {
	somKey, err := db.InsertOrUpdateSom(dbConn, somId, productId)
	if err != nil {
        return 0, fmt.Errorf("createTaskHandler: %w", err)
	}

	taskId, err := db.InsertTask(dbConn, somKey, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
        return 0, fmt.Errorf("CreateTask: %w", err)
	}

    return taskId, nil
}

// Queries for upto limit tasks in the db that are scheduled after scheduled time, starting from the id
// of the last task that was ran and wraps around if needed
func GetReadyTasks(myDB *sql.DB, id, limit int, scheduledTime time.Time) ([]int, error) {
    start := id+1
    taskIds, err := db.SelectTaskIds(myDB, db.TaskReady, &start, nil, &taskLimit, scheduledTime)
    if err != nil {
        return nil, fmt.Errorf("GetReadyTasks for %d onward: %w", id+1, err)
    }
    if len(taskIds) < taskLimit {
        start := id-1
        limit := taskLimit - len(taskIds)
        taskIds2, err := db.SelectTaskIds(myDB, db.TaskReady, nil, &start, &limit, scheduledTime)
        if err != nil {
            return nil, fmt.Errorf("GetReadyTasks for 1 to %d: %w", id-1, err)
        }
        taskIds = append(taskIds, taskIds2...)
    }
    return taskIds, nil
}

// TODO make backgroundTask sleep when there are no tasks, wake by new task post?
func backgroundTask() {
	var sem = make(chan int, 3)
	lastTaskId := 0
	for true {
		// Get ready tasks, starting from the lastTaskId, limited 1 per som
		// This implementation does not care about the order of tasks
		// To take into account order, would first need to get list of soms with ready tasks, then query the min for each
		taskIds, err := GetReadyTasks(myDB, lastTaskId, taskLimit, time.Now().UTC())
		if err != nil {
			log.Fatal("backgroundTask: ", err)
		}

		log.Printf("Loading %d ready tasks ids from the myDB\n", len(taskIds))
		if len(taskIds) == 0 {
			lastTaskId = 0
			time.Sleep(2 * time.Second)
			continue
		}
		nTasks := len(taskIds)
		lastTaskId = taskIds[nTasks-1]

		// TODO: Load additional requests in the background as tasks are processed
		for _, taskId := range taskIds {
			sem <- 1
			go func(id int) {
				processTask(id)
				<-sem
			}(taskId)
		}
	}
}

// TODO: Update the schedule time of the task if its been recently pinged and offline, ping fails or device is offile
func processTask(id int) {
	log.Println("processTask: process task ", id)
	task, err := db.SelectTask(myDB, id)
	if err != nil {
		log.Println("processTask:", err)
		return
	}
	// Consider pinging a som if its been more than n seconds since last check
	if !task.Som.LastOnline.Valid || time.Since(task.Som.LastOnline.Time) > 5*time.Minute {
		// Only ping a som if we have not pinged in n seconds
		if task.Som.LastPing.Valid && time.Since(task.Som.LastPing.Time) < 5*time.Minute {
			log.Printf("processTask: skipping task %d, som %s is not online\n", id, task.Som.SomId)
			return
		}
		log.Printf("processTask: pinging som %s\n", task.Som.SomId)
		online, err := particle.Ping(task.Som.SomId, task.Som.ProductId, particleToken)
		now := sql.NullTime{Time: time.Now(), Valid: true}
		if err != nil {
			log.Println("processTask:", err)
			err = db.UpdateSom(myDB, task.Som.Id, task.Som.ProductId, task.Som.LastOnline, now)
			if err != nil {
				log.Println("processTask: ", err)
			}
			return
		}
		if !online {
			log.Printf("processTask: som %s is offline\n", task.Som.SomId)
			err = db.UpdateSom(myDB, task.Som.Id, task.Som.ProductId, task.Som.LastOnline, now)
			// TODO: This and many places like this should never fail, so should the server crash here??
			if err != nil {
				log.Println("processTask: ", err)
			}
			return
		}
		err = db.UpdateSom(myDB, task.Som.Id, task.Som.ProductId, now, now)
		if err != nil {
			log.Println("processTask:", err)
			return
		}
	}

	log.Printf("processTask: runnning task %d\n", id)
	log.Printf("processTask: som %s is online\n", task.Som.SomId)
	// TODO: may want to get return value from function
	// TODO: may want to add some way to store error history in the database
	success, err := particle.CloudFunction(task.Som.SomId, task.Som.ProductId, task.CloudFunction, task.Argument, particleToken, task.DesiredReturnCode)
	fiveMinLater := time.Now().Add(5 * time.Minute)
	if err != nil {
		log.Println("processTask:", err)
		if task.Tries == 2 { // Task is considered failed on third attempt
			log.Printf("processTask task %d has failed due to exceeding max tries, err %v:\n", id, err)
			err = db.UpdateTask(myDB, id, task.ScheduledTime, db.TaskFailed, task.Tries+1)
		} else {
			err = db.UpdateTask(myDB, id, fiveMinLater, db.TaskReady, task.Tries+1)
		}
		if err != nil {
			log.Println("processTask:", err)
		}

		return
	}
	if !success {
		log.Printf("processTask task %d has failed due to mismatch in returned code\n", id)
		err = db.UpdateTask(myDB, id, task.ScheduledTime, db.TaskFailed, task.Tries+1)
	} else {
		log.Printf("processTask: task %d, success\n", id)
		err = db.UpdateTask(myDB, id, task.ScheduledTime, db.TaskComplete, task.Tries+1)
	}
	if err != nil {
		log.Println("processTask:", err)
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

	task, err := db.SelectTask(myDB, taskId)
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

	// TODO: validate the scheduled time
	scheduledTime := time.Now().UTC()
	if req.ScheduledTime != nil {
		scheduledTime = *req.ScheduledTime
        scheduledTime = scheduledTime.UTC()
	}
	argument := ""
	if req.Argument != nil {
		argument = *req.Argument
	}

    desiredReturnCode := sql.NullInt32{Int32: 0, Valid: false}
    if req.DesiredReturnCode != nil {
        desiredReturnCode = sql.NullInt32{Int32: int32(*req.DesiredReturnCode), Valid: true}
    }

    // TODO: I dont think the product id is strictly required, maybe I drop it
    taskId, err := CreateTask(myDB, req.SomId, req.ProductId, req.CloudFunction, argument, desiredReturnCode, scheduledTime)
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

func SetupDB(path string) (error) {
    var err error
	myDB, err = db.Connect(path)
	if err != nil {
        return fmt.Errorf("SetupDB: %w", err)
	}

	err = db.CreateTables(myDB)
	if err != nil {
        myDB.Close()
        return fmt.Errorf("SetupDB: %w", err)
	}
    return nil
}

func CloseDB() {
    myDB.Close()
}

func main() {
	fmt.Printf("Hello\n")
	var err error

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("main: Error loading .env file: %v", err)
	}

	// TODO: Test the token
	particleToken = os.Getenv("PARTICLE_TOKEN")
	if particleToken == "" {
		log.Fatalf("main: missing PARTICLE_TOKEN in .env file")
	}

    err = SetupDB("my.db3")
	if err != nil {
		log.Fatal("main: %w", err)
	}
	defer myDB.Close()

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
