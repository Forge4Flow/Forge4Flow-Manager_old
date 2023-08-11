# Stage 1: Build the backend
FROM golang:1.20 AS backend-builder

WORKDIR /forge4flow
COPY . .
RUN go mod download

RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -v -o forge4flow-manager main.go

# Stage 2: Create the final image with both backend and frontend
FROM alpine:3.18.3

RUN addgroup -S forge4flow-manager && adduser -S forge4flow-manager -G forge4flow-manager
USER forge4flow-manager

WORKDIR /forge4flow
COPY --from=backend-builder /forge4flow/forge4flow-manager .

ENTRYPOINT ["./forge4flow-manager"]

EXPOSE 8000