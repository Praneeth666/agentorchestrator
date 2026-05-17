package main

import (
	"fmt"
	"practice4/agentorchestrator"
)

func main() {
	cfg := agentorchestrator.Config{
		Closederrorrate:   5,
		Halfopenerrorrate: 90,
		// Openerrorrate:     90,
		Halfopentraffic: agentorchestrator.Trafficsplit{Claude: 5, Openapi: 95, Reject: 0},
		Closedtraffic:   agentorchestrator.Trafficsplit{Claude: 100, Openapi: 0, Reject: 0},
		Opentraffic:     agentorchestrator.Trafficsplit{Claude: 5, Openapi: 5, Reject: 90},
	}

	svc := agentorchestrator.Newservice(cfg)

	for i := 1; i <= 100; i++ {
		err := svc.Request(fmt.Sprintf("request %d", i))
		if err != nil {
			fmt.Printf("request %d error: %v\n", i, err)
		}
	}

	for i := 1; i <= 1000; i++ {
		err := svc.Request(fmt.Sprintf("request %d", i))
		if err != nil {
			fmt.Printf("request %d error: %v\n", i, err)
		}
	}
}
