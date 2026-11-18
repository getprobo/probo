-- Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

CREATE TYPE vendor_category AS ENUM (
    'ANALYTICS',
    'CLOUD_MONITORING',
    'CLOUD_PROVIDER',
    'COLLABORATION',
    'CUSTOMER_SUPPORT',
    'DATA_STORAGE_AND_PROCESSING',
    'DOCUMENT_MANAGEMENT',
    'EMPLOYEE_MANAGEMENT',
    'ENGINEERING',
    'FINANCE',
    'IDENTITY_PROVIDER',
    'IT',
    'MARKETING',
    'OFFICE_OPERATIONS',
    'OTHER',
    'PASSWORD_MANAGEMENT',
    'PRODUCT_AND_DESIGN',
    'PROFESSIONAL_SERVICES',
    'RECRUITING',
    'SALES',
    'SECURITY',
    'VERSION_CONTROL'
);

ALTER TABLE vendors DROP COLUMN IF EXISTS category;
ALTER TABLE vendors ADD COLUMN category vendor_category NOT NULL DEFAULT 'OTHER';

-- Update categories based on the vendor data.json mappings
UPDATE vendors SET category = 'ANALYTICS' WHERE name IN ('Segment', 'Plausible Analytics', 'PostHog');

UPDATE vendors SET category = 'CLOUD_MONITORING' WHERE name IN ('Datadog', 'New Relic', 'Better Stack');

UPDATE vendors SET category = 'CLOUD_PROVIDER' WHERE name IN (
    'Amazon Web Services (AWS)', 'Microsoft Azure', 'Google Cloud Platform', 'Cloudflare',
    'Vercel', 'Fly.io', 'Latitude.sh', 'Google', 'Heroku', 'Porter', 'OVHcloud global',
    'OVHcloud US', 'bunny.net'
);

UPDATE vendors SET category = 'COLLABORATION' WHERE name IN (
    'Slack', 'Microsoft 365', 'Cal.com', 'Zoom', 'Sendgrid', 'Google Workspace',
    'Signal', 'monday.com', 'Calendly', 'Twilio', 'Otter.ai', 'Claap', 'ClickUp'
);

UPDATE vendors SET category = 'CUSTOMER_SUPPORT' WHERE name = 'Pylon';

UPDATE vendors SET category = 'DATA_STORAGE_AND_PROCESSING' WHERE name IN (
    'Supabase', 'PlanetScale', 'ClickHouse', 'Airtable', 'Upstash'
);

UPDATE vendors SET category = 'DOCUMENT_MANAGEMENT' WHERE name = 'Notion';

UPDATE vendors SET category = 'EMPLOYEE_MANAGEMENT' WHERE name IN ('Gusto', 'Checkr', 'Rippling');

UPDATE vendors SET category = 'ENGINEERING' WHERE name IN (
    'Sentry', 'npm', 'Atlassian', 'Algolia', 'Mintlify', 'Linear', 'Langfuse',
    'OpenAI', 'Anthropic', 'Cursor', 'Together AI', 'Docker', 'jsDelivr'
);

UPDATE vendors SET category = 'FINANCE' WHERE name IN (
    'Stripe', 'Pulley', 'Mercury', 'Brex', 'Ramp', 'Carta', 'Qonto', 'Spendesk', 'Puzzle', 'Vouch'
);

UPDATE vendors SET category = 'IDENTITY_PROVIDER' WHERE name IN ('Auth0', 'Clerk');

UPDATE vendors SET category = 'IT' WHERE name IN ('Namecheap', 'Plesk');

UPDATE vendors SET category = 'MARKETING' WHERE name IN (
    'Hubspot', 'Mailchimp', 'Resend', 'Brevo', 'Loops', 'Perplexity'
);

UPDATE vendors SET category = 'PASSWORD_MANAGEMENT' WHERE name IN ('1Password', 'Bitwarden');

UPDATE vendors SET category = 'PRODUCT_AND_DESIGN' WHERE name IN (
    'Framer', 'Figma', 'Webflow', 'Pitch'
);

UPDATE vendors SET category = 'RECRUITING' WHERE name = 'Lever';

UPDATE vendors SET category = 'SALES' WHERE name IN ('Apollo.io', 'Pipedrive', 'folk');

UPDATE vendors SET category = 'SECURITY' WHERE name IN ('Tailscale', 'Probo');

UPDATE vendors SET category = 'VERSION_CONTROL' WHERE name = 'Github';
