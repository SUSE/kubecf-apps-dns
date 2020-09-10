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
// by resolving internal domain names. It satisfies plugin.Handler.ServeDNS.
func (sd *ServiceDiscovery) ServeDNS(
	ctx context.Context,
	rw dns.ResponseWriter,
	req *dns.Msg,
) (int, error) {
	qclass := req.Question[0].Qclass
	qtype := req.Question[0].Qtype
	qname := req.Question[0].Name

	if qclass == dns.ClassINET && (qtype == dns.TypeA || qtype == dns.TypeAAAA) {
		ips, err := sd.sdcClient.Discover(ctx, qname)
		if err != nil {
			sd.log.Error(err)
			return dns.RcodeServerFailure, err
		}

		if len(ips) > 0 {
			if err := sd.respond(rw, req, ips); err != nil {
				return dns.RcodeServerFailure, err
			}
			return dns.RcodeSuccess, nil
		}
	}

	return plugin.NextOrFailure(pluginName, sd.Next, ctx, rw, req)
}

func (sd *ServiceDiscovery) respond(rw dns.ResponseWriter, req *dns.Msg, ips []string) error {
	state := request.Request{W: rw, Req: req}
	qtype := state.QType()
	qname := state.Name()

	res := &dns.Msg{}
	res.SetReply(req)
	res.Authoritative = true

	answer := make([]dns.RR, 0, len(ips))
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if qtype == dns.TypeA && ip.To4() != nil {
			answer = append(answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   qname,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    sd.ttl,
				},
				A: ip,
			})
		} else if qtype == dns.TypeAAAA && ip.To16() != nil {
			answer = append(answer, &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   qname,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    sd.ttl,
				},
				AAAA: ip,
			})
		}
	}
	res.Answer = answer

	sd.log.Debugf("%s: %+v\n", qname, answer)

	return rw.WriteMsg(res)
}
