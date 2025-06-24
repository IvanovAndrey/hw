# Собираем в гошке
FROM golang:1.23 as build

ENV CODE_DIR /go/src

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
RUN make build

# На выходе тонкий образ
FROM alpine:3.9

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="calendar"
LABEL MAINTAINERS="student@otus.ru"

WORKDIR /opt/calendar

COPY --from=build /go/src/bin ./bin
COPY ./configs ./configs

CMD ["./bin/calendar", "--config=/configs/calendar_config.yaml"]
