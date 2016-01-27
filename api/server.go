package api

import (
	"net"
	"net/http"

	"github.com/cafebazaar/bahram/datasource"
	"github.com/cafebazaar/blacksmith/logging"
)

const (
	debugTag = "API/SERVER"
)

func Serve(listenAddr net.TCPAddr, datasource datasource.DataSource) error {
	logging.Log(debugTag, "Serving Rest API on %s", listenAddr)
	restApi := newRestServerAPI(datasource)
	handler, err := restApi.MakeHandler()
	if err != nil {
		return err
	}
	return http.ListenAndServe(listenAddr.String(), handler)
}
