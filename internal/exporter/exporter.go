package exporter

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/dsegura/bbox-exporter/internal/bbox"
)

// Exporter periodically pulls metrics from the BBox API and exposes them as Prometheus gauges.
type Exporter struct {
	client *bbox.Client
	g      gauges
}

type gauges struct {
	cpuTotal          prometheus.Gauge
	cpuUser           prometheus.Gauge
	cpuSystem         prometheus.Gauge
	cpuIdle           prometheus.Gauge
	cpuTemperature    prometheus.Gauge
	memFree           prometheus.Gauge
	memTotal          prometheus.Gauge
	wanRxBytes        prometheus.Gauge
	wanTxBytes        prometheus.Gauge
	wanRxContractual  prometheus.Gauge
	wanTxContractual  prometheus.Gauge
	wanIPState        prometheus.Gauge
	wanInternetState  prometheus.Gauge
	wanInterfaceState prometheus.Gauge
	wanCgnatEnabled   prometheus.Gauge
	wanInfo           *prometheus.GaugeVec
	lanRxBytes        prometheus.Gauge
	lanTxBytes        prometheus.Gauge
	wireless24RxBytes prometheus.Gauge
	wireless24TxBytes prometheus.Gauge
	wireless5RxBytes  prometheus.Gauge
	wireless5TxBytes  prometheus.Gauge
}

func New(client *bbox.Client) *Exporter {
	return &Exporter{
		client: client,
		g: gauges{
			cpuTotal:          promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_device_cpu_total", Help: "Total CPU time"}),
			cpuUser:           promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_device_cpu_user", Help: "User CPU time"}),
			cpuSystem:         promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_device_cpu_system", Help: "System CPU time"}),
			cpuIdle:           promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_device_cpu_idle", Help: "Idle CPU time"}),
			cpuTemperature:    promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_device_cpu_temperature_main", Help: "CPU temperature main sensor"}),
			memFree:           promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_device_mem_free", Help: "Free memory"}),
			memTotal:          promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_device_mem_total", Help: "Total memory"}),
			wanRxBytes:        promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wan_ip_stats_rx_bytes", Help: "WAN RX bytes"}),
			wanTxBytes:        promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wan_ip_stats_tx_bytes", Help: "WAN TX bytes"}),
			wanRxContractual:  promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wan_ip_stats_rx_contractual_bandwidth", Help: "WAN RX contractual bandwidth"}),
			wanTxContractual:  promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wan_ip_stats_tx_contractual_bandwidth", Help: "WAN TX contractual bandwidth"}),
			wanIPState:        promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wan_ip_state_up", Help: "WAN IP state (1=Up,0=Down)"}),
			wanInternetState:  promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wan_internet_state", Help: "WAN internet state code"}),
			wanInterfaceState: promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wan_interface_state", Help: "WAN interface state code"}),
			wanCgnatEnabled:   promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wan_cgnat_enabled", Help: "WAN CGNAT enabled flag"}),
			wanInfo: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "bb_wan_ip_info",
					Help: "WAN IP metadata (labels hold values, gauge is always 1)",
				},
				[]string{
					"address",
					"gateway",
					"dnsservers",
					"dnsserversv6",
					"subnet",
					"mac",
					"ip_state",
					"ip6_state",
					"ip6_addresses",
					"ip6_prefixes",
					"link_state",
					"link_type",
					"mapt_enable",
					"mtu",
				},
			),
			lanRxBytes:        promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_lan_stats_rx_bytes", Help: "LAN RX bytes"}),
			lanTxBytes:        promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_lan_stats_tx_bytes", Help: "LAN TX bytes"}),
			wireless24RxBytes: promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wireless_24_stats_rx_bytes", Help: "2.4GHz Wi-Fi RX bytes"}),
			wireless24TxBytes: promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wireless_24_stats_tx_bytes", Help: "2.4GHz Wi-Fi TX bytes"}),
			wireless5RxBytes:  promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wireless_5_stats_rx_bytes", Help: "5GHz Wi-Fi RX bytes"}),
			wireless5TxBytes:  promauto.NewGauge(prometheus.GaugeOpts{Name: "bb_wireless_5_stats_tx_bytes", Help: "5GHz Wi-Fi TX bytes"}),
		},
	}
}

