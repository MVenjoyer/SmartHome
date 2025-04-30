package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

var ErrInvalidSensor = errors.New("invalid sensor")

type SensorRepository struct {
	serialNumIDMap  map[int64]string
	sensorSerialNum map[string]*domain.Sensor
	mu              sync.RWMutex
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{serialNumIDMap: make(map[int64]string), sensorSerialNum: make(map[string]*domain.Sensor)}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if sensor == nil {
		return ErrInvalidSensor
	}
	sensor.RegisteredAt = time.Now()
	sensor.ID = int64(len(r.sensorSerialNum) + 1)
	r.serialNumIDMap[sensor.ID] = sensor.SerialNumber
	r.sensorSerialNum[sensor.SerialNumber] = sensor
	return nil
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sensors := make([]domain.Sensor, 0, len(r.sensorSerialNum))
	for _, sensor := range r.sensorSerialNum {
		if sensor == nil {
			continue
		}
		sensors = append(sensors, *sensor)
	}
	return sensors, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	num, ok := r.serialNumIDMap[id]
	if !ok {
		return nil, usecase.ErrSensorNotFound
	}
	return r.sensorSerialNum[num], nil
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sensor, ok := r.sensorSerialNum[sn]
	if !ok {
		return nil, usecase.ErrSensorNotFound
	}
	return sensor, nil
}
