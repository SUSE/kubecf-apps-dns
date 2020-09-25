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

package apps_dns_test

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"

	"github.com/miekg/dns"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SUSE/kubecf-apps-dns/tests/acceptance/apps_dns/fake"
)

var _ = Describe("AppsDns", func() {
	It("should return a DNS Server Failure when the Service Discovery Controller fails", func() {
		// Prepare
		domainName := dns.Fqdn("example.com")
		sdc.Handle(domainName, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		// Assert
		By("sending a question type A")
		res, err := dnsQuery(domainName, dns.TypeA)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Rcode).To(Equal(dns.RcodeServerFailure))
		Expect(res.Answer).To(BeEmpty())

		By("sending a question type AAAA")
		res, err = dnsQuery(domainName, dns.TypeAAAA)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Rcode).To(Equal(dns.RcodeServerFailure))
		Expect(res.Answer).To(BeEmpty())
	})

	It("should resolve a domain name outside of the Apps DNS authority", func() {
		// Prepare
		domainName := dns.Fqdn("github.com")
		ips := []net.IP{}
		sdc.Handle(domainName, fake.Handler(ips))

		// Assert
		By("sending a question type A")
		res, err := dnsQuery(domainName, dns.TypeA)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Rcode).To(Equal(dns.RcodeSuccess))
		Expect(res.Answer).ToNot(BeEmpty())
	})

	It("should resolve a k8s service domain name", func() {
		// Prepare
		domainName := dns.Fqdn(fmt.Sprintf("kubernetes.default.svc.%s", clusterDomain))
		ips := []net.IP{}
		sdc.Handle(domainName, fake.Handler(ips))

		// Assert
		By("sending a question type A")
		res, err := dnsQuery(domainName, dns.TypeA)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Rcode).To(Equal(dns.RcodeSuccess))
		Expect(res.Answer).ToNot(BeEmpty())
	})

	Context("resolves an App domain name discovered from the Service Discovery Controller", func() {
		It("should resolve to a single IPv4", func() {
			// Prepare
			domainName := dns.Fqdn(fmt.Sprintf("%d.apps.internal", rand.Uint64()))
			ips := []net.IP{net.ParseIP("10.11.12.13")}
			expectedAnswer := []dns.RR{
				&dns.A{
					Hdr: dns.RR_Header{
						Name:     domainName,
						Rrtype:   dns.TypeA,
						Class:    dns.ClassINET,
						Ttl:      ttl,
						Rdlength: 4,
					},
					A: ips[0].To4(),
				},
			}
			sdc.Handle(domainName, fake.Handler(ips))

			// Assert
			By("sending a question type A")
			res, err := dnsQuery(domainName, dns.TypeA)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Rcode).To(Equal(dns.RcodeSuccess))
			Expect(res.Answer).To(Equal(expectedAnswer))
		})

		It("should resolve to a single IPv6", func() {
			// Prepare
			domainName := dns.Fqdn(fmt.Sprintf("%d.apps.internal", rand.Uint64()))
			ips := []net.IP{net.ParseIP("2001:db8::68")}
			expectedAnswer := []dns.RR{
				&dns.AAAA{
					Hdr: dns.RR_Header{
						Name:     domainName,
						Rrtype:   dns.TypeAAAA,
						Class:    dns.ClassINET,
						Ttl:      ttl,
						Rdlength: 16,
					},
					AAAA: ips[0],
				},
			}
			sdc.Handle(domainName, fake.Handler(ips))

			// Assert
			By("sending a question type AAAA")
			res, err := dnsQuery(domainName, dns.TypeAAAA)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Rcode).To(Equal(dns.RcodeSuccess))
			Expect(res.Answer).To(Equal(expectedAnswer))
		})

		It("should resolve to multiple IPv4s", func() {
			// Prepare
			domainName := dns.Fqdn(fmt.Sprintf("%d.apps.internal", rand.Uint64()))
			ips := []net.IP{
				net.ParseIP("10.11.12.13"),
				net.ParseIP("10.11.12.14"),
			}
			expectedAnswer := []dns.RR{
				&dns.A{
					Hdr: dns.RR_Header{
						Name:     domainName,
						Rrtype:   dns.TypeA,
						Class:    dns.ClassINET,
						Ttl:      ttl,
						Rdlength: 4,
					},
					A: ips[0].To4(),
				},
				&dns.A{
					Hdr: dns.RR_Header{
						Name:     domainName,
						Rrtype:   dns.TypeA,
						Class:    dns.ClassINET,
						Ttl:      ttl,
						Rdlength: 4,
					},
					A: ips[1].To4(),
				},
			}
			sdc.Handle(domainName, fake.Handler(ips))

			// Assert
			By("sending a question type A")
			res, err := dnsQuery(domainName, dns.TypeA)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Rcode).To(Equal(dns.RcodeSuccess))
			Expect(res.Answer).To(ConsistOf(expectedAnswer))
		})

		It("should resolve to multiple IPv6s", func() {
			// Prepare
			domainName := dns.Fqdn(fmt.Sprintf("%d.apps.internal", rand.Uint64()))
			ips := []net.IP{
				net.ParseIP("2001:db8::68"),
				net.ParseIP("2001:db8::69"),
			}
			expectedAnswer := []dns.RR{
				&dns.AAAA{
					Hdr: dns.RR_Header{
						Name:     domainName,
						Rrtype:   dns.TypeAAAA,
						Class:    dns.ClassINET,
						Ttl:      ttl,
						Rdlength: 16,
					},
					AAAA: ips[0],
				},
				&dns.AAAA{
					Hdr: dns.RR_Header{
						Name:     domainName,
						Rrtype:   dns.TypeAAAA,
						Class:    dns.ClassINET,
						Ttl:      ttl,
						Rdlength: 16,
					},
					AAAA: ips[1],
				},
			}
			sdc.Handle(domainName, fake.Handler(ips))

			// Assert
			By("sending a question type AAAA")
			res, err := dnsQuery(domainName, dns.TypeAAAA)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Rcode).To(Equal(dns.RcodeSuccess))
			Expect(res.Answer).To(ConsistOf(expectedAnswer))
		})

		It("should resolve when the request is of type AAAA and the discovered service has an IPv4 address", func() {
			// Prepare
			domainName := dns.Fqdn(fmt.Sprintf("%d.apps.internal", rand.Uint64()))
			ips := []net.IP{net.ParseIP("10.11.12.13")}
			expectedAnswer := []dns.RR{
				&dns.AAAA{
					Hdr: dns.RR_Header{
						Name:     domainName,
						Rrtype:   dns.TypeAAAA,
						Class:    dns.ClassINET,
						Ttl:      ttl,
						Rdlength: 16,
					},
					AAAA: ips[0].To16(),
				},
			}
			sdc.Handle(domainName, fake.Handler(ips))

			// Assert
			By("sending a question type AAAA")
			res, err := dnsQuery(domainName, dns.TypeAAAA)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Rcode).To(Equal(dns.RcodeSuccess))
			Expect(res.Answer).To(Equal(expectedAnswer))
		})

		It("should resolve when the request is of type A and the discovered service has an IPv6 that's a valid IPv4", func() {
			// Prepare
			domainName := dns.Fqdn(fmt.Sprintf("%d.apps.internal", rand.Uint64()))
			ips := []net.IP{net.ParseIP("::ffff:a0b:c0d")}
			expectedAnswer := []dns.RR{
				&dns.A{
					Hdr: dns.RR_Header{
						Name:     domainName,
						Rrtype:   dns.TypeA,
						Class:    dns.ClassINET,
						Ttl:      ttl,
						Rdlength: 4,
					},
					A: ips[0].To4(),
				},
			}
			sdc.Handle(domainName, fake.Handler(ips))

			// Assert
			By("sending a question type A")
			res, err := dnsQuery(domainName, dns.TypeA)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Rcode).To(Equal(dns.RcodeSuccess))
			Expect(res.Answer).To(Equal(expectedAnswer))
		})

		It("should not resolve when the request is of type A and the discovered service has an IPv6 address that's not a valid IPv4", func() {
			// Prepare
			domainName := dns.Fqdn(fmt.Sprintf("%d.apps.internal", rand.Uint64()))
			ips := []net.IP{net.ParseIP("2001:db8::68")}
			sdc.Handle(domainName, fake.Handler(ips))

			// Assert
			By("sending a question type A")
			res, err := dnsQuery(domainName, dns.TypeA)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Rcode).To(Equal(dns.RcodeSuccess))
			Expect(res.Answer).To(BeEmpty())
		})
	})
})

// dnsQuery sends a query to the DNS for the domain name and type.
func dnsQuery(domainName string, questionType uint16) (*dns.Msg, error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(domainName, questionType)
	m.RecursionDesired = true
	r, _, err := c.Exchange(m, dnsAddr)
	return r, err
}
