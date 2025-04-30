package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

var ErrEventIsNil = errors.New("event is nil")

type EventRepository struct {
	events map[int64][]*domain.Event
	mu     sync.RWMutex
}

func NewEventRepository() *EventRepository {
	return &EventRepository{events: make(map[int64][]*domain.Event)}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if event == nil {
		return ErrEventIsNil
	}
	r.events[event.SensorID] = append(r.events[event.SensorID], event)
	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	eventLst, ok := r.events[id]
	if !ok || len(eventLst) == 0 {
		return nil, usecase.ErrEventNotFound
	}
	result := eventLst[0]
	for _, event := range eventLst {
		if event.Timestamp.After(result.Timestamp) {
			result = event
		}
	}
	return result, nil
}

func (r *EventRepository) GetSensorHistory(ctx context.Context, id int64, startDate, endDate time.Time) ([]domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if startDate.After(endDate) {
		return nil, usecase.ErrInvalidEventTimestamp
	}
	eventLst, ok := r.events[id]
	if !ok {
		return nil, usecase.ErrEventNotFound
	}
	events := make([]domain.Event, 0, len(eventLst))
	for _, event := range eventLst {
		if event.SensorID == id && startDate.Before(event.Timestamp) && endDate.After(event.Timestamp) {
			events = append(events, *event)
		}
	}
	if len(events) == 0 {
		return nil, usecase.ErrEventNotFound
	}
	return events, nil
}
