package main

import (
	"github.com/gouthams/play-with-go/events"
	"log"
)

func main() {
	// Sample demo of the events utils
	// Use the interface defined for the utility so that it could be composed to other implementation if need be.
	var event events.EventInterface

	// Polymorphic behaviour uses structural typing of the implicit interface defined in event struct
	event = &events.Event{}

	//NOTE: We are passing in the json format string but not initializing the events themselves to demonstrate the
	// go json parsing to struct. There is no technical reasons to do so.
	// Look at the eventutils_test TestPublishEventMultipleEntryConcurrentAndParallel for concurrent calls using go
	//-routines
	event1 := `{"event":"jump", "time":100}`
	event2 := `{"event":"run", "time":75}`
	event3 := `{"event":"jump", "time":200}`
	event4 := `{"event":"jump", "time":300}`

	err := event.PublishEvent(event1)
	if err != nil {
		log.Printf("Error publishing event: %v", err)
	}
	err = event.PublishEvent(event2)
	if err != nil {
		log.Printf("Error publishing event: %v", err)
	}

	event.GetStats()

	err = event.PublishEvent(event3)
	if err != nil {
		log.Printf("Error publishing event: %v", err)
	}
	err = event.PublishEvent(event4)
	if err != nil {
		log.Printf("Error publishing event: %v", err)
	}

	event.GetStats()
}
