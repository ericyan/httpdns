package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/miekg/dns"
)

func main() {
	dns.HandleFunc(".", handleRequest)
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

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	q := r.Question[0]
	if q.Qtype == dns.TypeA {
		clientIP := net.IP{}
		opt := r.IsEdns0()
		if opt != nil {
			for _, edns := range opt.Option {
				if edns.Option() == dns.EDNS0SUBNET {
					clientIP = edns.(*dns.EDNS0_SUBNET).Address
				}
			}
		}

		result, ttl, err := query(q.Name, clientIP)
		if err != nil {
			log.Printf("Error: %s", err)
			m.SetRcode(r, dns.RcodeServerFailure)
		} else {
			if len(result) > 0 {
				answer := make([]dns.RR, 0, len(result))
				for _, ip := range result {
					rr := new(dns.A)
					rr.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(ttl)}
					rr.A = ip

					answer = append(answer, rr)
				}

				m.Answer = answer
			} else {
				m.SetRcode(r, dns.RcodeNameError)
			}

			log.Printf("%s -> %v\n", q.Name, result)
		}
	} else {
		m.SetRcode(r, dns.RcodeNotImplemented)
	}

	w.WriteMsg(m)
}

func query(dn string, ip net.IP) ([]net.IP, int, error) {
	result := make([]net.IP, 0)
	ttl := 600

	qs := "http://119.29.29.29/d?dn=" + dn + "&ttl=1"
	if len(ip) > 0 {
		qs = qs + "&ip=" + ip.String()
	}
	resp, err := http.Get(qs)
	if err != nil {
		return result, ttl, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, ttl, err
	}
	data := strings.SplitN(string(body), ",", 2)

	for _, ip := range strings.Split(data[0], ";") {
		if ip := net.ParseIP(ip); ip != nil {
			result = append(result, ip)
		}
	}

	ttl, err = strconv.Atoi(data[1])
	if err != nil {
		ttl = 600
	}

	return result, ttl, nil
}