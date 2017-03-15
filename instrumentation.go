package main

import "github.com/miekg/dns"

func instrumentation(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	q := r.Question[0]
	if q.Name == "ping.httpdns." && q.Qtype == dns.TypeTXT {
		pong := new(dns.TXT)
		pong.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: uint32(60)}
		pong.Txt = []string{"pong"}

		m.Answer = []dns.RR{pong}
	} else {
		m.SetRcode(r, dns.RcodeNameError)
	}

	w.WriteMsg(m)
}
