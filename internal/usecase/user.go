package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
)

var ErrUserIsNil = errors.New("user is nil")

type User struct {
	sensorOwnerRepo SensorOwnerRepository
	userRepo        UserRepository
	sensorRepo      SensorRepository
}

func NewUser(ur UserRepository, sor SensorOwnerRepository, sr SensorRepository) *User {
	return &User{sensorOwnerRepo: sor, userRepo: ur, sensorRepo: sr}
}

func (u *User) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if user == nil {
		return nil, ErrUserIsNil
	}
	if user.Name == "" {
		return nil, ErrInvalidUserName
	}
	if err := u.userRepo.SaveUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) AttachSensorToUser(ctx context.Context, userID, sensorID int64) error {
	_, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	_, err = u.sensorRepo.GetSensorByID(ctx, sensorID)
	if err != nil {
		return err
	}
	sensorOwner := domain.SensorOwner{SensorID: sensorID, UserID: userID}
	return u.sensorOwnerRepo.SaveSensorOwner(ctx, sensorOwner)
}

func (u *User) GetUserSensors(ctx context.Context, userID int64) ([]domain.Sensor, error) {
	_, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sensors, err := u.sensorOwnerRepo.GetSensorsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sensorsIDs := make([]domain.Sensor, 0, len(sensors))
	for _, sensor := range sensors {
		sn, err := u.sensorRepo.GetSensorByID(ctx, sensor.SensorID)
		if err != nil {
			return nil, err
		}
		sensorsIDs = append(sensorsIDs, *sn)
	}
	return sensorsIDs, nil
}
