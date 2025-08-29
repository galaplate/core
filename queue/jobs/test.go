package jobs

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/galaplate/core/queue"
)

type Test struct {
}

func (e Test) MaxAttempts() int {
	return 3
}

func (e Test) RetryAfter() time.Duration {
	return 2 * time.Minute
}

func (Test) Type() string {
	return "test"
}

func (Test) Handle(payload json.RawMessage) error {
	fmt.Println("Test job is running with payload", string(payload))
	return nil
}

func init() {
	queue.RegisterJob(Test{})
}
