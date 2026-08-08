# Security Policy

## Supported Versions

catmint actively supports the latest minor release line. Older versions may not receive security updates. Users are encouraged to use the latest available release.

| Version | Supported |
|---------|-----------|
| 1.3.x   | Yes       |
| 1.2.x   | No        |
| 1.1.x   | No        |
| 1.0.x   | No        |

## Reporting a Vulnerability

If you discover a security vulnerability or a critical bug in catmint, please report it privately.

- Preferred method: use [GitHub Security Advisories](https://github.com/ferizco/catmint-project/security/advisories) to create a confidential security report.

We will investigate and respond as quickly as possible. Please provide as much detail as you can.

Do not disclose security issues in public GitHub issues or pull requests.

## Responsible Disclosure

We appreciate responsible disclosure of security vulnerabilities. When a vulnerability is confirmed and fixed, we will credit the reporter unless requested otherwise.

## Security Best Practices

While catmint is designed for file integrity and verification, you are responsible for:

- Verifying hash outputs and files from trusted sources only.
- Regularly updating to the latest version.
- Running catmint in secure environments, especially when handling sensitive or critical data.

## Cryptography Notice

catmint uses standard cryptographic hash algorithms. Prefer SHA3-256, SHA512, SHA256, or Blake3 for critical applications. MD5 and SHA1 are supported for compatibility but are considered weak for security-sensitive verification.

## Questions

For other security-related questions, please use [GitHub Issues](https://github.com/ferizco/catmint-project/issues).
