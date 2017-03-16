package main

import (
	"log"
	"net"

	"github.com/miekg/dns"
)

type handler struct {
	upstream UpstreamFunc
}

// ServeDNS implements dns.Handler interface
func (h handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
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

		answer, err := h.upstream(q.Name, clientIP)
		if err != nil {
			log.Printf("Error: %s", err)
			m.SetRcode(r, dns.RcodeServerFailure)
		} else {
			if len(answer.records) > 0 {
				m.Answer = answer.records
			} else {
				m.SetRcode(r, dns.RcodeNameError)
			}

			log.Printf("%s -> %v\n", q.Name, answer.records)
		}
	} else {
		m.SetRcode(r, dns.RcodeNotImplemented)
	}

	w.WriteMsg(m)
}
