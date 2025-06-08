package errors

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrDateBusy      = errors.New("event already exists at that time for the user")
	ErrEventNotFound = errors.New("event not found")
)

func MakeGrpcError(err error) error {
	if errors.Is(err, ErrEventNotFound) {
		return status.Errorf(codes.NotFound, "event not found: %v", err)
	}
	if errors.Is(err, ErrDateBusy) {
		return status.Errorf(codes.AlreadyExists, "date busy: %v", err)
	}
	return status.Errorf(codes.Internal, "failed to create event: %v", err)
}
