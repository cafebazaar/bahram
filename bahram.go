package main // import "github.com/cafebazaar/bahram"

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cafebazaar/bahram/api"
	"github.com/cafebazaar/bahram/datasource"
	"github.com/cafebazaar/bahram/smtp"
	"github.com/cafebazaar/blacksmith/logging"
	etcd "github.com/coreos/etcd/client"
)

const (
	debugTag = "MAIN"
)

var (
	versionFlag = flag.Bool("version", false, "Print version info and exit")
	debugFlag   = flag.Bool("debug", false, "Log more things that aren't directly related to booting a recognized client")
	etcdFlag    = flag.String("etcd", "", "Etcd endpoints")
	etcdDirFlag = flag.String("etcd-dir", "bahram", "Etcd path prefixe")

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

	// etcd config
	if *etcdFlag == "" || *etcdDirFlag == "" {
		// fmt.Fprint(os.Stderr, "\nPlease specify the etcd endpoints and prefix\n")
		// os.Exit(1)
		// TODO: remove these
		e1 := "http://aghajoon1.cafebazaar.ir:4001"
		e2 := "bahram"
		etcdFlag = &e1
		etcdDirFlag = &e2
	}

	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:               strings.Split(*etcdFlag, ","),
		HeaderTimeoutPerRequest: 5 * time.Second,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nCouldn't create etcd connection: %s\n", err)
		os.Exit(1)
	}
	kapi := etcd.NewKeysAPI(etcdClient)

	dataSource, err := datasource.NewDataSource(kapi, *etcdDirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nCouldn't create datasource: %s\n", err)
		os.Exit(1)
	}

	var apiAddr = net.TCPAddr{IP: net.IPv4zero, Port: 80}
	var smtpAddr = net.TCPAddr{IP: net.IPv4zero, Port: 25}

	go func() {
		err := api.Serve(apiAddr, dataSource)
		log.Fatalf("Error while serving api: %s\n", err)
	}()

	go func() {
		err := smtp.Serve(smtpAddr, dataSource)
		// log.Fatalf("Error while serving smtp: %s\n", err)
		log.Printf("Error while serving smtp: %s\n", err)
	}()

	logging.RecordLogs(log.New(os.Stderr, "", log.LstdFlags), *debugFlag)
}
