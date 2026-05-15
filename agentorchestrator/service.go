package agentorchestrator

import (
	"errors"
	"fmt"
	"sync"
)

type service interface {
	Request(body interface{}) error
}

type serviceImp struct {
	claudeerror  errorTracker
	openapierror errorTracker
	state        state
	config       Config
	transitions  map[state]map[Event]state
	lock         sync.RWMutex
}

func Newservice(config Config) service {
	transitions := make(map[state]map[Event]state)
	k := serviceImp{
		claudeerror:  errorTracker{currentBucketId: -1, bucketDurationSecs: 60},
		openapierror: errorTracker{currentBucketId: -1, bucketDurationSecs: 60},
		state:        closedState{name: "closed"},
		config:       config,
		transitions:  transitions,
	}
	k.addTransition(closedState{name: "closed"}, halfOpenState{name: "halfopen"}, FAILURE)
	k.addTransition(halfOpenState{name: "halfopen"}, closedState{name: "closed"}, SUCCESS)
	k.addTransition(halfOpenState{name: "halfopen"}, openState{name: "open"}, FAILURE)
	k.addTransition(openState{name: "open"}, halfOpenState{name: "halfopen"}, SUCCESS)
	return &k
}

func (l *serviceImp) addTransition(from state, to state, event Event) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	_, ok := l.transitions[from]
	if !ok {
		l.transitions[from] = make(map[Event]state)
	}
	l.transitions[from][event] = to
	return nil
}

func (l *serviceImp) Request(body interface{}) error {
	l.state.run(body, l)
	fmt.Println("state", l.state.getName(), l.getErrorRate())
	return nil
}

func weightedErrorRate(split Trafficsplit, claudeRate, openapiRate float32) float32 {
	total := split.Claude + split.Openapi
	if total == 0 {
		return 0
	}
	return (float32(split.Claude)*claudeRate + float32(split.Openapi)*openapiRate) / float32(total)
}

func (l *serviceImp) getErrorRate() float32 {
	switch l.state.(type) {
	case closedState:
		return l.claudeerror.getErrorRate()
	case halfOpenState:
		return weightedErrorRate(l.config.Halfopentraffic, l.claudeerror.getErrorRate(), l.openapierror.getErrorRate())
	case openState:
		return weightedErrorRate(l.config.Opentraffic, l.claudeerror.getErrorRate(), l.openapierror.getErrorRate())
	}
	return 0
}

func (l *serviceImp) triggerEvent(event Event) error {
	events, ok := l.transitions[l.state]
	if !ok {
		return errors.New("no transitions defined for current state")
	}
	next, ok := events[event]
	if !ok {
		return errors.New("no transition defined for event in current state")
	}
	l.state = next
	l.claudeerror.reset()
	l.openapierror.reset()
	return nil
}
