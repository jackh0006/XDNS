# Contributing to XDNS

Thanks for improving XDNS. Changes must preserve compatibility with existing
MasterDNS-family clients unless the pull request clearly documents an opt-in
protocol version or a migration path.

## Before opening a pull request

1. Keep changes small and explain the network trade-off: payload overhead,
   latency, loss recovery, CPU, memory, and censorship fingerprinting.
2. Add tests for protocol, parsing, configuration, or failure behaviour that
   changes. Do not make unmeasured claims about speed or DPI resistance.
3. Run `go fmt ./...`, `go vet ./...`, and `go test ./...`.
4. For installer edits, run `bash -n server_linux_install.sh
   client_linux_quick_install.sh client_linux_install.sh` and test the
   documented upgrade/uninstall path where safe.

## Design principles

- A DNS tunnel has limited bandwidth; avoid adding bytes or round trips without
  an evidence-backed benefit.
- Favor bounded queues, explicit timeouts, authentication, and graceful
  fallback under loss and hostile input.
- A transport that works in one network is not proof it works in another.
  Make optional behavior configurable and document failure modes.
- Never add secret keys, live domains, VPS addresses, or captured user traffic
  to the repository, tests, or issues.

## Pull-request expectations

Describe compatibility, defaults, security implications, and how you tested
the change. Maintainers may ask for benchmarks or packet captures with all
sensitive data removed.
