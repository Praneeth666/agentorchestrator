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

// Newservice constructs a serviceImp with the given configuration and registers
// all valid state transitions for the circuit breaker state machine.
// The returned service starts in closedState (healthy), routing all traffic to Claude.
// Both error trackers are initialized with a 60-second bucket duration.
//
// Parameters:
//   - config: Defines error-rate thresholds and traffic splits per state.
//
// Returns a service ready to handle requests via Request.
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

// addTransition registers a state transition: when event fires in from, the machine moves to to.
// Calling this with duplicate (from, event) pairs silently overwrites the previous target state.
// This method is called only during initialization in Newservice and is safe to call concurrently
// because it acquires the write lock before mutating the transitions map.
//
// Parameters:
//   - from:  The source state.
//   - to:    The destination state after the event fires.
//   - event: The event that triggers the transition.
//
// Returns nil; the error return exists for future validation needs.
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

// Request dispatches body to the appropriate backend based on the current circuit
// breaker state and logs the resulting state name and error rate.
// The active state's run() method handles routing, error tracking, and any
// resulting state transition (e.g., closedState → halfOpenState on high error rate).
//
// Parameters:
//   - body: The request payload forwarded to the backend; the concrete type is
//     interpreted by each state's run implementation.
//
// Returns nil; errors are tracked internally and surface through state transitions.
func (l *serviceImp) Request(body interface{}) error {
	l.state.run(body, l)
	fmt.Println("state", l.state.getName(), l.getErrorRate())
	return nil
}

// weightedErrorRate computes a traffic-weighted error rate across two backends.
// The weight of each backend is its share of the configured traffic split, so a
// backend carrying more traffic has proportionally greater influence on the result.
// Returns 0 if the total traffic weight is zero (no traffic configured).
//
// Parameters:
//   - split:       Traffic weights for Claude and Openapi backends.
//   - claudeRate:  Current error rate (0–100) observed for the Claude backend.
//   - openapiRate: Current error rate (0–100) observed for the OpenAI backend.
//
// Returns the combined error rate as a float32 in the same unit as the input rates.
func weightedErrorRate(split Trafficsplit, claudeRate, openapiRate float32) float32 {
	total := split.Claude + split.Openapi
	if total == 0 {
		return 0
	}
	return (float32(split.Claude)*claudeRate + float32(split.Openapi)*openapiRate) / float32(total)
}

// getErrorRate returns the effective error rate for the current circuit breaker state.
// In closedState only the Claude error tracker matters; in halfOpenState and openState
// the rate is weighted by the configured traffic split so that the dominant backend
// drives the threshold comparison used to trigger state transitions.
//
// Returns the error rate as a float32 percentage (0–100), or 0 if the state is unknown.
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

// triggerEvent advances the state machine by firing event from the current state.
// On a successful transition both error trackers are reset so stale rates from the
// previous state do not influence thresholds in the new state.
// This is called from within each state's run() method under the write lock, so
// callers must not re-acquire the lock.
//
// Parameters:
//   - event: SUCCESS or FAILURE signalling the outcome of the last request batch.
//
// Returns an error if no transitions are defined for the current state or if no
// transition exists for the given event in the current state.
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
