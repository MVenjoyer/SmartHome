package http

import (
	"context"
	"homework/internal/domain"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	useCases UseCases
	result   map[int64]chan *domain.Event
	mu       sync.RWMutex
	done     chan struct{}
}

func NewWebSocketHandler(useCases UseCases) *WebSocketHandler {
	return &WebSocketHandler{
		useCases: useCases,
		result:   make(map[int64]chan *domain.Event),
		done:     make(chan struct{}),
	}
}

func (h *WebSocketHandler) Handle(c *gin.Context, id int64) error {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{})
	if err != nil {
		return err
	}
	ctx := conn.CloseRead(c)
	h.mu.Lock()
	if _, ok := h.result[id]; !ok {
		h.result[id] = make(chan *domain.Event, 1)
		go h.produce(ctx, id)
	}
	h.mu.Unlock()
	go h.consume(ctx, conn, id)

	return nil
}

func (h *WebSocketHandler) Shutdown() error {
	close(h.done)
	h.mu.Lock()
	for id := range h.result {
		close(h.result[id])
		delete(h.result, id)
	}
	h.mu.Unlock()
	return nil
}

func (h *WebSocketHandler) produce(ctx context.Context, id int64) {
	h.mu.RLock()
	ch := h.result[id]
	h.mu.RUnlock()
	select {
	case <-ctx.Done():
		return
	case <-h.done:
		return
	case <-time.After(time.Second):
		event, err := h.useCases.Event.GetLastEventBySensorID(ctx, id)
		if err != nil || event == nil {
			return
		}
		select {
		case ch <- event:
		default:
		}
	}
}

func (h *WebSocketHandler) consume(ctx context.Context, conn *websocket.Conn, id int64) {
	defer conn.Close(websocket.StatusNormalClosure, "Connection closed")
	h.mu.RLock()
	ch := h.result[id]
	h.mu.RUnlock()
	for {
		select {
		case <-ctx.Done():
			return
		case <-h.done:
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			err := wsjson.Write(ctx, conn, event)
			if err != nil {
				return
			}
		}
	}
}
