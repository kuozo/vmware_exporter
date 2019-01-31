package pkg

import (
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// Exporter as prom exporter intstance for vmware
type Exporter struct {
	vsphereHost     string
	vsphereUser     string
	vspherePassword string
	ignoreSSL       bool
	metrics         map[string]*prometheus.GaugeVec
	descs           map[string]*prometheus.Desc
	metricsMtx      sync.RWMutex
}

type scrapeResult struct {
	Name   string
	Value  float64
	Labels map[string]string
}

// NewVMwareExporter returns a new exporter of vmware metrics.
func NewVMwareExporter(vsphereHost, vsphereUser, vspherePassword string, ignoreSSL bool) (*Exporter, error) {

	e := Exporter{
		vsphereHost:     vsphereHost,
		vsphereUser:     vsphereUser,
		vspherePassword: vspherePassword,
		ignoreSSL:       ignoreSSL,
	}
	// e.initGauges()
	return &e, nil
}

// NewVMDesc returns prom desc
func (e *Exporter) NewVMDesc(metricName, help string) *prometheus.Desc {

	name := GenerateMetricName(NAMESPECE, metricName)
	return prometheus.NewDesc(name, help, vmMetricLables, nil)
}

// Describe outputs VMware metric descriptions.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// for _, m := range e.metrics {
	// 	m.Describe(ch)
	// }
	ch <- scrapeDurationDesc
}

// Collect fetches new metrics from the VMware and updates the appropriate metrics.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrapeVMware(ch)
}

// scrapeVMware returns scrape vmware result
func (e *Exporter) scrapeVMware(ch chan<- prometheus.Metric) {

	now := time.Now()
	duration := time.Since(now)

	vmware, err := NewVMware(e.vsphereHost, e.vsphereUser, e.vspherePassword, e.ignoreSSL)
	if err != nil {
		log.Errorf("creating vmware failed (%s)", err)
		return
	}
	defer vmware.destroy()

	// Ready resources
	metrics := vmware.DeclareMetrics()
	spec := vmware.DeclareSpec()

	finder, datacenterList, err := vmware.GetDataCenterList()
	if err != nil {
		return
	}

	for _, datacenter := range datacenterList {

		vms, err := vmware.GetVirtaulMachineByDataCenter(*finder, datacenter)
		if err != nil {
			log.Errorf("get virtalMachine failed (%s)", err)
			continue
		}
		wg := sync.WaitGroup{}
		vmBasics, err := vmware.GetVirtaulMachineBasicsInfo(vms)
		for _, vm := range vmBasics {
			wg.Add(2)
			instance := vmware.GenLocalInstance(vm, datacenter.Name())

			// Collect VM Summary Config Metrics
			go func(vm mo.VirtualMachine) {
				defer wg.Done()
				e.scrapeBasicMetrics(ch, vm, *instance)
			}(vm)

			// // Collect VM Performance Metrics
			go func(vm mo.VirtualMachine) {
				defer wg.Done()
				entity := vm.Reference()
				objs := []types.ManagedObjectReference{entity}
				vmMetrics, err := vmware.GetVirtaulMachinePerformance(instance.uuid, objs, spec, metrics)
				if err != nil {
					log.Errorf("Get Virtual Machine(%s) performance failed (%s)", instance.uuid, err)
					return
				}
				e.scrapePerformanceMetrics(ch, *instance, vmMetrics)
			}(vm)
		}
		wg.Wait()
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), "vmware")

}

