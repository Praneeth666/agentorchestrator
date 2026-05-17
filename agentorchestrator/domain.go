package agentorchestrator

import (
	"errors"
	"math/rand"
	"time"
)

type state interface {
	run(body any, l *serviceImp) error
	getName() string
}

type Event string

const (
	SUCCESS Event = "success"
	FAILURE Event = "failure"
)

type openState struct {
	name string
}

func (s openState) getName() string { return s.name }

type closedState struct {
	name string
}

func (s closedState) getName() string { return s.name }

type halfOpenState struct {
	name string
}

func (s halfOpenState) getName() string { return s.name }

func (s openState) run(body any, l *serviceImp) error {
	split := l.config.Opentraffic
	total := split.Claude + split.Openapi + split.Reject
	r := rand.Intn(total)
	var e error

	if r < split.Claude {
		e = claude(body)
	} else if r < split.Claude+split.Openapi {
		e = openapi(body)
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if r < split.Claude {
		l.claudeerror.registerRequest()
		if e != nil {
			l.claudeerror.registerError()
		}
	} else if r < split.Claude+split.Openapi {
		l.openapierror.registerRequest()
		if e != nil {
			l.openapierror.registerError()
		}
	}

	errorRate := l.getErrorRate()
	if errorRate < float32(l.config.Halfopenerrorrate)/100 {
		l.triggerEvent(SUCCESS)
	}

	return e
}

func (s closedState) run(body any, l *serviceImp) error {
	e := claude(body)
	l.lock.Lock()
	defer l.lock.Unlock()
	l.claudeerror.registerRequest()
	if e != nil {
		l.claudeerror.registerError()
	}

	if l.claudeerror.getErrorRate() > float32(l.config.Closederrorrate)/100 {
		l.triggerEvent(FAILURE)
	}

	return e
}

func (s halfOpenState) run(body any, l *serviceImp) error {
	split := l.config.Halfopentraffic
	total := split.Claude + split.Openapi + split.Reject
	r := rand.Intn(total)
	var e error

	if r < split.Claude {
		e = claude(body)
	} else if r < split.Claude+split.Openapi {
		e = openapi(body)
	} else {
		e = errors.New("rejected before sending")
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if r < split.Claude {
		l.claudeerror.registerRequest()
		if e != nil {
			l.claudeerror.registerError()
		}
	} else if r < split.Claude+split.Openapi {
		l.openapierror.registerRequest()
		if e != nil {
			l.openapierror.registerError()
		}
	}

	errorRate := l.getErrorRate()
	if errorRate > float32(l.config.Halfopenerrorrate)/100 {
		l.triggerEvent(FAILURE)
	} else if errorRate < float32(l.config.Closederrorrate)/100 {
		l.triggerEvent(SUCCESS)
	}

	return e
}


type Trafficsplit struct {
	Claude  int
	Openapi int
	Reject  int
}

type Config struct {
	Closederrorrate   int // < 5
	Halfopenerrorrate int // 5 to 90
	// openerrorrate     int // > 90

	Halfopentraffic Trafficsplit // 5    95  0
	Closedtraffic   Trafficsplit // 100  0   0
	Opentraffic     Trafficsplit // 5    5   90
}

type errorTracker struct {
	previousBucketrequests int
	currentBucketrequests  int
	previousBucketerror    int
	currentBucketerror     int
	currentBucketId        int
	bucketDurationSecs     int64
}

func (l *errorTracker) reset() {
	l.previousBucketrequests = 0
	l.previousBucketerror = 0
	l.currentBucketrequests = 0
	l.currentBucketerror = 0
	l.currentBucketId = -1
}

func (l *errorTracker) getErrorRate() float32 {
	elapsedSecs := time.Now().Unix() % l.bucketDurationSecs
	previousFraction := float32(l.bucketDurationSecs-elapsedSecs) / float32(l.bucketDurationSecs)

	effectiveRequests := float32(l.currentBucketrequests) + float32(l.previousBucketrequests)*previousFraction
	effectiveErrors := float32(l.currentBucketerror) + float32(l.previousBucketerror)*previousFraction
	// fmt.Println("ddd", effectiveErrors, effectiveRequests)
	if effectiveRequests == 0 {
		return 0
	}
	return effectiveErrors / effectiveRequests
}

func (l *errorTracker) syncBucket() {
	newBucketId := int(time.Now().Unix() / l.bucketDurationSecs)
	if l.currentBucketId == -1 {
		l.currentBucketId = newBucketId
		return
	}
	if newBucketId == l.currentBucketId {
		return
	}
	if newBucketId == l.currentBucketId+1 {
		// one bucket elapsed: shift current → previous
		l.previousBucketrequests = l.currentBucketrequests
		l.previousBucketerror = l.currentBucketerror
		l.currentBucketrequests = 0
		l.currentBucketerror = 0
	} else {
		// more than one bucket elapsed: both buckets are stale
		l.previousBucketrequests = 0
		l.previousBucketerror = 0
		l.currentBucketrequests = 0
		l.currentBucketerror = 0
	}
	l.currentBucketId = newBucketId
}

func (l *errorTracker) registerRequest() {
	l.syncBucket()
	l.currentBucketrequests++
	// return l.getErrorRate()
}

func (l *errorTracker) registerError() {
	l.syncBucket()
	l.currentBucketerror++
	// return l.getErrorRate()
}
