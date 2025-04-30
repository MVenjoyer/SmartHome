package postgres

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEventIsNil = errors.New("event is nil")

const (
	saveEventRequest = `INSERT INTO events (timestamp,sensor_serial_number,sensor_id,payload) VALUES ($1,$2,$3,$4)`
	getEventRequest  = `SELECT timestamp,sensor_serial_number,sensor_id,payload FROM events WHERE sensor_id = $1`
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		pool,
	}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if event == nil {
		return ErrEventIsNil
	}
	_, err := r.pool.Exec(ctx, saveEventRequest, event.Timestamp, event.SensorSerialNumber, event.SensorID, event.Payload)
	if err != nil {
		return err
	}
	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, getEventRequest, id)
	if err != nil || !rows.Next() {
		return nil, usecase.ErrEventNotFound
	}
	defer rows.Close()
	result := domain.Event{}
	if err := rows.Scan(&result.Timestamp, &result.SensorSerialNumber, &result.SensorID, &result.Payload); err != nil {
		return nil, err
	}
	event := domain.Event{}
	for rows.Next() {
		if err := rows.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload); err == nil {
			if event.Timestamp.After(result.Timestamp) {
				result = event
			}
		}
	}
	return &result, nil
}

func (r *EventRepository) GetSensorHistory(ctx context.Context, id int64, startDate, endDate time.Time) ([]domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if startDate.After(endDate) {
		return nil, usecase.ErrInvalidEventTimestamp
	}
	rows, err := r.pool.Query(ctx, getEventRequest, id)
	if err != nil {
		return nil, usecase.ErrEventNotFound
	}
	defer rows.Close()
	events := make([]domain.Event, 0)
	event := domain.Event{}
	for rows.Next() {
		if err := rows.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload); err == nil {
			if startDate.Before(event.Timestamp) && endDate.After(event.Timestamp) {
				events = append(events, event)
			}
		}
	}
	if len(events) == 0 {
		return nil, usecase.ErrEventNotFound
	}
	return events, nil
}
