package axonx

import (
	"net/http"
	"net/http/pprof"
)

func DebugServer(address string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/cmd", pprof.Cmdline)
	mux.HandleFunc("/debug/index", pprof.Index)
	mux.HandleFunc("/debug/profile", pprof.Profile)
	mux.HandleFunc("/debug/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/trace", pprof.Trace)
	mux.Handle("/debug/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/debug/heap", pprof.Handler("heap"))
	mux.Handle("/debug/block", pprof.Handler("block"))
	mux.Handle("/debug/mutex", pprof.Handler("mutex"))
	http.ListenAndServe(address, mux)
}
