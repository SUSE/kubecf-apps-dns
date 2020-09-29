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
	"os"
	"testing"

	"github.com/kubernetes-sigs/minibroker/pkg/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/SUSE/kubecf-apps-dns/tests/acceptance/apps_dns/fake"
)

var (
	sdc           *fake.ServiceDiscoveryController
	dnsAddr       string
	serveShutdown chan<- struct{}
	serveErr      <-chan error
	clusterDomain string
)

const ttl = 300

func TestAppsDns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppsDns Suite")
}

var _ = BeforeSuite(func() {
	tlsCaPath := os.Getenv("TLS_CA_PATH")
	tlsCertPath := os.Getenv("TLS_CERT_PATH")
	tlsKeyPath := os.Getenv("TLS_KEY_PATH")
	listenAddr := os.Getenv("LISTEN_ADDRESS")
	sdc = fake.NewServiceDiscoveryController(
		tlsCaPath,
		tlsCertPath,
		tlsKeyPath,
		listenAddr,
	)

	dnsAddr = os.Getenv("DNS_ADDR")

	resolvConf, err := os.Open("/etc/resolv.conf")
	Expect(err).ToNot(HaveOccurred())
	clusterDomain, err = kubernetes.ClusterDomain(resolvConf)
	Expect(err).ToNot(HaveOccurred())

	serveShutdown, serveErr = sdc.Serve()
})

var _ = AfterSuite(func() {
	serveShutdown <- struct{}{}
	err := <-serveErr
	Expect(err).ToNot(HaveOccurred())
})
