FROM --platform=$BUILDPLATFORM node:20-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS backend-builder
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-builder /src/frontend/out ../frontend/out
RUN CGO_ENABLED=0 go build -o timeline-server .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /src/backend/timeline-server .
COPY --from=backend-builder /src/frontend/out ./frontend/out

EXPOSE 8080
CMD ["./timeline-server"]
