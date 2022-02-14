package Metrics

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	ExternalMetrics "server-agent/handlers/external"
)

type MemInfo struct {
	Total string
	Used  string
	Free  string
}
type SwapInfo struct {
	Total string
	Used  string
	Free  string
}
type Percent struct {
	CPU  string
	Mem  string
	Disk string
	Swap string
}
type Load struct {
	CPU  *load.AvgStat
	Swap SwapInfo
	Mem  MemInfo
}
type Host struct {
	Uptime uint64
}
type InterfaceInfo struct {
	Addrs    []string
	ByteSent uint64
	ByteRecv uint64
}

type DynamicMetrics struct {
	Percent     Percent
	Load        Load
	Host        Host
	Network     map[string]InterfaceInfo
	Container   []ExternalMetrics.DockerContainer
	MessageType string
}

func DynamicMetricsData() *DynamicMetrics {
	cc, _ := cpu.Percent(time.Second, false)
	v, _ := mem.VirtualMemory()
	vs, _ := mem.SwapMemory()
	d, _ := disk.Usage("/")
	n, _ := host.Info()
	nv, _ := net.IOCounters(true)

	ss := new(DynamicMetrics)
	ss.Percent.CPU = fmt.Sprintf("%.3g", cc[0])
	ss.Percent.Mem = fmt.Sprintf("%.3g", v.UsedPercent)
	ss.Percent.Disk = fmt.Sprintf("%.3g", d.UsedPercent)
	ss.Percent.Swap = fmt.Sprintf("%.3g", vs.UsedPercent)

	t_cpu_load, _ := load.Avg()
	ss.Load.CPU = t_cpu_load
	ss.Load.Swap.Total = fmt.Sprintf("%d", vs.Total/1024/1024)
	ss.Load.Swap.Used = fmt.Sprintf("%d", vs.Used/1024/1024)
	ss.Load.Swap.Free = fmt.Sprintf("%d", vs.Free/1024/1024)
	ss.Load.Mem.Total = fmt.Sprintf("%d", v.Total/1024/1024)
	ss.Load.Mem.Used = fmt.Sprintf("%d", v.Used/1024/1024)
	ss.Load.Mem.Free = fmt.Sprintf("%d", v.Free/1024/1024)

	ss.Host.Uptime = n.Uptime

	ss.Network = make(map[string]InterfaceInfo)
	for _, v := range nv {
		if v.Name != "lo" {
			var ii InterfaceInfo
			ii.ByteSent = v.BytesSent
			ii.ByteRecv = v.BytesRecv
			ss.Network[v.Name] = ii
		}
	}
	ss.Container = ExternalMetrics.ListDockerContainers()
	ss.MessageType = "state"

	return ss
}
