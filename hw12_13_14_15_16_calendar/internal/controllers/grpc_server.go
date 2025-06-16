package controllers

import (
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/configuration"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage"
)

type CalendarHandler struct {
	storage storage.Storage
	cfg     **configuration.Config
}

func NewCalendarHandler(storage storage.Storage, cfg **configuration.Config) *CalendarHandler {
	return &CalendarHandler{storage: storage, cfg: cfg}
}
