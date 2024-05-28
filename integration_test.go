package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"pcfs/db"
	"pcfs/server"
	// "pcfs/particle"
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

func (c Client) CreateTask(somId string, productId int, cloudFunction string, argument string, desiredReturnCode *int, scheduledTime *time.Time) (int, error) {
    data := server.CreateTaskRequest{
        SomId: somId,
        ProductId: productId,
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

func assertTask(task *db.Task, somId string, productId int, cloudFunction string, argument string, desiredReturnCode *int, scheduledTime *time.Time) (error) {
    if task.Som.SomId != somId {
        return fmt.Errorf("assertTask: somId, expected=%s, got=%s", somId, task.Som.SomId)
    }
    if task.Som.ProductId != productId {
        return fmt.Errorf("assertTask: productId, expected=%d, got=%d", productId, task.Som.ProductId)
    }
    if task.CloudFunction != cloudFunction {
        return fmt.Errorf("assertTask: cloudFunction, expected=%s, got=%s", cloudFunction, task.CloudFunction)
    }
    if task.Argument != argument {
        return fmt.Errorf("assertTask: argument, expected=%s, got=%s", argument, task.Argument)
    }
    if desiredReturnCode != nil {
        if !task.DesiredReturnCode.Valid {
            return fmt.Errorf("assertTask: desired return code: got invalid")
        } else if int(task.DesiredReturnCode.Int64) != *desiredReturnCode {
            return fmt.Errorf("assertTask: desired return code: expected=%d, got=%d", *desiredReturnCode, task.DesiredReturnCode.Int64)
        }
    }
    if scheduledTime != nil && task.ScheduledTime != *scheduledTime {
        return fmt.Errorf("assertTask: scheduled time: expected=%s, got=%s", scheduledTime, task.ScheduledTime)
    }
    return nil
}


func TestIntegration(t *testing.T) {
	t.Log("TestIntegration")
	go func() {
		err := runTestServer()
		if err != nil {
			// TODO: Fix this warning
			t.Fatalf("TestIntegration: %+v", err)
		}
	}()

    time.Sleep(50 * time.Millisecond)
	client := NewClient(8080)
	err := client.Ping()
	if err != nil {
		t.Fatalf("TestIntegration: %+v", err)
	}

    // Expect an error here for non existant task
    task, err := client.GetTask(1)
    if err == nil {
		t.Fatalf("TestIntegration: expected an error for non existant task got %+v", task)
    }

    somId := "som0"
    productId := 123
    cloudFunction := "func0"
    argument := ""
    var desiredReturnCode *int = nil
    var scheduledTime *time.Time = nil

    id, err := client.CreateTask(somId, productId, cloudFunction, argument, desiredReturnCode, scheduledTime)
    if err != nil {
		t.Fatalf("TestIntegration: %+v", err)
    }
    t.Logf("Created task %d\n", id)
    time.Sleep(1)

    task, err = client.GetTask(id)
    if err != nil {
        t.Fatalf("TestIntegration: %+v for id=%d", err, id)
    }

    err = assertTask(task, somId, productId, cloudFunction, argument, desiredReturnCode, scheduledTime)
    if err != nil {
        t.Fatalf("TestIntegration: %+v", err)
    }

    start := time.Now()
    complete := false
    for time.Since(start) < 10 * time.Second {
        task, err := client.GetTask(1)
        if err != nil {
            t.Logf("TestIntegration: expected an error for non existant task got %+v\n", task)
        } else if task.Status != db.TaskReady {
            t.Logf("TestIntegration: task %+v\n", task)
            complete = true
            break
        }
        time.Sleep(250* time.Millisecond)
    }
    if !complete {
        t.Fatalf("TestIntegration: expected task id to complete, final state %v", task)
    }
}


















