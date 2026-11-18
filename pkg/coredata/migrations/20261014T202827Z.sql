-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

-- Runs after 20261014T202826Z. That migration already copied a usable source
-- name onto empty accounts. Anything still blank has no source label. Store
-- the provider display name so the column is filled.
-- A provider with no registered display name keeps the provider value.
-- name_synced_at is left unchanged.
UPDATE connector_accounts AS account
SET
    name = COALESCE(label.display_name, connector.provider::text),
    updated_at = NOW()
FROM connectors AS connector
LEFT JOIN (
    VALUES
        ('SLACK', 'Slack'),
        ('GOOGLE_WORKSPACE', 'Google Workspace'),
        ('LINEAR', 'Linear'),
        ('LINEAR_SYNC', 'Linear Sync'),
        ('ONE_PASSWORD', '1Password'),
        ('HUBSPOT', 'HubSpot'),
        ('DOCUSIGN', 'DocuSign'),
        ('NOTION', 'Notion'),
        ('BREX', 'Brex'),
        ('TALLY', 'Tally'),
        ('CLOUDFLARE', 'Cloudflare'),
        ('GRAFANA', 'Grafana'),
        ('OPENAI', 'OpenAI'),
        ('POSTHOG', 'PostHog'),
        ('SENTRY', 'Sentry'),
        ('SIGNOZ', 'SigNoz'),
        ('SUPABASE', 'Supabase'),
        ('BETTER_STACK', 'Better Stack'),
        ('GITHUB', 'GitHub'),
        ('INTERCOM', 'Intercom'),
        ('RESEND', 'Resend'),
        ('SENDGRID', 'SendGrid'),
        ('MICROSOFT_365', 'Microsoft 365'),
        ('GITLAB', 'GitLab'),
        ('BITBUCKET', 'Bitbucket'),
        ('HEROKU', 'Heroku'),
        ('PAGERDUTY', 'PagerDuty'),
        ('ASANA', 'Asana'),
        ('NETLIFY', 'Netlify'),
        ('CLICKUP', 'ClickUp'),
        ('VERCEL', 'Vercel'),
        ('MONDAY', 'Monday.com'),
        ('METABASE', 'Metabase'),
        ('MONGODB_ATLAS', 'MongoDB Atlas'),
        ('TAILSCALE', 'Tailscale'),
        ('ANTHROPIC', 'Anthropic'),
        ('CURSOR', 'Cursor'),
        ('DATADOG', 'Datadog'),
        ('OKTA', 'Okta'),
        ('ZENDESK', 'Zendesk'),
        ('QOVERY', 'Qovery'),
        ('RENDER', 'Render'),
        ('NEON', 'Neon'),
        ('MERCURY', 'Mercury'),
        ('APOLLO', 'Apollo.io'),
        ('DEEPGRAM', 'Deepgram'),
        ('CLICKHOUSE', 'ClickHouse Cloud'),
        ('LANGFUSE', 'Langfuse'),
        ('PYLON', 'Pylon'),
        ('OPENROUTER', 'OpenRouter'),
        ('INCIDENT_IO', 'incident.io'),
        ('BREVO', 'Brevo'),
        ('SCALEWAY', 'Scaleway'),
        ('YOUSIGN', 'Yousign'),
        ('RAILWAY', 'Railway'),
        ('CRISP', 'Crisp'),
        ('DOTFILE', 'Dotfile'),
        ('SEGMENT', 'Segment'),
        ('SQUARE', 'Square'),
        ('GOOGLE_ANALYTICS', 'Google Analytics'),
        ('UPCLOUD', 'UpCloud'),
        ('NUKI', 'Nuki'),
        ('AUTHENTIK', 'authentik'),
        ('CAL_COM', 'Cal.com'),
        ('CALENDLY', 'Calendly'),
        ('AWS', 'Amazon Web Services'),
        ('GCP', 'Google Cloud'),
        ('ELEVENLABS', 'ElevenLabs'),
        ('NEW_RELIC', 'New Relic'),
        ('RETOOL', 'Retool'),
        ('TWINGATE', 'Twingate'),
        ('ATTIO', 'Attio'),
        ('AZURE', 'Microsoft Azure'),
        ('OVHCLOUD', 'OVHcloud')
) AS label(provider, display_name)
    ON label.provider = connector.provider::text
WHERE
    connector.id = account.connector_id
    AND account.name = '';
