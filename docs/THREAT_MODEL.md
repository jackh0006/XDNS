# XDNS threat model and operating limits

XDNS is a DNS tunnel for cases where a recursive resolver can reach a
delegated authoritative server but conventional proxy traffic is blocked. It
is not an anonymity system and it cannot promise universal DPI bypass,
unblockability, or a fixed speed.

## What the protocol protects

With an authenticated encryption method (AES-GCM), XDNS protects the tunnel
payload against modification and disclosure by parties that do not possess the
key. The client also has loss recovery, resolver/path selection, MTU probing,
and optional FEC to improve delivery on lossy links.

The DNS carrier still exposes metadata. A recursive resolver and network
observer may learn query timing, packet sizes, the tunnel domain, source IP,
and often the fact that DNS is being used unusually. The server can observe
traffic after it exits the tunnel. HTTPS or another end-to-end application
security layer remains necessary for destination confidentiality.

## What can block or degrade XDNS

An adversary may block or poison the delegated domain, block TCP/UDP port 53,
rate-limit or alter DNS replies, fingerprint query patterns, actively probe a
server, block DoT/DoH, or exhaust server resources. Encryption does not hide
all traffic analysis signals. Transport diversity and QNAME shaping can reduce
some brittle fingerprints, but are not guarantees.

## Practical choices

| Goal | Recommended choice | Trade-off |
| --- | --- | --- |
| New deployment security | AES-256-GCM and a high-entropy private key | Legacy clients that only support older methods cannot connect. |
| Fast path where ordinary DNS works | `RESOLVER_TRANSPORT = "auto"` with the `speed` preset | UDP can be filtered or poisoned by some networks. |
| Continuity during loss or resolver failure | `survival` preset and multiple independently configured tunnel domains | More redundancy can reduce effective throughput. |
| Networks that fingerprint plain DNS | Test DoT or DoH explicitly, one network at a time | These transports can themselves be blocked or stand out. |
| Reduce correlated failures | Configure independent resolvers/domains and retain default domain-diverse duplication | More endpoints increase operational complexity and metadata exposure. |

## Evidence and boundaries

DNS wire behavior follows [RFC 1035](https://www.rfc-editor.org/rfc/rfc1035),
DNS-over-TLS follows [RFC 7858](https://www.rfc-editor.org/rfc/rfc7858), and
DNS-over-HTTPS follows [RFC 8484](https://www.rfc-editor.org/rfc/rfc8484).
Those standards describe transports, not censorship-bypass guarantees. Measure
your own network with the included benchmark and test against a controlled
server before relying on any profile during a disruption.
