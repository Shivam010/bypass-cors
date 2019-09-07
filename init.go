package main

import (
	"flag"
)

const (
	shortHand = " (short hand)"

	defaultPort = "8080"
	usagePort   = "PORT at which the server will run"
)

var (
	// PORT at which the server will run (default: 8080),
	// can be modified using flags:
	// 	`-port 80` or `-p 80`
	PORT string
)

func init() {
	flag.StringVar(&PORT, "port", defaultPort, usagePort)
	flag.StringVar(&PORT, "p", defaultPort, usagePort+shortHand)
}
