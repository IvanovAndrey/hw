package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSender_ForwardsNotificationToRabbit(t *testing.T) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		t.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	queue := "notifications"
	msgs, err := ch.Consume(
		queue, "", true, false, false, false, nil,
	)
	if err != nil {
		t.Fatalf("failed to consume: %v", err)
	}

	eventTime := time.Now()
	notifyBefore := "1m"

	user := uuid.NewString()
	payload := proto.CreateEventReq{
		Title:        "Notify me",
		Date:         timestamppb.New(eventTime),
		EndTime:      timestamppb.New(eventTime.Add(30 * time.Minute)),
		User:         user,
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
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected response: %d — %s", resp.StatusCode, string(b))
	}

	found := false
	for {
		select {
		case <-ctx.Done():
			t.Fatal("timeout waiting for notification from sender")
		case msg := <-msgs:
			t.Logf("msg :%v", string(msg.Body))
			var note struct {
				EventID string `json:"eventId"`
				Title   string `json:"title"`
				UserID  string `json:"userId"`
			}
			if err := json.Unmarshal(msg.Body, &note); err != nil {
				t.Logf("invalid message: %s", msg.Body)
				continue
			}
			t.Logf("got notification: %v", note)

			if note.Title == "Notify me" && note.UserID == user {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	assert.True(t, found, "notification was not received by sender in time")
}
