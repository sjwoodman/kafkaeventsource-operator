package controller

import (
	"github.com/rh-event-flow-incubator/kafkaeventsource-operator/pkg/controller/kafkaeventsource"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, kafkaeventsource.Add)
}
