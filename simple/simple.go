package simple

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("simple", simple)
}

func simple(w http.ResponseWriter, r *http.Request) {
	dumpReq, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatalf("failed to dump HTTP request; %v", err.Error())
	}
	log.Print("dump HTTP request")
	log.Print(string(dumpReq))

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Fatalf("failed to write response; %v", err.Error())
	}
}
