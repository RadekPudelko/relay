package test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"relay/client"
	"relay/db"
	"relay/server"
	// "relay/particle"
)

type TestTask struct {
	Id     int
	DeviceId  string
	DRC    int
	Status db.TaskStatus
}

func assertTask(
	task *db.Task,
	config server.Config,
	somId string,
	cloudFunction string,
	argument string,
	desiredReturnCode *int,
	scheduledTime *time.Time,
) error {
	if task.Device.DeviceId != somId {
		return fmt.Errorf("assertTask: somId, expected=%s, got=%s", somId, task.Device.DeviceId)
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
	if task.Tries > config.MaxRetries {
		return fmt.Errorf("assertTask: tries=%d exceeds max tries=%d\n", task.Tries, config.MaxRetries)
	}
	return nil
}

func generateTask(nSoms int) (string, int, db.TaskStatus) {
	somNum := rand.Intn(nSoms)
	somId := fmt.Sprintf("som_%d", somNum)
	drc := rand.Intn(3) + 1
	if drc == 3 {
		return somId, drc, db.TaskComplete
	} else {
		return somId, drc, db.TaskFailed
	}
	// return &TestTask{
	//     Id: somNum,
	//     Status: db.TaskStatus(drc),
	// }
	// id, err := client.CreateTask(somId, cloudFunction, argument, &drc, scheduledTime)
}

func TestIntegration(t *testing.T) {
	t.Log("TestIntegration")
	config := server.Config{
		Host:              "localhost",
		Port:              "8080",
		MaxRoutines:       3,
		TaskLimit:         10,
		PingRetryDuration: 15 * time.Second,
		CFRetryDuration:   10 * time.Second,
		MaxRetries:        3,
	}
	go func() {
		err := runTestServer(config)
		if err != nil {
			// TODO: Fix this warning
			t.Fatalf("TestIntegration: %+v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)
	client := client.NewClient(8080)
	err := client.Ping()
	if err != nil {
		t.Fatalf("TestIntegration: %+v", err)
	}

	// Expect an error here for non existant task
	task, err := client.GetTask(1)
	if err == nil {
		t.Fatalf("TestIntegration: expected an error for non existant task got %+v", task)
	}

	cloudFunction := "func0"
	argument := ""
	var scheduledTime *time.Time = nil

	nTasks := 1000
	nSoms := 20
	testTasks := make([]TestTask, nTasks)

	// TODO: use goroutines to hit multiple requests as faster
	for i := 0; i < nTasks; i++ {
		somId, drc, status := generateTask(nSoms)
		id, err := client.CreateTask(somId, cloudFunction, argument, &drc, scheduledTime)
		if err != nil {
			t.Fatalf("TestIntegration: %+v", err)
		}
		testTasks[i].Id = id
		testTasks[i].DeviceId = somId
		testTasks[i].DRC = drc
		testTasks[i].Status = status
	}

	// TODO: add extra routine to spam the service with gets
	for i := 0; i < nTasks; i++ {
		for {
			task, err := client.GetTask(testTasks[i].Id)
			if err != nil {
				t.Logf("TestIntegration: expected an error for non existant task got %+v\n", task)
			} else if task.Status == db.TaskReady {
				time.Sleep(100 * time.Millisecond)
				continue
			} else if task.Status != testTasks[i].Status {
				t.Fatalf("TestIntegration: task status mismatch, want=%d, got=%d, task=%+v\n", int(testTasks[i].Status), int(task.Status), task)
			} else {
				err = assertTask(task, config, testTasks[i].DeviceId, cloudFunction, argument, &testTasks[i].DRC, scheduledTime)
				if err != nil {
					t.Fatalf("TestIntegration: %+v", err)
				}
				break
			}
		}
	}
}
