package memorystorage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/errors"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/models"
	"github.com/google/uuid"
)

type LocalStorage struct {
	mu     sync.RWMutex
	events map[string]*models.Event
	logger logger.Logger
}

func NewLocalStorage(logger logger.Logger) *LocalStorage {
	return &LocalStorage{
		events: make(map[string]*models.Event),
		logger: logger,
	}
}

func (s *LocalStorage) EventCreate(_ context.Context, req *models.CreateEventReq) (*models.Event, error) {
	s.logger.Debug("EventCreate called")

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ev := range s.events {
		if ev.User == req.User && rangesOverlap(req.Date, req.EndTime, ev.Date, ev.EndTime) {
			s.logger.Error(fmt.Sprintf("conflict on create for user=%s at %s-%s", req.User, req.Date, req.EndTime))
			return nil, fmt.Errorf("event create: %w", errors.ErrDateBusy)
		}
	}

	id := uuid.New().String()
	event := &models.Event{
		ID:           id,
		Title:        req.Title,
		Date:         req.Date,
		EndTime:      req.EndTime,
		Description:  req.Description,
		User:         req.User,
		NotifyBefore: req.NotifyBefore,
	}

	s.events[id] = event
	s.logger.Debug("event created id=" + id)
	return event, nil
}

func (s *LocalStorage) EventEdit(_ context.Context, req *models.EditEventReq) (*models.Event, error) {
	s.logger.Debug("EventEdit called")

	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.events[req.ID]
	if !ok {
		s.logger.Error("event not found id=" + req.ID)
		return nil, fmt.Errorf("event edit: %w", errors.ErrEventNotFound)
	}

	updated := *event

	if req.Title != nil {
		updated.Title = *req.Title
	}
	if req.Date != nil {
		updated.Date = *req.Date
	}
	if req.EndTime != nil {
		updated.EndTime = *req.EndTime
	}
	if req.Description != nil {
		updated.Description = req.Description
	}
	if req.User != nil {
		updated.User = *req.User
	}
	if req.NotifyBefore != nil {
		updated.NotifyBefore = req.NotifyBefore
	}

	for _, ev := range s.events {
		if ev.ID != req.ID && ev.User == updated.User && rangesOverlap(updated.Date, updated.EndTime, ev.Date, ev.EndTime) {
			s.logger.Error("conflict on edit id=" + req.ID)
			return nil, fmt.Errorf("event edit: %w", errors.ErrDateBusy)
		}
	}

	s.events[req.ID] = &updated
	s.logger.Debug("event edited id=" + req.ID)
	return &updated, nil
}

func (s *LocalStorage) EventDelete(_ context.Context, req *models.EventIDReq) error {
	s.logger.Debug("EventDelete called id=" + req.ID)

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.events, req.ID)
	s.logger.Debug("event deleted id=" + req.ID)
	return nil
}

func (s *LocalStorage) EventGet(_ context.Context, req *models.EventIDReq) (*models.Event, error) {
	s.logger.Debug("EventGet called id=" + req.ID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.events[req.ID]
	if !ok {
		s.logger.Error("event not found id=" + req.ID)
		return nil, fmt.Errorf("event get: %w", errors.ErrEventNotFound)
	}

	cpy := *event
	s.logger.Debug("event fetched id=" + req.ID)
	return &cpy, nil
}

func (s *LocalStorage) EventGetList(_ context.Context, req *models.CreateEventReq) (*models.GetEventListResp, error) {
	s.logger.Debug("EventGetList called for user=" + req.User)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.Event
	for _, ev := range s.events {
		if ev.User == req.User {
			result = append(result, *ev)
		}
	}

	s.logger.Debug(fmt.Sprintf("event list returned user=%s count=%d", req.User, len(result)))
	return &models.GetEventListResp{Data: result}, nil
}

func rangesOverlap(start1Str, end1Str, start2Str, end2Str string) bool {
	start1, err1 := time.Parse(time.RFC3339, start1Str)
	end1, err2 := time.Parse(time.RFC3339, end1Str)
	start2, err3 := time.Parse(time.RFC3339, start2Str)
	end2, err4 := time.Parse(time.RFC3339, end2Str)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return false
	}
	return start1.Before(end2) && start2.Before(end1)
}
