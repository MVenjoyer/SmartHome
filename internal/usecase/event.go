package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
	"time"
)

var ErrInvalidEvent = errors.New("invalid event")

type Event struct {
	eventRepo  EventRepository
	sensorRepo SensorRepository
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{eventRepo: er, sensorRepo: sr}
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	if time.Time.IsZero(event.Timestamp) {
		return ErrInvalidEventTimestamp
	}
	sensor, err := e.sensorRepo.GetSensorBySerialNumber(ctx, event.SensorSerialNumber)
	if err != nil {
		return err
	} else if sensor == nil {
		return ErrSensorNotFound
	}
	event.SensorID = sensor.ID
	if err = e.eventRepo.SaveEvent(ctx, event); err != nil {
		return err
	}
	sensor.CurrentState = event.Payload
	sensor.LastActivity = event.Timestamp
	return e.sensorRepo.SaveSensor(ctx, sensor)
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	events, err := e.eventRepo.GetLastEventBySensorID(ctx, id)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (e *Event) GetSensorHistory(ctx context.Context, id int64, startDate, endDate time.Time) ([]domain.Event, error) {
	events, err := e.eventRepo.GetSensorHistory(ctx, id, startDate, endDate)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, ErrEventNotFound
	}
	return events, nil
}