// Refresh performs a full login -> scrape -> logout cycle and updates gauges.
func (e *Exporter) Refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := e.client.Login(ctx); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	defer func() {
		if err := e.client.Logout(ctx); err != nil {
			log.Printf("logout failed: %v", err)
		}
	}()

	cpu, err := e.client.FetchCPU(ctx)
	if err != nil {
		return fmt.Errorf("fetch cpu: %w", err)
	}
	mem, err := e.client.FetchMem(ctx)
	if err != nil {
		return fmt.Errorf("fetch mem: %w", err)
	}
	wanInfo, err := e.client.FetchWanIPInfo(ctx)
	if err != nil {
		return fmt.Errorf("fetch wan info: %w", err)
	}
	wanStats, err := e.client.FetchWanIPStats(ctx)
	if err != nil {
		return fmt.Errorf("fetch wan stats: %w", err)
	}
	lanStats, err := e.client.FetchLanStats(ctx)
	if err != nil {
		return fmt.Errorf("fetch lan stats: %w", err)
	}
	wireless24Stats, err := e.client.FetchWireless24Stats(ctx)
	if err != nil {
		return fmt.Errorf("fetch wireless 2.4 stats: %w", err)
	}
	wireless5Stats, err := e.client.FetchWireless5Stats(ctx)
	if err != nil {
		return fmt.Errorf("fetch wireless 5 stats: %w", err)
	}

	e.g.cpuTotal.Set(float64(cpu.Device.CPU.Time.Total))
	e.g.cpuUser.Set(float64(cpu.Device.CPU.Time.User))
	e.g.cpuSystem.Set(float64(cpu.Device.CPU.Time.System))
	e.g.cpuIdle.Set(float64(cpu.Device.CPU.Time.Idle))
	e.g.cpuTemperature.Set(float64(cpu.Device.CPU.Temperature.Main))

	e.g.memTotal.Set(float64(mem.Device.Mem.Total))
	e.g.memFree.Set(float64(mem.Device.Mem.Free))

	e.g.wanRxBytes.Set(float64(wanStats.Wan.IP.Stats.Rx.Bytes))
	e.g.wanTxBytes.Set(float64(wanStats.Wan.IP.Stats.Tx.Bytes))
	e.g.wanRxContractual.Set(float64(wanStats.Wan.IP.Stats.Rx.ContractualBandwidth))
	e.g.wanTxContractual.Set(float64(wanStats.Wan.IP.Stats.Tx.ContractualBandwidth))
	e.g.wanInternetState.Set(float64(wanInfo.Wan.Internet.State))
	e.g.wanInterfaceState.Set(float64(wanInfo.Wan.Interface.State))
	if strings.EqualFold(wanInfo.Wan.IP.State, "up") {
		e.g.wanIPState.Set(1)
	} else {
		e.g.wanIPState.Set(0)
	}
	e.g.wanCgnatEnabled.Set(float64(wanInfo.Wan.IP.CgnatEnable))

	ipv6Addrs := make([]string, 0, len(wanInfo.Wan.IP.IP6Address))
	for _, a := range wanInfo.Wan.IP.IP6Address {
		ipv6Addrs = append(ipv6Addrs, a.IPAddress)
	}
	ipv6Prefixes := make([]string, 0, len(wanInfo.Wan.IP.IP6Prefix))
	for _, p := range wanInfo.Wan.IP.IP6Prefix {
		ipv6Prefixes = append(ipv6Prefixes, p.Prefix)
	}

	e.g.wanInfo.Reset()
	e.g.wanInfo.WithLabelValues(
		wanInfo.Wan.IP.Address,
		wanInfo.Wan.IP.Gateway,
		wanInfo.Wan.IP.DNSServers,
		wanInfo.Wan.IP.DNSServersV6,
		wanInfo.Wan.IP.Subnet,
		wanInfo.Wan.IP.Mac,
		wanInfo.Wan.IP.State,
		wanInfo.Wan.IP.IP6State,
		strings.Join(ipv6Addrs, ","),
		strings.Join(ipv6Prefixes, ","),
		wanInfo.Wan.Link.State,
		wanInfo.Wan.Link.Type,
		fmt.Sprintf("%d", wanInfo.Wan.IP.MaptEnable),
		fmt.Sprintf("%d", wanInfo.Wan.IP.MTU),
	).Set(1)

	e.g.lanRxBytes.Set(float64(lanStats.Lan.Stats.Rx.Bytes))
	e.g.lanTxBytes.Set(float64(lanStats.Lan.Stats.Tx.Bytes))

	e.g.wireless24RxBytes.Set(float64(wireless24Stats.Wireless.SSID.Stats.Rx.Bytes))
	e.g.wireless24TxBytes.Set(float64(wireless24Stats.Wireless.SSID.Stats.Tx.Bytes))
	e.g.wireless5RxBytes.Set(float64(wireless5Stats.Wireless.SSID.Stats.Rx.Bytes))
	e.g.wireless5TxBytes.Set(float64(wireless5Stats.Wireless.SSID.Stats.Tx.Bytes))

	return nil
}
