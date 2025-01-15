package extension

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HistoryReqItem represents the raw history item from Chrome browser
type HistoryReqItem struct {
	ID            string  `json:"id"`
	LastVisitTime float64 `json:"lastVisitTime"`
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	VisitCount    int     `json:"visitCount"`
}

// HistoryItem represents our internal history item with proper time type
type HistoryItem struct {
	ID            string    `json:"id"`
	LastVisitTime time.Time `json:"lastVisitTime"`
	Title         string    `json:"title"`
	URL           string    `json:"url"`
	VisitCount    int       `json:"visitCount"`
}

// ProcessFunc is the function type for processing history items in batch
type ProcessFunc func(items []HistoryItem) error

// toHistoryItem converts a BrowserHistoryReqItem to HistoryItem
func toHistoryItem(req HistoryReqItem) HistoryItem {
	// Chrome's lastVisitTime is in milliseconds since epoch
	lastVisit := time.UnixMilli(int64(req.LastVisitTime))

	return HistoryItem{
		ID:            req.ID,
		LastVisitTime: lastVisit,
		Title:         req.Title,
		URL:           req.URL,
		VisitCount:    req.VisitCount,
	}
}

// Server represents the history server
type Server struct {
	server    *http.Server
	wg        sync.WaitGroup
	processor ProcessFunc
}

// New creates a new history server instance
func New(processor ProcessFunc) *Server {
	if processor == nil {
		panic("processor function cannot be nil")
	}

	return &Server{
		server: &http.Server{
			Addr: "localhost:8765",
		},
		processor: processor,
	}
}

// processHistoryItems processes the history items using the processor function
func (s *Server) processHistoryItems(items []HistoryItem) {
	if err := s.processor(items); err != nil {
		fmt.Printf("Error processing items: %v\n", err)
	}
}

// Start initializes and starts the server. The server will be shutdown when the context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Root endpoint handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// Browser history endpoint handler
	mux.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var reqItems []HistoryReqItem
		if err := json.NewDecoder(r.Body).Decode(&reqItems); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Convert request items to internal format
		historyItems := make([]HistoryItem, len(reqItems))
		for i, item := range reqItems {
			historyItems[i] = toHistoryItem(item)
		}

		// Process the items
		s.processHistoryItems(historyItems)

		w.WriteHeader(http.StatusOK)
	})

	s.server.Handler = mux

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		if err := s.Close(); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		}
	}()

	return nil
}

// Close gracefully shuts down the server
func (s *Server) Close() error {
	if err := s.server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}
	s.wg.Wait()
	return nil
}
