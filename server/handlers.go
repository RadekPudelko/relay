package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"pcfs/db"
)

func HandleGetRoot() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			getRoot(w, r)
		},
	)
}

func HandleCreateTask(dbConn *sql.DB) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			createTaskHandler(dbConn, w, r)
		},
	)
}

func HandleGetTask(dbConn *sql.DB) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			getTaskHandler(dbConn, w, r)
		},
	)
}

// TODO: Rename this to something w/o verb
type CreateTaskRequest struct {
	SomId             string  `json:"som_id"`
	CloudFunction     string  `json:"cloud_function"`
	Argument          *string `json:"argument,omitempty"`
	DesiredReturnCode *int    `json:"desired_return_code,omitempty"`
	// TODO time comes in a as a string need to parse
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
}

func (p CreateTaskRequest) String() string {
	str := fmt.Sprintf("som: %s, function: %s", p.SomId, p.CloudFunction)
	if p.Argument != nil {
		str += fmt.Sprintf(", argument: %s", *p.Argument)
	}
	if p.DesiredReturnCode != nil {
		str += fmt.Sprintf(", desired return code: %d", *p.DesiredReturnCode)
	}
	return str
}

// TODO: Want to add some sort of id to these logs so that I can know whats going on if there are multiple requests at once
func createTaskHandler(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	if req.SomId == "" || req.CloudFunction == "" {
		log.Println("createTaskHandler: Atleast one field in the post payload was blank or invalid")
		http.Error(w, "som_id and cloud_function are required fields",
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

	desiredReturnCode := sql.NullInt64{Int64: 0, Valid: false}
	if req.DesiredReturnCode != nil {
		desiredReturnCode = sql.NullInt64{Int64: int64(*req.DesiredReturnCode), Valid: true}
	}

	taskId, err := CreateTask(dbConn, req.SomId, req.CloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		log.Println("createTaskHandler:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("createTaskHandler: new task created, id: %d scheduled for %s\n", taskId, scheduledTime.String())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
    // TODO: Send json?
	io.WriteString(w, fmt.Sprintf("%d", taskId))
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, HTTP!\n")
}

func getTaskHandler(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
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

func CreateTask(dbConn *sql.DB, somId string, cloudFunction string, argument string, desiredReturnCode sql.NullInt64, scheduledTime time.Time) (int, error) {
	somKey, err := db.InsertOrUpdateSom(dbConn, somId)
	if err != nil {
		return 0, fmt.Errorf("createTaskHandler: %w", err)
	}

	taskId, err := db.InsertTask(dbConn, somKey, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		return 0, fmt.Errorf("CreateTask: %w", err)
	}

	return taskId, nil
}

