<!-- SPDX-License-Identifier: MIT -->
# XDNS — Privacy Policy

A DNS tunnel moves **names you look up**. Understand what observers see.

## What observers see (inherent to DNS tunneling)

- **Your resolvers** (ISP, public DNS, or configured DoT/DoH endpoints):
  every tunneled query — timing, sizes, frequency, and the outer QNAMEs.
  Payloads are encrypted, but traffic shape is visible.
- **Your VPS provider**: VM, server address, traffic sizes/timing,
  destinations of decapsulated traffic. They always can.
- **Your domain's DNS chain**: parent zones and certificate logs see the
  tunnel domain's existence and its certificates.

## What stays protected (when configured right)

- Query **contents** (AES-GCM/ChaCha20 per `internal/security`), SOCKS
  credentials (server-side auth), config files (keep `0600`).

## What this project collects

Nothing. No analytics, no telemetry, no phone-home in `cmd/` or
`internal/`. Server access logs (if you enable them) are yours — rotate
and protect them; they map clients to query times.

## Operator duties

- Run your own resolver chain where possible; prefer DoT/DoH upstreams.
- Assume the tunnel domain is fingerprinted eventually — rotate domains,
  follow `docs/THREAT_MODEL.md`.
- Never paste real server addresses, keys, or client configs in issues.

Questions: open an issue (no keys, no real addresses).
