package Metrics

import (
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

type CPUInfo struct {
	ModelName string
	Cores     int
}

type SystemInfo struct {
	Os             string
	OsVersion      string
	Architecture   string
	Virtualization string
}

type StaticMetrics struct {
	CPU             CPUInfo
	System          SystemInfo
	OuterIPAddr     string
	EnableContainer bool
	MessageType     string
}

// Get preferred outbound ip of this machine
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
	// FIXME: another solution, use udp detect cannot break container.
	// req, _ := http.NewRequest("GET", "http://ip.sb/", nil)
	// req.Header.Set("User-Agent", "curl/7.74.0")
	// resp, err := (&http.Client{}).Do(req)
	// if err != nil {
	// 	return ""
	// }
	// defer resp.Body.Close()
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return ""
	// }
	// bs := string(body)
	// return bs
}

func StaticMetricsData() *StaticMetrics {
	enable_docker := os.Getenv("AGENT_ENABLE_CONTAINERS")
	ss := new(StaticMetrics)

	// psutil - cpu
	c, _ := cpu.Info()
	t_cpu := make([]CPUInfo, len(c))
	for i, ci := range c {
		t_cpu[i].ModelName = ci.ModelName
	}
	ss.CPU.ModelName = t_cpu[0].ModelName
	ss.CPU.Cores = len(c)

	// psutil - host
	n, _ := host.Info()
	ss.System.Os = n.Platform
	ss.System.OsVersion = n.PlatformVersion
	ss.System.Architecture = runtime.GOARCH
	ss.System.Virtualization = n.VirtualizationSystem

	ss.OuterIPAddr = strings.ReplaceAll(getOutboundIP(), "\n", "")

	ss.MessageType = "info"

	if enable_docker == "docker" {
		ss.EnableContainer = true
	}

	return ss
}
