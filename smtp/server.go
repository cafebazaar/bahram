package smtp

import (
	"fmt"
	"log"
	"net"

	"github.com/cafebazaar/bahram/datasource"
	"github.com/cafebazaar/blacksmith/logging"
)

const (
	debugTag = "SMTP"
)

func logln(level int, s string) {
	if level == 2 {
		log.Fatalf(s)
	} else if level == 1 {
		logging.Log(debugTag, s)
	} else {
		logging.Debug(debugTag, s)
	}
}

func Serve(listenAddr net.TCPAddr, datasource datasource.DataSource) error {
	addr := listenAddr.String()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logln(2, fmt.Sprintf("Cannot listen on port, %v\n", err))
	} else {
		logln(1, fmt.Sprintf("Listening on tcp %s\n", addr))
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			logln(1, fmt.Sprintf("Accept error: %s\n", err))
			continue
		}
		logln(1, conn.RemoteAddr().String())
		logln(1, fmt.Sprintf("\nSalam\n"))
	}

	return nil
}
