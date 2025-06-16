package models

type Event struct {
	ID           string
	Title        string
	Date         string
	EndTime      string
	Description  *string
	User         string
	NotifyBefore *string
}

type CreateEventReq struct {
	Title        string
	Date         string
	EndTime      string
	Description  *string
	User         string
	NotifyBefore *string
}

type EditEventReq struct {
	ID           string
	Title        *string
	Date         *string
	EndTime      *string
	Description  *string
	User         *string
	NotifyBefore *string
}

type EventIDReq struct {
	ID string
}

type GetEventListReq struct {
	Filter string // TODO refactor
}

type GetEventListResp struct {
	Data []Event
}
