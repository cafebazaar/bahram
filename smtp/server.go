package smtp

import (
	"net"

	"github.com/cafebazaar/bahram/datasource"
)

const (
	debugTag = "SMTP"
)

func Serve(listenAddr net.TCPAddr, datasource *datasource.DataSource) error {
	return nil
}
