package main
import(
	"log"
    "time"
	"database/sql"
    "fmt"

    "pcfs/particle"
	"pcfs/db"
)

// TODO make backgroundTask sleep when there are no tasks, wake by new task post?
func BackgroundTask(config Config, dbConn *sql.DB, particle particle.ParticleProvider) {
	var sem = make(chan int, config.MaxRoutines)
	lastTaskId := 0
	for true {
		// Get ready tasks, starting from the lastTaskId, limited 1 per som
		// This implementation does not care about the order of tasks
		// To take into account order, would first need to get list of soms with ready tasks, then query the min for each
        taskIds, err := GetReadyTasks(dbConn, lastTaskId, config.TaskLimit, time.Now().UTC())
		if err != nil {
			log.Fatal("backgroundTask: ", err)
		}

		log.Printf("Loading %d ready tasks ids from the dbConn\n", len(taskIds))
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
				processTask(config, dbConn, particle, id)
				<-sem
			}(taskId)
		}
	}
}

// TODO: Update the schedule time of the task if its been recently pinged and offline, ping fails or device is offile
func processTask(config Config, dbConn *sql.DB, particle particle.ParticleProvider, id int) {
	log.Println("processTask: process task ", id)
	task, err := db.SelectTask(dbConn, id)
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
		online, err := particle.Ping(task.Som.SomId, task.Som.ProductId)
		now := sql.NullTime{Time: time.Now(), Valid: true}
		if err != nil {
			log.Println("processTask:", err)
			err = db.UpdateSom(dbConn, task.Som.Id, task.Som.ProductId, task.Som.LastOnline, now)
			if err != nil {
				log.Println("processTask: ", err)
			}
			return
		}
		if !online {
			log.Printf("processTask: som %s is offline\n", task.Som.SomId)
			err = db.UpdateSom(dbConn, task.Som.Id, task.Som.ProductId, task.Som.LastOnline, now)
			// TODO: This and many places like this should never fail, so should the server crash here??
			if err != nil {
				log.Println("processTask: ", err)
			}
			return
		}
		err = db.UpdateSom(dbConn, task.Som.Id, task.Som.ProductId, now, now)
		if err != nil {
			log.Println("processTask:", err)
			return
		}
	}

	log.Printf("processTask: runnning task %d\n", id)
	log.Printf("processTask: som %s is online\n", task.Som.SomId)
	// TODO: may want to get return value from function
	// TODO: may want to add some way to store error history in the database
	success, err := particle.CloudFunction(task.Som.SomId, task.Som.ProductId, task.CloudFunction, task.Argument, task.DesiredReturnCode)
	fiveMinLater := time.Now().Add(5 * time.Minute)
	if err != nil {
		log.Println("processTask:", err)
		if task.Tries == config.MaxRetries {
			log.Printf("processTask task %d has failed due to exceeding max tries, err %v:\n", id, err)
			err = db.UpdateTask(dbConn, id, task.ScheduledTime, db.TaskFailed, task.Tries+1)
		} else {
			err = db.UpdateTask(dbConn, id, fiveMinLater, db.TaskReady, task.Tries+1)
		}
		if err != nil {
			log.Println("processTask:", err)
		}

		return
	}
	if !success {
		log.Printf("processTask task %d has failed due to mismatch in returned code\n", id)
		err = db.UpdateTask(dbConn, id, task.ScheduledTime, db.TaskFailed, task.Tries+1)
	} else {
		log.Printf("processTask: task %d, success\n", id)
		err = db.UpdateTask(dbConn, id, task.ScheduledTime, db.TaskComplete, task.Tries+1)
	}
	if err != nil {
		log.Println("processTask:", err)
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

