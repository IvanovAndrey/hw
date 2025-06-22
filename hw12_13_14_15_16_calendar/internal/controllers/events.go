package controllers

import (
	"context"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/models"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (c CalendarHandler) CreateEvent(ctx context.Context, req *proto.CreateEventReq) (*proto.Event, error) {
	res, err := c.storage.EventCreate(ctx, &models.CreateEventReq{
		Title:        req.Title,
		Date:         req.Date,
		EndTime:      req.EndTime,
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
		EndTime:      res.EndTime,
		Description:  res.Description,
		Date:         res.Date,
		Title:        res.Title,
	}, nil
}

func (c CalendarHandler) EditEvent(ctx context.Context, req *proto.EditEventReq) (*proto.Event, error) {
	res, err := c.storage.EventEdit(ctx, &models.EditEventReq{
		Title:        req.Title,
		Date:         req.Date,
		EndTime:      req.EndTime,
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
		EndTime:      res.EndTime,
		Description:  res.Description,
		Date:         res.Date,
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
		EndTime:      res.EndTime,
		Description:  res.Description,
		Date:         res.Date,
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

func (c CalendarHandler) GetEventList(ctx context.Context, _ *proto.GetEventListReq) (*proto.GetEventListRes, error) {
	res, err := c.storage.EventGetList(ctx, &models.GetEventListReq{})
	if err != nil {
		return nil, err
	}

	data := make([]*proto.Event, 0, len(res.Data))

	for _, e := range res.Data {
		data = append(data, &proto.Event{
			Id:           e.ID,
			User:         e.User,
			NotifyBefore: e.NotifyBefore,
			EndTime:      e.EndTime,
			Description:  e.Description,
			Date:         e.Date,
			Title:        e.Title,
		})
	}
	return &proto.GetEventListRes{
		Data: data,
	}, nil
}
