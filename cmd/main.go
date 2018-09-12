package main

import (
	"os"
	"flag"
	"runtime"
	"strconv"
	"net/http"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/klnchu/vmware_exporter/pkg"

	log "github.com/sirupsen/logrus"
)

var (
	vsphereHost       = flag.String("vsphere_host", "", "")
	vsphereUser       = flag.String("vsphere_user", "", "")
	vspherePassword   = flag.String("vsphere_password", "", "")
	ignoreSSL         = flag.Bool("ignore_ssl", true, "ssl")
	port              = flag.String("port", ":9272", "VMware exporter http port")
	metricPath        = flag.String("metric_path", "/metrics", "Metrics Path")
	vmwareMetricsOnly = flag.Bool("vmware-only-metrics", getEnvBool("VMWARE_EXPORTER_ONLY_METRICS"), "Whether to export go runtime metrics also")
	VERSION           = "<<< filled in by build >>>"
	COMMIT_SHA1       = "<<< filled in by build >>>"
	BUILD_DATE        = "<<< filled in by build >>>"
)

func init(){
	log.SetFormatter(&log.JSONFormatter{})
}

func getEnvBool(key string) (envValBool bool) {
	if envVal, ok := os.LookupEnv(key); ok {
		envValBool, _ = strconv.ParseBool(envVal)
	}else{
		envValBool = true
	}
	return
}

func main(){

	flag.Parse()

	exp, err := pkg.NewVMwareExporter(*vsphereHost, *vsphereUser, *vspherePassword, *ignoreSSL)
	
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vmware_exporter_build_info",
		Help: "vmware exporter build_info",
	}, []string{"version", "commit_sha", "build_date", "golang_version"})
	buildInfo.WithLabelValues(VERSION, COMMIT_SHA1, BUILD_DATE, runtime.Version()).Set(1)

	if *vmwareMetricsOnly{
		registry := prometheus.NewRegistry()
		registry.Register(exp)
		registry.Register(buildInfo)
		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		http.Handle(*metricPath, handler)
	}else{
		prometheus.MustRegister(exp)
		prometheus.MustRegister(buildInfo)
		http.Handle(*metricPath, promhttp.Handler())
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>VMWare Exporter v` + VERSION + `</title></head>
<body>
<h1>VMWare Exporter ` + VERSION + `</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>`))
	})
	log.Printf("Providing metrics at %s%s", *port, *metricPath)
	http.ListenAndServe(":9272", nil)
}
