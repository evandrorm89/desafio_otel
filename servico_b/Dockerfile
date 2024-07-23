FROM golang:latest as build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags="-w -s" -o bin/main ./cmd/microservice

FROM alpine:latest

COPY --from=build /app/bin/main /app/bin/main

EXPOSE 8181

ENTRYPOINT ["/app/bin/main"]
