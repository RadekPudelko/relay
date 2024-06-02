package server

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"relay/internal/models"
	"relay/internal/particle"
)

// TODO: make backgroundTask sleep when there are no relays, wake by new relay post?
// TODO: reduce logs
func BackgroundTask(config Config, dbConn *sql.DB, particle particle.ParticleAPI) {
	var sem = make(chan int, config.MaxRoutines)
	lastRelayId := 0
	lastNRelays := -1
	for true {
		err := ProcessCancellations(dbConn)
		if err != nil {
			// Fatal?
			log.Fatal("backgroundTask: ", err)
		}

		// Get ready relays, starting from the lastRelayId, limited 1 per device
		// This implementation does not care about the order of relays
		// To take into account order, would first need to get list of devices with ready relays, then query the min for each
		relayIds, err := GetReadyRelays(dbConn, lastRelayId, config.RelayLimit, time.Now().UTC())
		if err != nil {
			// Fatal?
			log.Fatal("backgroundTask: ", err)
		}

		nRelays := len(relayIds)
		if lastNRelays != 0 || nRelays != 0 {
			log.Printf("Loading %d ready relays ids from the dbConn\n", nRelays)
		}
		lastNRelays = nRelays
		if nRelays == 0 {
			lastRelayId = 0
			continue
		}
		lastRelayId = relayIds[nRelays-1]

		// TODO: Load additional requests in the background as relays are processed - need to be careful with this to ignore already loaded relays, otherwise may load already completed relays
		i := 0
		var wg sync.WaitGroup
		for _, relayId := range relayIds {
			sem <- 1
			wg.Add(1)
			go func(id int) {
				processRelay(config, dbConn, particle, id)
				<-sem
				wg.Done()
			}(relayId)
			i++
		}
		wg.Wait()
	}
}

func ProcessCancellations(dbConn *sql.DB) error {
	// Handle cancellations 100 at a time until they are all processed
	for {
		cancellations, err := models.SelectCancellations(dbConn, 100)
		if err != nil {
			return fmt.Errorf("ProcessCancellations: %w", err)
		}
		if len(cancellations) == 0 {
			return nil
		}
		for _, cancellation := range cancellations {
			err := models.UpdateRelayStatus(dbConn, cancellation.RelayId, models.RelayCancelled)
			if err != nil {
				return fmt.Errorf("ProcessCancellations: %w on cancellation %+v", err, cancellation)
			}
			err = models.DeleteCancellation(dbConn, cancellation.Id)
			if err != nil {
				return fmt.Errorf("ProcessCancellations: %w on cancellation %+v", err, cancellation)
			}
		}
	}
}

// TODO: Update the schedule time of the relay if its been recently pinged and offline, ping fails or device is offile
func processRelay(config Config, dbConn *sql.DB, particle particle.ParticleAPI, id int) {
	relay, err := models.SelectRelay(dbConn, id)
	if err != nil {
		log.Printf("processRelay: id=%d, %+v\n", id, err)
		return
	}
	// Consider pinging a device if its been more than n seconds since last check
	// TODO: define a config for how long a last ping is valid for
	// TODO: update online time on good communication from cf
	// if !relay.Device.LastOnline.Valid || time.Since(relay.Device.LastOnline.Time) > config.PingRetryDuration {
	if !relay.Device.LastOnline.Valid {
		// Only ping a device if we have not pinged in n seconds
		log.Printf("processRelay: id=%d, pinging device %s\n", id, relay.Device.DeviceId)
		online, err := particle.Ping(relay.Device.DeviceId)
		now := sql.NullTime{Time: time.Now(), Valid: true}
		if err != nil || !online {
            if err != nil {
                log.Printf("processRelay: %+v for relay id=%d, device %s \n", err, id, relay.Device.DeviceId)
            } else {
                log.Printf("processRelay: id=%d, device %s is offline\n", id, relay.Device.DeviceId)
            }
            later := time.Now().Add(config.PingRetryDuration).UTC()
			err = models.UpdateRelay(dbConn, id, later, relay.Status, relay.Tries)
			if err != nil {
                // TODO: This and many places like this should never fail, so should the server crash here??
				log.Printf("processRelay: id=%d, %+v\n", id, err)
			}
			return
		}
		err = models.UpdateDevice(dbConn, relay.Device.Id, now)
		if err != nil {
			log.Printf("processRelay: relay id=%d, %+v\n", id, err)
			return
		}
	}

	log.Printf("processRelay: id=%d, device %s is online\n", id, relay.Device.DeviceId)
	// TODO: may want to get return value from function
	// TODO: may want to add some way to store error history in the database
	success, err := particle.CloudFunction(relay.Device.DeviceId, relay.CloudFunction, relay.Argument, relay.DesiredReturnCode)
	later := time.Now().Add(config.CFRetryDuration).UTC()
	if err != nil {
		log.Printf("processRelay: id=%d, tries=%d, %+v", id, relay.Tries, err)
		if relay.Tries >= config.MaxRetries-1 { // start from 0
			log.Printf("processRelay: id=%d has failed due to max failed tries\n", id)
			err = models.UpdateRelay(dbConn, id, relay.ScheduledTime, models.RelayFailed, relay.Tries+1)
		} else {
			log.Printf("processRelay: id=%d has failed, try again at %s\n", id, later)
			err = models.UpdateRelay(dbConn, id, later, models.RelayReady, relay.Tries+1)
		}
		if err != nil {
			log.Printf("processRelay: relay=%d, %+v\n", id, err)
		}

		return
	}

	if !success {
		log.Printf("processRelay: id=%d has failed due to mismatch in returned code\n", id)
		err = models.UpdateRelay(dbConn, id, relay.ScheduledTime, models.RelayFailed, relay.Tries+1)
	} else {
		log.Printf("processRelay: id=%d, success\n", id)
		err = models.UpdateRelay(dbConn, id, relay.ScheduledTime, models.RelayComplete, relay.Tries+1)
	}
	if err != nil {
		log.Printf("processRelay: relay=%d, %+v\n", id, err)
	}
}

// Queries for upto limit relays in the db that are scheduled after scheduled time from id to id - 1 (inclusive)
func GetReadyRelays(dbConn *sql.DB, id, limit int, scheduledTime time.Time) ([]int, error) {
	relayIds, err := models.SelectRelayIds(dbConn, models.RelayReady, &id, nil, &limit, scheduledTime)
	if err != nil {
		return nil, fmt.Errorf("GetReadyRelays for %d onward: %w", id+1, err)
	}
	// TODO: If we don't get enough relays, get the relays upto id (exclusive) and try to add them to the list (need to check for unique soms)
	// if len(relayIds) < limit && id > 1 {
	// 	end := id - 1
	// 	limit := limit - len(relayIds)
	// 	taskIds2, err := db.SelectTaskIds(dbConn, db.TaskReady, nil, &end, &limit, scheduledTime)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("GetReadyRelays for 1 to %d: %w", id-1, err)
	// 	}
	// 	relayIds = append(relayIds, taskIds2...)
	// }
	return relayIds, nil
}
