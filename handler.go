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
	ecsIPv4Mask  int
	ecsIPv6Mask  int
}

func newHandler(upstream string, ecsOverrides string, ecsIPv4Mask int, ecsIPv6Mask int) handler {
	var upstreamFunc UpstreamFunc

	switch upstream {
	case "dnspod":
		upstreamFunc = dnspod
	default:
		log.Fatalf("Unsupported upstream: %s\n", upstream)
	}

	return handler{upstreamFunc, ecsOverrides, ecsIPv4Mask, ecsIPv6Mask}
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

	if len(ecs) == 0 {
		return ecs
	}

	if h.ecsOverrides != "" {
		override, err := ioutil.ReadFile(h.ecsOverrides + "/" + ecs.String())
		if err == nil {
			ecsOverride := net.ParseIP(strings.TrimSpace(string(override)))
			log.Printf("ECS override: %s -> %s\n", ecs, ecsOverride)
			ecs = ecsOverride
		} else {
			log.Printf("Error loading ECS override: %s\n", err)
		}
	}

	if ip := ecs.To4(); len(ip) == net.IPv4len {
		ecs = ecs.Mask(net.CIDRMask(h.ecsIPv4Mask, 32))
		log.Printf("ECS: %s/%d\n", ecs, h.ecsIPv4Mask)
	} else {
		ecs = ecs.Mask(net.CIDRMask(h.ecsIPv6Mask, 128))
		log.Printf("ECS: %s/%d\n", ecs, h.ecsIPv6Mask)
	}

	return ecs
}
