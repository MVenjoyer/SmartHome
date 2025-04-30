package main

import (
	"context"
	"errors"
	"homework/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	httpGateway "homework/internal/gateways/http"
	eventRepository "homework/internal/repository/event/postgres"
	sensorRepository "homework/internal/repository/sensor/postgres"
	userRepository "homework/internal/repository/user/postgres"
)

type Config struct {
	DBURL       string
	ServerPort  int
	ReadTimeout time.Duration
}

func loadConfig() Config {
	cfg := Config{
		DBURL:       "postgres://postgres:postgres@127.0.0.1:5432/db",
		ServerPort:  8080,
		ReadTimeout: 10 * time.Second,
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DBURL = dbURL
	}

	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.ServerPort = port
		}
	}

	return cfg
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	cfg := loadConfig()
	config, err := pgxpool.ParseConfig(cfg.DBURL)
	if err != nil {
		log.Fatalf("can't parse pgxpool config")
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("can't create new pool")
	}
	defer pool.Close()

	er := eventRepository.NewEventRepository(pool)
	sr := sensorRepository.NewSensorRepository(pool)
	ur := userRepository.NewUserRepository(pool)
	sor := userRepository.NewSensorOwnerRepository(pool)

	useCases := httpGateway.UseCases{
		Event:  usecase.NewEvent(er, sr),
		Sensor: usecase.NewSensor(sr),
		User:   usecase.NewUser(ur, sor, sr),
	}

	r := httpGateway.NewServer(useCases)
	if err := r.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error during server shutdown: %v", err)
	}
}
