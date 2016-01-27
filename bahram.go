package main // import "github.com/cafebazaar/bahram"

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/cafebazaar/bahram/api"
	"github.com/cafebazaar/bahram/datasource"
	"github.com/cafebazaar/bahram/ldap"
	"github.com/cafebazaar/bahram/smtp"
	"github.com/cafebazaar/blacksmith/logging"
	//	etcd "github.com/coreos/etcd/client"
)

const (
	debugTag = "MAIN"
)

var (
	versionFlag = flag.Bool("version", false, "Print version info and exit")
	debugFlag   = flag.Bool("debug", false, "Log more things that aren't directly related to booting a recognized client")

	version   string
	commit    string
	buildTime string
)

func init() {
	// If version, commit, or build time are not set, make that clear.
	if version == "" {
		version = "unknown"
	}
	if commit == "" {
		commit = "unknown"
	}
	if buildTime == "" {
		buildTime = "unknown"
	}
}

func main() {
	var err error
	flag.Parse()

	fmt.Printf("Bahram (%s)\n", version)
	fmt.Printf("  Commit:        %s\n", commit)
	fmt.Printf("  Build Time:    %s\n", buildTime)

	if *versionFlag {
		os.Exit(0)
	}

	etcdDataSource, err := datasource.NewEtcdDataSource(nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nCouldn't create datasource: %s\n", err)
		os.Exit(1)
	}

	var apiAddr = net.TCPAddr{IP: net.IPv4zero, Port: 80}
	var ldapAddr = net.TCPAddr{IP: net.IPv4zero, Port: 389}
	var smtpAddr = net.TCPAddr{IP: net.IPv4zero, Port: 25}

	go func() {
		logging.RecordLogs(log.New(os.Stderr, "", log.LstdFlags), *debugFlag)
	}()

	// serving api
	go func() {
		err := api.Serve(apiAddr, etcdDataSource)
		log.Fatalf("\nError while serving api: %s\n", err)
	}()

	// serving ldap
	go func() {
		err := ldap.Serve(ldapAddr, etcdDataSource)
		log.Fatalf("\nError while serving ldap: %s\n", err)
	}()

	// serving smtp
	go func() {
		err := smtp.Serve(smtpAddr, etcdDataSource)
		log.Fatalf("\nError while serving smtp: %s\n", err)
	}()
}
