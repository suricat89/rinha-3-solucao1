FROM golang:1.24.1-alpine3.21 AS build
WORKDIR /app
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/main ./cmd/main.go

FROM alpine:3.21.0 AS final
WORKDIR /app
COPY --from=build /app/bin/main /app
ENV GOGC=1000
CMD [ "/app/main" ]
