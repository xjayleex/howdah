package howdah_agent

import (
	"github.com/golang-collections/go-datastructures/queue"
)

type ActionQueue struct{
	q *queue.Queue
}

func NewActionQueue(queue *queue.Queue) *ActionQueue {
	return &ActionQueue{
		q: queue,
	}
}

func (aq *ActionQueue) Put (commands []*Command) {
}

func (aq *ActionQueue) Run () error {
	return nil
}

func (aq *ActionQueue) processCommand(command Command) {
}





type Command interface {}

type CommandStatusMap struct {
	currentState map[string]interface{}
	global *global
}
