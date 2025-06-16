package storage

import (
	"context"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/models"
)

type Storage interface {
	EventCreate(ctx context.Context, req *models.CreateEventReq) (*models.Event, error)

	EventEdit(ctx context.Context, req *models.EditEventReq) (*models.Event, error)

	EventDelete(ctx context.Context, req *models.EventIDReq) error

	EventGet(ctx context.Context, req *models.EventIDReq) (*models.Event, error)

	EventGetList(ctx context.Context, req *models.GetEventListReq) (*models.GetEventListResp, error)

	EventsToNotify(ctx context.Context) ([]models.Event, error)

	DeleteOldEvents(ctx context.Context, cutoff time.Time) error
}
