package http

import (
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/generated"
	"homework/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/strfmt"
)

func setupRouter(r *gin.Engine, us UseCases, handler *WebSocketHandler) {
	r.HandleMethodNotAllowed = true
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
	r.NoMethod(handleNoMethod)
	routes := []struct {
		method, path string
		handler      gin.HandlerFunc
	}{
		{"POST", "/users", createUser(us.User)},
		{"GET", "/users/:id/sensors", getUser(us.User)},
		{"HEAD", "/users/:id/sensors", headUser(us.User)},
		{"POST", "/users/:id/sensors", createUserId(us.User)},
		{"GET", "/sensors", getSensors(us.Sensor)},
		{"HEAD", "/sensors", headSensor(us.Sensor)},
		{"POST", "/sensors", createSensor(us.Sensor)},
		{"GET", "/sensors/:id", getSensorById(us.Sensor)},
		{"HEAD", "/sensors/:id", headSensorId(us.Sensor)},
		{"POST", "/events", createEvent(us.Event)},
		{"GET", "/sensors/:id/history", getHistory(us.Event)},
	}
	r.GET("/sensors/:id/events", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		s, err := us.Sensor.GetSensorByID(c.Request.Context(), id)
		if err != nil && errors.Is(err, usecase.ErrSensorNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if err = handler.Handle(c, s.ID); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
	})
	for _, route := range routes {
		r.Handle(route.method, route.path, route.handler)
	}
	optionsRoutes := []string{"/users", "/events", "/sensors", "/sensors/:id", "/users/:id/sensors"}
	for _, path := range optionsRoutes {
		r.OPTIONS(path, optionsHandler(path))
	}
}

func handleNoMethod(c *gin.Context) {
	if c.FullPath() == "/users" {
		c.Header("Allow", "OPTIONS, POST")
	}
	c.Status(http.StatusMethodNotAllowed)
}

func optionsHandler(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		allow := "OPTIONS,POST"
		if path != "/users" && path != "/events" {
			allow = "OPTIONS,GET,HEAD,POST"
		}
		c.Header("Allow", allow)
		c.Status(http.StatusNoContent)
	}
}

func headUser(user *usecase.User) gin.HandlerFunc {
	return head(func(context *gin.Context) error {
		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.AbortWithStatus(http.StatusUnprocessableEntity)
		}
		_, err = user.GetUserSensors(context.Request.Context(), int64(id))
		return err
	})
}

func headSensorId(sensor *usecase.Sensor) gin.HandlerFunc {
	return head(func(c *gin.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
		}
		_, err = sensor.GetSensorByID(c.Request.Context(), int64(id))
		return err
	})
}

func headSensor(_ *usecase.Sensor) gin.HandlerFunc {
	return head(func(_ *gin.Context) error { return nil })
}

func head(foo func(ctx *gin.Context) error) gin.HandlerFunc {
	return func(context *gin.Context) {
		if context.GetHeader("accept") != "application/json" {
			context.AbortWithStatus(http.StatusNotAcceptable)
			return
		}
		if foo(context) != nil {
			context.AbortWithStatus(http.StatusNotFound)
			return
		}
		context.Header("Content-Length", "0")
		context.Status(http.StatusOK)
	}
}

func getHistory(event *usecase.Event) gin.HandlerFunc {
	return get(func(c *gin.Context) ([]*generated.SensorHistory, error) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return nil, err
		}
		start := c.Query("start_date")
		startTime, err := time.Parse(time.RFC3339, start)
		if err != nil && start != "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return nil, err
		} else if start == "" {
			startTime = time.Time{}
		}
		end := c.Query("end_date")
		endTime, err := time.Parse(time.RFC3339, end)
		if err != nil && end != "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return nil, err
		} else if end == "" {
			now := time.Now()
			_, offset := now.Zone()
			endTime = now.Add(time.Second * time.Duration(offset))
		}
		if startTime.After(endTime) {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return nil, nil
		}
		events, err := event.GetSensorHistory(c.Request.Context(), id, startTime, endTime)
		if err != nil && errors.Is(err, usecase.ErrEventNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return nil, err
		}
		result := make([]*generated.SensorHistory, 0, len(events))
		for _, event := range events {
			result = append(result, &generated.SensorHistory{
				Payload:   &event.Payload,
				Timestamp: (*strfmt.DateTime)(&event.Timestamp),
			})
		}
		return result, nil
	}, http.StatusNotFound)
}

func getSensorById(sensor *usecase.Sensor) gin.HandlerFunc {
	return get(func(ctx *gin.Context) (*domain.Sensor, error) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
		}
		return sensor.GetSensorByID(ctx.Request.Context(), int64(id))
	}, http.StatusNotFound)
}

