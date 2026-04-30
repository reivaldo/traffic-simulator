FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
ENV GOTOOLCHAIN=auto
RUN go mod tidy && go build -o admin-api .

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/admin-api /admin-api
COPY config.yaml /config.yaml
EXPOSE 8086
ENTRYPOINT ["/admin-api"]
