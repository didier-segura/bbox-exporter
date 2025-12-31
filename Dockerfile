FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -o /out/bb_exporter ./cmd/bb_exporter

FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=build /out/bb_exporter /app/bb_exporter
COPY appsettings.example.json /app/appsettings.json

EXPOSE 9100
ENTRYPOINT ["/app/bb_exporter"]
