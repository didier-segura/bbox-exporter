package bbox

import "fmt"

// DeviceCPU mirrors /api/v1/device/cpu payload.
type DeviceCPU struct {
	Device CPUDevice `json:"device"`
}

type CPUDevice struct {
	CPU CPUData `json:"cpu"`
}

type CPUData struct {
	Time        CPUTimes       `json:"time"`
	Process     CPUProcess     `json:"process"`
	Temperature CPUTemperature `json:"temperature"`
}

type CPUTimes struct {
	Total  int `json:"total"`
	User   int `json:"user"`
	Nice   int `json:"nice"`
	System int `json:"system"`
	IO     int `json:"io"`
	Idle   int `json:"idle"`
	IRQ    int `json:"irq"`
}

type CPUProcess struct {
	Created int `json:"created"`
	Running int `json:"running"`
	Blocked int `json:"blocked"`
}

type CPUTemperature struct {
	Main int `json:"main"`
}

// DeviceMem mirrors /api/v1/device/mem payload.
type DeviceMem struct {
	Device MemDevice `json:"device"`
}

type MemDevice struct {
	Mem MemoryStats `json:"mem"`
}

type MemoryStats struct {
	Total       int `json:"total"`
	Free        int `json:"free"`
	Cached      int `json:"cached"`
	CommittedAs int `json:"committedas"`
}

// WanIPStats mirrors /api/v1/wan/ip/stats payload.
type WanIPStats struct {
	Wan Wan `json:"wan"`
}

type Wan struct {
	IP WanIP `json:"ip"`
}

type WanIP struct {
	Stats WanIPThroughput `json:"stats"`
}

// WanIPInfo mirrors /api/v1/wan/ip payload.
type WanIPInfo struct {
	Wan WanIPDetails `json:"wan"`
}

type WanIPDetails struct {
	Internet  WanInternet  `json:"internet"`
	Interface WanInterface `json:"interface"`
	IP        WanIPAddress `json:"ip"`
	Link      WanLink      `json:"link"`
}

type WanInternet struct {
	State int `json:"state"`
}

type WanInterface struct {
	ID      int `json:"id"`
	Default int `json:"default"`
	State   int `json:"state"`
}

type WanIPAddress struct {
	Address      string          `json:"address"`
	CgnatEnable  int             `json:"cgnatenable"`
	MaptEnable   int             `json:"maptenable"`
	State        string          `json:"state"`
	Gateway      string          `json:"gateway"`
	DNSServers   string          `json:"dnsservers"`
	Subnet       string          `json:"subnet"`
	DNSServersV6 string          `json:"dnsserversv6"`
	IP6State     string          `json:"ip6state"`
	IP6Address   []WanIP6Address `json:"ip6address"`
	IP6Prefix    []WanIP6Prefix  `json:"ip6prefix"`
	Mac          string          `json:"mac"`
	MTU          int             `json:"mtu"`
}

type WanIP6Address struct {
	IPAddress string `json:"ipaddress"`
	Status    string `json:"status"`
	Valid     string `json:"valid"`
	Preferred string `json:"preferred"`
}

type WanIP6Prefix struct {
	Prefix    string `json:"prefix"`
	Status    string `json:"status"`
	Valid     string `json:"valid"`
	Preferred string `json:"preferred"`
}

type WanLink struct {
	State string `json:"state"`
	Type  string `json:"type"`
}

type WanIPThroughput struct {
	Rx WanRx `json:"rx"`
	Tx WanTx `json:"tx"`
}

type WanRx struct {
	Packets              int         `json:"packets"`
	Bytes                FlexibleInt `json:"bytes"`
	PacketsErrors        int         `json:"packetserrors"`
	PacketsDiscards      int         `json:"packetsdiscards"`
	Occupation           float64     `json:"occupation"`
	Bandwidth            int         `json:"bandwidth"`
	MaxBandwidth         int         `json:"maxBandwidth"`
	ContractualBandwidth int         `json:"contractualBandwidth"`
}

type WanTx struct {
	Packets              int         `json:"packets"`
	Bytes                FlexibleInt `json:"bytes"`
	PacketsErrors        int         `json:"packetserrors"`
	PacketsDiscards      int         `json:"packetsdiscards"`
	Occupation           float64     `json:"occupation"`
	Bandwidth            int         `json:"bandwidth"`
	MaxBandwidth         int         `json:"maxBandwidth"`
	ContractualBandwidth int         `json:"contractualBandwidth"`
}

// LanStats mirrors /api/v1/lan/stats payload.
type LanStats struct {
	Lan Lan `json:"lan"`
}

type Lan struct {
	Stats LanThroughput `json:"stats"`
}

type LanThroughput struct {
	Rx LanRx `json:"rx"`
	Tx LanTx `json:"tx"`
}

type LanRx struct {
	Bytes           FlexibleInt `json:"bytes"`
	Packets         int         `json:"packets"`
	PacketsErrors   int         `json:"packetserrors"`
	PacketsDiscards int         `json:"packetsdiscards"`
}

type LanTx struct {
	Bytes           FlexibleInt `json:"bytes"`
	Packets         int         `json:"packets"`
	PacketsErrors   int         `json:"packetserrors"`
	PacketsDiscards int         `json:"packetsdiscards"`
}

// WirelessStats mirrors /api/v1/wireless/{band}/stats payload.
type WirelessStats struct {
	Wireless Wireless `json:"wireless"`
}

type Wireless struct {
	SSID WirelessSSID `json:"ssid"`
}

type WirelessSSID struct {
	ID    int                `json:"id"`
	Stats WirelessThroughput `json:"stats"`
}

type WirelessThroughput struct {
	Rx WirelessRx `json:"rx"`
	Tx WirelessTx `json:"tx"`
}

type WirelessRx struct {
	Bytes           FlexibleInt `json:"bytes"`
	Packets         int         `json:"packets"`
	PacketsErrors   int         `json:"packetserrors"`
	PacketsDiscards int         `json:"packetsdiscards"`
}

type WirelessTx struct {
	Bytes           FlexibleInt `json:"bytes"`
	Packets         int         `json:"packets"`
	PacketsErrors   int         `json:"packetserrors"`
	PacketsDiscards int         `json:"packetsdiscards"`
}

// FlexibleInt handles APIs that sometimes return numbers as strings.
type FlexibleInt int64

func (f *FlexibleInt) UnmarshalJSON(b []byte) error {
	// Trim surrounding quotes if present.
	if len(b) > 1 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	var v int64
	if _, err := fmt.Sscan(string(b), &v); err != nil {
		return err
	}
	*f = FlexibleInt(v)
	return nil
}
