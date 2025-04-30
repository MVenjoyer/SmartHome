package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
)

var ErrSensorIsNil = errors.New("sensor is nil")

type Sensor struct {
	sensorRepo SensorRepository
}

func NewSensor(sr SensorRepository) *Sensor {
	return &Sensor{sensorRepo: sr}
}

func (s *Sensor) RegisterSensor(ctx context.Context, sensor *domain.Sensor) (*domain.Sensor, error) {
	if sensor == nil {
		return nil, ErrSensorIsNil
	}
	if sensor.Type != domain.SensorTypeADC && sensor.Type != domain.SensorTypeContactClosure {
		return nil, ErrWrongSensorType
	}
	if len(sensor.SerialNumber) != 10 {
		return nil, ErrWrongSensorSerialNumber
	}
	existingSensor, err := s.sensorRepo.GetSensorBySerialNumber(ctx, sensor.SerialNumber)
	if err != nil {
		if errors.Is(err, ErrSensorNotFound) {
			if err := s.sensorRepo.SaveSensor(ctx, sensor); err != nil {
				return nil, err
			}
			return sensor, nil
		}
		return nil, err
	}
	return existingSensor, nil
}

func (s *Sensor) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	return s.sensorRepo.GetSensors(ctx)
}

func (s *Sensor) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	return s.sensorRepo.GetSensorByID(ctx, id)
}
