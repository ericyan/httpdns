# httpdns

httpdns is an recursive DNS server that utilises HTTP(S) to communicate
with its upstream. This makes it (mostly) immune to DNS hijacking. For
the optimal name resolution, it supports EDNS Client Subnet as defined
in [RFC 7871].

Due to limitations of [DNSPod], the upstream, only querying `A` records
is supported.

[RFC 7871]: https://datatracker.ietf.org/doc/rfc7871/
[DNSPod]: https://www.dnspod.cn/httpdns
