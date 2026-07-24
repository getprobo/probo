---
name: compliance-portal-commitments
description: >-
  Create or update the public commitments (commitment groups and their commitments) shown on a Probo
  compliance portal, grounded strictly in the organization's own published Probo
  policies and written in a factual, understated engineering voice. Use this skill whenever the user
  wants to add, write, draft, edit, rewrite, curate, trim, or publish compliance portal commitments,
  "security commitments", or a compliance portal security section for a company on Probo; asks what
  commitments to show for a company doing SOC 2 or ISO 27001; gives an organization name or ID and asks
  to build out its compliance portal commitments; or wants existing commitments reworded to sound less like
  marketing. Always pull the real published policies first, surface only the specific and differentiating
  controls a skeptical security reviewer would find useful, and create or update them through the Probo
  MCP. Never invent controls that are not in the published policies.
---

# Probo compliance portal commitments

A Probo compliance portal can display **commitments**: short, public statements about the security controls a
company actually operates. They live in **commitment groups**. Put all of a company's commitments under a
**single group titled "Security at <company name>"** (for example "Security at Captain") with a one-line
description; each commitment is one card in that group. Do not split commitments across multiple groups.

Each commitment has four fields:

- **icon** — one value from a fixed set (see `references/portal-mechanics.md`)
- **eyebrow** — a short category label above the title (e.g. "Encryption", "Authentication")
- **title** — a short heading, under six words
- **description** — one plain sentence (two only if the second adds a genuinely separate fact) describing
  what is true of the system, in the company's own voice ("We encrypt your data..."). One idea per card.
  Describe the security property, not the policy or the procedure behind it. Do not stack facts, and do not
  reference documents ("a formal X Plan that defines...") or internal ceremony (approval chains, ticket
  logging).

The reader you are writing for is a **technical buyer skimming your public trust page**, not an auditor
cross-checking evidence. They are skeptical of marketing but they are reading fast. Every claim must still
be true and verifiable, but write it plainly and in the company's own voice, speaking to the reader: first
person for what the company does ("We enforce MFA"), second person for what the reader gets ("your data").
Aim for the way the security pages at routine.co/security and supabase.com/security read: short cards, one
outcome each, no policy prose. That register drives the whole workflow below.

Work in four stages: **get context → draft → filter → publish**. Do not skip straight to publishing.

---

## 1. Get context (read the real policies first)

Commitments must trace to controls the company genuinely has. The source of truth is the organization's
**published policies** in Probo, not general knowledge about SOC 2 or ISO 27001.

1. **Find the Probo MCP and the organization.** This environment may expose more than one Probo MCP
   server (for example a US and an EU instance). Call `listOrganizations` on each until you find the one
   that returns the target company, and use that server for every later call. Match the organization the
   user named and capture its `id`.
2. **List the published policies.** Call `listDocuments` with `document_types: ["POLICY"]`. Policies with
   a `current_published_major` are published.
3. **Read the actual content.** `getDocument` returns metadata only. To get the text, call
   `listDocumentVersions` for each policy with `filter: {statuses: ["PUBLISHED"]}`,
   `order_by: {field: "CREATED_AT", direction: "DESC"}`, `size: 1`. The returned version includes the
   `title` and the full `content`. Fetch the policies in parallel.

Read the substance, not just the titles. The specific, quotable facts live inside the statements: exact
algorithms (AES-256), protocols (TLS, SSH, VPN), tools (SAST, secret scanning), cadences (quarterly
access reviews, annual penetration test, daily backups retained 30 days), and mechanisms (signed commits,
protected main branch, SSO, MFA). These specifics are what make a commitment credible.

Do not pull details from the web and do not assume a control exists because the framework expects it. If a
policy says "TLS", write "TLS", not "TLS 1.3". If it says data is retained per contract, do not invent a
fixed retention window.

---

## 2. Draft candidate commitments

All commitments go under one group titled "Security at <company name>", so there is no theming decision to
make. Draft a single ordered list of the strongest cards, and order them so related ones sit next to each
other (for example encryption and backups, then access, then development, then operations). Do not create
multiple groups. Give the group one light line of description, or leave it off; keep it in the same direct
voice, not a policy heading.

For each candidate commitment, write down which policy statement backs it. If you cannot point to a
sentence in a published policy, drop the commitment. This is the grounding check and it is not optional.

Then apply the voice in `references/voice.md`. The tone is calm, factual, and technical, the way internal
engineering documentation reads. Load that file before writing any titles or descriptions; the difference
between a good and a bad commitment here is almost entirely tone.

---

## 3. Filter: what earns a place

The instinct is to publish everything. Resist it. A compliance portal that lists ten generic commitments is
weaker than one that lists five specific ones, because the generic entries signal "marketing" and make the
reader trust the whole page less.

**Aim for roughly 6 to 8 short cards. Hard cap: at most 10 commitments in the group.** The
reference pages run 7 to 9 flat cards; that is the target feel. These caps are ceilings, not targets, and
fewer is usually better. If you have more strong candidates than fit, prioritize by importance to a
technical buyer and drop or fold the rest, then tell the user what you left out and offer to swap.

**Keep a commitment when it is:**

- **Specific** — names a real technology, cadence, or mechanism (AES-256, quarterly reviews, signed commits).
- **Differentiating** — not every SaaS company does it, or does it this concretely.
- **Verifiable** — an auditor could confirm it from evidence.

**Cut a commitment when it is:**

- **Generic** — a sentence that could appear unchanged on a hundred SaaS security pages.
- **Table-stakes with nothing specific to add** — the fact is expected and you have no concrete detail
  that makes it interesting.
- **Thin or off-audience** — internal-culture or legal items that a vendor-security reviewer would skip
  (code of conduct, office badge procedures, cookie policy).
- **Already shown elsewhere on the portal** — the compliance portal separately displays certifications and
  frameworks (SOC 2, ISO 27001) and published documents. Do not add a "Certifications" commitment or
  otherwise restate a badge, framework, or document the page already surfaces; it is redundant. Spend the
  card on a control the portal does not already show.

There is no fixed list of "always cut" topics. The same topic can be worth keeping for one company and not
another. For example, disaster recovery is worth publishing if the company rehearses failovers and can
state a concrete objective; it is worth cutting if the entry would just say "we have a DR plan and test it
yearly", which every vendor claims. Judge each candidate against the three keep-criteria above.

Before publishing, show the user the proposed commitments and say briefly what you cut and why.
Let them adjust. They know which controls they want to lead with.

---

## 4. Create or update in the compliance portal

Once the user has agreed on the set, write it through the Probo MCP. The exact tool sequence, the icon
enum, ordering/rank behavior, and the common pitfalls (read/write scope errors, reusing existing empty
groups instead of duplicating them, deleting a whole group vs individual commitments) are in
`references/portal-mechanics.md`. Read it before making any write calls.

Key habits:

- Publishing to a compliance portal is **public-facing**. Confirm the final copy with the user before writing,
  and treat creates, updates, and deletes as changes that change what visitors see.
- Create the single "Security at <company name>" group first (or reuse it if it already exists), capture
  its returned `id`, then attach every commitment to that id.
- Never create a second group. If the portal already has other groups from a previous run, fold their
  commitments into the one group and remove the extras (confirm with the user before deleting).
- After writing, give the user a compact recap of the live state (groups and their commitments).

---

## Reference files

- `references/voice.md` — the tone-of-voice rules and before/after examples. Load this before writing any
  commitment copy.
- `references/portal-mechanics.md` — the Probo MCP tool sequence, icon enum, and gotchas. Load this before
  any create/update/delete call.
