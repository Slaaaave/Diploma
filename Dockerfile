FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o scheduler .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/scheduler .
COPY --from=builder /app/web ./web

ENV TODO_PORT=7540
ENV TODO_DBFILE=scheduler.db

EXPOSE 7540

CMD ["./scheduler"]
