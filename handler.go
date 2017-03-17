package main

import (
	"io/ioutil"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
)

type handler struct {
	upstream     UpstreamFunc
	ecsOverrides string
}

// ServeDNS implements dns.Handler interface
func (h handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	q := r.Question[0]
	if q.Qtype == dns.TypeA {
		clientIP := h.getECS(r)
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

func (h handler) getECS(r *dns.Msg) net.IP {
	ecs := net.IP{}

	opt := r.IsEdns0()
	if opt != nil {
		for _, edns := range opt.Option {
			if edns.Option() == dns.EDNS0SUBNET {
				ecs = edns.(*dns.EDNS0_SUBNET).Address
			}
		}
	}

	if h.ecsOverrides != "" && len(ecs) > 0 {
		override, err := ioutil.ReadFile(h.ecsOverrides + "/" + ecs.String())
		if err == nil {
			ecsOverride := net.ParseIP(strings.TrimSpace(string(override)))
			log.Printf("ECS override: %s -> %s\n", ecs, ecsOverride)
			ecs = ecsOverride
		} else {
			log.Printf("Error loading ECS override: %s\n", err)
		}
	}

	log.Printf("ECS: %s\n", ecs)
	return ecs
}
