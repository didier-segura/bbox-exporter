# BBox Prometheus Exporter

Expose Bouygues BBox gateway statistics as Prometheus metrics and scrape them on a configurable interval.

## Configuration

The exporter expects a JSON config file (default `appsettings.json`, override with `-config <path>`):

```json
{
  "BBoxAPIURL": "https://mabbox.bytel.fr",
  "BBoxPassword": "<admin_password>",
  "BBoxAPIRefreshTime": 60,
  "MetricsServerListeningPort": 9100
}
```

- `BBoxAPIURL`: Base URL of the BBox web UI/API (HTTPS recommended).
- `BBoxPassword`: Gateway admin password used to authenticate requests.
- `BBoxAPIRefreshTime`: Polling interval in seconds.
- `MetricsServerListeningPort`: Port where `/metrics` is exposed.

An example file lives at `appsettings.example.json`. Keep real credentials out of version control by copying that file and filling in your values.

`BBoxPassword` can also be provided through the `BBOX_PASSWORD` environment variable; if set, it overrides the value in the config file. This is recommended for Docker usage so the password stays out of the config.

## Running locally

```bash
go test ./...
go run ./cmd/bb_exporter -config appsettings.json
```

Metrics will be served at `http://localhost:<MetricsServerListeningPort>/metrics`.

## Docker

Build locally:

The image ships with a default `appsettings.json` baked in. Provide the password via env; mount your own config only if you need custom settings.

```bash
docker build -t bb_exporter:local .
docker run --rm -p 9100:9100 -e BBOX_PASSWORD=your_password bb_exporter:local

# If you need a custom config, override the default:
docker run --rm -p 9100:9100 -e BBOX_PASSWORD=your_password \
  -v $(pwd)/appsettings.json:/app/appsettings.json:ro bb_exporter:local
```

Or pull from GHCR after releases (see CI/CD):

```bash
docker run --rm -p 9100:9100 -e BBOX_PASSWORD=your_password \
  ghcr.io/${GITHUB_USER_OR_ORG}/bb_exporter:<tag>

# With a custom config:
docker run --rm -p 9100:9100 -e BBOX_PASSWORD=your_password \
  -v /path/to/appsettings.json:/app/appsettings.json:ro \
  ghcr.io/${GITHUB_USER_OR_ORG}/bb_exporter:<tag>
```

## Metrics exported

- CPU: `bb_device_cpu_total`, `bb_device_cpu_user`, `bb_device_cpu_system`, `bb_device_cpu_idle`, `bb_device_cpu_temperature_main`
- Memory: `bb_device_mem_total`, `bb_device_mem_free`
- WAN: `bb_wan_ip_stats_rx_bytes`, `bb_wan_ip_stats_tx_bytes`, `bb_wan_ip_stats_rx_contractual_bandwidth`, `bb_wan_ip_stats_tx_contractual_bandwidth`, `bb_wan_ip_state_up`, `bb_wan_internet_state`, `bb_wan_interface_state`, `bb_wan_cgnat_enabled`, `bb_wan_ip_info{...}`
- LAN: `bb_lan_stats_rx_bytes`, `bb_lan_stats_tx_bytes`
- Wi‑Fi: `bb_wireless_24_stats_rx_bytes`, `bb_wireless_24_stats_tx_bytes`, `bb_wireless_5_stats_rx_bytes`, `bb_wireless_5_stats_tx_bytes`

## Grafana

Import `grafana/BBox_Exporter.json` into Grafana and point it at your Prometheus datasource.

## CI/CD

- `.github/workflows/ci.yml` runs `go test`, `go vet`, and a build on pushes and pull requests.
- `.github/workflows/release.yml` builds binaries for linux/darwin/windows on amd64 and arm64, publishes a GitHub Release for tags matching `v*`, and pushes multi-arch container images to GHCR.

Tag a release (`git tag vX.Y.Z && git push origin vX.Y.Z`) to publish artifacts.

## Development

Standard Go module workflow applies:

```bash
go fmt ./...
go test ./...
go build ./cmd/bb_exporter
```

Feel free to adjust scrape intervals and add new metrics in `internal/exporter`.
