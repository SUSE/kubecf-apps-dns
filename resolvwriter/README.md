# resolvwriter

The `resolvwriter` is a tool to resolve a nameserver hostname and write an
`/etc/resolv.conf`-like file with only `nameserver` instructions for the
returned IPs. It's useful for generating a file to be used by the `forward`
plugin to use the Quarks DNS as an upstream server.

The program keeps trying to resolve the hostname until it succeeds.
