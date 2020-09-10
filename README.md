# KubeCF Apps DNS

This is the Apps DNS for KubeCF. It's a CoreDNS plugin that performs service
discovery using the Service Discovery Controller and responds with discovered
routes. It replaces the `bosh-dns-adapter`.

This repository also contains the Dockerfile for building a minimal CoreDNS
binary with only the required plugins for KubeCF and a tool for generating
`/etc/resolv.conf`-like files with only `nameserver` instructions to be used by
a `forward` CoreDNS plugin.

For examples on how to perform container-to-container (c2c) networking and
service discovering, please refer to [the documentation example](
https://github.com/cloudfoundry-attic/cf-networking-examples/blob/8e3be5faa135746cfd1e05bc618fe18053eab2b5/docs/c2c-with-service-discovery.md).
