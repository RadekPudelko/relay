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

	"relay/internal/models"
)


// TODO: Rename this to something w/o verb
type CreateRelayRequest struct {
	DeviceId          string  `json:"device_id"`
	CloudFunction     string  `json:"cloud_function"`
	Argument          *string `json:"argument,omitempty"`
	DesiredReturnCode *int    `json:"desired_return_code,omitempty"`
	// TODO time comes in a as a string need to parse
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
}

func (p CreateRelayRequest) String() string {
	str := fmt.Sprintf("device: %s, function: %s", p.DeviceId, p.CloudFunction)
	if p.Argument != nil {
		str += fmt.Sprintf(", argument: %s", *p.Argument)
	}
	if p.DesiredReturnCode != nil {
		str += fmt.Sprintf(", desired return code: %d", *p.DesiredReturnCode)
	}
	return str
}

func HandleGetRoot() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			getRoot(w, r)
		},
	)
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, HTTP!\n")
}

func HandleGetRelay(dbConn *sql.DB) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handleGetRelay(dbConn, w, r)
		},
	)
}

func handleGetRelay(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
	relayIdStr := r.PathValue("id")

	if relayIdStr == "" {
		log.Println("handleGetRelay: missing relay id in url: ", r.URL.Path)
		http.Error(w, "Missing relay id", http.StatusBadRequest)
		return
	}

	relayId, err := strconv.Atoi(relayIdStr)
	if err != nil {
		log.Println("handleGetRelay: invalid relay id: ", relayIdStr)
		http.Error(w, "Invalid relay id", http.StatusBadRequest)
		return
	}

	log.Printf("handleGetRelay: request for relay %d\n", relayId)

	relay, err := models.SelectRelay(dbConn, relayId)
	if err != nil {
		log.Println("handleGetRelay: ", err)
		http.Error(w, "Error in getting relay", http.StatusInternalServerError)
		return
	}

	if relay == nil {
		msg := fmt.Sprintf("handleGetRelay: relay %d does not exist", relayId)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(relay)
	if err != nil {
		log.Println("handleGetRelay: json.Marshal: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func HandleCreateRelay(dbConn *sql.DB) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handleCreateRelay(dbConn, w, r)
		},
	)
}

func HandleCancelRelay(dbConn *sql.DB) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handleCreateRelay(dbConn, w, r)
		},
	)
}

func handleCancelRelay(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
	relayIdStr := r.PathValue("id")

	if relayIdStr == "" {
		log.Println("handleCancelRelay: missing relay id in url: ", r.URL.Path)
		http.Error(w, "Missing relay id", http.StatusBadRequest)
		return
	}

	relayId, err := strconv.Atoi(relayIdStr)
	if err != nil {
		log.Println("handleCancelRelay: invalid relay id: ", relayIdStr)
		http.Error(w, "Invalid relay id", http.StatusBadRequest)
		return
	}

	log.Printf("handleCancelRelay: request for relay %d\n", relayId)

	relay, err := models.SelectRelay(dbConn, relayId)
	if err != nil {
		log.Println("handleCancelRelay: ", err)
		http.Error(w, "Error in getting relay", http.StatusInternalServerError)
		return
	}

    if relay == nil {
		log.Println("handleCancelRelay: relay id=%d does not exist", relayId)
		http.Error(w, fmt.Sprintf("Relay %d does not exist", relayId), http.StatusUnprocessableEntity)
		return
    }

    if relay.Status != models.RelayReady {
        log.Println("handleCancelRelay: relay id=%d is not cancellatble, status=%d", relayId, relay.Status)
        if relay.Status == models.RelayFailed {
            http.Error(w, fmt.Sprintf("Relay %d has already failed", relayId), http.StatusUnprocessableEntity)
        } else {
            http.Error(w, fmt.Sprintf("Relay %d has already succeeded", relayId), http.StatusUnprocessableEntity)
        }
		return
    }

    id, err := models.InsertCancellation(dbConn, relayId)
    if err != nil {
		log.Println("handleCancelRelay: %w for relay=%d", err, relayId)
		http.Error(w, fmt.Sprintf("Relay %d does not exist", relayId), http.StatusUnprocessableEntity)
    }

    if id == 0 {
		log.Println("handleCancelRelay: cancellation request already exists for relay=%d", relayId)
		http.Error(w, fmt.Sprintf("Cancellation already exists for relay %d", relayId), http.StatusUnprocessableEntity)
    }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

// TODO: Want to add some sort of id to these logs so that I can know whats going on if there are multiple requests at once
func handleCreateRelay(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("handleCreateRelay: io.ReadAll:", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var req CreateRelayRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Println("handleCreateRelay: json.Unmarshal:", err)
		log.Println("request body:", string(body))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("handleCreateRelay: received request body: %s\n", req)
	if req.DeviceId == "" || req.CloudFunction == "" {
		log.Println("handleCreateRelay: Atleast one field in the post payload was blank or invalid")
		http.Error(w, "device_id and cloud_function are required fields",
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

	relayId, err := CreateRelay(dbConn, req.DeviceId, req.CloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		log.Println("handleCreateRelay:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("handleCreateRelay: new relay created, id: %d scheduled for %s\n", relayId, scheduledTime.String())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// TODO: Send json?
	io.WriteString(w, fmt.Sprintf("%d", relayId))
}



func CreateRelay(dbConn *sql.DB, deviceId string, cloudFunction string, argument string, desiredReturnCode sql.NullInt64, scheduledTime time.Time) (int, error) {
	deviceKey, err := models.InsertOrUpdateDevice(dbConn, deviceId)
	if err != nil {
		return 0, fmt.Errorf("CreateRelay: %w", err)
	}

	relayId, err := models.InsertRelay(dbConn, deviceKey, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		return 0, fmt.Errorf("CreateRelay: %w", err)
	}

	return relayId, nil
}
