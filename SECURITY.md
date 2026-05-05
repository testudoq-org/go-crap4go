# Security Policy

## Supported versions

Only the latest release is actively maintained. Security fixes are not
back-ported to older versions.

| Version | Supported |
|---------|-----------|
| latest  | ✅        |
| < latest | ❌       |

## Reporting a vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report privately using GitHub's built-in private advisory mechanism:

👉 https://github.com/testudoq-org/go-crap4go/security/advisories/new

We aim to:
- Acknowledge your report within **5 business days**
- Confirm or decline the vulnerability within **15 business days**
- Release a patch within **30 days** of confirmation for valid vulnerabilities

Once the vulnerability is patched and a release is published, a public GitHub
Security Advisory will be created to credit the reporter (unless anonymity is
requested).

## Scope

`crap4go` is a local CLI tool — it reads Go source and coverage profile files
from disk and writes to stdout. It makes no outbound network connections during
normal use.

In-scope:
- Vulnerabilities in direct or transitive dependencies that affect users of the
  published binary
- Path traversal or unexpected file access via `--coverprofile` / `--config`
  flags

Out of scope:
- Vulnerabilities only exploitable by the machine owner supplying malicious
  input files (local privilege escalation is not a realistic threat model here)
