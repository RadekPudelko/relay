package server

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"relay/db"
	"relay/particle"
)

// TODO make backgroundTask sleep when there are no tasks, wake by new task post?
func BackgroundTask(config Config, dbConn *sql.DB, particle particle.ParticleAPI) {
	var sem = make(chan int, config.MaxRoutines)
	lastTaskId := 0
	lastNTasks := 0
	for true {
		// Get ready tasks, starting from the lastTaskId, limited 1 per som
		// This implementation does not care about the order of tasks
		// To take into account order, would first need to get list of soms with ready tasks, then query the min for each
		taskIds, err := GetReadyTasks(dbConn, lastTaskId, config.TaskLimit, time.Now().UTC())
		if err != nil {
			log.Fatal("backgroundTask: ", err)
		}

		nTasks := len(taskIds)
		if lastNTasks != 0 || nTasks != 0 {
			log.Printf("Loading %d ready tasks ids from the dbConn\n", nTasks)
		}
		lastNTasks = nTasks
		if nTasks == 0 {
			lastTaskId = 0
			continue
		}
		lastTaskId = taskIds[nTasks-1]

		// TODO: Load additional requests in the background as tasks are processed - need to be careful with this to ignore already loaded tasks, otherwise may load already completed tasks
		i := 0
		var wg sync.WaitGroup
		for _, taskId := range taskIds {
			sem <- 1
			wg.Add(1)
			go func(id int) {
				processTask(config, dbConn, particle, id)
				<-sem
				wg.Done()
			}(taskId)
			i++
		}
		wg.Wait()
	}
}

// TODO: Update the schedule time of the task if its been recently pinged and offline, ping fails or device is offile
func processTask(config Config, dbConn *sql.DB, particle particle.ParticleAPI, id int) {
	task, err := db.SelectTask(dbConn, id)
	if err != nil {
		log.Printf("processTask: id=%d, %+v\n", id, err)
		return
	}
	// Consider pinging a som if its been more than n seconds since last check
	// TODO: define a config for how long a last ping is valid for
	// TODO: update online time on good communication from cf
	if !task.Som.LastOnline.Valid || time.Since(task.Som.LastOnline.Time) > config.PingRetryDuration {
		// Only ping a som if we have not pinged in n seconds
		if task.Som.LastPing.Valid && time.Since(task.Som.LastPing.Time) < config.PingRetryDuration {
			log.Printf("processTask: id=%d, om %s is not online, skipping\n", id, task.Som.SomId)
			return
		}
		log.Printf("processTask: id=%d, pinging som %s\n", id, task.Som.SomId)
		online, err := particle.Ping(task.Som.SomId)
		now := sql.NullTime{Time: time.Now(), Valid: true}
		if err != nil {
			log.Printf("processTask: id=%d, %+v\n", id, err)
			err = db.UpdateSom(dbConn, task.Som.Id, task.Som.LastOnline, now)
			if err != nil {
				log.Printf("processTask: id=%d, %+v\n", id, err)
			}
			return
		}
		if !online {
			log.Printf("processTask: id=%d, som %s is offline\n", id, task.Som.SomId)
			err = db.UpdateSom(dbConn, task.Som.Id, task.Som.LastOnline, now)
			// TODO: This and many places like this should never fail, so should the server crash here??
			if err != nil {
				log.Printf("processTask: id=%d, %+v\n", id, err)
			}
			return
		}
		err = db.UpdateSom(dbConn, task.Som.Id, now, now)
		if err != nil {
			log.Printf("processTask: id=%d, %+v\n", id, err)
			return
		}
	}

	log.Printf("processTask: id=%d, som %s is online\n", id, task.Som.SomId)
	// TODO: may want to get return value from function
	// TODO: may want to add some way to store error history in the database
	success, err := particle.CloudFunction(task.Som.SomId, task.CloudFunction, task.Argument, task.DesiredReturnCode)
	later := time.Now().Add(config.CFRetryDuration).UTC()
	if err != nil {
		log.Printf("processTask: id=%d, tries=%d, %+v", id, task.Tries, err)
		if task.Tries >= config.MaxRetries-1 { // start from 0
			log.Printf("processTask: id=%d has failed due to max failed tries\n", id)
			err = db.UpdateTask(dbConn, id, task.ScheduledTime, db.TaskFailed, task.Tries+1)
		} else {
			log.Printf("processTask: id=%d has failed, try again at %s\n", id, later)
			err = db.UpdateTask(dbConn, id, later, db.TaskReady, task.Tries+1)
		}
		if err != nil {
			log.Printf("processTask: task=%d, %+v\n", id, err)
		}

		return
	}

	if !success {
		log.Printf("processTask: id=%d has failed due to mismatch in returned code\n", id)
		err = db.UpdateTask(dbConn, id, task.ScheduledTime, db.TaskFailed, task.Tries+1)
	} else {
		log.Printf("processTask: id=%d, success\n", id)
		err = db.UpdateTask(dbConn, id, task.ScheduledTime, db.TaskComplete, task.Tries+1)
	}
	if err != nil {
		log.Printf("processTask: task=%d, %+v\n", id, err)
	}
}

// Queries for upto limit tasks in the db that are scheduled after scheduled time from id to id - 1 (inclusive)
func GetReadyTasks(dbConn *sql.DB, id, limit int, scheduledTime time.Time) ([]int, error) {
	taskIds, err := db.SelectTaskIds(dbConn, db.TaskReady, &id, nil, &limit, scheduledTime)
	if err != nil {
		return nil, fmt.Errorf("GetReadyTasks for %d onward: %w", id+1, err)
	}
	// TODO: If we don't get enough tasks, get the tasks upto id (exclusive) and try to add them to the list (need to check for unique soms)
	// if len(taskIds) < limit && id > 1 {
	// 	end := id - 1
	// 	limit := limit - len(taskIds)
	// 	taskIds2, err := db.SelectTaskIds(dbConn, db.TaskReady, nil, &end, &limit, scheduledTime)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("GetReadyTasks for 1 to %d: %w", id-1, err)
	// 	}
	// 	taskIds = append(taskIds, taskIds2...)
	// }
	return taskIds, nil
}
