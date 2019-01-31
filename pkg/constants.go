package pkg

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// NAMESPECE metirc namespace
	NAMESPECE = "vmware"
	// VM Metrics Ojbects
	VM = "VirtualMachine"
)

var vmProperties = []string{
	"summary",
	"guest",
	"runtime",
}

var vmMetricLables = []string{
	"uuid",
	"instance",
}

// vm metrics mapping(per)
var vmMetricMap = map[string]string{
	// CPU
	"cpu.usage.average":  "vm_cpu_usage_average",
	"cpu.idle.summation": "vm_cpu_idle_summation",
	//"cpu.usagemhz.average": "vm_cpu_usagemhz_average",

	// Memory
	"mem.usage.average":    "vm_mem_usage_average",
	"mem.active.average":   "vm_mem_active_average",
	"mem.consumed.average": "vm_mem_consumed_average",

	// Network
	"net.usage.average":       "vm_net_usage_average",
	"net.transmitted.average": "vm_net_transmitted_average",
	"net.received.average":    "vm_net_received_average",

	// Disk
	"disk.write.average": "vm_disk_write_average",
	"disk.read.average":  "vm_disk_read_average",
}

var scrapeDurationDesc = prometheus.NewDesc(
	prometheus.BuildFQName(NAMESPECE, "scrape", "collector_duration_seconds"),
	"vmware_exporter: Duration of a collector scrape.",
	[]string{"collector"},
	nil,
)
