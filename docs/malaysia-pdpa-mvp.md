# Malaysia PDPA MVP

## Purpose

This extension adapts Probo for Malaysian SMEs and managed service providers.
It keeps Probo's organization as the tenant-facing compliance boundary and adds
Malaysia-specific assessments, deadlines, evidence, and reporting without
renaming GDPR concepts or weakening the existing tenant isolation model.

The software supports compliance work. It does not make a legal determination
on behalf of a customer and must show the source, assessment date, assessor,
and recorded evidence for every automated conclusion.

## Regulatory baseline

The first release implements these Malaysia-specific rules:

1. A data controller or data processor must appoint a DPO when any of these
   conditions applies:
   - personal data processing exceeds 20,000 data subjects;
   - sensitive personal data, including financial information, exceeds 10,000
     data subjects; or
   - processing requires regular and systematic monitoring.
2. The DPO appointment is notified to the Commissioner within 21 days of the
   appointment.
3. A notifiable personal data breach is reported to the Commissioner as soon as
   practicable and no later than 72 hours. Where significant harm is likely,
   affected data subjects are notified no later than seven days after the
   Commissioner is notified.
4. Cross-border transfers record the destination, recipient, transfer basis,
   safeguards, and review evidence.
5. High-risk processing is routed through a DPIA workflow.

Primary references:

- DPO guideline: <https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2025/08/GP_DPO_ENG.pdf>
- DPO registration manual: <https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2025/07/Manual_Pengguna_Pendaftaran_DPO_EN.pdf>
- Data breach notification guideline: <https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2025/08/GP_DBN_ENG.pdf>
- Cross-border transfer guideline: <https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2025/08/GP_CBPDT_EN-1.pdf>
- DPIA guideline: <https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2026/04/Data-Protection-Impact-Assessment-Guideline-DPIA.pdf>

The regulatory content must be versioned. A future rule update must not silently
rewrite conclusions recorded under an earlier version.

## Tenant model

- A Probo organization represents one customer compliance boundary.
- Every Malaysia PDPA record carries `tenant_id` and `organization_id` where
  applicable.
- MSP users receive cross-customer views only through explicit MSP-level APIs;
  core organization queries remain tenant-scoped.
- Server-side authorization and database scoping are mandatory. Organization
  identifiers supplied by a browser are never sufficient authorization.

## Phase 1: organization profile and DPO assessment

Each organization has one Malaysia PDPA profile containing:

- estimated total number of data subjects;
- estimated number involving sensitive personal or financial data;
- whether regular and systematic monitoring is performed;
- server-calculated DPO requirement and triggering reasons;
- assessor and assessment timestamp;
- appointed DPO, appointment date, Commissioner notification date, and
  notification reference.

The threshold comparison is strictly `>` rather than `>=`. A total of exactly
20,000 data subjects or exactly 10,000 sensitive/financial data subjects does
not trigger the corresponding volume criterion by itself.

The server recalculates the result on every update. Clients cannot submit or
override `dpo_required` or the triggering reasons. Sensitive/financial data
subjects cannot exceed the total data-subject count. Referenced membership
profiles must belong to the same organization.

## Phase 2: data breach workflow

Each incident record contains discovery time, awareness time, impact assessment,
notification decision, evidence, and immutable status history. The product
calculates the 72-hour Commissioner deadline and, where applicable, the
seven-day data-subject deadline from recorded regulatory trigger times. It must
warn before deadlines but never submit a notification automatically without an
authorized human confirmation.

The server recommendation distinguishes the two official notification tests:

- significant scale means more than 1,000 affected data subjects and recommends
  notification to the Commissioner only;
- significant harm recommends notification to both the Commissioner and the
  affected data subjects.

Exactly 1,000 affected data subjects does not meet the significant-scale test.
When an initial Commissioner notification is recorded, the workflow also
calculates the 30-day phased-information deadline. A late Commissioner notice
requires both a recorded reason and supporting evidence. Status history is
append-only, and incident records are retained for the organization's breach
register; the guideline requires that register to be kept for at least two
years.

## Phase 3: transfer and DPIA localization

Extend the existing TIA, processing activity, vendor, and DPIA features rather
than creating duplicate Malaysia-only modules. Add the Malaysia transfer basis,
destination jurisdiction, recipient, safeguards, approval, and review fields.
Add a Malaysia high-risk screening questionnaire that can open an existing
DPIA.

## Phase 4: cybersecurity evidence

Integrate the existing OpenList security scanner as an isolated agent/service.
The service submits normalized findings and evidence through an authenticated
API. The core platform owns tenants, authorization, remediation tasks, and
audit history; the scanner never receives cross-tenant database access.

Initial checks include MFA, exposed services, TLS, token lifetime, backup
evidence, privileged accounts, and configuration weaknesses. Findings map to
Malaysia PDPA controls but retain the original scanner evidence and rule
version.

## Deferred from the first release

- subscriptions, invoicing, and payment collection;
- white-label domains and email branding;
- regulator portal submission automation;
- automated legal advice or a guaranteed-compliance score;
- a cross-tenant MSP dashboard before organization-level authorization tests
  are complete.

## Delivery order

1. DPO assessment domain rules and tests.
2. Organization-level PDPA profile persistence and service validation.
3. Authorized GraphQL query and mutation.
4. Console settings page with assessment explanation and evidence prompts.
5. Data breach workflow and deadline engine.
6. Transfer/DPIA localization.
7. Scanner ingestion API.
8. MSP views, billing, and white-label features.
