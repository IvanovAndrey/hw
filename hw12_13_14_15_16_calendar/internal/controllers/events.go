package controllers

import (
	"context"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/models"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"go.openly.dev/pointy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c CalendarHandler) CreateEvent(ctx context.Context, req *proto.CreateEventReq) (*proto.Event, error) {
	res, err := c.storage.EventCreate(ctx, &models.CreateEventReq{
		Title:        req.Title,
		Date:         req.Date.AsTime(),
		EndTime:      req.EndTime.AsTime(),
		Description:  req.Description,
		User:         req.User,
		NotifyBefore: req.NotifyBefore,
	})
	if err != nil {
		return nil, err
	}
	return &proto.Event{
		Id:           res.ID,
		User:         res.User,
		NotifyBefore: res.NotifyBefore,
		EndTime:      timestamppb.New(res.EndTime),
		Description:  res.Description,
		Date:         timestamppb.New(res.Date),
		Title:        res.Title,
	}, nil
}

func (c CalendarHandler) EditEvent(ctx context.Context, req *proto.EditEventReq) (*proto.Event, error) {
	res, err := c.storage.EventEdit(ctx, &models.EditEventReq{
		ID:           req.Id,
		Title:        req.Title,
		Date:         TimePtr(req.Date),
		EndTime:      TimePtr(req.EndTime),
		Description:  req.Description,
		User:         req.User,
		NotifyBefore: req.NotifyBefore,
	})
	if err != nil {
		return nil, err
	}
	return &proto.Event{
		Id:           res.ID,
		User:         res.User,
		NotifyBefore: res.NotifyBefore,
		EndTime:      TimestampPtr(&res.EndTime),
		Description:  res.Description,
		Date:         TimestampPtr(&res.Date),
		Title:        res.Title,
	}, nil
}

func (c CalendarHandler) GetEvent(ctx context.Context, req *proto.EventByIdReq) (*proto.Event, error) {
	res, err := c.storage.EventGet(ctx, &models.EventIDReq{
		ID: req.EventId,
	})
	if err != nil {
		return nil, err
	}
	return &proto.Event{
		Id:           res.ID,
		User:         res.User,
		NotifyBefore: res.NotifyBefore,
		EndTime:      TimestampPtr(&res.EndTime),
		Description:  res.Description,
		Date:         TimestampPtr(&res.Date),
		Title:        res.Title,
	}, nil
}

func (c CalendarHandler) DeleteEvent(ctx context.Context, req *proto.EventByIdReq) (*emptypb.Empty, error) {
	err := c.storage.EventDelete(ctx, &models.EventIDReq{
		ID: req.EventId,
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c CalendarHandler) GetEventList(ctx context.Context, req *proto.GetEventListReq) (*proto.GetEventListRes, error) {
	var start, end time.Time
	var err error

	if req.Start != nil {
		start, err = time.Parse(time.RFC3339, *req.Start)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid start time: %v", err)
		}
	}

	if req.End != nil {
		end, err = time.Parse(time.RFC3339, *req.End)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid end time: %v", err)
		}
	}

	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return nil, status.Errorf(codes.InvalidArgument, "start time must not be after end time")
	}
	res, err := c.storage.EventGetList(ctx, &models.GetEventListReq{
		Start: pointy.Pointer(start),
		End:   pointy.Pointer(end),
	})
	if err != nil {
		return nil, err
	}

	data := make([]*proto.Event, 0, len(res.Data))

	for _, e := range res.Data {
		data = append(data, &proto.Event{
			Id:           e.ID,
			User:         e.User,
			NotifyBefore: e.NotifyBefore,
			EndTime:      TimestampPtr(&e.EndTime),
			Description:  e.Description,
			Date:         TimestampPtr(&e.Date),
			Title:        e.Title,
		})
	}
	return &proto.GetEventListRes{
		Data: data,
	}, nil
}

func TimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func TimestampPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
