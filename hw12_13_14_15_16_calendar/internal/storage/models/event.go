package models

import "time"

type Event struct {
	ID           string
	Title        string
	Date         time.Time
	EndTime      time.Time
	Description  *string
	User         string
	NotifyBefore *string
}

type CreateEventReq struct {
	Title        string
	Date         time.Time
	EndTime      time.Time
	Description  *string
	User         string
	NotifyBefore *string
}

type EditEventReq struct {
	ID           string
	Title        *string
	Date         *time.Time
	EndTime      *time.Time
	Description  *string
	User         *string
	NotifyBefore *string
}

type EventIDReq struct {
	ID string
}

type GetEventListReq struct {
	Start *time.Time
	End   *time.Time
}

type GetEventListResp struct {
	Data []Event
}
