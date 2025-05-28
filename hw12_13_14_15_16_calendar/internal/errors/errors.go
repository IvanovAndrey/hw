package errors

import "errors"

var ErrDateBusy = errors.New("event already exists at that time for the user")
var ErrNotFound = errors.New("event not found")
