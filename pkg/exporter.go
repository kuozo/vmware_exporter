package pkg

import (
	"sync"
	//"context"

	//"github.com/vmware/govmomi"
	"github.com/prometheus/client_golang/prometheus"
)


type Exporter struct {
	vsphereHost string
	vsphereUser string
	vspherePassword string
	ignoreSSL bool
	metrics  map[string]*prometheus.GaugeVec
	sync.RWMutex
	namespace string
}
// NewVMwareExporter returns a new exporter of vmware metrics.
func NewVMwareExporter(vsphereHost, vsphereUser, vspherePassword string, ignoreSSL bool)(*Exporter, error){

	e := Exporter{
		vsphereHost: vsphereHost,
		vsphereUser: vsphereUser, 
		vspherePassword: vspherePassword,
		ignoreSSL: ignoreSSL,
	}
	e.initGauges()
	return &e, nil
}

func (e *Exporter) initGauges(){
	e.metrics = map[string]*prometheus.GaugeVec{}
	// e.metrics[""] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		
	// })
}

// Describe outputs VMware metric descriptions.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	for _, m := range e.metrics {
		m.Describe(ch)
	}

	// ch <- e.duration.Desc()
	// ch <- e.totalScrapes.Desc()
	// ch <- e.scrapeErrors.Desc()
}

// Collect fetches new metrics from the RedisHost and updates the appropriate metrics.
func (e *Exporter)Collect(ch chan<- prometheus.Metric){

}

func (e *Exporter) scrapeVMware(){

}