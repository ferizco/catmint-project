# Security Policy

## Supported Versions

catmint actively supports the latest minor release (e.g., v1.0.x). Older versions may not receive security updates. Users are encouraged to use the latest version.

| Version    | Supported         |
|------------|------------------|
| 1.0.x      | :white_check_mark: |


## Reporting a Vulnerability

If you discover a security vulnerability or a critical bug in catmint, please report it **privately**.

- **Preferred method:**  
  Use [GitHub Security Advisories](https://github.com/ferizco/catmint/security/advisories) to create a confidential security report.

We will investigate and respond as quickly as possible. Please provide as much detail as you can.

**Do not disclose security issues in public GitHub issues or pull requests.**

## Responsible Disclosure

We appreciate responsible disclosure of security vulnerabilities. When a vulnerability is confirmed and fixed, we will credit the reporter unless requested otherwise.

## Security Best Practices

While catmint is designed for file integrity and verification, you are responsible for:

- Verifying hash outputs and files from trusted sources only.
- Regularly updating to the latest version.
- Running catmint in secure environments, especially when handling sensitive or critical data.

## Cryptography Notice

catmint uses standard cryptographic hash algorithms. Always use the strongest supported algorithm (e.g., SHA3-256 or SHA512) for critical applications. MD5 and SHA1 are supported for compatibility but are considered weak for security purposes.

## Questions

For any other security-related questions, please use [GitHub Issues](https://github.com/ferizco/catmint-project/issues) 
