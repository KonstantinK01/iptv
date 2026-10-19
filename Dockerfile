FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o /server ./cmd/app

FROM alpine:3.20
COPY --from=build /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
