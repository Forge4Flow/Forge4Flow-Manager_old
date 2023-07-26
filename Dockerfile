# Stage 1: Build the backend
FROM golang:1.20 AS backend-builder

WORKDIR /forge4flow
COPY . .
RUN go mod download

RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -v -o forge4flow-manager main.go

# Stage 2: Build the frontend
FROM node:18 AS frontend-builder

WORKDIR /forge4flow/web
COPY web .
RUN npm run build

# Stage 3: Create the final image with both backend and frontend
FROM alpine:latest

RUN apk add --update nodejs npm
RUN npm install -g concurrently

RUN addgroup -S forge4flow-manager && adduser -S forge4flow-manager -G forge4flow-manager
USER forge4flow-manager

WORKDIR /forge4flow
COPY --from=backend-builder /forge4flow/forge4flow-manager .
COPY --from=frontend-builder /forge4flow/web/.next ./web/.next

EXPOSE 8000
EXPOSE 3000

CMD concurrently "./forge4flow-core" "npm start --prefix ./web"