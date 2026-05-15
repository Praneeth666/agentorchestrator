package agentorchestrator

import (
	"errors"
	"fmt"
	"math/rand"
)

func openapi(body any) error {
	fmt.Println("openapi called", body)
	if rand.Intn(100) < 50 {
		return errors.New("openapi error")
	}
	return nil
}

func claude(body any) error {
	fmt.Println("claude called", body)
	if rand.Intn(100) < 25 {
		return errors.New("claude error")
	}
	return nil
}
