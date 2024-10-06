FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/server ./cmd/server

FROM scratch

COPY --from=builder /app/bin/server /server

LABEL maintainer="vadimdominik2005@gmailcom"
LABEL version="1.0.0"
LABEL description="Online Song Library API"

ENTRYPOINT [ "/server" ]
