package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/miekg/dns"
)

func main() {
	dns.Handle(".", handler{dnspod})
	dns.HandleFunc("httpdns.", instrumentation)

	go func() {
		srv := &dns.Server{Addr: ":8653", Net: "udp"}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set udp listener %s\n", err.Error())
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Fatalf("Signal (%v) received, stopping\n", s)
}
