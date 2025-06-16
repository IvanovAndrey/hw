package app

import (
	"context"
	"errors"
	"testing"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) CreateEvent(ctx context.Context, req *proto.CreateEventReq) (*proto.Event, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*proto.Event), args.Error(1)
}

func (m *mockStorage) EditEvent(ctx context.Context, req *proto.EditEventReq) (*proto.Event, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*proto.Event), args.Error(1)
}

func (m *mockStorage) GetEvent(ctx context.Context, req *proto.EventByIdReq) (*proto.Event, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*proto.Event), args.Error(1)
}

func (m *mockStorage) DeleteEvent(ctx context.Context, req *proto.EventByIdReq) (*emptypb.Empty, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *mockStorage) GetEventList(ctx context.Context, req *proto.GetEventListReq) (*proto.GetEventListRes, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*proto.GetEventListRes), args.Error(1)
}

type noopLogger struct{}

func (noopLogger) Error(_ string) {}

func newTestApp() (*App, *mockStorage) {
	st := new(mockStorage)
	return New(noopLogger{}, st), st
}

func TestGetLiveZ(t *testing.T) {
	a, _ := newTestApp()
	res, err := a.GetLiveZ(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestCreateEvent(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		a, _ := newTestApp()
		_, err := a.CreateEvent(context.Background(), &proto.CreateEventReq{})
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("success", func(t *testing.T) {
		a, st := newTestApp()
		req := &proto.CreateEventReq{
			Date:    "1",
			User:    "1",
			EndTime: "1",
			Title:   "test",
		}
		expected := &proto.Event{Id: "1"}
		st.On("CreateEvent", mock.Anything, req).Return(expected, nil)

		resp, err := a.CreateEvent(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, expected, resp)
	})

	t.Run("storage error", func(t *testing.T) {
		a, st := newTestApp()
		req := &proto.CreateEventReq{
			Date:    "1",
			User:    "1",
			EndTime: "1",
			Title:   "test",
		}
		st.On("CreateEvent", mock.Anything, req).Return(&proto.Event{}, errors.New("db error"))

		_, err := a.CreateEvent(context.Background(), req)
		stErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, stErr.Code())
	})
}

func TestEditEvent(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		a, _ := newTestApp()
		_, err := a.EditEvent(context.Background(), &proto.EditEventReq{})
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("success", func(t *testing.T) {
		a, st := newTestApp()
		req := &proto.EditEventReq{Id: "123"}
		expected := &proto.Event{Id: "123"}
		st.On("EditEvent", mock.Anything, req).Return(expected, nil)

		resp, err := a.EditEvent(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, expected, resp)
	})
}

func TestGetEvent(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		a, _ := newTestApp()
		_, err := a.GetEvent(context.Background(), &proto.EventByIdReq{})
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("success", func(t *testing.T) {
		a, st := newTestApp()
		req := &proto.EventByIdReq{EventId: "event-id"}
		expected := &proto.Event{Id: "event-id"}
		st.On("GetEvent", mock.Anything, req).Return(expected, nil)

		resp, err := a.GetEvent(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, expected, resp)
	})
}

func TestDeleteEvent(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		a, _ := newTestApp()
		_, err := a.DeleteEvent(context.Background(), &proto.EventByIdReq{})
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("success", func(t *testing.T) {
		a, st := newTestApp()
		req := &proto.EventByIdReq{EventId: "del-id"}
		expected := &emptypb.Empty{}
		st.On("DeleteEvent", mock.Anything, req).Return(expected, nil)

		resp, err := a.DeleteEvent(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, expected, resp)
	})
}

func TestGetEventList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		a, st := newTestApp()
		req := &proto.GetEventListReq{}
		expected := &proto.GetEventListRes{}
		st.On("GetEventList", mock.Anything, req).Return(expected, nil)

		resp, err := a.GetEventList(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, expected, resp)
	})
}
