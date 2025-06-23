package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/errors"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const baseURL = "http://calendar:8080"

var notifyBefore = "30m"

type Event struct {
	Title        string `json:"title"`
	Date         string `json:"date"`
	EndTime      string `json:"end_time"` //nolint
	Description  string `json:"description,omitempty"`
	User         string `json:"user"`
	NotifyBefore string `json:"notify_before,omitempty"` //nolint
}

func TestCreateEvent_Success(t *testing.T) {
	now := time.Now().Add(2 * time.Hour).UTC()
	payload := proto.CreateEventReq{
		Title:        "test event",
		Date:         timestamppb.New(now),
		EndTime:      timestamppb.New(now.Add(1 * time.Hour)),
		User:         uuid.NewString(),
		NotifyBefore: &notifyBefore,
	}
	body, _ := protojson.Marshal(&payload)
	t.Logf("body %v", string(body))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/event", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Logf("warning: failed to close response body: %v", cerr)
		}
	}()
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
}

func TestCreateEvent_Conflict(t *testing.T) {
	now := time.Now().Add(3 * time.Hour).UTC()
	payload := &proto.CreateEventReq{
		Title:        "test event",
		Date:         timestamppb.New(now),
		EndTime:      timestamppb.New(now.Add(1 * time.Hour)),
		User:         uuid.NewString(),
		NotifyBefore: &notifyBefore,
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()

	resp1, body1 := doPostJSON(t, ctx1, baseURL+"/api/v1/event", payload)
	defer func() {
		if cerr := resp1.Body.Close(); cerr != nil {
			t.Logf("warning: failed to close response body: %v", cerr)
		}
	}()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("unexpected first response: %d, body: %s", resp1.StatusCode, string(body1))
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	resp2, body2 := doPostJSON(t, ctx2, baseURL+"/api/v1/event", payload)
	defer func() {
		if cerr := resp2.Body.Close(); cerr != nil {
			t.Logf("warning: failed to close response body: %v", cerr)
		}
	}()
	if resp2.StatusCode == http.StatusOK {
		t.Fatalf("expected conflict error, got 200")
	}

	assert.Contains(t, string(body2), errors.ErrDateBusy.Error())
}

func TestCreateEvent_ValidationError(t *testing.T) {
	now := time.Now().Add(3 * time.Hour).UTC()
	payload := proto.CreateEventReq{
		Title:        "test event",
		Date:         timestamppb.New(now),
		EndTime:      timestamppb.New(now.Add(1 * time.Hour)),
		NotifyBefore: &notifyBefore,
	}
	body, _ := protojson.Marshal(&payload)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/event", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Logf("warning: failed to close response body: %v", cerr)
		}
	}()

	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected conflict error, got 200")
	}
	if resp.Body == nil {
		t.Fatalf("body is nil")
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	t.Logf("response body: %v", string(respBody))
	assert.True(t, strings.Contains(string(respBody), "validation error"))
}

func TestCreateEvent_EventNotFound(t *testing.T) {
	title := "new title"
	payload := proto.EditEventReq{
		Id:    uuid.NewString(),
		Title: &title,
	}
	body, _ := protojson.Marshal(&payload)
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, baseURL+"/api/v1/event", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Logf("warning: failed to close response body: %v", cerr)
		}
	}()
	if err != nil {
		t.Fatalf("failed to send patch request: %v", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected conflict error, got 200")
	}
	t.Logf("response body: %v", string(respBody))
	assert.True(t, strings.Contains(string(respBody), errors.ErrEventNotFound.Error()))
}

func createEvent(t *testing.T, e Event) {
	t.Helper()
	body, _ := json.Marshal(e)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/event", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Logf("warning: failed to close response body: %v", cerr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status: %d, body: %s", resp.StatusCode, string(b))
	}
}

func getEvents(t *testing.T, start, end time.Time) []Event {
	t.Helper()
	u, _ := url.Parse(baseURL + "/api/v1/events")
	q := u.Query()
	q.Set("start", start.Format(time.RFC3339))
	q.Set("end", end.Format(time.RFC3339))
	u.RawQuery = q.Encode()

	t.Logf("req: %v", u.String())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		t.Fatalf("failed to create GET request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send GET request: %v", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Logf("warning: failed to close response body: %v", cerr)
		}
	}()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	t.Logf("response body: %v", string(respBody))
	var res struct {
		Data []Event `json:"data"`
	}
	if err := json.Unmarshal(respBody, &res); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	return res.Data
}

func TestListEvents_Day(t *testing.T) {
	now := time.Now().UTC()
	createEvent(t, Event{
		Title:   "Today Event",
		Date:    now.Format(time.RFC3339),
		EndTime: now.Add(1 * time.Hour).Format(time.RFC3339),
		User:    uuid.NewString(),
	})

	events := getEvents(t, now.Truncate(24*time.Hour), now.Add(24*time.Hour))
	if len(events) == 0 {
		t.Errorf("expected at least 1 event for today")
	}
}

func TestListEvents_Week(t *testing.T) {
	start := time.Now().AddDate(0, 0, 2).UTC()
	createEvent(t, Event{
		Title:   "This Week Event",
		Date:    start.Format(time.RFC3339),
		EndTime: start.Add(1 * time.Hour).Format(time.RFC3339),
		User:    uuid.NewString(),
	})

	from := time.Now().Truncate(24 * time.Hour)
	to := from.AddDate(0, 0, 7)

	events := getEvents(t, from, to)
	if len(events) == 0 {
		t.Errorf("expected at least 1 event for this week")
	}
}

func TestListEvents_Month(t *testing.T) {
	start := time.Now().AddDate(0, 0, 20).UTC()
	createEvent(t, Event{
		Title:   "Month Event",
		Date:    start.Format(time.RFC3339),
		EndTime: start.Add(2 * time.Hour).Format(time.RFC3339),
		User:    uuid.NewString(),
	})

	from := time.Now().Truncate(24 * time.Hour)
	to := from.AddDate(0, 1, 0)

	events := getEvents(t, from, to)
	if len(events) == 0 {
		t.Errorf("expected at least 1 event for this month")
	}
}

func doPostJSON(t *testing.T, ctx context.Context, url string, body *proto.CreateEventReq) (*http.Response, []byte) {
	t.Helper()

	jsonBody, err := protojson.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Fatalf("can't close body")
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	return resp, respBody
}
