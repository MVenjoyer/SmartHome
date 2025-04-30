package postgres

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidSensor = errors.New("invalid sensor")

const saveSensorRequest = `
INSERT INTO sensors (serial_number, type, current_state, description, is_active, registered_at, last_activity)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (serial_number) 
DO UPDATE SET 
    type = EXCLUDED.type,
    current_state = EXCLUDED.current_state,
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    registered_at = EXCLUDED.registered_at,
    last_activity = EXCLUDED.last_activity
`

const getSensorByIdRequest = `SELECT id,serial_number,type,current_state,description,is_active,registered_at,last_activity
FROM sensors WHERE id = $1`

const getSensorBySerialNumRequest = `SELECT id,serial_number,type,current_state,description,is_active,registered_at,last_activity
FROM sensors WHERE serial_number = $1`

type SensorRepository struct {
	pool *pgxpool.Pool
}

func NewSensorRepository(pool *pgxpool.Pool) *SensorRepository {
	return &SensorRepository{
		pool: pool,
	}
}

func compress(sensor *domain.Sensor) []any {
	return []any{
		sensor.SerialNumber, sensor.Type,
		sensor.CurrentState, sensor.Description,
		sensor.IsActive, sensor.RegisteredAt, sensor.LastActivity,
	}
}

func scanSensor[T interface{ Scan(...interface{}) error }](rows T, sensor *domain.Sensor) error {
	return rows.Scan(
		&sensor.ID,
		&sensor.SerialNumber,
		&sensor.Type,
		&sensor.CurrentState,
		&sensor.Description,
		&sensor.IsActive,
		&sensor.RegisteredAt,
		&sensor.LastActivity,
	)
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if sensor == nil {
		return ErrInvalidSensor
	}
	sensor.RegisteredAt = time.Now()
	_, err := r.pool.Exec(ctx, saveSensorRequest, compress(sensor)...)
	if err != nil {
		return err
	}
	return nil
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT * FROM sensors`)
	if err != nil {
		return nil, err
	}
	sensors := make([]domain.Sensor, 0)
	for rows.Next() {
		var sensor domain.Sensor
		if err := scanSensor(rows, &sensor); err == nil {
			sensors = append(sensors, sensor)
		}
	}
	return sensors, nil
}

func get[T int64 | string](ctx context.Context, r *SensorRepository, request string, input T) (*domain.Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, request, input)
	sensor := domain.Sensor{}
	err := scanSensor(row, &sensor)
	if err != nil {
		return nil, usecase.ErrSensorNotFound
	}
	return &sensor, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	return get(ctx, r, getSensorByIdRequest, id)
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	return get(ctx, r, getSensorBySerialNumRequest, sn)
}
