# Security Policy

## Supported Versions

Only the latest version receives security updates.

| Version | Supported |
| ------- | --------- |
| Latest  | ✅        |
| Older   | ❌        |

## Reporting a Vulnerability

If you believe you have found a security vulnerability in Probo,
please report it responsibly by emailing
[security@getprobo.com](mailto:security@getprobo.com).

**Please do NOT create public GitHub issues for security vulnerabilities.**

### What to Include in Your Report
- A clear description of the vulnerability
- Steps to reproduce the issue
- Affected version(s)
- Potential impact of the vulnerability
- Any suggested fix (optional but appreciated)

## Scope

### In Scope
- `getprobo.com` and all subdomains
- Probo open source codebase (this repository)
- Authentication & authorization issues
- Data exposure vulnerabilities
- API security issues
- Injection vulnerabilities (SQLi, XSS, CSRF, etc.)

### Out of Scope
- Denial of Service (DoS/DDoS) attacks
- Social engineering attacks
- Physical security attacks
- Vulnerabilities in third-party services
- Issues already known or previously reported
- Automated scanner reports without proof of exploitability

## Response Process

| Timeline | Action |
| -------- | ------ |
| 48 hours | Acknowledgement of your report |
| 5 days   | Initial assessment and severity rating |
| 30 days  | Target resolution for critical/high issues |
| 90 days  | Target resolution for medium/low issues |

We follow responsible disclosure — once a fix is released,
we will notify you and you are free to publish your findings.

## Severity Ratings

We use severity ratings aligned with
**ISO/IEC 27001:2022** and **CVSS v3.1**:

| Severity | CVSS Score | Description |
| -------- | ---------- | ----------- |
| 🔴 Critical | 9.0 – 10.0 | RCE, authentication bypass, direct data breach |
| 🟠 High | 7.0 – 8.9 | Privilege escalation, significant data exposure |
| 🟡 Medium | 4.0 – 6.9 | Limited data exposure, CSRF, open redirects |
| 🟢 Low | 0.1 – 3.9 | Minor issues, information disclosure |
| ℹ️ Info | 0.0 | Best practice improvements |

## Our Commitment to Researchers

- We will not take legal action against researchers
  who follow responsible disclosure guidelines
- We will keep your report confidential until resolved
- We will credit you for your finding (if you wish)
- We will work collaboratively to understand and fix the issue

## Hall of Fame

We appreciate security researchers who help keep Probo secure.
Responsible disclosures will be acknowledged here. 🙏

*Be the first to be listed here!*

---

*Last updated: March 2026*
*Aligned with ISO/IEC 27001:2022 Information Security Standards*
