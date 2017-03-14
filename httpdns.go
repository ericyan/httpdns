package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/miekg/dns"
)

func main() {
	dns.HandleFunc(".", handleRequest)

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

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	// TODO: Handle incoming queries properly
	log.Printf("Question: %s\n", r.Question[0].String())
	m.SetRcode(r, dns.RcodeServerFailure)

	w.WriteMsg(m)
}
