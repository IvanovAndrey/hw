package controllers

import (
	"context"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (c CalendarHandler) CreateEvent(ctx context.Context, req *proto.CreateEventReq) (*proto.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (c CalendarHandler) EditEvent(ctx context.Context, req *proto.EditEventReq) (*proto.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (c CalendarHandler) GetEvent(ctx context.Context, req *proto.EventByIdReq) (*proto.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (c CalendarHandler) DeleteEvent(ctx context.Context, req *proto.EventByIdReq) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (c CalendarHandler) GetEventList(ctx context.Context, req *proto.GetEventListReq) (*proto.GetEventListRes, error) {
	//TODO implement me
	panic("implement me")
}
