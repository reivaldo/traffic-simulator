FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
ENV GOTOOLCHAIN=auto
RUN go mod tidy && go build -o ingestor .

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/ingestor /ingestor
COPY config.yaml /config.yaml
EXPOSE 8082
ENTRYPOINT ["/ingestor"]
