-- Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

ALTER TABLE vendors
  ADD COLUMN business_associate_agreement_url TEXT,
  ADD COLUMN subprocessors_list_url TEXT;

UPDATE vendors SET
  business_associate_agreement_url = CASE name
    WHEN 'Notion' THEN 'https://www.notion.so/notion/Business-Associate-Agreement-909d9f4ccca041b1a23d0fe6e56fa111'
    WHEN 'OpenAI' THEN 'https://openai.com/policies/business-terms/'
    WHEN 'Amazon Web Services (AWS)' THEN 'https://aws.amazon.com/compliance/hipaa-compliance/'
    WHEN 'Google Cloud Platform' THEN 'https://cloud.google.com/terms/hipaa-baa'
    WHEN 'Microsoft Azure' THEN 'https://learn.microsoft.com/en-us/compliance/regulatory/offering-hipaa-hitech'
    WHEN 'Cal.com' THEN 'https://app.cal.com/apps/baa-for-hipaa'
    WHEN 'Fly.io' THEN 'https://fly.io/documents'
    WHEN 'New Relic' THEN 'https://newrelic.com/termsandconditions/hipaabaafaq'
    WHEN 'Gusto' THEN 'https://support.gusto.com/article/152166121100000/The-Health-Insurance-Portability-and-Accountability-Act-of-1996-HIPAA'
    WHEN 'Microsoft 365' THEN 'https://learn.microsoft.com/en-us/compliance/regulatory/offering-hipaa-hitech'
    WHEN 'Monday.com' THEN 'https://monday.com/l/privacy/hipaa-baa'
    WHEN 'Atlassian' THEN 'https://www.atlassian.com/legal/business-associate-agreement'
    WHEN 'Resend' THEN 'https://newrelic.com/termsandconditions/hipaabaafaq'
    END,
  subprocessors_list_url = CASE name
    WHEN 'Github' THEN 'https://docs.github.com/en/site-policy/privacy-policies/github-subprocessors'
    WHEN 'Notion' THEN 'https://www.notion.so/notion/Notion-s-List-of-Subprocessors-268fa5bcfa0f46b6bc29436b21676734'
    WHEN 'Mintlify' THEN 'https://security.mintlify.com'
    WHEN 'Segment' THEN 'https://www.twilio.com/en-us/legal/sub-processors'
    WHEN 'Heroku' THEN 'https://www.salesforce.com/en-us/wp-content/uploads/sites/4/documents/legal/misc/salesforce-infrastructure-and-subprocessors.pdf'
    WHEN 'Linear' THEN 'https://linear.app/dpa'
    WHEN 'Hubspot' THEN 'https://legal.hubspot.com/sub-processors-page'
    WHEN 'Google Workspace' THEN 'https://workspace.google.com/terms/subprocessors/'
    WHEN 'Algolia' THEN 'https://www.algolia.com/policies/infrastructure-and-sub-processors'
    WHEN 'PostHog' THEN 'https://posthog.com/dpa'
    WHEN 'Airtable' THEN 'https://www.airtable.com/company/subprocessors'
    WHEN 'OpenAI' THEN 'https://openai.com/policies/sub-processor-list/'
    WHEN 'Anthropic' THEN 'https://trust.anthropic.com/subprocessors'
    WHEN 'Amazon Web Services (AWS)' THEN 'https://aws.amazon.com/compliance/sub-processors/'
    WHEN 'Calendly' THEN 'https://help.calendly.com/hc/en-us/articles/360047345493-Calendly-sub-processors-GDPR-CCPA'
    WHEN 'jsDelivr' THEN 'https://www.jsdelivr.com/terms/sub-processors'
    WHEN 'Clerk' THEN 'https://clerk.com/legal/subprocessors'
    WHEN 'Plausible Analytics' THEN 'https://plausible.io/privacy'
    WHEN 'Docker' THEN 'https://www.docker.com/trust/privacy/sub-processors/'
    WHEN 'Brex' THEN 'https://trust-portal.brex.com'
    WHEN 'Ramp' THEN 'https://trust.ramp.com/'
    WHEN 'ClickHouse' THEN 'https://clickhouse.com/legal/agreements/subprocessors'
    WHEN 'Tailscale' THEN 'https://tailscale.com/dpa-subprocessors'
    WHEN 'Probo' THEN 'https://www.getprobo.com/subprocessors'
    WHEN 'folk' THEN 'https://www.folk.app/privacy-policy#3-list-of-sub-processors'
    WHEN 'Claap' THEN 'https://www.claap.io/legal/privacy-policy#sub-processors'
    WHEN 'Pitch' THEN 'https://pitch.com/dpa#annex-7-2'
    WHEN 'Spendesk' THEN 'https://www.spendesk.com/legals/subprocessors/'
    WHEN 'Lever' THEN 'https://www.lever.co/subprocessors-archive-june-2023/'
    WHEN 'Google Cloud Platform' THEN 'https://cloud.google.com/terms/subprocessors'
    WHEN 'Microsoft Azure' THEN 'https://www.microsoft.com/en-us/download/details.aspx?id=50426'
    WHEN 'Fly.io' THEN 'https://fly.io/legal/sub-processors/'
    WHEN 'Upstash' THEN 'https://trust.upstash.com/subprocessors'
    WHEN 'bunny.net' THEN 'https://bunny.net/gdpr/sub-processors/'
    WHEN 'ClickUp' THEN 'https://clickup.com/terms/dpa/subprocessors'
    WHEN 'New Relic' THEN 'https://newrelic.com/sub-processors'
    WHEN 'Pylon' THEN 'https://usepylon.com/subprocessors'
    WHEN 'Loops' THEN 'https://loops.so/subprocessors'
    WHEN 'Perplexity' THEN 'https://trust.perplexity.ai/'
    WHEN 'Mailchimp' THEN 'https://mailchimp.com/legal/subprocessors/'
    WHEN 'Otter.ai' THEN 'https://otter.ai/subprocessors'
    WHEN 'Resend' THEN 'https://newrelic.com/sub-processors'
    WHEN 'OVHcloud global' THEN 'https://www.ovh.ie/support/termsofservice/OVH%20IE%20Sub%20processors.pdf'
    WHEN 'OVHcloud US' THEN 'https://storage.gra.cloud.ovh.net/v1/AUTH_325716a587c64897acbef9a4a4726e38/contracts/30c4ab8-OVH_Sub_processors-IE-2.0.pdf'
    WHEN 'Pipedrive' THEN 'https://www.pipedrive.com/en/subprocessors'
    WHEN 'Plesk' THEN 'https://www.plesk.com/wp-content/uploads/legal/gdpr3.pdf'
    WHEN 'Brevo' THEN 'https://www.brevo.com/legal/termsofuse/#sub-processors'
    WHEN 'Better Stack' THEN 'https://betterstack.com/dpa/schedules'
    WHEN 'PlanetScale' THEN 'https://planetscale.com/legal/subprocessors'
    WHEN 'Microsoft 365' THEN 'https://www.microsoft.com/en-us/download/details.aspx?id=50426'
    WHEN 'Bitwarden' THEN 'https://bitwarden.com/help/subprocessors/'
    WHEN 'Framer' THEN 'https://www.framer.com/legal/sub-processors/'
    WHEN 'Cal.com' THEN 'https://security.cal.com/?itemUid=e3fae2ca-94a9-416b-b577-5c90e382df57'
    WHEN 'monday.com' THEN 'https://monday.com/l/privacy/sub-processors-subsidiaries-support'
    WHEN 'Carta' THEN 'https://carta.com/legal/carta-europe/data-processor-agreement/'
    WHEN 'Apollo.io' THEN 'https://trust.apollo.io/'
    WHEN 'Atlassian' THEN 'https://www.atlassian.com/legal/sub-processors'
    WHEN 'Webflow' THEN 'https://uploads-ssl.webflow.com/5d80c03f1edd7bd68fcdb623/657c411c020d1ee5f047f6ab_20231207%20-%20List%20of%20approved%20Sub-processors.pdf'
    WHEN '1Password' THEN 'https://1passwordstatic.com/files/legal-center/notice-of-updates-to-the-1password-subprocessor-list.pdf'
    WHEN 'Checkr' THEN 'https://checkr.com/sub-processor-list'
    WHEN 'Rippling' THEN 'https://app.rippling.com/legal/subprocessors'
    WHEN 'Twilio' THEN 'https://www.twilio.com/en-us/legal/sub-processors'
    WHEN 'Cursor' THEN 'https://trust.cursor.com/subprocessors'
    WHEN 'Pulley' THEN 'https://large-walk-bd5.notion.site/6449d3ecd2854af0a313b6e8679b5795?v=d61abc36355842bb939d49544f18dc95'
    END;
