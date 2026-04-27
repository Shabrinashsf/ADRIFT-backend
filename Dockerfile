FROM golang:1.24.4-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/utils/mailer/template ./utils/mailer/template

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8888

ENV TZ=Asia/Jakarta
ENV APP_ENV=production
ENV GIN_MODE=release

CMD ["./main"]