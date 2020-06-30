package events

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log"
	"runtime"
	"sync"
	"testing"
)

//WaitGroup for go-routine to access PublishEvent and GetStats
var waitGrp sync.WaitGroup

var event EventInterface

func init() {
	event = Event{}
}

type MockAction struct {
	mock.Mock
}

func TestPublishEventSingleEntry(t *testing.T) {
	//Clean up after this test
	defer deleteActionsMap()

	event1 := `{"event":"jump", "time":100}`
	expectedResult := `[{"event":"jump","avg":100}]`

	err := event.PublishEvent(event1)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, event.GetStats())
}

func TestPublishEventMultipleEntry(t *testing.T) {
	//Clean up after this test
	defer deleteActionsMap()

	event1 := `{"event":"jump", "time":100}`
	action2 := `{"event":"run", "time":75}`
	action3 := `{"event":"jump", "time":200}`
	action4 := `{"event":"jump", "time":300}`
	expectedResult1 := `{"event":"jump","avg":200}`
	expectedResult2 := `{"event":"run","avg":75}`

	err1 := event.PublishEvent(event1)
	err2 := event.PublishEvent(action2)
	err3 := event.PublishEvent(action3)
	err4 := event.PublishEvent(action4)

	assert.Nil(t, err1, err2, err3, err4)
	assert.Contains(t, event.GetStats(), expectedResult1)
	assert.Contains(t, event.GetStats(), expectedResult2)
}

func TestPublishEventMultipleEntryConcurrent(t *testing.T) {
	//Clean up after this test
	defer deleteActionsMap()

	// Sum of 1+2+3+.....+100 = 50.5
	expectedResult1 := `{"event":"jump","avg":50}`
	expectedResult2 := `{"event":"run","avg":10}`

	times := 100
	//Wait unit two for loop go-routines are done
	waitGrp.Add(2 * times)

	// Calls 100 times to insert the jump with i values
	for i := 1; i <= times; i++ {
		action := fmt.Sprintf(`{"event":"jump", "time":%d}`, i)

		go wrapPublishEvent(action)
	}

	// Calls 100 times to insert run with 10
	for i := 1; i <= times; i++ {
		action := `{"event":"run", "time":10}`

		go wrapPublishEvent(action)
	}

	//wait until all go-routines are done
	waitGrp.Wait()

	assert.Contains(t, event.GetStats(), expectedResult1)
	assert.Contains(t, event.GetStats(), expectedResult2)
}

func TestPublishEventMultipleEntryConcurrentAndParallel(t *testing.T) {
	//Clean up after this test
	defer deleteActionsMap()

	//This allows go-routines to run parallel on multiple logical cores available
	log.Printf("CPUs to use for parallel processing: %d", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Sum of (1000000 * 1000001) / 2 = 500000.5
	expectedResult1 := `{"event":"jump","avg":500000}`
	expectedResult2 := `{"event":"run","avg":10}`

	times := 1000 * 1000
	//Wait unit 3 loops of go-routines are done
	waitGrp.Add(3 * times)

	// Calls number of times to insert the jump with i values
	for i := 1; i <= times; i++ {
		action := fmt.Sprintf(`{"event":"jump", "time":%d}`, i)

		go wrapPublishEvent(action)
	}

	// Calls number of times to insert run with 10
	for i := 1; i <= times; i++ {
		action := `{"event":"run", "time":10}`

		go wrapPublishEvent(action)
	}

	// Call getstats specified number of times
	for i := 1; i <= times; i++ {
		go wrapGetStats()
	}

	//wait until all go-routines are done
	waitGrp.Wait()

	assert.Contains(t, event.GetStats(), expectedResult1)
	assert.Contains(t, event.GetStats(), expectedResult2)
}

func TestJsonExtraAndDuplicateData(t *testing.T) {
	event1 := `{"event":"jump", "fox":"ran", "5":34, "time":150, "Valid": "string", "time":200}`
	expectedResult := `[{"event":"jump","avg":200}]`

	err := event.PublishEvent(event1)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, event.GetStats())
}

// Negative test cases
func TestInvalidJsonAction(t *testing.T) {

	event1 := `{"abcdef":"jump", "time":100}`
	expectedResult := `[{"event":"jump","avg":100}]`

	err := event.PublishEvent(event1)
	assert.NotContains(t, event.GetStats(), expectedResult)
	if err != nil {
		assert.Contains(t, err.Error(), "Invaild event:")
	}
}

func TestInvalidJsonActionEmpty(t *testing.T) {

	event1 := `{"event":""}`
	unknownResult := `[{"event":"jump","avg":100}]`

	err := event.PublishEvent(event1)
	assert.NotContains(t, event.GetStats(), unknownResult)
	if err != nil {
		assert.Contains(t, err.Error(), "Invaild event")
	}
}

func TestInvalidJsonTime(t *testing.T) {

	event1 := `{"event":"jump", "time":-100}`
	expectedResult := `[{"event":"jump","avg":100}]`

	err := event.PublishEvent(event1)
	assert.NotContains(t, event.GetStats(), expectedResult)
	if err != nil {
		assert.Contains(t, err.Error(), "Invaild time value:")
	}
}

func TestInvalidJsonTimeLimit(t *testing.T) {

	event1 := `{"event":"jump", "time":922337203685477580900}`
	expectedResult := `[{"event":"jump","avg":100}]`

	err := event.PublishEvent(event1)
	assert.NotContains(t, event.GetStats(), expectedResult)
	if err != nil {
		assert.Contains(t, err.Error(), "json: cannot unmarshal number 922337203685477580900")
	}
}

func TestInvalidJsonTimeDouble(t *testing.T) {

	event1 := `{"event":"jump", "time":1.123}`
	expectedResult := `[{"event":"jump","avg":100}]`

	err := event.PublishEvent(event1)
	assert.NotContains(t, event.GetStats(), expectedResult)
	if err != nil {
		assert.Contains(t, err.Error(), "json: cannot unmarshal number 1.123")
	}
}

func (mc *MockAction) GetStats() string {
	args := mc.Called()
	return args.String(0)
}

func TestMockGetStats(t *testing.T) {
	mc := new(MockAction)
	mc.On("GetStats").Return(`{"event":"jump","avg":100}`)
	stat := mc.GetStats()

	assert.Equal(t, `{"event":"jump","avg":100}`, stat)
	mc.AssertExpectations(t)
}

// Internal private helper functions for unit tests

func wrapPublishEvent(actionStr string) error {
	//Indicate the completion via deferred done
	defer waitGrp.Done()

	return event.PublishEvent(actionStr)
}

func wrapGetStats() string {
	//Indicate the completion via deferred done
	defer waitGrp.Done()

	return event.GetStats()
}

func deleteActionsMap() {
	for entry := range eventMap.data {
		delete(eventMap.data, entry)
	}
}
