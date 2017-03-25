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
	upstream := flag.String("upstream", "dnspod", "upstream to use")
	ecsOverrides := flag.String("ecs-overrides", "", "path to EDNS Client Subnet overrides")
	ecsIPv4Mask := flag.Int("ecs-ipv4-prefix", 24, "prefix-length of IPv4 EDNS Client Subnet")
	ecsIPv6Mask := flag.Int("ecs-ipv6-prefix", 56, "prefix-length of IPv6 EDNS Client Subnet")
	flag.Parse()

	dns.Handle(".", newHandler(*upstream, *ecsOverrides, *ecsIPv4Mask, *ecsIPv6Mask))
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
