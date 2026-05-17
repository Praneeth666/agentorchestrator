package agentorchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
)

// Controller holds a reference to the service layer and exposes HTTP handlers.
type Controller struct {
	svc service
}

// NewController constructs a Controller backed by the given service.
func NewController(svc service) *Controller {
	return &Controller{svc: svc}
}

// HandleRequest decodes a JSON request body and forwards it to the service layer.
// Responds with 400 on decode failure, 500 if the service returns an error, and 200 on success.
func (c *Controller) HandleRequest(w http.ResponseWriter, r *http.Request) {
	var body interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := c.svc.Request(body); err != nil {
		fmt.Printf("service error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Start registers routes and begins listening for HTTP requests on addr (e.g. ":8080").
// It blocks until SIGINT/SIGTERM is received, then gracefully drains in-flight requests.
func (c *Controller) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/request", c.HandleRequest)

	srv := &http.Server{Addr: addr, Handler: mux}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		fmt.Printf("listening on %s\n", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("listen error: %v\n", err)
		}
	}()

	<-ctx.Done()
	return srv.Shutdown(context.Background())
}
