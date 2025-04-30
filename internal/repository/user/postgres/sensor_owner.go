package postgres

import (
	"context"
	"homework/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	saveSensorOwnerRequest = `INSERT INTO sensors_users (sensor_id,user_id) VALUES ($1,$2)`
	getSensorOwnerRequest  = `SELECT sensor_id,user_id FROM sensors_users WHERE user_id = $1`
)

type SensorOwnerRepository struct {
	pool *pgxpool.Pool
}

func NewSensorOwnerRepository(pool *pgxpool.Pool) *SensorOwnerRepository {
	return &SensorOwnerRepository{
		pool: pool,
	}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, saveSensorOwnerRequest, sensorOwner.SensorID, sensorOwner.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, getSensorOwnerRequest, userID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.SensorOwner, 0)
	for rows.Next() {
		var sensorOwner domain.SensorOwner
		if err := rows.Scan(&sensorOwner.SensorID, &sensorOwner.UserID); err == nil {
			result = append(result, sensorOwner)
		}
	}
	return result, nil
}
