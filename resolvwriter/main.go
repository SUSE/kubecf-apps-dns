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

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	var upstreamDNSHost string
	var outPath string

	flag.StringVar(&upstreamDNSHost, "upstream-dns-host", "", "The upstream DNS host to be resolved.")
	flag.StringVar(&outPath, "out", "", "The output path for the rendered resolv.conf.")
	flag.Parse()

	ips, err := resolve(upstreamDNSHost)
	if err != nil {
		log.Fatal(err)
	}

	if err := render(ips, outPath); err != nil {
		log.Fatal(err)
	}
}

// resolve resolves the host and returns all IPs.
func resolve(host string) ([]net.IP, error) {
	var ips []net.IP
	for {
		var err error
		ips, err = net.LookupIP(host)
		if err != nil {
			dnsErr, ok := err.(*net.DNSError)
			// We keep trying to resolve indefinitely because the host is expected to
			// exist at some point.
			if ok && (dnsErr.Temporary() || dnsErr.IsNotFound) {
				time.Sleep(3 * time.Second)
				continue
			}
			return nil, fmt.Errorf("failed to resolve %q: %w", host, err)
		}
		break
	}
	return ips, nil
}

// render renders the output file with the provided nameservers. The output is
// in the same format as /etc/resolv.conf but only with nameserver
// instructions.
func render(nameservers []net.IP, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}
	defer f.Close()

	for _, nameserver := range nameservers {
		fmt.Fprintf(f, "nameserver %s\n", nameserver)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}

	return nil
}
