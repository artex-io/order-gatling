package cmd

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sylr.dev/fix/config"
)

func InitHTTP() error {
	options := config.GetOptions()

	if !options.Metrics && !options.PProf {
		return nil
	}

	mux := http.NewServeMux()

	if options.Metrics {
		mux.Handle("/metrics", promhttp.Handler())
	}
	if options.PProf {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	go http.ListenAndServe(fmt.Sprintf(":%d", options.HTTPPort), mux)

	return nil
}
