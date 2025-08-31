# builder
FROM golang:1.20-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/mock-service


# final
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/mock-service /bin/mock-service
EXPOSE 8080
ENTRYPOINT ["/bin/mock-service"]
