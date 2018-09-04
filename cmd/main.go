package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	vsphereHost = flag.String("vsphere_host", "", "")
	metricPath  = flag.String("metric_path", "/metrics", "Metrics Path")
	VERSION = "<<< filled in by build >>>"
)

func main(){

	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
<html>
<head><title>VMWare Exporter v` + VERSION + `</title></head>
<body>
<h1>Redis Exporter ` + VERSION + `</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
						`))
	})
	
	http.ListenAndServe(":9272", nil)
}