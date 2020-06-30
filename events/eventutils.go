package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
)

//json hints helps while de-serialization from the struct
type Event struct {
	Event string `json:"event"`
	Time  int    `json:"time"`
}

//struct for the serialization of json for output
type AvgEvent struct {
	Event string `json:"event"`
	Avg   int    `json:"avg"`
}

type AvgEvents []AvgEvent

//Use a map with the RWMutex to optimize for the multiple read and writes
// This is in global scope like static in other languages
var eventMap = struct {
	sync.RWMutex
	data map[string][]int
}{data: make(map[string][]int)}

//PublishEvent taken a json serialized string. Keep track of the events with the map.
//This thread-safe map is used to calculate the average of the events
func (e Event) PublishEvent(input string) error {
	incomingEvent := Event{}
	err := json.Unmarshal([]byte(input), &incomingEvent)
	if err != nil {
		log.Printf("Invalid json. Failed to unmarshal! %s", err.Error())
		return err
	}

	validationError := validateEvent(incomingEvent)
	if validationError != nil {
		return validationError
	}

	//Redundant call with overhead of another go-routine (processChannel)
	//But this will be helpful to bulk process the event in single json. Over-head will be negligible for those cases.
	//Use a boolean channel as a locking mechanism to avoid race condition on write data.
	//This channel is used to wait till a signal is sent back from the go-routine when it completes.
	//NOTE: Here the unbuffered boolean channel is uses as a synchronous block only to process the writes on the map.
	//       Concurrent reads are fine. This enables to process the event concurrent to the caller.
	processChannel := make(chan bool)
	go processEvent(incomingEvent, processChannel)
	isComplete := <-processChannel

	//If the channel is done close it
	if isComplete {
		close(processChannel)
	}

	//Processed successfully, return nil
	return nil
}

//Helper function to process the events
func processEvent(event Event, channel chan bool) {
	//Sanitize the incoming data to lowercase
	eventKey := strings.ToLower(event.Event)

	//Check if the incoming event exists, if so sum it up and add the counts
	// else add the time value as new entry
	//Takes a write lock
	eventMap.Lock()
	if val, ok := eventMap.data[eventKey]; ok {
		eventMap.data[eventKey][0] = val[0] + event.Time
		eventMap.data[eventKey][1] += 1
	} else {
		eventMap.data[eventKey] = []int{event.Time, 1}
	}
	eventMap.Unlock()

	// Done processing the events to map, send signal via channel
	channel <- true

}

func validateEvent(incomingEvent Event) error {
	if incomingEvent.Event == "" {
		log.Printf("Invaild event: %s", incomingEvent.Event)
		return errors.New(fmt.Sprintf("Invaild event: %s", incomingEvent.Event))
	}

	if incomingEvent.Time < 0 {
		log.Printf("Invaild time value: %d", incomingEvent.Time)
		return errors.New(fmt.Sprintf("Invaild time value: %d", incomingEvent.Time))
	}

	//If no validation issue, return nil
	return nil
}

func (e Event) GetStats() string {
	//Collect the map data into array of AvgEvent
	avgStructs := AvgEvents{}

	//Takes a read Lock, multiple read locks can be used simultaneously
	eventMap.RLock()
	for key, val := range eventMap.data {
		avgStructs = append(avgStructs, AvgEvent{key, val[0] / val[1]})
	}
	eventMap.RUnlock()

	//Marshal the data to structured json format
	avgJson, err := json.Marshal(avgStructs)
	if err != nil {
		log.Println("Marshalling error", err)
	}

	log.Printf("Stats: %s", avgJson)
	//Send the string value from the marshalled data
	return string(avgJson)
}
