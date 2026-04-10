FROM node:20-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.23-alpine AS backend-build
RUN apk add --no-cache gcc musl-dev
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-build /app/backend/static ./static
RUN CGO_ENABLED=1 go build -o /server ./cmd/app/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-build /server ./server
COPY --from=backend-build /app/backend/static ./static
ENV STATIC_DIR=static
ENV SQLITE_PATH=data/job_fit.db
EXPOSE 8080
CMD ["./server"]
