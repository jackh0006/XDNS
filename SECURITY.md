# Security policy

XDNS is designed for hostile and unreliable networks. Treat it as security-
sensitive software: a deployment can expose its domain, server address,
resolver choices, traffic timing, and server resources even when payloads are
encrypted.

## Supported versions

Only the latest release on `main` is supported. Upgrade promptly when a
security release is published.

## Reporting a vulnerability

Do **not** put exploit details, private keys, real server addresses, or client
configuration in a public issue. Use the repository's **Security** tab and
choose **Report a vulnerability** when private reporting is enabled. If that
option is unavailable, contact the maintainer through the GitHub profile and
ask for a private channel before sending technical details.

Include the affected version and platform, a minimal reproduction, impact,
and whether the issue can reveal traffic, bypass authentication, crash a
server, or exhaust resources. Allow reasonable time for a fix before public
disclosure.

## Deployment baseline

- Use an authenticated cipher (AES-GCM); do not use legacy XOR or ChaCha20
  compatibility modes for a new deployment.
- Keep `encrypt_key.txt` private, back it up securely, and rotate it after a
  suspected compromise. Never commit it.
- Restrict the server's egress when possible, keep the host patched, and use
  the built-in connection/admission limits for public endpoints.
- Review the installer before piping it to a shell, pin a release for
  production changes, and verify release checksums.

See the [threat model](docs/THREAT_MODEL.md) for what XDNS does and does not
protect.
