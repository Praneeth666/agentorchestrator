# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the application
go run main.go

# Build
go build ./...

# Test all / single test
go test ./...
go test ./agentorchestrator/... -run TestFunctionName

# Format (enforced)
gofmt -w .

# Lint
golangci-lint run
```

## Code Style

- Use `fmt.Printf` for all logging.
- Run `gofmt` before committing — formatting is enforced.
- Functions must stay under 50 lines.
- Every function must have a comment at the top describing what it does.

## Architecture

This is a circuit breaker implementation that gates traffic between two backends: **Claude** and **OpenAI**.

### State machine (`domain.go`)

Three states implement the `state` interface (`run`, `getName`):

- **`closedState`** — healthy; all traffic goes to Claude. Transitions to `halfOpenState` on `FAILURE`.
- **`halfOpenState`** — degraded; traffic split per `Config.Halfopentraffic`. Transitions to `closedState` on `SUCCESS`, or `openState` on `FAILURE`.
- **`openState`** — critical; traffic split per `Config.Opentraffic` (mostly rejected). Transitions to `halfOpenState` on `SUCCESS`.

State transitions are registered in `Newservice` via `addTransition(from, to, event)` and fired inside each state's `run()` method by calling `triggerEvent(event)`. On every transition, both error trackers are reset.

### Error tracking (`domain.go`)

`errorTracker` uses a two-bucket sliding window (60s buckets). `getErrorRate()` returns a weighted average of the current and previous bucket, proportional to how much of the current bucket has elapsed. Each state's `run()` calls `registerRequest()` and conditionally `registerError()` after every call.

`serviceImp.getErrorRate()` returns the **weighted** error rate for the current state, using the traffic split weights for `halfOpenState` and `openState`, and plain Claude rate for `closedState`.

### Configuration (`domain.go`)

```go
Config{
    Closederrorrate:   5,   // % — threshold to leave closedState
    Halfopenerrorrate: 90,  // % — threshold to leave halfOpenState → openState

    Closedtraffic:   Trafficsplit{Claude: 100, Openapi: 0,  Reject: 0},
    Halfopentraffic: Trafficsplit{Claude: 5,   Openapi: 95, Reject: 0},
    Opentraffic:     Trafficsplit{Claude: 5,   Openapi: 5,  Reject: 90},
}
```

### Concurrency

`serviceImp` uses a `sync.RWMutex` (`lock`). Each state's `run()` acquires a write lock before mutating error tracker state.
