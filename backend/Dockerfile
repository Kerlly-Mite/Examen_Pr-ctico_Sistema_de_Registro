# ---- Etapa 1: compilar el binario de Go ----
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api .

# ---- Etapa 2: imagen final pequeña ----
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
COPY --from=build /out/api /app/api
USER app
EXPOSE 3000
CMD ["/app/api"]
