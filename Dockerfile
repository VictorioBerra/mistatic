# Stage 1: Build frontend
FROM node:24-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./

ARG APP_DOMAIN=app.mistatic.local
ENV PUBLIC_APP_DOMAIN=${APP_DOMAIN}

ARG ROOT_DOMAIN=mistatic.local
ENV PUBLIC_ROOT_DOMAIN=${ROOT_DOMAIN}

ENV NODE_ENV=production

RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.27-alpine AS go-builder
RUN apk add --no-cache tzdata
WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
# Copy the built frontend into the go-builder where it will be embedded
COPY --from=frontend-builder /app/frontend/build ./ui/dist/
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/mistatic -ldflags="-s -w" .

# Stage 3: Runtime
FROM alpine:3.22
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=go-builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go-builder /app/mistatic /app/mistatic
COPY --from=go-builder /app/pb_migrations /app/pb_migrations

EXPOSE 8090

ENTRYPOINT ["/app/mistatic"]
CMD ["serve", "--http=0.0.0.0:8090", "--dir=/pb/pb_data"]
