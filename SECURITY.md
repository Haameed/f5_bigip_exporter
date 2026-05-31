# Security Policy

## Supported Versions

The latest released version receives security updates.

## Reporting a Vulnerability

**Please do not open public issues for security vulnerabilities.**

Instead, report them privately via
[GitHub Security Advisories](https://github.com/Haameed/f5_bigip_exporter/security/advisories/new)
or by contacting the maintainer directly.

We will acknowledge your report within a reasonable timeframe and keep you
updated on the remediation progress.

## Security Best Practices

When deploying this exporter:

- **Protect the config file** — it contains plaintext credentials.
  Use `chmod 600` and/or mount it as a read-only secret.
- **Use a least-privilege BIG-IP account** — a read-only / auditor role is enough.
- **Keep TLS verification enabled** in production (avoid `-insecure`); trust the
  BIG-IP CA instead.
- **Restrict network exposure** — keep the exporter on a trusted management network.
- **Avoid logging credentials** — they are never placed in URLs by design.
