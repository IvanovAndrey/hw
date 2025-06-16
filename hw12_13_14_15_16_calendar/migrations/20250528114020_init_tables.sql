-- +goose Up
-- +goose StatementBegin
create schema if not exists calendar;

create table if not exists calendar.events
(
    id            uuid primary key default gen_random_uuid(),
    title         varchar(64) not null,
    start_time    timestamp with time zone not null,
    end_time      timestamp with time zone not null,
    description   varchar(1000),
    user_id       uuid not null,
    notify_before interval
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists calendar.events;
-- +goose StatementEnd
