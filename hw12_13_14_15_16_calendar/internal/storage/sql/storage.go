package sqlstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/configuration"
	calendarErrors "github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/errors"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/IvanovAndrey/hw/hw12_13_14_15_calendar/internal/storage/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type DBStorage struct {
	DB     *pgxpool.Pool
	logger logger.Logger
}

func NewStorage(ctx context.Context, cfg *configuration.DatabaseConf, logger logger.Logger) (*DBStorage, error) {
	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	dbPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		logger.Error("failed to create db pool: " + err.Error())
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		logger.Error("failed to ping db: " + err.Error())
		return nil, fmt.Errorf("ping db: %w", err)
	}

	logger.Debug("connected to database")
	return &DBStorage{DB: dbPool, logger: logger}, nil
}

func (s *DBStorage) EventCreate(ctx context.Context, req *models.CreateEventReq) (*models.Event, error) {
	checkSQL := `
		SELECT EXISTS (
			SELECT 1 FROM calendar.events
			WHERE user_id = $1
			  AND tstzrange(start_time, end_time) && tstzrange($2::timestamptz, $3::timestamptz)
		)`
	s.logger.Debug("SQL: " + checkSQL)

	var exists bool
	if err := s.DB.QueryRow(ctx, checkSQL, req.User, req.Date, req.EndTime).Scan(&exists); err != nil {
		s.logger.Error("check overlap failed: " + err.Error())
		return nil, fmt.Errorf("check overlap: %w", err)
	}
	if exists {
		s.logger.Error("conflict: user=" + req.User + " date=" +
			req.Date.Format(time.RFC822Z) + " end=" + req.EndTime.Format(time.RFC822Z))
		return nil, fmt.Errorf("event conflict: %w", calendarErrors.ErrDateBusy)
	}

	insertSQL := `
		INSERT INTO calendar.events (title, start_time, end_time, description, user_id, notify_before)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	s.logger.Debug("SQL: " + insertSQL)

	var id string
	if err := s.DB.QueryRow(
		ctx,
		insertSQL,
		req.Title,
		req.Date,
		req.EndTime,
		req.Description,
		req.User,
		req.NotifyBefore,
	).Scan(&id); err != nil {
		s.logger.Error("insert failed: " + err.Error())
		return nil, fmt.Errorf("insert event: %w", err)
	}

	s.logger.Debug("event created id=" + id)
	return &models.Event{
		ID:           id,
		Title:        req.Title,
		Date:         req.Date,
		EndTime:      req.EndTime,
		Description:  req.Description,
		User:         req.User,
		NotifyBefore: req.NotifyBefore,
	}, nil
}

func (s *DBStorage) EventEdit(ctx context.Context, req *models.EditEventReq) (*models.Event, error) {
	event, err := s.EventGet(ctx, &models.EventIDReq{ID: req.ID})
	if err != nil {
		s.logger.Error("edit get failed: " + err.Error())
		return nil, fmt.Errorf("get event for edit: %w", err)
	}

	event = checkRequest(req, event)

	checkSQL := `
		SELECT EXISTS (
			SELECT 1 FROM calendar.events
			WHERE user_id = $1
			  AND tstzrange(start_time, end_time) && tstzrange($2::timestamptz, $3::timestamptz)
			  AND id <> $4
		)`
	s.logger.Debug("SQL: " + checkSQL)

	var exists bool
	if err := s.DB.QueryRow(ctx, checkSQL, event.User, event.Date, event.EndTime, event.ID).Scan(&exists); err != nil {
		s.logger.Error("edit overlap check failed: " + err.Error())
		return nil, fmt.Errorf("check overlap: %w", err)
	}
	if exists {
		s.logger.Error("conflict on edit: id=" + event.ID)
		return nil, fmt.Errorf("event conflict: %w", calendarErrors.ErrDateBusy)
	}

	updateSQL := `
		UPDATE calendar.events
		SET title = $1, start_time = $2, end_time = $3, description = $4, user_id = $5, notify_before = $6
		WHERE id = $7`
	s.logger.Debug("SQL: " + updateSQL)

	if _, err := s.DB.Exec(
		ctx,
		updateSQL,
		event.Title,
		event.Date,
		event.EndTime,
		event.Description,
		event.User,
		event.NotifyBefore,
		event.ID,
	); err != nil {
		s.logger.Error("edit update failed: " + err.Error())
		return nil, fmt.Errorf("update event: %w", err)
	}

	s.logger.Debug("event edited id=" + event.ID)
	return event, nil
}

func checkRequest(req *models.EditEventReq, event *models.Event) *models.Event {
	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Date != nil {
		event.Date = *req.Date
	}
	if req.EndTime != nil {
		event.EndTime = *req.EndTime
	}
	if req.Description != nil {
		event.Description = req.Description
	}
	if req.User != nil {
		event.User = *req.User
	}
	if req.NotifyBefore != nil {
		event.NotifyBefore = req.NotifyBefore
	}
	return event
}

func (s *DBStorage) EventDelete(ctx context.Context, req *models.EventIDReq) error {
	sql := `DELETE FROM calendar.events WHERE id = $1`
	s.logger.Debug("SQL: " + sql)

	if _, err := s.DB.Exec(ctx, sql, req.ID); err != nil {
		s.logger.Error("delete failed: " + err.Error())
		return fmt.Errorf("delete event: %w", err)
	}

	s.logger.Debug("event deleted id=" + req.ID)
	return nil
}

func (s *DBStorage) EventGet(ctx context.Context, req *models.EventIDReq) (*models.Event, error) {
	sql := `
		SELECT id, title, start_time, end_time, description, user_id, notify_before
		FROM calendar.events
		WHERE id = $1`
	s.logger.Debug("SQL: " + sql)

	var e models.Event
	var notify pgtype.Interval
	if err := s.DB.QueryRow(ctx, sql, req.ID).Scan(
		&e.ID, &e.Title, &e.Date, &e.EndTime, &e.Description, &e.User, &notify,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn("event not found id=" + req.ID)
			return nil, calendarErrors.ErrEventNotFound
		}
		s.logger.Error("get failed: " + err.Error())
		return nil, fmt.Errorf("get event: %w", err)
	}
	e.NotifyBefore = IntervalToDurationString(notify)
	s.logger.Debug("event fetched id=" + req.ID)
	return &e, nil
}

func (s *DBStorage) EventGetList(ctx context.Context, req *models.GetEventListReq) (*models.GetEventListResp, error) {
	query := `SELECT id, title, start_time, end_time, description, user_id, notify_before
              FROM calendar.events WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if req.Start != nil {
		query += fmt.Sprintf(" AND end_time >= $%d", argID)
		args = append(args, *req.Start)
		argID++
	}

	if req.End != nil {
		query += fmt.Sprintf(" AND start_time <= $%d", argID)
		args = append(args, *req.End)
	}

	query += " ORDER BY start_time"
	s.logger.Debug("SQL: " + query)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		s.logger.Error("list query failed: " + err.Error())
		return nil, fmt.Errorf("get event list: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		var notify pgtype.Interval
		err := rows.Scan(&e.ID, &e.Title, &e.Date, &e.EndTime, &e.Description, &e.User, &notify)
		if err != nil {
			s.logger.Error("scan failed: " + err.Error())
			return nil, fmt.Errorf("scan event: %w", err)
		}
		e.NotifyBefore = IntervalToDurationString(notify)
		events = append(events, e)
	}

	return &models.GetEventListResp{Data: events}, nil
}

