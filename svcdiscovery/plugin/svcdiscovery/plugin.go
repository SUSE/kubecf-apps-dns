/*
Copyright 2020 SUSE

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package svcdiscovery

import (
	"context"
	"net"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// ServiceDiscovery is a CoreDNS plugin for handling Cloud Foundry App Service
// Discovery.
type ServiceDiscovery struct {
	Next plugin.Handler

	log       clog.P
	sdcClient *SDCClient
	ttl       uint32
}

// Name satisfies plugin.Handler.Name.
func (sd *ServiceDiscovery) Name() string { return pluginName }

// ServeDNS handles domain name requests for Cloud Foundry App Service Discovery
// by resolving internal domain names. It satisfies satisfies
// plugin.Handler.ServeDNS.
func (sd *ServiceDiscovery) ServeDNS(
	ctx context.Context,
	res dns.ResponseWriter,
	req *dns.Msg,
) (int, error) {
	state := request.Request{W: res, Req: req}
	qclass := state.QClass()
	qtype := state.QType()
	qname := state.Name()

	if qclass == dns.ClassINET && (qtype == dns.TypeA || qtype == dns.TypeAAAA) {
		ips, err := sd.sdcClient.Discover(qname)
		if err != nil {
			sd.log.Error(err)
			return dns.RcodeServerFailure, err
		}

		if len(ips) > 0 {
			res = &ServiceDiscoveryResponseWriter{
				ResponseWriter: res,
				ips:            ips,
				ttl:            sd.ttl,
			}
		}
	}

	return plugin.NextOrFailure(pluginName, sd.Next, ctx, res, req)
}

// ServiceDiscoveryResponseWriter wraps a dns.ResponseWriter so it can respond
// with discovered service addresses.
type ServiceDiscoveryResponseWriter struct {
	dns.ResponseWriter

	ips []string
	ttl uint32
}

// WriteMsg writes the response for the discovered service.
func (sdrw *ServiceDiscoveryResponseWriter) WriteMsg(res *dns.Msg) error {
	answer := make([]dns.RR, 0)
	qtype := res.Question[0].Qtype
	qname := res.Question[0].Name
	for _, host := range sdrw.ips {
		ip := net.ParseIP(host)
		if qtype == dns.TypeA && ip.To4() != nil {
			answer = append(answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   qname,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    sdrw.ttl,
				},
				A: ip,
			})
		} else if qtype == dns.TypeAAAA && ip.To16() != nil {
			answer = append(answer, &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   qname,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    sdrw.ttl,
				},
				AAAA: ip,
			})
		}
	}
	res.MsgHdr.Rcode = dns.RcodeSuccess
	res.Answer = answer
	res.Ns = make([]dns.RR, 0)
	res.Extra = make([]dns.RR, 0)
	return sdrw.ResponseWriter.WriteMsg(res)
}
