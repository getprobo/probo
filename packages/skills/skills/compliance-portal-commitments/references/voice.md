# Voice: how commitments should read

Write copy that sounds like it was written by a senior infrastructure engineer, not a marketing team.
The page is a public trust page, read by a technical buyer skimming fast. They are skeptical of marketing,
but they are not an auditor reading your evidence line by line. Write short, plain cards that state one
security outcome each. For the target feel, read routine.co/security and supabase.com/security: a card is a
heading and a sentence, never a paragraph of policy.

Write in the company's own voice, speaking to the reader. Use the first person ("we") for what the company
does and the second person ("your data", "your workloads") for what the reader gets. This is the single
biggest lever against generic copy: "We encrypt your data at rest with AES-256" reads direct and specific,
where "Data is encrypted at rest with AES-256" reads like a policy clause. Prefer the active, direct form.
It stays factual; it is not marketing.

Two tests for every sentence:
- **If it could appear unchanged on a hundred SaaS security pages, rewrite it.**
- **If it describes a document or a procedure instead of what is true of the system, rewrite it.**

## Principles

- State facts, not aspirations. First person is for what is already true ("We enforce MFA"), never for
  intentions ("We are committed to MFA", "We aim to..."). If it is not in place today, leave it off.
- Describe the security property, not the policy or the procedure. Say what is true of the system, not that
  a document exists or what steps a process follows. Cut references to "a formal X Plan/Policy that
  defines...", approval chains, and ticket logging.
- One idea per card. Do not stack two or three facts into one description; split them or drop the weaker one.
- Use specific technologies and practices where the policy supports them: TLS, AES-256, SAST, MFA, SSO,
  signed commits, SSH, VPN, penetration testing. Naming the actual tool or provider (Snyk, the cloud
  provider's SOC 2, Stripe for PCI) is concrete and reads well, when the policy supports it.
- Avoid adjectives unless they are measurable. "Encrypted with AES-256" is measurable; "industry-leading
  encryption" is not.
- Body copy explains how something works, not why it is impressive.
- Calm and understated beats persuasive. Do not try to sound clever.
- Every claim must still be true and defensible, but write it for a reader skimming, not for an audit file.

## Hard rules

- No slogans, buzzwords, or inspirational language.
- No metaphors or analogies. ("No public door to production" is a metaphor; cut it.)
- No punchy two-sentence headlines.
- Headings under six words.
- Banned phrases and their kin: "built in, not bolted on", "best-in-class", "enterprise-grade", "we hunt
  for threats", "around the clock", "bank-grade", "military-grade", "peace of mind", "always".
- Do not overclaim with absolutes. Prefer a plain statement of the control over "there is no way for X".

## Titles

Titles are short, up to six words. A plain noun label ("Least privilege by default") or a short active
statement ("MFA is enforced", "Every change is reviewed") both work. Not punchy marketing headlines, not
two-sentence slogans. The description carries the detail.

- Prefer "MFA is enforced" over "SSO first. MFA always."
- Prefer "Data encrypted in transit and at rest" over "Encrypted at rest. Encrypted in transit. Always."
- Prefer "Every change is reviewed" over "Every commit signed. Every change reviewed."
- Prefer "Production access" over "No public door to production."

## Descriptions

One sentence is the default. A second only if it carries a genuinely separate fact. State the outcome and,
where the policy gives one, a specific mechanism or cadence, then stop. Keep a card to roughly 25 words.
One idea per card: if you are writing "X. Y. Z." with three separate controls, split them into separate
cards or drop the weakest. Read like a single line on a trust page a customer skims, not like a paragraph
lifted from the policy.

## Before / after

Three ways a card goes wrong: it reads like marketing, it reads like the policy it came from, or it is
written impersonally in the third person. The "Ship" line is short, direct, and in the company's own voice
(we/your).

**Encryption** (marketing)
- Before: "Customer data is protected with industry-leading encryption from storage to transit."
- Ship: "Data encrypted in transit and at rest" / "Your data is encrypted in transit with TLS and at rest with AES-256."

**Authentication** (impersonal)
- Before: "MFA is required for privileged accounts and critical systems."
- Ship: "MFA is enforced" / "We enforce MFA for privileged accounts and critical systems."

**Least privilege** (policy prose)
- Before: "Access is granted on a least-privilege basis. Requests are approved by the system owner and logged. Access rights are reviewed quarterly."
- Ship: "Least privilege by default" / "We grant the minimum access needed and review it quarterly."

**Secure development** (three facts in one card, so split it)
- Before: "Code is scanned with SAST and secret scanning. Container images are scanned before deployment. The production environment is penetration tested once a year."
- Ship, card 1: "We scan code and images" / "We scan code and container images before every deploy."
- Ship, card 2: "Pen-tested every year" / "We run an external penetration test once a year."

**Incident response** (describes a document, not the system)
- Before: "Critical and security-related issues are handled through a formal Incident Response Plan that defines roles, procedures, and escalation."
- Ship: "Incident response" / "We investigate security incidents on a defined escalation path."

**Monitoring** (policy prose)
- Before: "Application logs are retained for at least 30 days and aggregated to a central platform. Production outages trigger alerts to on-call engineers."
- Ship: "We monitor production" / "We centralize our logs and keep them 30 days. Outages page our on-call engineers."

## Group description

There is one group, titled "Security at <company name>". Keep its description to one short line in the same
direct voice, or leave it off. "How we protect your data and our platform" is fine; a line like "How access
to systems and production is granted and reviewed" reads like a policy table of contents, which is the tone
to avoid.

## A note on em dashes

Do not use em dashes (—) in the copy. Use periods, commas, colons, or parentheses instead. This is a
standing preference of the person this skill was built for.
