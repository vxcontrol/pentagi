package profiling

import (
	"net/http"
	"net/http/pprof"

	"github.com/sirupsen/logrus"
)

const profilerAddress = ":7777"

func Start() {
	router := http.NewServeMux()
	router.HandleFunc("/profiler/", pprof.Index)
	router.HandleFunc("/profiler/profile", pprof.Profile)
	router.HandleFunc("/profiler/cmdline", pprof.Cmdline)
	router.HandleFunc("/profiler/symbol", pprof.Symbol)
	router.HandleFunc("/profiler/trace", pprof.Trace)
	router.HandleFunc("/profiler/allocs", pprof.Handler("allocs").ServeHTTP)
	router.HandleFunc("/profiler/block", pprof.Handler("block").ServeHTTP)
	router.HandleFunc("/profiler/goroutine", pprof.Handler("goroutine").ServeHTTP)
	router.HandleFunc("/profiler/heap", pprof.Handler("heap").ServeHTTP)
	router.HandleFunc("/profiler/mutex", pprof.Handler("mutex").ServeHTTP)
	router.HandleFunc("/profiler/threadcreate", pprof.Handler("threadcreate").ServeHTTP)

	logrus.WithField("component", "init").Infof("start profiling server on %s", profilerAddress)

	if err := http.ListenAndServe(profilerAddress, router); err != nil { //nolint:gosec
		logrus.WithField("component", "init").WithError(err).Error("profiling monitor was exited")
	}
}
