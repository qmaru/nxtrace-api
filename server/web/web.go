package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"nxtrace-api/server/common"
)

type TraceData struct {
	Region string   `json:"region"`
	Host   string   `json:"host"`
	Params []string `json:"params"`
}

func StringHandle(code int, writer http.ResponseWriter, a ...any) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(code)
	if len(a) == 1 {
		switch v := a[0].(type) {
		case []byte:
			_, _ = writer.Write(v)
			return
		case error:
			_, _ = io.WriteString(writer, v.Error())
			return
		case string:
			_, _ = io.WriteString(writer, v)
			return
		}
	}
	fmt.Fprint(writer, a...)
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
				fmt.Printf("Recovered from panic: %v\n", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func TraceHandle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var traceData TraceData
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&traceData); err != nil {
		StringHandle(http.StatusBadRequest, w, "invalid request: "+err.Error())
		return
	}

	if traceData.Host == "" {
		StringHandle(http.StatusBadRequest, w, "missing host")
		return
	}

	host := traceData.Host
	params := traceData.Params

	output, err := common.RunTrace(host, params)
	if err != nil {
		StringHandle(503, w, err)
		return
	}

	StringHandle(200, w, output)
}

func Run() error {
	webCfg := common.NxtConfig.GetWebConfig()

	listenAddr := fmt.Sprintf("%s:%d", webCfg.ServerHost, webCfg.ServerPort)
	log.Printf("Listenning: %s\n", listenAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("/trace", TraceHandle)

	return http.ListenAndServe(listenAddr, RecoveryMiddleware(mux))
}
