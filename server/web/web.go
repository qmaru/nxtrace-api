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
	fmt.Fprintln(writer, a...)
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		StringHandle(400, w, "invalid request")
		return
	}

	var traceData TraceData

	err = json.Unmarshal(body, &traceData)
	if err != nil {
		StringHandle(400, w, "invalid request")
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
	config := new(common.Config)
	webCfg := config.NewWebConfig()

	listenAddr := fmt.Sprintf("%s:%s", webCfg.ServerHost, webCfg.ServerPort)
	log.Printf("Listenning: %s\n", listenAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("/trace", TraceHandle)

	return http.ListenAndServe(listenAddr, RecoveryMiddleware(mux))
}
