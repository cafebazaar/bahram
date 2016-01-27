package ldap

import (
	"net"

	"github.com/cafebazaar/bahram/datasource"
)

const (
	debugTag = "LDAP"
)

func Serve(listenAddr net.TCPAddr, datasource datasource.DataSource) error {
	return nil
}
