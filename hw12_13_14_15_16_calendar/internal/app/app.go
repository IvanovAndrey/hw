package app

import (
	"context"

	calendarErrors "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/errors"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type App struct {
	logger       Logger
	eventHandler Storage
	proto.UnimplementedCalendarServer
}

type Logger interface {
	Error(msg string)
}

type Storage interface {
	CreateEvent(ctx context.Context, req *proto.CreateEventReq) (*proto.Event, error)
	EditEvent(ctx context.Context, req *proto.EditEventReq) (*proto.Event, error)
	GetEvent(ctx context.Context, req *proto.EventByIdReq) (*proto.Event, error)
	DeleteEvent(ctx context.Context, req *proto.EventByIdReq) (*emptypb.Empty, error)
	GetEventList(ctx context.Context, req *proto.GetEventListReq) (*proto.GetEventListRes, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger:       logger,
		eventHandler: storage,
	}
}

func (a *App) GetLiveZ(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return new(emptypb.Empty), nil
}

func (a *App) CreateEvent(ctx context.Context, req *proto.CreateEventReq) (*proto.Event, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	res, err := a.eventHandler.CreateEvent(ctx, req)
	if err != nil {
		return nil, calendarErrors.MakeGrpcError(err)
	}
	return res, nil
}

func (a *App) EditEvent(ctx context.Context, req *proto.EditEventReq) (*proto.Event, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	res, err := a.eventHandler.EditEvent(ctx, req)
	if err != nil {
		return nil, calendarErrors.MakeGrpcError(err)
	}
	return res, nil
}

func (a *App) GetEvent(ctx context.Context, req *proto.EventByIdReq) (*proto.Event, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	res, err := a.eventHandler.GetEvent(ctx, req)
	if err != nil {
		return nil, calendarErrors.MakeGrpcError(err)
	}
	return res, nil
}

func (a *App) DeleteEvent(ctx context.Context, req *proto.EventByIdReq) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	res, err := a.eventHandler.DeleteEvent(ctx, req)
	if err != nil {
		return nil, calendarErrors.MakeGrpcError(err)
	}
	return res, nil
}

func (a *App) GetEventList(ctx context.Context, req *proto.GetEventListReq) (*proto.GetEventListRes, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	res, err := a.eventHandler.GetEventList(ctx, req)
	if err != nil {
		return nil, calendarErrors.MakeGrpcError(err)
	}
	return res, nil
}
