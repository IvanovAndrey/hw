package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSender_ForwardsNotificationToRabbit(t *testing.T) {
	ch, msgs := mustConsumeFromQueue(t, "notifications")
	defer closeChannel(t, ch)

	user := uuid.NewString()
	title := "Notify me"
	notifyBefore := "1m"
	eventTime := time.Now().UTC()

	createEventReq := &proto.CreateEventReq{
		Title:        title,
		Date:         timestamppb.New(eventTime),
		EndTime:      timestamppb.New(eventTime.Add(30 * time.Minute)),
		User:         user,
		NotifyBefore: &notifyBefore,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, respBody := doPostJSON(t, ctx, baseURL+"/api/v1/event", createEventReq)
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			t.Logf("warning: failed to close response body: %v", cerr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %d — %s", resp.StatusCode, string(respBody))
	}

	found := false
	for {
		select {
		case <-ctx.Done():
			t.Fatal("timeout waiting for notification from sender")
		case msg := <-msgs:
			var note struct {
				EventID string `json:"eventId"`
				Title   string `json:"title"`
				UserID  string `json:"userId"`
			}
			if err := json.Unmarshal(msg.Body, &note); err != nil {
				t.Logf("invalid message: %s", msg.Body)
				continue
			}
			if note.Title == title && note.UserID == user {
				found = true
			}
		}
		if found {
			break
		}
	}

	assert.True(t, found, "notification was not received by sender in time")
}

func mustConsumeFromQueue(t *testing.T, queue string) (*amqp.Channel, <-chan amqp.Delivery) {
	t.Helper()

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		t.Fatalf("rabbitmq dial: %v", err)
	}

	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("close conn: %v", err)
		}
	})

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("open channel: %v", err)
	}

	msgs, err := ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}

	return ch, msgs
}

func closeChannel(t *testing.T, ch *amqp.Channel) {
	t.Helper()
	if err := ch.Close(); err != nil {
		t.Errorf("close channel: %v", err)
	}
}
