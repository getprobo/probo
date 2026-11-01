# BlackShield Security Audit Report — `getprobo/probo`

- **Scanned Commit:** `b45ea069e54b8c99e3fd824af81121368840caf3`
- **Actionable Findings:** `4` (`3 Critical`, `1 High`, `5 Medium`, `1 Low`)
- **BlackShield Platform & Continuous Zero-Cost PR Protection:** [https://blackshield.chaplau.com](https://blackshield.chaplau.com)

## Detailed Findings & Remediation Guide

### 1. [CRITICAL] `ec-private-key` — Elliptic Curve (EC) Private Key Detected

- **Category:** `secret`
- **Location:** `.env.example:313`
- **Details:** Potential hardcoded credential or secret detected in `.env.example:313` (rule `ec-private-key`). Raw value redacted (`***REDACTED***`) for safe disclosure.
- **Recommended Fix:** Revoke leaked credential immediately and rotate in secret manager.

### 2. [CRITICAL] `private-key` — Private Cryptographic Key Detected

- **Category:** `secret`
- **Location:** `.env.example:285`
- **Details:** Potential hardcoded credential or secret detected in `.env.example:285` (rule `private-key`). Raw value redacted (`***REDACTED***`) for safe disclosure.
- **Recommended Fix:** Revoke leaked credential immediately and rotate in secret manager.

### 3. [CRITICAL] `private-key` — Private Cryptographic Key Detected

- **Category:** `secret`
- **Location:** `.env.example:293`
- **Details:** Potential hardcoded credential or secret detected in `.env.example:293` (rule `private-key`). Raw value redacted (`***REDACTED***`) for safe disclosure.
- **Recommended Fix:** Revoke leaked credential immediately and rotate in secret manager.

### 4. [HIGH] `CVE-2022-25883` — CVE-2022-25883 in semver@6.3.1

- **Category:** `vulnerability`
- **Location:** `package-lock.json`
- **Affected Package:** `semver@6.3.1`
- **Details:** Regular Expression Denial of Service (ReDoS) in semver range parser.
- **Recommended Fix:** Upgrade `semver` to `7.5.2` or newer

