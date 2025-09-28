FROM golang:1.16-alpine as build
WORKDIR /app
COPY ./ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api ./src

FROM build as development
RUN go install github.com/cosmtrek/air@latest
# RUN go install github.com/air-verse/air@latest
CMD ["/go/bin/air", "-c", ".air.toml"]

FROM alpine:latest as production
COPY --from=build /app/api ./
CMD ["./api"]
