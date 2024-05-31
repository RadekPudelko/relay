package test

import (
	"fmt"
	"testing"
	"time"
    "math/rand"

	"relay/client"
	"relay/db"
	// "relay/particle"
)

type TestTask struct {
    Id int
    Status db.TaskStatus
}

func assertTask(task *db.Task, somId string, cloudFunction string, argument string, desiredReturnCode *int, scheduledTime *time.Time) error {
	if task.Som.SomId != somId {
		return fmt.Errorf("assertTask: somId, expected=%s, got=%s", somId, task.Som.SomId)
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

func generateTask(nSoms int) (string, int) {
    somNum := rand.Intn(nSoms)
    somId := fmt.Sprintf("som_%d", somNum)
    drc := rand.Intn(3)
    return somId, drc
    // return &TestTask{
    //     Id: somNum,
    //     Status: db.TaskStatus(drc),
    // }
    // id, err := client.CreateTask(somId, cloudFunction, argument, &drc, scheduledTime)
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

	somId := "som0"
	cloudFunction := "func0"
	argument := ""
	var desiredReturnCode *int = nil
	var scheduledTime *time.Time = nil

	id, err := client.CreateTask(somId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		t.Fatalf("TestIntegration: %+v", err)
	}
	t.Logf("Created task %d\n", id)
	time.Sleep(1)

	task, err = client.GetTask(id)
	if err != nil {
		t.Fatalf("TestIntegration: %+v for id=%d", err, id)
	}

	err = assertTask(task, somId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		t.Fatalf("TestIntegration: %+v", err)
	}

	start := time.Now()
	complete := false
	for time.Since(start) < 10*time.Second {
		task, err := client.GetTask(1)
		if err != nil {
			t.Logf("TestIntegration: expected an error for non existant task got %+v\n", task)
		} else if task.Status != db.TaskReady {
			t.Logf("TestIntegration: task %+v\n", task)
			complete = true
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if !complete {
		t.Fatalf("TestIntegration: expected task id to complete, final state %v", task)
	}


    nTasks := 100
    nSoms := 3
    testTasks := make([]TestTask, nTasks)

    // TODO: use goroutines to hit multiple requests as fast as possible
    for i:= 0; i < nTasks; i++ {
        somId, drc := generateTask(nSoms)
        id, err := client.CreateTask(somId, cloudFunction, argument, &drc, scheduledTime)
        if err != nil {
            t.Fatalf("TestIntegration: %+v", err)
        }
        testTasks[i].Id = id
        testTasks[i].Status = db.TaskStatus(drc)
    }

    for i := 0; i < nTasks; i++ {
        for {
            time.Sleep(1 * time.Second)
            task, err := client.GetTask(testTasks[i].Id)
            if err != nil {
                t.Logf("TestIntegration: expected an error for non existant task got %+v\n", task)
            } else if task.Status != testTasks[i].Status {
                continue
            } else if task.Status != testTasks[i].Status {
               t.Fatalf("TestIntegration: task status mismatch, want=%d, got=%d, task=%+v\n", int(testTasks[i].Status), int(task.Status), task)
            } else {
                break
            }
        }
    }
    // TODO: Assert no task has more than the max tries
}


















