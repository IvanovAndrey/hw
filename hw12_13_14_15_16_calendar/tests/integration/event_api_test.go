package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/errors"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const baseURL = "http://calendar:8080"

func TestCreateEvent_Success(t *testing.T) {
	now := time.Now().Add(2 * time.Hour).UTC()
	notifyBefore := "30m"
	payload := proto.CreateEventReq{
		Title:        "test event",
		Date:         now.Format(time.RFC3339),
		EndTime:      now.Add(1 * time.Hour).Format(time.RFC3339),
		User:         uuid.NewString(),
		NotifyBefore: &notifyBefore,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/api/v1/event", "application/json", bytes.NewReader(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to create event: %v, code: %d", err, resp.StatusCode)
	}
}

func TestCreateEvent_Conflict(t *testing.T) {
	now := time.Now().Add(3 * time.Hour).UTC()
	notifyBefore := "30m"
	payload := proto.CreateEventReq{
		Title:        "test event",
		Date:         now.Format(time.RFC3339),
		EndTime:      now.Add(1 * time.Hour).Format(time.RFC3339),
		User:         uuid.NewString(),
		NotifyBefore: &notifyBefore,
	}
	body, _ := json.Marshal(payload)

	// First creation
	_, _ = http.Post(baseURL+"/api/v1/event", "application/json", bytes.NewReader(body))
	// Second creation (should fail)
	resp, _ := http.Post(baseURL+"/api/v1/event", "application/json", bytes.NewReader(body))

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected conflict error, got 200")
	}
	if resp.Body == nil {
		t.Fatalf("body is nil")
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	assert.True(t, strings.Contains(string(respBody), errors.ErrDateBusy.Error()))
}

func TestCreateEvent_ValidationError(t *testing.T) {
	now := time.Now().Add(3 * time.Hour).UTC()
	notifyBefore := "30m"
	payload := proto.CreateEventReq{
		Title:        "test event",
		Date:         now.Format(time.RFC3339),
		EndTime:      now.Add(1 * time.Hour).Format(time.RFC3339),
		NotifyBefore: &notifyBefore,
	}
	body, _ := json.Marshal(payload)

	resp, _ := http.Post(baseURL+"/api/v1/event", "application/json", bytes.NewReader(body))

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected conflict error, got 200")
	}
	if resp.Body == nil {
		t.Fatalf("body is nil")
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	assert.True(t, strings.Contains(string(respBody), "validation error"))
}

func TestCreateEvent_EventNotFound(t *testing.T) {
	title := "new title"
	payload := proto.EditEventReq{
		Id:    uuid.NewString(),
		Title: &title,
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPatch, baseURL+"/api/v1/event", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send patch request: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected conflict error, got 200")
	}
}
