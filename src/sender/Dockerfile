FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
ENV GOTOOLCHAIN=auto
RUN go mod tidy && go build -o sender .

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/sender /sender
COPY config.yaml /config.yaml
EXPOSE 8084
ENTRYPOINT ["/sender"]
