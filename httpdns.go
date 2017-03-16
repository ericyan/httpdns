package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/miekg/dns"
)

func main() {
	bind := flag.String("bind", "127.0.0.1", "interface to bind")
	port := flag.Int("port", 8653, "port to run on")
	flag.Parse()

	dns.Handle(".", handler{dnspod})
	dns.HandleFunc("httpdns.", instrumentation)

	go func() {
		srv := &dns.Server{Addr: *bind + ":" + strconv.Itoa(*port), Net: "udp"}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set udp listener %s\n", err.Error())
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Fatalf("Signal (%v) received, stopping\n", s)
}
