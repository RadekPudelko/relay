package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"pcfs/db"
	"pcfs/server"
)

type Client struct {
	url string
}

func NewClient(port int) Client {
	return Client{fmt.Sprintf("http://localhost:%d", port)}
}

func (c Client) Ping() error {
	resp, err := http.Get(fmt.Sprintf("%s/", c.url))
	if err != nil {
		return fmt.Errorf("Ping: http.Get %+w", err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Ping: io.ReadAll %+w", err)
	}
	fmt.Printf("Ping: msg recieved: %s\n", out)
	return nil
}

func (c Client) GetTask(id int) (*db.Task, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/tasks/%d", c.url, id))
	if err != nil {
		return nil, fmt.Errorf("GetTask: http.Get %+w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetTask: io.ReadAll %+w", err)
	}

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("GetTask: response status: %d, body %s", resp.StatusCode, body)
    }

    var task db.Task
	err = json.Unmarshal(body, &task)
	if err != nil {
        return nil, fmt.Errorf("GetTask: json.Unmarshal: %+w", err)
    }

	return &task, nil
}

func (c Client) CreateTask(somId string, cloudFunction string, argument string, desiredReturnCode *int, scheduledTime *time.Time) (int, error) {
    data := server.CreateTaskRequest{
        SomId: somId,
        CloudFunction: cloudFunction,
        Argument: &argument,
        DesiredReturnCode: desiredReturnCode,
        ScheduledTime: scheduledTime,
    }

	jsonData, err := json.Marshal(data)
	if err != nil {
        return 0, fmt.Errorf("CreateTask: json.Marshal: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/tasks", c.url), bytes.NewBuffer(jsonData))
	if err != nil {
        return 0, fmt.Errorf("CreateTask: http.NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
        return 0, fmt.Errorf("CreateTask: client.Do: %w", err)
	}
	defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return 0, fmt.Errorf("CreateTask: io.ReadAll: %w", err)
    }

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
        return 0, fmt.Errorf("CreateTask: request error, status code=%d, body=%s", resp.StatusCode, body)
	}

    fmt.Printf("Task body %s\n", body)
    id, err := strconv.ParseInt(string(body), 10, 64)
    if err != nil {
        return 0, fmt.Errorf("CreateTask: strconv.ParseInt: %w on %s", err, string(body))
    }
    return int(id), nil
}