// scrapeBasicMetrics returns scraping basic metrics.
func (e *Exporter) scrapeBasicMetrics(ch chan<- prometheus.Metric, vm mo.VirtualMachine, instance Instance) {

	summary := vm.Summary

	// CPU Usage
	vmCPUUsage := CalMetricPercent(
		float64(summary.QuickStats.OverallCpuUsage),
		float64(summary.Runtime.MaxCpuUsage),
	)
	labelValues := instance.LableValues("")

	// CPU
	ch <- prometheus.MustNewConstMetric(
		e.NewVMDesc("vm_cpu_usage", "vmware VM CPU Usage(percent) from summary"),
		prometheus.GaugeValue,
		vmCPUUsage,
		labelValues...,
	)

	// Memory
	ch <- prometheus.MustNewConstMetric(
		e.NewVMDesc("vm_mem_total", "vmware VM Memory Total(MB) from summary"),
		prometheus.GaugeValue,
		float64(summary.Config.MemorySizeMB),
		labelValues...,
	)
	memoryUsage := float64(summary.QuickStats.GuestMemoryUsage) * 1024 * 1024
	memoryCapacity := float64(summary.Runtime.MaxMemoryUsage) * 1024 * 1024
	memoryUsagePercent := CalMetricPercent(memoryUsage, memoryCapacity)

	ch <- prometheus.MustNewConstMetric(
		e.NewVMDesc("vm_mem_usage", "vmware VM Usage from summary"),
		prometheus.GaugeValue,
		memoryUsage,
		labelValues...,
	)
	ch <- prometheus.MustNewConstMetric(
		e.NewVMDesc("vm_mem_capacity", "vmware VM Memory Capacity from summary"),
		prometheus.GaugeValue,
		memoryCapacity,
		labelValues...,
	)
	ch <- prometheus.MustNewConstMetric(
		e.NewVMDesc("vm_mem_usage_percent", "vmware VM Usage Percent(Usage/Capacity*100) from summary"),
		prometheus.GaugeValue,
		memoryUsagePercent,
		labelValues...,
	)
	wg := sync.WaitGroup{}
	for _, disk := range vm.Guest.Disk {
		wg.Add(1)
		go func(disk types.GuestDiskInfo) {
			defer wg.Done()
			diskPath := strings.Replace(disk.DiskPath, `\`, "", -1)
			diskvalues := instance.LableValues(diskPath)
			free := disk.FreeSpace
			capacity := disk.Capacity
			freePercent := CalMetricPercent(float64(free), float64(capacity))
			ch <- prometheus.MustNewConstMetric(
				e.NewVMDesc("vm_guest_disk_free", "vmware VM Disk Free from summary"),
				prometheus.GaugeValue,
				float64(free),
				diskvalues...,
			)
			ch <- prometheus.MustNewConstMetric(
				e.NewVMDesc("vm_guest_disk_capacity", "vmware VM Disk Capacity from summary"),
				prometheus.GaugeValue,
				float64(free),
				diskvalues...,
			)
			ch <- prometheus.MustNewConstMetric(
				e.NewVMDesc("vm_guest_disk_free_percent", "vmware VM Disk Free Percent from summary"),
				prometheus.GaugeValue,
				freePercent,
				diskvalues...,
			)
		}(disk)
	}
	wg.Wait()
}

// scrapePerformanceMetrics scrape MetricSeries from Performance
func (e *Exporter) scrapePerformanceMetrics(ch chan<- prometheus.Metric, instance Instance, metrics []performance.EntityMetric) {

	for _, metric := range metrics {

		for _, m := range metric.Value {
			if m.Instance != "" {
				continue
			}
			name := ""
			if measure, ok := vmMetricMap[m.Name]; ok {
				name = measure
			} else {
				name = strings.Replace(m.Name, ".", "_", -1)
			}
			labelValues := instance.LableValues(m.Instance)
			ch <- prometheus.MustNewConstMetric(
				e.NewVMDesc(name, name),
				prometheus.GaugeValue,
				float64(m.Value[0]),
				labelValues...,
			)
		}
	}
}

func (e *Exporter) setMetrics(scrapes chan scrapeResult) {

	for item := range scrapes {
		name := item.Name
		if _, ok := e.metrics[name]; !ok {
			e.metricsMtx.Lock()
			e.metrics[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: NAMESPECE,
				Name:      name,
				Help:      name + "metric", // needs to be set for prometheus >= 2.3.1
			}, vmMetricLables)
			e.metricsMtx.Unlock()
		}
		var labels prometheus.Labels = item.Labels
		e.metrics[name].With(labels).Set(float64(item.Value))
	}
}

func (e *Exporter) collectMetrics(metrics chan<- prometheus.Metric) {
	for _, m := range e.metrics {
		m.Collect(metrics)
	}
}
