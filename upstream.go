package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

// The UpstreamFunc type is an adapter to allow the use of ordinary
// functions as upstreams.
type UpstreamFunc func(string, net.IP) (*answer, error)

type answer struct {
	qname   string
	records []dns.RR
}

func newAnswer(qname string) *answer {
	return &answer{qname, make([]dns.RR, 0)}
}

func (a *answer) addRecord(ip string, ttl int) {
	if ip := net.ParseIP(ip); ip != nil {
		rr := new(dns.A)
		rr.Hdr = dns.RR_Header{Name: a.qname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(ttl)}
		rr.A = ip

		a.records = append(a.records, rr)
	}
}

func dnspod(dn string, ip net.IP) (*answer, error) {
	answer := newAnswer(dn)

	qs := "http://119.29.29.29/d?dn=" + dn + "&ttl=1"
	if len(ip) > 0 {
		qs = qs + "&ip=" + ip.String()
	}
	resp, err := http.Get(qs)
	if err != nil {
		return answer, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return answer, err
	}
	data := strings.SplitN(string(body), ",", 2)

	ttl, err := strconv.Atoi(data[1])
	if err != nil {
		ttl = 600
	}

	for _, ip := range strings.Split(data[0], ";") {
		answer.addRecord(ip, ttl)
	}

	return answer, nil
}

func dns114(dn string, ip net.IP) (*answer, error) {
	answer := newAnswer(dn)

	qs := "http://114.114.114.114/d?dn=" + dn + "&type=a&ttl=y"
	if len(ip) > 0 {
		qs = qs + "&ip=" + ip.String()
	}
	resp, err := http.Get(qs)
	if err != nil {
		return answer, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return answer, err
	}

	for _, record := range strings.Split(string(body), ";") {
		record := strings.Split(record, ",")

		ttl, err := strconv.Atoi(record[1])
		if err != nil {
			ttl = 600
		}

		answer.addRecord(record[0], ttl)
	}

	return answer, nil
}