func getUser(user *usecase.User) gin.HandlerFunc {
	return get(func(ctx *gin.Context) ([]domain.Sensor, error) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnprocessableEntity)
		}
		return user.GetUserSensors(ctx.Request.Context(), int64(id))
	}, http.StatusNotFound)
}

func getSensors(sensor *usecase.Sensor) gin.HandlerFunc {
	return get(func(ctx *gin.Context) ([]domain.Sensor, error) {
		return sensor.GetSensors(ctx.Request.Context())
	}, http.StatusInternalServerError)
}

type getResponse interface {
	[]domain.Sensor | *domain.Sensor | *domain.Event | []*generated.SensorHistory
}

func get[T getResponse](foo func(ctx *gin.Context) (T, error), errStat int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("accept") != "application/json" {
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}
		s, err := foo(c)
		if err != nil {
			c.AbortWithStatus(errStat)
			return
		}
		c.JSON(http.StatusOK, s)
	}
}

func createEvent(event *usecase.Event) gin.HandlerFunc {
	return create(event, func(ctx *gin.Context, uc *usecase.Event, req *generated.SensorEvent) (*domain.Event, error) {
		return nil, uc.ReceiveEvent(ctx.Request.Context(), &domain.Event{
			SensorSerialNumber: *req.SensorSerialNumber,
			Payload:            *req.Payload,
			Timestamp:          time.Now(),
		})
	}, func(_ *domain.Event, req *generated.SensorEvent) *generated.SensorEvent {
		return req
	}, http.StatusCreated, http.StatusInternalServerError)
}

func createSensor(sensor *usecase.Sensor) gin.HandlerFunc {
	return create(
		sensor,
		func(ctx *gin.Context, uc *usecase.Sensor, req *generated.SensorToCreate) (*domain.Sensor, error) {
			return uc.RegisterSensor(ctx.Request.Context(), &domain.Sensor{
				SerialNumber: *req.SerialNumber,
				Type:         domain.SensorType(*req.Type), Description: *req.Description, IsActive: *req.IsActive,
			})
		},
		func(ds *domain.Sensor, _ *generated.SensorToCreate) *generated.Sensor {
			return &generated.Sensor{
				ID:           &ds.ID,
				SerialNumber: &ds.SerialNumber,
				Type:         (*string)(&ds.Type),
				Description:  &ds.Description,
				IsActive:     &ds.IsActive,
				RegisteredAt: (*strfmt.DateTime)(&ds.RegisteredAt),
			}
		}, http.StatusOK, http.StatusInternalServerError)
}

func createUserId(user *usecase.User) gin.HandlerFunc {
	return create(user, func(ctx *gin.Context, uc *usecase.User, req *generated.SensorToUserBinding) (*domain.User, error) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			return nil, err
		}
		return nil, uc.AttachSensorToUser(ctx.Request.Context(), int64(id), *req.SensorID)
	}, func(_ *domain.User, req *generated.SensorToUserBinding) *generated.SensorToUserBinding {
		return req
	}, http.StatusCreated, http.StatusNotFound)
}

func createUser(uc *usecase.User) gin.HandlerFunc {
	return create(uc, func(ctx *gin.Context, uc *usecase.User, req *generated.UserToCreate) (*domain.User, error) {
		return uc.RegisterUser(ctx.Request.Context(), &domain.User{Name: *req.Name})
	}, func(ds *domain.User, _ *generated.UserToCreate) *generated.User {
		return &generated.User{
			ID:   &ds.ID,
			Name: &ds.Name,
		}
	}, http.StatusOK, http.StatusInternalServerError)
}

type (
	domainStructs interface {
		domain.Sensor | domain.User | domain.Event
	}
	useCase interface {
		usecase.Sensor | usecase.User | usecase.Event
	}
	generatedRequest interface {
		*generated.UserToCreate | *generated.SensorToCreate | *generated.SensorEvent | *generated.SensorToUserBinding
		Validate(formats strfmt.Registry) error
	}
	generatedResponse interface {
		*generated.User | *generated.Sensor | *generated.SensorEvent | *generated.SensorToUserBinding
	}
)

func create[
	T useCase,
	Q domainStructs,
	Req generatedRequest,
	Res generatedResponse,
](
	uc *T,
	foo func(ctx *gin.Context, uc *T, r Req) (*Q, error),
	resFoo func(*Q, Req) Res,
	resStat, errStat int,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if c.GetHeader("content-type") != "application/json" {
			c.AbortWithStatus(http.StatusUnsupportedMediaType)
			return
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if req == nil {
			return
		}
		if err := req.Validate(nil); err != nil {
			c.AbortWithStatus(http.StatusUnprocessableEntity)
			return
		}
		ds, err := foo(c, uc, req)
		if err != nil {
			fmt.Println("Oshibochka", err)
			c.AbortWithStatus(errStat)
			return
		}
		c.JSON(resStat, resFoo(ds, req))
	}
}
