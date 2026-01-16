package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/victorzhuk/go-ent/internal/execution"
)

type contextKey string

const sessionKey contextKey = "session"

const (
	defaultChannelSize = 100
	channelTimeout     = 1 * time.Millisecond
)

type sessionState struct {
	startTime time.Time
	toolName  string
}

type Collector struct {
	store *Store
	ch    chan Metric
	done  chan struct{}
	wg    sync.WaitGroup

	mu     sync.Mutex
	starts map[string]sessionState
}

func NewCollector(store *Store) *Collector {
	c := &Collector{
		store:  store,
		ch:     make(chan Metric, defaultChannelSize),
		done:   make(chan struct{}),
		starts: make(map[string]sessionState),
	}

	c.wg.Add(1)
	go c.run()

	return c
}

func (c *Collector) StartExecution(ctx context.Context, toolName string) string {
	sessionID := uuid.Must(uuid.NewV7()).String()

	c.mu.Lock()
	c.starts[sessionID] = sessionState{
		startTime: time.Now(),
		toolName:  toolName,
	}
	c.mu.Unlock()

	return sessionID
}

func (c *Collector) EndExecution(ctx context.Context, result *execution.Result) error {
	sessionID, ok := ctx.Value(sessionKey).(string)
	if !ok || sessionID == "" {
		return fmt.Errorf("no session: %w", ErrNoSession)
	}

	c.mu.Lock()
	state, ok := c.starts[sessionID]
	if ok {
		delete(c.starts, sessionID)
	}
	c.mu.Unlock()

	if !ok {
		return fmt.Errorf("session %q not started: %w", sessionID, ErrSessionNotStarted)
	}

	metric := Metric{
		SessionID: sessionID,
		ToolName:  state.toolName,
		TokensIn:  result.TokensIn,
		TokensOut: result.TokensOut,
		Duration:  result.Duration,
		Success:   result.Success,
		ErrorMsg:  result.Error,
		Timestamp: time.Now(),
	}

	select {
	case c.ch <- metric:
	case <-time.After(channelTimeout):
		slog.Warn("metrics channel full, dropping metric", "session", sessionID)
	}

	return nil
}

func (c *Collector) Shutdown(ctx context.Context) error {
	close(c.done)

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown: %w", ctx.Err())
	}
}

func (c *Collector) run() {
	defer c.wg.Done()

	for {
		select {
		case <-c.done:
			c.drain()
			return

		case m := <-c.ch:
			if err := c.store.Add(m); err != nil {
				slog.Error("add metric", "error", err)
			}
		}
	}
}

func (c *Collector) drain() {
	for {
		select {
		case m := <-c.ch:
			if err := c.store.Add(m); err != nil {
				slog.Error("add metric during drain", "error", err)
			}
		default:
			return
		}
	}
}

func ContextWithSession(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionKey, sessionID)
}

func SessionFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionKey).(string)
	return sessionID, ok
}
