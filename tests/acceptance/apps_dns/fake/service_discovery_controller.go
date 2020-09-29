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

package fake

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

// ServiceDiscoveryController is a fake service-discovery-controller for
// performing assertions on the acceptance tests. It's useful for faking the
// responses from the service-discovery-controller in a controllable way.
type ServiceDiscoveryController struct {
	caPath     string
	certPath   string
	keyPath    string
	listenAddr string
	handlers   map[string]func(http.ResponseWriter, *http.Request)
}

// NewServiceDiscoveryController creates a new ServiceDiscoveryController.
func NewServiceDiscoveryController(
	caPath string,
	certPath string,
	keyPath string,
	listenAddr string,
) *ServiceDiscoveryController {
	return &ServiceDiscoveryController{
		caPath:     caPath,
		certPath:   certPath,
		keyPath:    keyPath,
		listenAddr: listenAddr,
		handlers:   make(map[string]func(http.ResponseWriter, *http.Request)),
	}
}

// Serve configures an mTLS server with a given handler for performing
// assertions.
func (sdc *ServiceDiscoveryController) Serve() (chan<- struct{}, <-chan error) {
	chShutdown := make(chan struct{}, 1)
	chErr := make(chan error, 1)

	router := mux.NewRouter()

	router.HandleFunc("/v1/registration/{domain}", func(w http.ResponseWriter, r *http.Request) {
		if f, ok := w.(http.Flusher); ok {
			defer f.Flush()
		}
		vars := mux.Vars(r)
		handler, ok := sdc.handlers[vars["domain"]]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		handler(w, r)
	})

	caCert, err := ioutil.ReadFile(sdc.caPath)
	if err != nil {
		chErr <- fmt.Errorf("failed to serve: %w", err)
		return chShutdown, chErr
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		TLSConfig: tlsConfig,
		Handler:   router,
	}

	listener, err := net.Listen("tcp", sdc.listenAddr)
	if err != nil {
		chErr <- fmt.Errorf("failed to serve: %w", err)
		return chShutdown, chErr
	}

	go func() {
		<-chShutdown
		server.Shutdown(context.Background())
	}()

	go func() {
		defer close(chErr)

		if err := server.ServeTLS(listener, sdc.certPath, sdc.keyPath); err != nil && err != http.ErrServerClosed {
			chErr <- fmt.Errorf("failed to serve: %w", err)
		}
		chErr <- nil
	}()

	return chShutdown, chErr
}

// Handle adds a handler that matches the provided domain.
func (sdc *ServiceDiscoveryController) Handle(
	domain string,
	handler func(w http.ResponseWriter, r *http.Request),
) {
	sdc.handlers[domain] = handler
}

// Handler returns a helper handler for responding fake services.
func Handler(ips []net.IP) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		hosts := make([]ServiceDiscoveryControllerHost, len(ips))
		for i, ip := range ips {
			hosts[i] = ServiceDiscoveryControllerHost{IPAddress: ip.String()}
		}
		registration := ServiceDiscoveryControllerRegistration{Hosts: hosts}

		encoder := json.NewEncoder(w)
		encoder.Encode(registration)
	}
}

// ServiceDiscoveryControllerRegistration mimics the original unexported
// registration struct:
// https://github.com/cloudfoundry/cf-networking-release/blob/7fe3693f06aabe554620bc41e33e5da7dd040ba8/src/service-discovery-controller/routes/server.go#L43
type ServiceDiscoveryControllerRegistration struct {
	Hosts   []ServiceDiscoveryControllerHost `json:"hosts"`
	Env     string                           `json:"env"`
	Service string                           `json:"service"`
}

// ServiceDiscoveryControllerHost mimics the original unexported host struct:
// https://github.com/cloudfoundry/cf-networking-release/blob/7fe3693f06aabe554620bc41e33e5da7dd040ba8/src/service-discovery-controller/routes/server.go#L33
type ServiceDiscoveryControllerHost struct {
	IPAddress       string                 `json:"ip_address"`
	LastCheckIn     string                 `json:"last_check_in"`
	Port            uint16                 `json:"port"`
	Revision        string                 `json:"revision"`
	Service         string                 `json:"service"`
	ServiceRepoName string                 `json:"service_repo_name"`
	Tags            map[string]interface{} `json:"tags"`
}