func (s *DBStorage) EventsToNotify(ctx context.Context) ([]models.Event, error) {
	query := `
		SELECT id, title, start_time, end_time, description, user_id, notify_before
		FROM calendar.events
		WHERE notify_before IS NOT NULL
		AND start_time - notify_before <= $1
	`

	now := time.Now()
	s.logger.Debug(fmt.Sprintf("SQL: %v%v", query, now))
	rows, err := s.DB.Query(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var ev models.Event
		var notify pgtype.Interval
		err := rows.Scan(&ev.ID, &ev.Title, &ev.Date, &ev.EndTime, &ev.Description, &ev.User, &notify)
		if err != nil {
			return nil, err
		}
		ev.NotifyBefore = IntervalToDurationString(notify)
		events = append(events, ev)
	}
	s.logger.Debug(fmt.Sprintf("events: %v", events))
	return events, rows.Err()
}

func (s *DBStorage) DeleteOldEvents(ctx context.Context, cutoff time.Time) error {
	_, err := s.DB.Exec(ctx, `
		DELETE FROM calendar.events
		WHERE end_time < $1
	`, cutoff)
	return err
}

func IntervalToDurationString(iv pgtype.Interval) *string {
	d := time.Duration(iv.Microseconds) * time.Microsecond
	if iv.Months != 0 || iv.Days != 0 {
		res := fmt.Sprintf("%dmo %dd %s", iv.Months, iv.Days, d)
		return &res
	}
	x := d.String()
	return &x
}
