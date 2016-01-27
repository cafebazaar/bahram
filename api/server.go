package api

import (
	"net"

	"github.com/cafebazaar/bahram/datasource"
)

const (
	debugTag = "API"
)

func Serve(listenAddr net.TCPAddr, datasource datasource.DataSource) error {
	return nil
}
