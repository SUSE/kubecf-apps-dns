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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

const pluginName = "svcdiscovery"
const sdcEndpointBase = "/v1/registration/"

// Register the plugin.
func init() {
	plugin.Register(pluginName, func(c *caddy.Controller) error {
		// Ignore svcdiscovery token.
		c.Next()

		var sdcClient *SDCClient
		var tlsCAPath string
		var tlsClientCertPath string
		var tlsClientKeyPath string
		var sdcHost string
		var sdcPort uint16
		var ttl uint32
		for c.NextBlock() {
			switch c.Val() {
			case "tls_ca_path":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return plugin.Error(pluginName, c.ArgErr())
				}
				tlsCAPath = args[0]
			case "tls_client_cert_path":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return plugin.Error(pluginName, c.ArgErr())
				}
				tlsClientCertPath = args[0]
			case "tls_client_key_path":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return plugin.Error(pluginName, c.ArgErr())
				}
				tlsClientKeyPath = args[0]
			case "sdc_host":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return plugin.Error(pluginName, c.ArgErr())
				}
				sdcHost = args[0]
			case "sdc_port":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return plugin.Error(pluginName, c.ArgErr())
				}
				u, err := strconv.ParseUint(args[0], 10, 16)
				if err != nil {
					return plugin.Error(pluginName, c.Errf("failed to convert sdc_port: %w", err))
				}
				sdcPort = uint16(u)
			case "ttl":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return plugin.Error(pluginName, c.ArgErr())
				}
				u, err := strconv.ParseUint(args[0], 10, 32)
				if err != nil {
					return plugin.Error(pluginName, c.Errf("failed to convert TTL: %w", err))
				}
				ttl = uint32(u)
			}
		}

		httpClient, err := newHTTPClient(tlsCAPath, tlsClientCertPath, tlsClientKeyPath)
		if err != nil {
			return plugin.Error(pluginName, c.Errf("failed to construct new HTTP client: %w", err))
		}

		sdcURLBase := (&url.URL{
			Scheme: "https",
			Host:   fmt.Sprintf("%s:%d", sdcHost, sdcPort),
			Path:   sdcEndpointBase,
		}).String()

		sdcClient = &SDCClient{
			httpClient: httpClient,
			sdcURLBase: sdcURLBase,
		}

		dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
			return &ServiceDiscovery{
				Next:      next,
				log:       clog.NewWithPlugin(pluginName),
				sdcClient: sdcClient,
				ttl:       ttl,
			}
		})

		return nil
	})
}

func newHTTPClient(tlsCAPath, tlsClientCertPath, tlsClientKeyPath string) (*http.Client, error) {
	caCert, err := ioutil.ReadFile(tlsCAPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(tlsClientCertPath, tlsClientKeyPath)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout:   time.Second * 2,
		KeepAlive: time.Minute * 10,
	}
	transport := &http.Transport{
		Dial:                dialer.Dial,
		MaxIdleConnsPerHost: 1024,
		TLSHandshakeTimeout: time.Second,
		TLSClientConfig: &tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{cert},
		},
	}
	client := &http.Client{Transport: transport}

	return client, nil
}
