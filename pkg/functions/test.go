package functions

import (
	"fmt"
	"github.com/adjust/rmq/v5"
)

type Test struct {
	workerIndex int
}

func NewTest(workerIndex int) *Test {
	return &Test{
		workerIndex: workerIndex,
	}
}

func (t *Test) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()

	fmt.Println(fmt.Sprintf("execute worker test index %d with payload %s", t.workerIndex, payload))
	if err := delivery.Ack(); err != nil {
		panic(err)
	}
}
