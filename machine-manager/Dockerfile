# syntax=docker/dockerfile:1

FROM golang:1.17.5 AS test

WORKDIR /code

COPY src/go.mod src/go.sum ./

RUN go mod download

COPY src ./

RUN go build -o /app

ENTRYPOINT ["go", "test"]

FROM test AS app

ENTRYPOINT /app