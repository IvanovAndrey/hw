package memorystorage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/errors"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() logger.Logger {
	return *logger.NewLogger("calendar", "test", "error")
}

func newCreateReq(user, title string, start, end time.Time) *models.CreateEventReq {
	return &models.CreateEventReq{
		Title:       title,
		Date:        start.Format(time.RFC3339),
		EndTime:     end.Format(time.RFC3339),
		Description: nil,
		User:        user,
	}
}

func TestCreateAndGetEvent(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	start := time.Now()
	end := start.Add(time.Hour)

	req := newCreateReq("user1", "Meeting", start, end)
	created, err := store.EventCreate(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "Meeting", created.Title)

	got, err := store.EventGet(ctx, &models.EventIDReq{ID: created.ID})
	assert.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestCreateEventConflict(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	start := time.Now()
	end := start.Add(time.Hour)

	req1 := newCreateReq("user1", "Event1", start, end)
	_, err := store.EventCreate(ctx, req1)
	assert.NoError(t, err)

	req2 := newCreateReq("user1", "Event2", start.Add(30*time.Minute), end.Add(30*time.Minute))
	_, err = store.EventCreate(ctx, req2)
	assert.ErrorIs(t, err, errors.ErrDateBusy)
}

func TestEditEvent(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	start := time.Now()
	end := start.Add(time.Hour)

	req := newCreateReq("user1", "Event1", start, end)
	event, err := store.EventCreate(ctx, req)
	assert.NoError(t, err)

	newTitle := "Updated"
	editReq := &models.EditEventReq{
		ID:    event.ID,
		Title: &newTitle,
	}

	updated, err := store.EventEdit(ctx, editReq)
	assert.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
}

func TestDeleteEvent(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	start := time.Now()
	end := start.Add(time.Hour)

	req := newCreateReq("user1", "Event1", start, end)
	event, err := store.EventCreate(ctx, req)
	assert.NoError(t, err)

	err = store.EventDelete(ctx, &models.EventIDReq{ID: event.ID})
	assert.NoError(t, err)

	_, err = store.EventGet(ctx, &models.EventIDReq{ID: event.ID})
	assert.ErrorIs(t, err, errors.ErrEventNotFound)
}

func TestGetEventList(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	start := time.Now()
	end := start.Add(time.Hour)

	req1 := newCreateReq("user1", "Event1", start, end)
	req2 := newCreateReq("user1", "Event2", start.Add(2*time.Hour), end.Add(2*time.Hour))

	_, _ = store.EventCreate(ctx, req1)
	_, _ = store.EventCreate(ctx, req2)

	list, err := store.EventGetList(ctx, req1)
	assert.NoError(t, err)
	assert.Len(t, list.Data, 2)
}

func TestGetEvent_NotFound(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	_, err := store.EventGet(ctx, &models.EventIDReq{ID: "non-existent-id"})
	assert.ErrorIs(t, err, errors.ErrEventNotFound)
}

func TestEditEvent_NotFound(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	newTitle := "NoSuch"
	req := &models.EditEventReq{
		ID:    "non-existent-id",
		Title: &newTitle,
	}

	_, err := store.EventEdit(ctx, req)
	assert.ErrorIs(t, err, errors.ErrEventNotFound)
}

func TestDeleteEvent_NotFound(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	err := store.EventDelete(ctx, &models.EventIDReq{ID: "non-existent-id"})
	assert.NoError(t, err)
}

func TestConcurrentAccess(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	start := time.Now()

	var wg sync.WaitGroup
	eventCount := 100

	for i := 0; i < eventCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			startTime := start.Add(time.Minute * time.Duration(i*2))
			endTime := startTime.Add(time.Minute)
			req := newCreateReq("user-concurrent", fmt.Sprintf("Event-%d", i), startTime, endTime)
			_, err := store.EventCreate(ctx, req)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	list, err := store.EventGetList(ctx, &models.CreateEventReq{User: "user-concurrent"})
	assert.NoError(t, err)
	assert.Len(t, list.Data, eventCount)
}

func TestConcurrentCreateAndRead(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	start := time.Now()

	var wg sync.WaitGroup
	count := 50

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			startTime := start.Add(time.Minute * time.Duration(i*2))
			endTime := startTime.Add(time.Minute)

			req := newCreateReq("user-cr-read", fmt.Sprintf("Event-%d", i), startTime, endTime)
			_, err := store.EventCreate(ctx, req)
			assert.NoError(t, err)
		}(i)
	}

	for i := 0; i < count/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.EventGetList(ctx, &models.CreateEventReq{User: "user-cr-read"})
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

func TestConcurrentEdit(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	start := time.Now()
	end := start.Add(time.Hour)

	event, err := store.EventCreate(ctx, newCreateReq("user-edit", "Initial", start, end))
	require.NoError(t, err)

	var wg sync.WaitGroup
	count := 10

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			newTitle := fmt.Sprintf("Title-%d", i)
			req := &models.EditEventReq{
				ID:    event.ID,
				Title: &newTitle,
			}
			_, err := store.EventEdit(ctx, req)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	updated, err := store.EventGet(ctx, &models.EventIDReq{ID: event.ID})
	require.NoError(t, err)
	assert.Contains(t, updated.Title, "Title-")
}

func TestConcurrentDeleteAndRead(t *testing.T) {
	store := NewLocalStorage(testLogger())
	ctx := context.Background()

	event, err := store.EventCreate(
		ctx,
		newCreateReq("user-del-read",
			"ToDelete",
			time.Now(),
			time.Now().Add(time.Minute)))
	require.NoError(t, err)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = store.EventDelete(ctx, &models.EventIDReq{ID: event.ID})
	}()

	go func() {
		defer wg.Done()
		_, _ = store.EventGet(ctx, &models.EventIDReq{ID: event.ID})
	}()

	wg.Wait()
}
