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

func (s *LocalStorage) EventGetList(_ context.Context, req *models.GetEventListReq) (*models.GetEventListResp, error) {
	s.logger.Debug("EventGetList called")

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Event, 0, len(s.events))
	for _, ev := range s.events {
		if req.Start != nil && ev.EndTime.Before(*req.Start) {
			continue
		}
		if req.End != nil && ev.Date.After(*req.End) {
			continue
		}

		result = append(result, *ev)
	}

	s.logger.Debug(fmt.Sprintf("event list returned count=%d", len(result)))
	return &models.GetEventListResp{Data: result}, nil
}

func (s *LocalStorage) EventsToNotify(_ context.Context) ([]models.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	result := make([]models.Event, 0, len(s.events))
	for _, event := range s.events {
		var notifyBefore time.Duration
		var err error
		if event.NotifyBefore != nil {
			notifyBefore, err = time.ParseDuration(*event.NotifyBefore)
			if err != nil {
				notifyBefore = 0
			}
		}
		notifyAt := event.Date.Add(-notifyBefore)
		if now.After(notifyAt) && now.Before(event.Date) {
			result = append(result, *event)
		}
	}
	return result, nil
}

func (s *LocalStorage) DeleteOldEvents(_ context.Context, cutoff time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, event := range s.events {
		if event.EndTime.Before(cutoff) {
			delete(s.events, id)
		}
	}
	return nil
}

func rangesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}
