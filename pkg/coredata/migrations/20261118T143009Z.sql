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

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: pg_stat_statements; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_stat_statements WITH SCHEMA public;


--
-- Name: EXTENSION pg_stat_statements; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_stat_statements IS 'track planning and execution statistics of all SQL statements executed';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: unaccent; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS unaccent WITH SCHEMA public;


--
-- Name: EXTENSION unaccent; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION unaccent IS 'text search dictionary that removes accents';


--
-- Name: access_review_campaign_source_fetch_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.access_review_campaign_source_fetch_status AS ENUM (
    'QUEUED',
    'FETCHING',
    'SUCCESS',
    'FAILED'
);


--
-- Name: access_review_campaign_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.access_review_campaign_status AS ENUM (
    'DRAFT',
    'IN_PROGRESS',
    'PENDING_ACTIONS',
    'COMPLETED',
    'CANCELLED'
);


--
-- Name: access_review_entry_decision; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.access_review_entry_decision AS ENUM (
    'PENDING',
    'APPROVED',
    'REVOKE',
    'DEFER',
    'ESCALATE'
);


--
-- Name: access_review_entry_flag; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.access_review_entry_flag AS ENUM (
    'NONE',
    'ORPHANED',
    'INACTIVE',
    'EXCESSIVE',
    'ROLE_MISMATCH',
    'NEW'
);


--
-- Name: agent_input_purpose; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.agent_input_purpose AS ENUM (
    'USER'
);


--
-- Name: ai_system_risk_classifications; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.ai_system_risk_classifications AS ENUM (
    'HIGH_RISK',
    'LIMITED',
    'MINIMAL',
    'GPAI'
);


--
-- Name: ai_system_statuses; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.ai_system_statuses AS ENUM (
    'ACTIVE',
    'IN_DEVELOPMENT',
    'DECOMMISSIONED'
);


--
-- Name: asset_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.asset_type AS ENUM (
    'PHYSICAL',
    'VIRTUAL'
);


--
-- Name: audit_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.audit_state AS ENUM (
    'TO_BOOK',
    'AUDIT_BOOKED',
    'NOT_STARTED',
    'IN_PROGRESS',
    'COMPLETED',
    'REJECTED',
    'OUTDATED'
);


--
-- Name: auth_method; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.auth_method AS ENUM (
    'SSO',
    'PASSWORD',
    'API_KEY',
    'SERVICE_ACCOUNT',
    'UNKNOWN',
    'OAUTH2',
    'SSH'
);


--
-- Name: authz_role; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.authz_role AS ENUM (
    'OWNER',
    'ADMIN',
    'EMPLOYEE',
    'VIEWER',
    'AUDITOR',
    'COMPLIANCE_PORTAL_MANAGER',
    'COMPLIANCE_PORTAL_ACCESS_MANAGER'
);


--
-- Name: bot_message_purpose; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.bot_message_purpose AS ENUM (
    'POST',
    'UPDATE'
);


--
-- Name: business_function_classifications; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.business_function_classifications AS ENUM (
    'CRITICAL',
    'IMPORTANT',
    'SECONDARY',
    'STANDARD'
);


--
-- Name: business_impact; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.business_impact AS ENUM (
    'LOW',
    'MEDIUM',
    'HIGH',
    'CRITICAL'
);


--
-- Name: common_third_party_review; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.common_third_party_review AS ENUM (
    'UNREVIEWED',
    'VALIDATED',
    'REJECTED'
);


--
-- Name: common_tracker_pattern_attribution; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.common_tracker_pattern_attribution AS ENUM (
    'UNDETERMINED',
    'THIRD_PARTY',
    'FIRST_PARTY',
    'NOT_ATTRIBUTABLE'
);


--
-- Name: compliance_portal_document_access_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.compliance_portal_document_access_status AS ENUM (
    'REQUESTED',
    'GRANTED',
    'REJECTED',
    'REVOKED'
);


--
-- Name: compliance_portal_visibility; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.compliance_portal_visibility AS ENUM (
    'RESTRICTED',
    'PUBLIC'
);


--
-- Name: connector_protocol; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.connector_protocol AS ENUM (
    'OAUTH2',
    'API_KEY',
    'GITHUB_APP',
    'WORKLOAD_IDENTITY'
);


--
-- Name: connector_provider; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.connector_provider AS ENUM (
    'SLACK',
    'GOOGLE_WORKSPACE',
    'LINEAR',
    'FIGMA',
    'ONE_PASSWORD',
    'HUBSPOT',
    'DOCUSIGN',
    'NOTION',
    'BREX',
    'TALLY',
    'CLOUDFLARE',
    'OPENAI',
    'SENTRY',
    'SUPABASE',
    'GITHUB',
    'INTERCOM',
    'RESEND',
    'MICROSOFT_365',
    'GITLAB',
    'BITBUCKET',
    'HEROKU',
    'PAGERDUTY',
    'ASANA',
    'NETLIFY',
    'CLICKUP',
    'VERCEL',
    'MONDAY',
    'TAILSCALE',
    'METABASE',
    'GRAFANA',
    'CURSOR',
    'ANTHROPIC',
    'POSTHOG',
    'CLERK',
    'SIGNOZ',
    'DATADOG',
    'SENDGRID',
    'ZENDESK',
    'OKTA',
    'BETTER_STACK',
    'QOVERY',
    'RENDER',
    'NEON',
    'LANGFUSE',
    'APOLLO',
    'CLICKHOUSE',
    'MERCURY',
    'DEEPGRAM',
    'BREVO',
    'PYLON',
    'OPENROUTER',
    'INCIDENT_IO',
    'CRISP',
    'SCALEWAY',
    'RAILWAY',
    'YOUSIGN',
    'DOTFILE',
    'SEGMENT',
    'SQUARE',
    'GOOGLE_ANALYTICS',
    'UPCLOUD',
    'NUKI',
    'AUTHENTIK',
    'CAL_COM',
    'CALENDLY',
    'AWS',
    'GCP',
    'AZURE',
    'RETOOL',
    'ELEVENLABS',
    'NEW_RELIC',
    'TWINGATE',
    'ATTIO',
    'MONGODB_ATLAS',
    'LINEAR_SYNC',
    'OVHCLOUD',
    'OVHCLOUD_US',
    'OVHCLOUD_CA',
    'FRONT',
    'PLANETSCALE',
    'DAYTONA',
    'ELASTIC_CLOUD',
    'UNIFI'
);


--
-- Name: control_maturity_level; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.control_maturity_level AS ENUM (
    'NONE',
    'INITIAL',
    'MANAGED',
    'DEFINED',
    'QUANTITATIVELY_MANAGED',
    'OPTIMIZING'
);


--
-- Name: cookie_banner_scan_failure_reason; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_banner_scan_failure_reason AS ENUM (
    'UNREACHABLE',
    'BLOCKED',
    'NO_PAGES_LOADED',
    'INTERRUPTED',
    'INTERNAL',
    'BANNER_NOT_INSTALLED',
    'DIFFERENT_BANNER'
);


--
-- Name: cookie_banner_scan_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_banner_scan_state AS ENUM (
    'PENDING',
    'RUNNING',
    'COMPLETED',
    'FAILED',
    'PAUSED'
);


--
-- Name: cookie_banner_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_banner_state AS ENUM (
    'ACTIVE',
    'INACTIVE'
);


--
-- Name: cookie_banner_version_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_banner_version_state AS ENUM (
    'DRAFT',
    'PUBLISHED'
);


--
-- Name: cookie_category_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_category_kind AS ENUM (
    'NORMAL',
    'NECESSARY',
    'UNCATEGORISED'
);


--
-- Name: cookie_consent_action; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_consent_action AS ENUM (
    'ACCEPT_ALL',
    'REJECT_ALL',
    'CUSTOMIZE',
    'GPC',
    'ACKNOWLEDGE'
);


--
-- Name: cookie_consent_mode; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_consent_mode AS ENUM (
    'OPT_IN',
    'OPT_OUT'
);


--
-- Name: cookie_pattern_match_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_pattern_match_type AS ENUM (
    'EXACT',
    'PREFIX',
    'GLOB'
);


--
-- Name: cookie_same_site; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_same_site AS ENUM (
    'STRICT',
    'LAX',
    'NONE'
);


--
-- Name: cookie_source; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.cookie_source AS ENUM (
    'SCRIPT',
    'PRE_EXISTING',
    'HTTP',
    'EXTENSION'
);


--
-- Name: country_code; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.country_code AS ENUM (
    'GLOBAL',
    'AD',
    'AE',
    'AF',
    'AG',
    'AI',
    'AL',
    'AM',
    'AO',
    'AQ',
    'AR',
    'AS',
    'AT',
    'AU',
    'AW',
    'AX',
    'AZ',
    'BA',
    'BB',
    'BD',
    'BE',
    'BF',
    'BG',
    'BH',
    'BI',
    'BJ',
    'BL',
    'BM',
    'BN',
    'BO',
    'BQ',
    'BR',
    'BS',
    'BT',
    'BV',
    'BW',
    'BY',
    'BZ',
    'CA',
    'CC',
    'CD',
    'CF',
    'CG',
    'CH',
    'CI',
    'CK',
    'CL',
    'CM',
    'CN',
    'CO',
    'CR',
    'CU',
    'CV',
    'CW',
    'CX',
    'CY',
    'CZ',
    'DE',
    'DJ',
    'DK',
    'DM',
    'DO',
    'DZ',
    'EC',
    'EE',
    'EG',
    'EH',
    'ER',
    'ES',
    'ET',
    'EU',
    'FI',
    'FJ',
    'FK',
    'FM',
    'FO',
    'FR',
    'GA',
    'GB',
    'GD',
    'GE',
    'GF',
    'GG',
    'GH',
    'GI',
    'GL',
    'GM',
    'GN',
    'GP',
    'GQ',
    'GR',
    'GT',
    'GU',
    'GW',
    'GY',
    'HK',
    'HM',
    'HN',
    'HR',
    'HT',
    'HU',
    'ID',
    'IE',
    'IL',
    'IM',
    'IN',
    'IO',
    'IQ',
    'IR',
    'IS',
    'IT',
    'JE',
    'JM',
    'JO',
    'JP',
    'KE',
    'KG',
    'KH',
    'KI',
    'KM',
    'KN',
    'KP',
    'KR',
    'KW',
    'KY',
    'KZ',
    'LA',
    'LB',
    'LC',
    'LI',
    'LK',
    'LR',
    'LS',
    'LT',
    'LU',
    'LV',
    'LY',
    'MA',
    'MC',
    'MD',
    'ME',
    'MF',
    'MG',
    'MH',
    'MK',
    'ML',
    'MM',
    'MN',
    'MO',
    'MP',
    'MQ',
    'MR',
    'MS',
    'MT',
    'MU',
    'MV',
    'MW',
    'MX',
    'MY',
    'MZ',
    'NA',
    'NC',
    'NE',
    'NF',
    'NG',
    'NI',
    'NL',
    'NO',
    'NP',
    'NR',
    'NU',
    'NZ',
    'OM',
    'PA',
    'PE',
    'PF',
    'PG',
    'PH',
    'PK',
    'PL',
    'PM',
    'PN',
    'PR',
    'PS',
    'PT',
    'PW',
    'PY',
    'QA',
    'RE',
    'RO',
    'RS',
    'RU',
    'RW',
    'SA',
    'SB',
    'SC',
    'SD',
    'SE',
    'SG',
    'SH',
    'SI',
    'SJ',
    'SK',
    'SL',
    'SM',
    'SN',
    'SO',
    'SR',
    'SS',
    'ST',
    'SV',
    'SX',
    'SY',
    'SZ',
    'TC',
    'TD',
    'TF',
    'TG',
    'TH',
    'TJ',
    'TK',
    'TL',
    'TM',
    'TN',
    'TO',
    'TR',
    'TT',
    'TV',
    'TW',
    'TZ',
    'UA',
    'UG',
    'UM',
    'US',
    'UY',
    'UZ',
    'VA',
    'VC',
    'VE',
    'VG',
    'VI',
    'VN',
    'VU',
    'WF',
    'WS',
    'YE',
    'YT',
    'ZA',
    'ZM',
    'ZW'
);


--
-- Name: criticity_level; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.criticity_level AS ENUM (
    'LOW',
    'MEDIUM',
    'HIGH'
);


--
-- Name: custom_domain_ssl_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.custom_domain_ssl_status AS ENUM (
    'PENDING',
    'PROVISIONING',
    'ACTIVE',
    'RENEWING',
    'EXPIRED',
    'FAILED'
);


--
-- Name: data_classification_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.data_classification_type AS ENUM (
    'PUBLIC',
    'INTERNAL',
    'CONFIDENTIAL',
    'SECRET'
);


--
-- Name: data_sensitivity; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.data_sensitivity AS ENUM (
    'NONE',
    'LOW',
    'MEDIUM',
    'HIGH',
    'CRITICAL'
);


--
-- Name: detected_tracker_observer; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.detected_tracker_observer AS ENUM (
    'SDK',
    'SCANNER'
);


--
-- Name: device_platform; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.device_platform AS ENUM (
    'DARWIN',
    'LINUX',
    'FREEBSD',
    'WINDOWS'
);


--
-- Name: device_posture_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.device_posture_status AS ENUM (
    'PASS',
    'FAIL',
    'UNKNOWN',
    'NOT_APPLICABLE'
);


--
-- Name: device_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.device_state AS ENUM (
    'PENDING',
    'ACTIVE',
    'REVOKED'
);


--
-- Name: document_classification; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.document_classification AS ENUM (
    'PUBLIC',
    'INTERNAL',
    'CONFIDENTIAL',
    'SECRET'
);


--
-- Name: document_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.document_status AS ENUM (
    'ACTIVE',
    'ARCHIVED'
);


--
-- Name: document_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.document_type AS ENUM (
    'OTHER',
    'GOVERNANCE',
    'POLICY',
    'PROCEDURE',
    'PLAN',
    'REGISTER',
    'RECORD',
    'REPORT',
    'TEMPLATE',
    'STATEMENT_OF_APPLICABILITY'
);


--
-- Name: document_version_approval_decision_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.document_version_approval_decision_state AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED',
    'VOIDED'
);


--
-- Name: document_version_approval_quorum_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.document_version_approval_quorum_status AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED',
    'VOIDED'
);


--
-- Name: document_version_orientation; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.document_version_orientation AS ENUM (
    'PORTRAIT',
    'LANDSCAPE'
);


--
-- Name: document_version_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.document_version_status AS ENUM (
    'PENDING_APPROVAL',
    'PUBLISHED',
    'REJECTED',
    'VOIDED'
);


--
-- Name: document_write_mode; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.document_write_mode AS ENUM (
    'AUTHORED',
    'GENERATED'
);


--
-- Name: electronic_signature_document_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.electronic_signature_document_type AS ENUM (
    'NDA',
    'DPA',
    'MSA',
    'SOW',
    'SLA',
    'TOS',
    'PRIVACY_POLICY',
    'OTHER',
    'GOVERNANCE',
    'POLICY',
    'PROCEDURE',
    'PLAN',
    'REGISTER',
    'RECORD',
    'REPORT',
    'TEMPLATE',
    'STATEMENT_OF_APPLICABILITY'
);


--
-- Name: electronic_signature_event_source; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.electronic_signature_event_source AS ENUM (
    'CLIENT',
    'SERVER'
);


--
-- Name: electronic_signature_event_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.electronic_signature_event_type AS ENUM (
    'DOCUMENT_VIEWED',
    'CONSENT_GIVEN',
    'FULL_NAME_TYPED',
    'SIGNATURE_ACCEPTED',
    'SIGNATURE_COMPLETED',
    'SEAL_COMPUTED',
    'TIMESTAMP_REQUESTED',
    'CERTIFICATE_GENERATED',
    'PROCESSING_ERROR'
);


--
-- Name: electronic_signature_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.electronic_signature_status AS ENUM (
    'PENDING',
    'ACCEPTED',
    'PROCESSING',
    'COMPLETED',
    'FAILED'
);


--
-- Name: evidence_description_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.evidence_description_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'COMPLETED',
    'FAILED'
);


--
-- Name: evidence_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.evidence_state AS ENUM (
    'FULFILLED',
    'REQUESTED'
);


--
-- Name: evidence_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.evidence_type AS ENUM (
    'FILE',
    'LINK'
);


--
-- Name: export_jobs_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.export_jobs_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'COMPLETED',
    'FAILED'
);


--
-- Name: export_jobs_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.export_jobs_type AS ENUM (
    'FRAMEWORK',
    'DOCUMENT',
    'AUDIT_LOG',
    'SCIM_EVENT'
);


--
-- Name: findings_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.findings_kind AS ENUM (
    'MINOR_NONCONFORMITY',
    'OBSERVATION',
    'EXCEPTION',
    'MAJOR_NONCONFORMITY'
);


--
-- Name: findings_priority; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.findings_priority AS ENUM (
    'LOW',
    'MEDIUM',
    'HIGH'
);


--
-- Name: findings_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.findings_status AS ENUM (
    'OPEN',
    'IN_PROGRESS',
    'CLOSED',
    'RISK_ACCEPTED',
    'MITIGATED',
    'FALSE_POSITIVE'
);


--
-- Name: framework_export_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.framework_export_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed'
);


--
-- Name: iam_oidc_provider; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.iam_oidc_provider AS ENUM (
    'GOOGLE',
    'MICROSOFT',
    'COMPLIANCE_PORTAL'
);


--
-- Name: internal_control_event_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.internal_control_event_type AS ENUM (
    'CREATED',
    'UPDATED',
    'DELETED'
);


--
-- Name: internal_control_implementation_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.internal_control_implementation_status AS ENUM (
    'NOT_IMPLEMENTED',
    'IN_PROGRESS',
    'IMPLEMENTED',
    'OPERATING'
);


--
-- Name: internal_control_nature; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.internal_control_nature AS ENUM (
    'MANUAL'
);


--
-- Name: internal_control_operating_mode; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.internal_control_operating_mode AS ENUM (
    'CONTINUOUS',
    'EVENT',
    'PERIODIC'
);


--
-- Name: internal_control_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.internal_control_type AS ENUM (
    'PREVENTIVE',
    'DETECTIVE',
    'CORRECTIVE'
);


--
-- Name: mailing_list_subscriber_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.mailing_list_subscriber_status AS ENUM (
    'PENDING',
    'CONFIRMED'
);


--
-- Name: mailing_list_update_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.mailing_list_update_status AS ENUM (
    'DRAFT',
    'ENQUEUED',
    'PROCESSING',
    'SENT'
);


--
-- Name: membership_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.membership_state AS ENUM (
    'PENDING',
    'ACTIVE',
    'DEACTIVATED'
);


--
-- Name: mfa_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.mfa_status AS ENUM (
    'ENABLED',
    'DISABLED',
    'UNKNOWN'
);


--
-- Name: mitigation_importance; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.mitigation_importance AS ENUM (
    'MANDATORY',
    'PREFERRED',
    'ADVANCED'
);


--
-- Name: mitigation_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.mitigation_state AS ENUM (
    'NOT_STARTED',
    'IN_PROGRESS',
    'NOT_APPLICABLE',
    'IMPLEMENTED',
    'UNKNOWN',
    'NOT_IMPLEMENTED'
);


--
-- Name: notification_email_delivery; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.notification_email_delivery AS ENUM (
    'NONE',
    'EACH',
    'REMINDERS',
    'WEEKLY'
);


--
-- Name: notification_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.notification_status AS ENUM (
    'PENDING',
    'RESOLVED'
);


--
-- Name: notification_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.notification_type AS ENUM (
    'DOCUMENT_SIGNATURE_REQUESTED',
    'DOCUMENT_APPROVAL_REQUESTED',
    'TASK_ASSIGNED',
    'TASK_MENTIONED'
);


--
-- Name: oauth2_client_token_endpoint_auth_method; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.oauth2_client_token_endpoint_auth_method AS ENUM (
    'client_secret_basic',
    'client_secret_post',
    'none'
);


--
-- Name: oauth2_client_visibility; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.oauth2_client_visibility AS ENUM (
    'private',
    'public'
);


--
-- Name: oauth2_device_code_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.oauth2_device_code_status AS ENUM (
    'pending',
    'authorized',
    'denied',
    'expired'
);


--
-- Name: obligation_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.obligation_type AS ENUM (
    'LEGAL',
    'CONTRACTUAL'
);


--
-- Name: obligations_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.obligations_status AS ENUM (
    'NON_COMPLIANT',
    'PARTIALLY_COMPLIANT',
    'COMPLIANT'
);


--
-- Name: people_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.people_kind AS ENUM (
    'EMPLOYEE',
    'CONTRACTOR',
    'SERVICE_ACCOUNT'
);


--
-- Name: policy_version_signature_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.policy_version_signature_state AS ENUM (
    'REQUESTED',
    'SIGNED'
);


--
-- Name: processing_activities_data_protection_impact_assessment; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.processing_activities_data_protection_impact_assessment AS ENUM (
    'NEEDED',
    'NOT_NEEDED'
);


--
-- Name: processing_activities_lawful_basis; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.processing_activities_lawful_basis AS ENUM (
    'LEGITIMATE_INTEREST',
    'CONSENT',
    'CONTRACTUAL_NECESSITY',
    'LEGAL_OBLIGATION',
    'VITAL_INTERESTS',
    'PUBLIC_TASK'
);


--
-- Name: processing_activities_special_or_criminal_data; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.processing_activities_special_or_criminal_data AS ENUM (
    'YES',
    'NO',
    'POSSIBLE'
);


--
-- Name: processing_activities_transfer_impact_assessment; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.processing_activities_transfer_impact_assessment AS ENUM (
    'NEEDED',
    'NOT_NEEDED'
);


--
-- Name: processing_activities_transfer_safeguards; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.processing_activities_transfer_safeguards AS ENUM (
    'STANDARD_CONTRACTUAL_CLAUSES',
    'BINDING_CORPORATE_RULES',
    'ADEQUACY_DECISION',
    'DEROGATIONS',
    'CODES_OF_CONDUCT',
    'CERTIFICATION_MECHANISMS'
);


--
-- Name: processing_activity_dpia_residual_risk; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.processing_activity_dpia_residual_risk AS ENUM (
    'LOW',
    'MEDIUM',
    'HIGH'
);


--
-- Name: processing_activity_role; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.processing_activity_role AS ENUM (
    'CONTROLLER',
    'PROCESSOR'
);


--
-- Name: regulation_source; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.regulation_source AS ENUM (
    'DETECTED',
    'DEFAULT'
);


--
-- Name: rights_request_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.rights_request_state AS ENUM (
    'TODO',
    'IN_PROGRESS',
    'DONE',
    'REJECTED'
);


--
-- Name: rights_request_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.rights_request_type AS ENUM (
    'ACCESS',
    'DELETION',
    'RECTIFICATION',
    'PORTABILITY',
    'OBJECTION',
    'COMPLAINT'
);


--
-- Name: risk_analysis_node_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.risk_analysis_node_type AS ENUM (
    'ENTITY',
    'ASSET',
    'DATA'
);


--
-- Name: risk_tier; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.risk_tier AS ENUM (
    'CRITICAL',
    'SIGNIFICANT',
    'GENERAL'
);


--
-- Name: risk_treatment; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.risk_treatment AS ENUM (
    'MITIGATED',
    'TRANSFERRED',
    'AVOIDED',
    'ACCEPTED'
);


--
-- Name: saml_enforcement_policy; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.saml_enforcement_policy AS ENUM (
    'OFF',
    'OPTIONAL',
    'REQUIRED'
);


--
-- Name: search_engine_indexing; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.search_engine_indexing AS ENUM (
    'INDEXABLE',
    'NOT_INDEXABLE'
);


--
-- Name: service_criticality; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.service_criticality AS ENUM (
    'LOW',
    'MEDIUM',
    'HIGH'
);


--
-- Name: session_auth_method; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.session_auth_method AS ENUM (
    'PASSWORD',
    'SAML',
    'MAGIC_LINK',
    'OIDC'
);


--
-- Name: session_expire_reason; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.session_expire_reason AS ENUM (
    'idle_timeout',
    'revoked',
    'closed'
);


--
-- Name: slack_delivery_operation_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.slack_delivery_operation_kind AS ENUM (
    'POST_MESSAGE',
    'ADD_REACTION'
);


--
-- Name: slack_message_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.slack_message_type AS ENUM (
    'TRUST_CENTER_ACCESS_REQUEST',
    'WELCOME'
);


--
-- Name: slackbot_installation_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.slackbot_installation_status AS ENUM (
    'ACTIVE',
    'DISABLED'
);


--
-- Name: task_activity_field; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.task_activity_field AS ENUM (
    'NAME',
    'STATE',
    'PRIORITY',
    'ASSIGNED_TO',
    'INTERNAL_CONTROL',
    'DEADLINE',
    'TIME_ESTIMATE'
);


--
-- Name: task_activity_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.task_activity_type AS ENUM (
    'CREATED',
    'UPDATED',
    'COMMENTED'
);


--
-- Name: task_priority; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.task_priority AS ENUM (
    'URGENT',
    'HIGH',
    'MEDIUM',
    'LOW'
);


--
-- Name: task_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.task_state AS ENUM (
    'BACKLOG',
    'TODO',
    'IN_PROGRESS',
    'DONE',
    'CANCELED',
    'DUPLICATE'
);


--
-- Name: third_party_category; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.third_party_category AS ENUM (
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


--
-- Name: third_party_vetting_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.third_party_vetting_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'COMPLETED',
    'FAILED'
);


--
-- Name: tracker_resource_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.tracker_resource_type AS ENUM (
    'SCRIPT',
    'IFRAME',
    'IMAGE',
    'STYLESHEET',
    'FONT',
    'BEACON',
    'FETCH',
    'MEDIA',
    'SERVICE_WORKER'
);


--
-- Name: tracker_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.tracker_type AS ENUM (
    'COOKIE',
    'LOCAL_STORAGE',
    'SESSION_STORAGE',
    'INDEXED_DB',
    'CACHE_STORAGE'
);


--
-- Name: treatment_plan_event_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.treatment_plan_event_type AS ENUM (
    'CREATED',
    'UPDATED',
    'DELETED',
    'INTERNAL_CONTROL_LINKED',
    'INTERNAL_CONTROL_UNLINKED'
);


--
-- Name: trust_center_access_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.trust_center_access_state AS ENUM (
    'ACTIVE',
    'DEACTIVATED'
);


--
-- Name: webhook_event_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.webhook_event_status AS ENUM (
    'PENDING',
    'SUCCEEDED',
    'FAILED'
);


--
-- Name: webhook_event_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.webhook_event_type AS ENUM (
    'meeting:created',
    'meeting:updated',
    'meeting:deleted',
    'third-party:created',
    'third-party:updated',
    'third-party:deleted',
    'user:created',
    'user:updated',
    'user:deleted',
    'obligation:created',
    'obligation:updated',
    'obligation:deleted',
    'document:created',
    'document:updated',
    'document:archived',
    'document:unarchived',
    'document:deleted',
    'document-version:created',
    'document-version:updated',
    'document-version:published',
    'document-version:rejected',
    'document-version:deleted',
    'document-version-signature:requested',
    'document-version-signature:signed',
    'document-version-signature:cancelled',
    'document-version-approval-quorum:requested',
    'document-version-approval-quorum:updated',
    'document-version-approval-quorum:approved',
    'document-version-approval-quorum:rejected',
    'document-version-approval-quorum:voided',
    'right-request:created',
    'right-request:updated',
    'right-request:deleted',
    'task:created',
    'task:updated',
    'task:deleted',
    'task-comment:created',
    'task-comment:updated',
    'task-comment:deleted'
);


--
-- Name: decode_base64_unpadded(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.decode_base64_unpadded(input_text text) RETURNS bytea
    LANGUAGE plpgsql
    AS $$
DECLARE
    padded_text text;
    mod_length integer;
    normalized_text text;
BEGIN
    -- Replace URL-safe characters with standard base64 characters
    normalized_text := translate(input_text, '-_', '+/');
    
    -- Calculate how many padding characters we need to add
    mod_length := length(normalized_text) % 4;
    
    -- Add the required padding
    IF mod_length = 0 THEN
        padded_text := normalized_text;
    ELSIF mod_length = 1 THEN
        -- Invalid base64 - length mod 4 can't be 1
        RAISE EXCEPTION 'Invalid base64 length';
    ELSIF mod_length = 2 THEN
        padded_text := normalized_text || '==';
    ELSIF mod_length = 3 THEN
        padded_text := normalized_text || '=';
    END IF;
    
    -- Decode the padded base64
    RETURN decode(padded_text, 'base64');
END;
$$;


--
-- Name: extract_entity_type(bytea); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.extract_entity_type(gid bytea) RETURNS integer
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN get_byte(gid, 8) * 256 + get_byte(gid, 9);
END;
$$;


--
-- Name: extract_tenant_id(bytea); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.extract_tenant_id(gid bytea) RETURNS bytea
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN substring(gid from 1 for 8);
END;
$$;


--
-- Name: extract_timestamp(bytea); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.extract_timestamp(gid bytea) RETURNS timestamp without time zone
    LANGUAGE plpgsql IMMUTABLE
    AS $$
DECLARE
    ms bigint;
BEGIN
    ms := (get_byte(gid, 10)::bigint << 56) |
          (get_byte(gid, 11)::bigint << 48) |
          (get_byte(gid, 12)::bigint << 40) |
          (get_byte(gid, 13)::bigint << 32) |
          (get_byte(gid, 14)::bigint << 24) |
          (get_byte(gid, 15)::bigint << 16) |
          (get_byte(gid, 16)::bigint << 8) |
          get_byte(gid, 17)::bigint;
          
    RETURN to_timestamp(ms / 1000.0);
END;
$$;


--
-- Name: generate_gid(bytea, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.generate_gid(tenant_id bytea, entity_type integer) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    id bytea;
    timestamp_ms_bytes bytea;
    entity_type_bytes bytea;
    random_bytes bytea;
BEGIN
    -- Validate tenant_id
    IF tenant_id IS NULL OR octet_length(tenant_id) != 8 THEN
        RAISE EXCEPTION 'Invalid tenant ID: must be 8 bytes';
    END IF;

    -- 1. Start with tenant ID (8 bytes)
    id := tenant_id;
    
    -- 2. Add entity type (2 bytes)
    entity_type_bytes := substring(int4send(entity_type) from 3 for 2);
    id := id || entity_type_bytes;
    
    -- 3. Add timestamp in milliseconds (8 bytes)
    -- Extract milliseconds since epoch
    timestamp_ms_bytes := int8send(
        (extract(epoch from now()) * 1000)::bigint
    );
    id := id || timestamp_ms_bytes;
    
    -- 4. Add random bytes for uniqueness (6 bytes)
    random_bytes := gen_random_bytes(6);
    id := id || random_bytes;
    
    -- Return as URL-safe base64
    RETURN rtrim(translate(encode(id, 'base64'), '+/', '-_'), '=');
END;
$$;


--
-- Name: generate_machine_id(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.generate_machine_id() RETURNS bytea
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
    machine_id bytea;
BEGIN
    -- Generate 3 random bytes for machine ID
    machine_id := decode(encode(gen_random_bytes(3), 'hex'), 'hex');
    RETURN machine_id;
END;
$$;


--
-- Name: generate_tenant_id(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.generate_tenant_id() RETURNS bytea
    LANGUAGE plpgsql
    AS $$
DECLARE
    id bytea;
    machine_id bytea;
    timestamp_bytes bytea;
    counter_bytes bytea;
    counter_val int;
BEGIN
    -- 1. Get machine ID (3 bytes)
    machine_id := decode(current_setting('app.machine_id'), 'hex');
    
    -- 2. Get timestamp bytes (3 bytes - Unix time in seconds)
    timestamp_bytes := substring(int8send(extract(epoch from now())::bigint) from 6 for 3);
    
    -- 3. Get counter (2 bytes)
    counter_val := nextval('tenant_id_counter_seq') % 65536; -- 2^16
    counter_bytes := substring(int4send(counter_val) from 3 for 2);
    
    -- 4. Combine all parts
    id := machine_id || timestamp_bytes || counter_bytes;
    
    RETURN id;
END;
$$;


--
-- Name: gid_to_string(bytea); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.gid_to_string(gid bytea) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    -- Make sure to handle padding properly - remove trailing '=' characters
    RETURN rtrim(translate(encode(gid, 'base64'), '+/', '-_'), '=');
END;
$$;


--
-- Name: parse_gid(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.parse_gid(encoded text) RETURNS bytea
    LANGUAGE plpgsql IMMUTABLE
    AS $$
DECLARE
    decoded bytea;
    padded_input text;
    padding_needed int;
BEGIN
    -- Add proper padding for base64 decoding
    padding_needed := (4 - (length(encoded) % 4)) % 4;
    padded_input := encoded || repeat('=', padding_needed);
    
    -- Replace URL-safe chars and decode
    BEGIN
        decoded := decode(translate(padded_input, '-_', '+/'), 'base64');
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION 'Invalid base64 encoding in GID';
    END;
    
    -- Validate length
    IF octet_length(decoded) != 24 THEN
        RAISE EXCEPTION 'Invalid GID length: got %, want 24', octet_length(decoded);
    END IF;
    
    RETURN decoded;
END;
$$;


--
-- Name: parse_tenant_id(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.parse_tenant_id(encoded text) RETURNS bytea
    LANGUAGE plpgsql IMMUTABLE
    AS $$
DECLARE
    decoded bytea;
    padded_input text;
    padding_needed int;
BEGIN
    -- Add proper padding for base64 decoding
    padding_needed := (4 - (length(encoded) % 4)) % 4;
    padded_input := encoded || repeat('=', padding_needed);
    
    -- Replace URL-safe chars and decode
    BEGIN
        decoded := decode(translate(padded_input, '-_', '+/'), 'base64');
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION 'Invalid base64 encoding in tenant ID';
    END;
    
    -- Validate length
    IF octet_length(decoded) != 8 THEN
        RAISE EXCEPTION 'Invalid tenant ID length: got %, want 8', octet_length(decoded);
    END IF;
    
    RETURN decoded;
END;
$$;


--
-- Name: section_title_sort_key(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.section_title_sort_key(text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE STRICT
    AS $_$
DECLARE
    result text := '';
    matches text[];
    remainder text := $1;
BEGIN
    WHILE remainder ~ '\d+' LOOP
        -- Extract text before the number
        result := result || substring(remainder FROM '^[^\d]*');

        -- Extract and pad the next number
        matches := regexp_matches(remainder, '(\d+)', '');  -- captures the first number
        IF matches IS NOT NULL THEN
            result := result || lpad(matches[1], 10, '0');
            -- Remove processed part from remainder
            remainder := substring(remainder FROM '\d+(.*)$');
        ELSE
            EXIT;
        END IF;
    END LOOP;

    -- Append any remaining non-digit text
    result := result || remainder;

    RETURN result;
END;
$_$;


--
-- Name: FUNCTION section_title_sort_key(text); Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON FUNCTION public.section_title_sort_key(text) IS 'Converts numbers in strings to zero-padded format for natural sorting';


--
-- Name: tenant_id_to_string(bytea); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.tenant_id_to_string(tenant_id bytea) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    -- Make sure to handle padding properly - remove trailing '=' characters
    RETURN rtrim(translate(encode(tenant_id, 'base64'), '+/', '-_'), '=');
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: access_review_campaign_source_fetch_attempts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.access_review_campaign_source_fetch_attempts (
    id text NOT NULL,
    tenant_id text NOT NULL,
    access_review_campaign_source_id text CONSTRAINT access_review_campaign_sour_access_review_campaign_sou_not_null NOT NULL,
    attempt_number integer CONSTRAINT access_review_campaign_source_fetch_att_attempt_number_not_null NOT NULL,
    status public.access_review_campaign_source_fetch_status NOT NULL,
    fetched_accounts_count integer CONSTRAINT access_review_campaign_source__fetched_accounts_count_not_null1 NOT NULL,
    error text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone CONSTRAINT access_review_campaign_source_fetch_attempt_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT access_review_campaign_source_fetch_attempt_updated_at_not_null NOT NULL,
    organization_id text CONSTRAINT access_review_campaign_source_fetch_at_organization_id_not_null NOT NULL
);


--
-- Name: access_review_campaign_sources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.access_review_campaign_sources (
    id text NOT NULL,
    tenant_id text NOT NULL,
    access_review_campaign_id text CONSTRAINT access_review_campaign_sour_access_review_campaign_id_not_null1 NOT NULL,
    access_review_source_id text,
    name text NOT NULL,
    connector_id text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    csv_data text,
    organization_id text NOT NULL,
    connector_account_id text
);


--
-- Name: access_review_campaigns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.access_review_campaigns (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    name text NOT NULL,
    status public.access_review_campaign_status NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    description text NOT NULL
);


--
-- Name: access_review_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.access_review_entries (
    id text CONSTRAINT access_entries_id_not_null NOT NULL,
    tenant_id text CONSTRAINT access_entries_tenant_id_not_null NOT NULL,
    access_review_campaign_id text CONSTRAINT access_entries_access_review_campaign_id_not_null NOT NULL,
    identity_id text,
    email text CONSTRAINT access_entries_email_not_null NOT NULL,
    full_name text CONSTRAINT access_entries_full_name_not_null NOT NULL,
    job_title text CONSTRAINT access_entries_job_title_not_null NOT NULL,
    is_admin boolean,
    mfa_status public.mfa_status CONSTRAINT access_entries_mfa_status_not_null NOT NULL,
    auth_method public.auth_method CONSTRAINT access_entries_auth_method_not_null NOT NULL,
    last_login timestamp with time zone,
    account_created_at timestamp with time zone,
    external_id text CONSTRAINT access_entries_external_id_not_null NOT NULL,
    account_key text CONSTRAINT access_entries_account_key_not_null NOT NULL,
    decision public.access_review_entry_decision CONSTRAINT access_entries_decision_not_null NOT NULL,
    decision_note text,
    decided_by text,
    decided_at timestamp with time zone,
    created_at timestamp with time zone CONSTRAINT access_entries_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT access_entries_updated_at_not_null NOT NULL,
    account_type text DEFAULT 'USER'::text CONSTRAINT access_entries_account_type_not_null NOT NULL,
    flags text[] DEFAULT '{}'::text[] CONSTRAINT access_entries_flags_not_null NOT NULL,
    flag_reasons text[] DEFAULT '{}'::text[] CONSTRAINT access_entries_flag_reasons_not_null NOT NULL,
    organization_id text CONSTRAINT access_entries_organization_id_not_null NOT NULL,
    active boolean CONSTRAINT access_entries_active_not_null NOT NULL,
    access_review_campaign_source_id text CONSTRAINT access_entries_access_review_campaign_source_id_not_null NOT NULL,
    roles text[] DEFAULT '{}'::text[] NOT NULL
);


--
-- Name: access_review_entry_decision_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.access_review_entry_decision_history (
    id text CONSTRAINT access_entry_decision_history_id_not_null NOT NULL,
    tenant_id text CONSTRAINT access_entry_decision_history_tenant_id_not_null NOT NULL,
    access_review_entry_id text CONSTRAINT access_entry_decision_history_access_entry_id_not_null NOT NULL,
    decision public.access_review_entry_decision CONSTRAINT access_entry_decision_history_decision_not_null NOT NULL,
    decision_note text,
    decided_by text,
    decided_at timestamp with time zone CONSTRAINT access_entry_decision_history_decided_at_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT access_entry_decision_history_created_at_not_null NOT NULL,
    organization_id text CONSTRAINT access_entry_decision_history_organization_id_not_null NOT NULL
);


--
-- Name: access_review_sources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.access_review_sources (
    id text CONSTRAINT access_sources_id_not_null NOT NULL,
    tenant_id text CONSTRAINT access_sources_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT access_sources_organization_id_not_null NOT NULL,
    connector_id text,
    name text CONSTRAINT access_sources_name_not_null NOT NULL,
    csv_data text,
    created_at timestamp with time zone CONSTRAINT access_sources_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT access_sources_updated_at_not_null NOT NULL,
    name_synced_at timestamp with time zone,
    name_sync_attempts integer DEFAULT 0 NOT NULL,
    name_sync_next_attempt_at timestamp with time zone,
    connector_account_id text
);


--
-- Name: agent_execution_anchors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.agent_execution_anchors (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    agent_execution_id text CONSTRAINT agent_execution_anchors_agent_run_id_not_null NOT NULL,
    provider text NOT NULL,
    external_conversation_id text NOT NULL,
    external_message_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: agent_executions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.agent_executions (
    id text CONSTRAINT agent_runs_id_not_null NOT NULL,
    tenant_id text CONSTRAINT agent_runs_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT agent_runs_organization_id_not_null NOT NULL,
    start_agent_name text CONSTRAINT agent_runs_start_agent_name_not_null NOT NULL,
    status text CONSTRAINT agent_runs_status_not_null NOT NULL,
    checkpoint jsonb,
    error_message text,
    started_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() CONSTRAINT agent_runs_created_at_not_null NOT NULL,
    updated_at timestamp with time zone DEFAULT now() CONSTRAINT agent_runs_updated_at_not_null NOT NULL,
    source text,
    session_key text,
    source_coordinates jsonb,
    session_messages jsonb CONSTRAINT agent_runs_session_messages_not_null NOT NULL,
    processing_owner_token text,
    processing_heartbeat_at timestamp with time zone,
    processing_input_ids text[] CONSTRAINT agent_runs_processing_input_ids_not_null NOT NULL,
    attempt_count integer CONSTRAINT agent_runs_attempt_count_not_null NOT NULL,
    max_attempts integer CONSTRAINT agent_runs_max_attempts_not_null NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    dead_lettered_at timestamp with time zone
);


--
-- Name: agent_inputs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.agent_inputs (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    agent_execution_id text CONSTRAINT agent_inputs_agent_run_id_not_null NOT NULL,
    source text NOT NULL,
    source_event_id text,
    message jsonb NOT NULL,
    processed_at timestamp with time zone,
    attempt_count integer NOT NULL,
    max_attempts integer NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    dead_lettered_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    purpose public.agent_input_purpose NOT NULL,
    identity_id text,
    source_coordinates jsonb
);


--
-- Name: ai_systems; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_systems (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    name text NOT NULL,
    version text,
    company_roles text[] NOT NULL,
    status public.ai_system_statuses NOT NULL,
    owner_id text,
    source text,
    purpose text,
    intended_use_cases text,
    autonomy_level text,
    human_oversight_mechanism text,
    risk_classification public.ai_system_risk_classifications NOT NULL,
    key_stakeholders text,
    data_sources_and_type text,
    deployment_date date,
    last_review_date date,
    next_review_date date,
    notes text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: applicability_statements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.applicability_statements (
    id text CONSTRAINT states_of_applicability_controls_id_not_null NOT NULL,
    statement_of_applicability_id text CONSTRAINT states_of_applicability_cont_state_of_applicability_id_not_null NOT NULL,
    control_id text CONSTRAINT states_of_applicability_controls_control_id_not_null NOT NULL,
    organization_id text CONSTRAINT states_of_applicability_controls_organization_id_not_null NOT NULL,
    tenant_id text CONSTRAINT states_of_applicability_controls_tenant_id_not_null NOT NULL,
    applicability boolean CONSTRAINT states_of_applicability_controls_applicability_not_null NOT NULL,
    justification text,
    created_at timestamp with time zone CONSTRAINT states_of_applicability_controls_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT states_of_applicability_controls_updated_at_not_null NOT NULL
);


--
-- Name: asana_webhook_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.asana_webhook_events (
    delivery_id text NOT NULL,
    envelope jsonb NOT NULL,
    processing_owner_token text,
    processing_started_at timestamp with time zone,
    processed_at timestamp with time zone,
    attempt_count integer NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    dead_lettered_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: asana_webhook_handshakes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.asana_webhook_handshakes (
    nonce text NOT NULL,
    connector_id text NOT NULL,
    secret_ciphertext bytea NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: asset_third_parties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.asset_third_parties (
    asset_id text CONSTRAINT asset_vendors_asset_id_not_null NOT NULL,
    third_party_id text CONSTRAINT asset_vendors_vendor_id_not_null NOT NULL,
    tenant_id text CONSTRAINT asset_vendors_tenant_id_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT asset_vendors_created_at_not_null NOT NULL,
    organization_id text CONSTRAINT asset_vendors_organization_id_not_null NOT NULL
);


--
-- Name: assets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assets (
    id text NOT NULL,
    tenant_id text NOT NULL,
    name text NOT NULL,
    amount integer NOT NULL,
    organization_id text NOT NULL,
    asset_type public.asset_type NOT NULL,
    data_types_stored text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    owner_profile_id text NOT NULL
);


--
-- Name: audit_log_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_log_entries (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    actor_id text NOT NULL,
    actor_type text NOT NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: audits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audits (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    valid_from date,
    valid_until date,
    state public.audit_state NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text,
    report_file_id text,
    audit_start_date date,
    audit_end_date date,
    firm text
);


--
-- Name: audits_frameworks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audits_frameworks (
    audit_id text NOT NULL,
    framework_id text NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: bot_delivery_destinations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bot_delivery_destinations (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    provider text NOT NULL,
    target_namespace text NOT NULL,
    target_key text NOT NULL,
    external_destination_id text NOT NULL,
    external_name text NOT NULL,
    verified_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: bot_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bot_messages (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    capability text NOT NULL,
    message_type text NOT NULL,
    attributes jsonb NOT NULL,
    subject_namespace text NOT NULL,
    subject_key text NOT NULL,
    event_key text NOT NULL,
    purpose public.bot_message_purpose NOT NULL,
    processing_started_at timestamp with time zone,
    processed_at timestamp with time zone,
    attempt_count integer NOT NULL,
    max_attempts integer NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    dead_lettered_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    processing_owner_token uuid
);


--
-- Name: bot_thread_subjects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bot_thread_subjects (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    provider text NOT NULL,
    external_conversation_id text NOT NULL,
    external_message_id text NOT NULL,
    capability text NOT NULL,
    message_type text NOT NULL,
    attributes jsonb NOT NULL,
    subject_namespace text NOT NULL,
    subject_key text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: business_function_assets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.business_function_assets (
    tenant_id text NOT NULL,
    business_function_id text NOT NULL,
    asset_id text NOT NULL,
    organization_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: business_function_third_parties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.business_function_third_parties (
    tenant_id text NOT NULL,
    business_function_id text NOT NULL,
    third_party_id text NOT NULL,
    organization_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: business_functions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.business_functions (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    reference_id text NOT NULL,
    name text NOT NULL,
    classification public.business_function_classifications NOT NULL,
    mtd_minutes integer NOT NULL,
    rto_minutes integer NOT NULL,
    rpo_minutes integer NOT NULL,
    impact_tolerance text,
    notes text,
    owner_id text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: cached_certificates; Type: TABLE; Schema: public; Owner: -
--

CREATE UNLOGGED TABLE public.cached_certificates (
    domain public.citext NOT NULL,
    certificate_pem text NOT NULL,
    private_key_pem text NOT NULL,
    certificate_chain text,
    expires_at timestamp with time zone NOT NULL,
    cached_at timestamp with time zone NOT NULL,
    certificate_id text NOT NULL
);


--
-- Name: certificates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.certificates (
    id text NOT NULL,
    tenant_id text NOT NULL,
    hostname public.citext NOT NULL,
    ssl_certificate bytea,
    encrypted_ssl_private_key bytea,
    ssl_certificate_chain text,
    status public.custom_domain_ssl_status NOT NULL,
    ssl_expires_at timestamp with time zone,
    ssl_retry_count integer NOT NULL,
    ssl_last_attempt_at timestamp with time zone,
    http_challenge_token text,
    http_challenge_key_auth text,
    http_challenge_url text,
    http_order_url text,
    provisioning_error text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: cmp_domain_surfaces; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cmp_domain_surfaces (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    cmp_domain_id text NOT NULL,
    origin text NOT NULL,
    status text NOT NULL,
    discovered_by_cookie_banner_id text,
    last_cookie_banner_scan_id text,
    last_seen_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT cmp_domain_surfaces_status_check CHECK ((status = ANY (ARRAY['UNCONFIRMED'::text, 'IGNORED'::text, 'VENDOR_RUN'::text])))
);


--
-- Name: cmp_domains; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cmp_domains (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    domain public.citext NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: collaboration_actors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.collaboration_actors (
    tenant_id text NOT NULL,
    document_id text NOT NULL,
    organization_id text NOT NULL,
    actor_id bytea NOT NULL,
    identity_id text NOT NULL,
    profile_id text,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: collaboration_changes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.collaboration_changes (
    tenant_id text NOT NULL,
    document_id text NOT NULL,
    organization_id text NOT NULL,
    revision bigint NOT NULL,
    change_hash bytea NOT NULL,
    change_bytes bytea NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT collaboration_changes_change_hash_check CHECK ((octet_length(change_hash) = 32))
);


--
-- Name: collaboration_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.collaboration_documents (
    tenant_id text NOT NULL,
    document_id text NOT NULL,
    organization_id text NOT NULL,
    snapshot bytea NOT NULL,
    heads bytea NOT NULL,
    revision bigint NOT NULL,
    snapshot_revision bigint NOT NULL,
    change_revision bigint NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: common_gvl_snapshots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.common_gvl_snapshots (
    id text NOT NULL,
    vendor_list_version integer NOT NULL,
    gvl_specification_version integer NOT NULL,
    tcf_policy_version integer NOT NULL,
    last_updated timestamp with time zone NOT NULL,
    fetched_at timestamp with time zone NOT NULL,
    payload jsonb NOT NULL
);


--
-- Name: common_gvl_state; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.common_gvl_state (
    singleton boolean DEFAULT true NOT NULL,
    latest_vendor_list_version integer,
    etag text,
    cache_max_age_seconds integer,
    last_fetched_at timestamp with time zone NOT NULL,
    CONSTRAINT common_gvl_state_singleton_check CHECK (singleton)
);


--
-- Name: common_gvl_vendors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.common_gvl_vendors (
    id text NOT NULL,
    iab_vendor_id integer NOT NULL,
    vendor_list_version integer NOT NULL,
    name text NOT NULL,
    deleted_date timestamp with time zone,
    purposes integer[] DEFAULT '{}'::integer[] NOT NULL,
    leg_int_purposes integer[] DEFAULT '{}'::integer[] NOT NULL,
    flexible_purposes integer[] DEFAULT '{}'::integer[] NOT NULL,
    special_purposes integer[] DEFAULT '{}'::integer[] NOT NULL,
    features integer[] DEFAULT '{}'::integer[] NOT NULL,
    special_features integer[] DEFAULT '{}'::integer[] NOT NULL,
    policy_url text,
    uses_cookies boolean,
    cookie_refresh boolean,
    uses_non_cookie_access boolean,
    cookie_max_age_seconds integer,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: common_ip_country_blocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.common_ip_country_blocks (
    cidr cidr NOT NULL,
    country_code character(2) NOT NULL
);


--
-- Name: common_ip_location_blocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.common_ip_location_blocks (
    address_family smallint NOT NULL,
    ip_start inet NOT NULL,
    ip_end inet NOT NULL,
    country_code character(2) NOT NULL,
    subdivision_code character varying(15),
    CONSTRAINT common_ip_location_blocks_address_family_check CHECK ((address_family = ANY (ARRAY[4, 6]))),
    CONSTRAINT common_ip_location_blocks_check CHECK ((family(ip_start) = address_family)),
    CONSTRAINT common_ip_location_blocks_check1 CHECK ((family(ip_end) = address_family)),
    CONSTRAINT common_ip_location_blocks_check2 CHECK ((ip_start <= ip_end))
);


--
-- Name: common_third_parties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.common_third_parties (
    id text NOT NULL,
    name text NOT NULL,
    category public.third_party_category NOT NULL,
    headquarter_address text,
    legal_name text,
    website_url text,
    privacy_policy_url text,
    service_level_agreement_url text,
    service_software_agreement_url text,
    data_processing_agreement_url text,
    business_associate_agreement_url text,
    subprocessors_list_url text,
    certifications text[],
    status_page_url text,
    terms_of_service_url text,
    security_page_url text,
    trust_page_url text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    logo_file_id text,
    slug text NOT NULL,
    enrichment_requested_at timestamp with time zone,
    enrichment jsonb,
    enrichment_attempts integer NOT NULL,
    last_enrichment_attempt_at timestamp with time zone,
    review public.common_third_party_review NOT NULL,
    rejected_verdict public.common_tracker_pattern_attribution,
    reviewed_at timestamp with time zone
);


--
-- Name: common_third_party_domains; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.common_third_party_domains (
    id text NOT NULL,
    common_third_party_id text NOT NULL,
    domain public.citext NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: common_tracker_patterns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.common_tracker_patterns (
    id text NOT NULL,
    common_third_party_id text,
    tracker_type public.tracker_type NOT NULL,
    pattern text NOT NULL,
    match_type public.cookie_pattern_match_type NOT NULL,
    description text NOT NULL,
    max_age_seconds integer,
    confidence real NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    enrichment_requested_at timestamp with time zone,
    last_enrichment_attempt_at timestamp with time zone,
    enrichment jsonb,
    enrichment_attempts integer NOT NULL,
    attribution public.common_tracker_pattern_attribution NOT NULL
);


--
-- Name: compliance_custom_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.compliance_custom_links (
    id text CONSTRAINT compliance_external_urls_id_not_null NOT NULL,
    tenant_id text CONSTRAINT compliance_external_urls_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT compliance_external_urls_organization_id_not_null NOT NULL,
    compliance_portal_id text CONSTRAINT compliance_external_urls_compliance_portal_id_not_null NOT NULL,
    name text CONSTRAINT compliance_external_urls_name_not_null NOT NULL,
    url text CONSTRAINT compliance_external_urls_url_not_null NOT NULL,
    rank integer CONSTRAINT compliance_external_urls_rank_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT compliance_external_urls_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT compliance_external_urls_updated_at_not_null NOT NULL
);


--
-- Name: compliance_frameworks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.compliance_frameworks (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    compliance_portal_id text NOT NULL,
    framework_id text NOT NULL,
    rank integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: compliance_portal_commitment_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.compliance_portal_commitment_groups (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    compliance_portal_id text CONSTRAINT compliance_portal_commitment_groups_compliance_portal_id_not_nu NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    rank integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: compliance_portal_commitments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.compliance_portal_commitments (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    compliance_portal_id text NOT NULL,
    group_id text NOT NULL,
    icon text NOT NULL,
    eyebrow text NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    rank integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: compliance_portals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.compliance_portals (
    id text NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    active boolean NOT NULL,
    slug text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    non_disclosure_agreement_file_id text,
    logo_file_id text,
    dark_logo_file_id text,
    mailing_list_id text NOT NULL,
    search_engine_indexing public.search_engine_indexing NOT NULL,
    default_domain_id text,
    custom_domain_id text,
    description text,
    website_url text,
    email public.citext,
    headquarter_address text,
    entity_name text NOT NULL,
    capabilities jsonb NOT NULL,
    CONSTRAINT compliance_portals_slug_check CHECK ((slug ~ '^[a-z0-9_-]+$'::text))
);


--
-- Name: connector_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.connector_accounts (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    connector_id text NOT NULL,
    external_account_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name_synced_at timestamp with time zone,
    name_sync_attempts integer DEFAULT 0 NOT NULL,
    name_sync_next_attempt_at timestamp with time zone
);


--
-- Name: connector_install_state_claims; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.connector_install_state_claims (
    state_digest bytea NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    processing_token uuid,
    processing_started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: connectors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.connectors (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    encrypted_connection bytea NOT NULL,
    protocol public.connector_protocol NOT NULL,
    provider public.connector_provider NOT NULL,
    settings jsonb,
    name text NOT NULL
);


--
-- Name: controls; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.controls (
    id text NOT NULL,
    tenant_id text NOT NULL,
    framework_id text NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    section_title text CONSTRAINT controls_reference_id_not_null NOT NULL,
    organization_id text NOT NULL,
    best_practice boolean NOT NULL,
    not_implemented_justification text,
    maturity_level public.control_maturity_level NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, translate(((((COALESCE(section_title, ''::text) || ' '::text) || COALESCE(name, ''::text)) || ' '::text) || COALESCE(description, ''::text)), '-+'::text, '  '::text))) STORED
);


--
-- Name: controls_audits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.controls_audits (
    control_id text NOT NULL,
    audit_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    organization_id text NOT NULL
);


--
-- Name: controls_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.controls_documents (
    control_id text CONSTRAINT controls_policies_control_id_not_null NOT NULL,
    document_id text CONSTRAINT controls_policies_policy_id_not_null NOT NULL,
    tenant_id text CONSTRAINT controls_policies_tenant_id_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT controls_policies_created_at_not_null NOT NULL,
    organization_id text NOT NULL
);


--
-- Name: controls_internal_controls; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.controls_internal_controls (
    control_id text CONSTRAINT controls_mitigations_control_id_not_null NOT NULL,
    internal_control_id text CONSTRAINT controls_mitigations_mitigation_id_not_null NOT NULL,
    tenant_id text CONSTRAINT controls_mitigations_tenant_id_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT controls_mitigations_created_at_not_null NOT NULL,
    organization_id text NOT NULL
);


--
-- Name: controls_obligations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.controls_obligations (
    control_id text NOT NULL,
    obligation_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: cookie_banner_daily_page_loads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookie_banner_daily_page_loads (
    cookie_banner_id text NOT NULL,
    day date NOT NULL,
    browser_family text NOT NULL,
    presentation text NOT NULL,
    consent_action text NOT NULL,
    tenant_id text NOT NULL,
    count integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: cookie_banner_discovery_stats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookie_banner_discovery_stats (
    cookie_banner_id text NOT NULL,
    tenant_id text NOT NULL,
    chrome_page_loads integer DEFAULT 0 NOT NULL,
    edge_page_loads integer DEFAULT 0 NOT NULL,
    firefox_page_loads integer DEFAULT 0 NOT NULL,
    safari_page_loads integer DEFAULT 0 NOT NULL,
    other_page_loads integer DEFAULT 0 NOT NULL,
    frozen_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: cookie_banner_gvl_vendors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookie_banner_gvl_vendors (
    cookie_banner_id text NOT NULL,
    iab_vendor_id integer NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: cookie_banner_scans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookie_banner_scans (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    state public.cookie_banner_scan_state NOT NULL,
    failure_reason public.cookie_banner_scan_failure_reason,
    truncated boolean NOT NULL,
    pages_scanned integer NOT NULL,
    cookies_found integer NOT NULL,
    storage_found integer NOT NULL,
    resources_found integer NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    checkpoint jsonb
);


--
-- Name: cookie_banner_translations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookie_banner_translations (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    language text NOT NULL,
    translations jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: cookie_banner_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookie_banner_versions (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    version integer NOT NULL,
    state public.cookie_banner_version_state NOT NULL,
    snapshot jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: cookie_banners; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookie_banners (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    name text NOT NULL,
    origin text NOT NULL,
    state public.cookie_banner_state NOT NULL,
    privacy_policy_url text,
    consent_expiry_days integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    show_branding boolean NOT NULL,
    default_language text NOT NULL,
    cookie_policy_url text NOT NULL,
    pattern_analysis_requested_at timestamp with time zone,
    policy_document_id text,
    policy_generation_requested_at timestamp with time zone,
    capabilities jsonb NOT NULL,
    publisher_country_code character varying(2) NOT NULL,
    force_opt_in boolean NOT NULL,
    cmp_domain_id text,
    source_promotion_requested_at timestamp with time zone,
    source_promotion_since timestamp with time zone,
    source_promotion_attempts integer NOT NULL,
    source_promotion_attempted_at timestamp with time zone
);


--
-- Name: cookie_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookie_categories (
    id text NOT NULL,
    tenant_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    rank integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    organization_id text NOT NULL,
    kind public.cookie_category_kind NOT NULL,
    slug text NOT NULL,
    gcm_consent_types text[] NOT NULL,
    posthog_consent boolean NOT NULL,
    tcf_purpose_ids integer[] NOT NULL
);


--
-- Name: cookie_consent_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookie_consent_records (
    id text NOT NULL,
    tenant_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    visitor_id text NOT NULL,
    ip_address text,
    user_agent text,
    consent_data jsonb NOT NULL,
    action public.cookie_consent_action NOT NULL,
    created_at timestamp with time zone NOT NULL,
    organization_id text NOT NULL,
    cookie_banner_version_id text NOT NULL,
    sdk_version text NOT NULL,
    regulation text,
    country_code text,
    consent_mode public.cookie_consent_mode,
    regulation_source public.regulation_source,
    subdivision_code character varying(15),
    tc text,
    origin text
);


--
-- Name: cp_accesses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cp_accesses (
    id text NOT NULL,
    tenant_id text NOT NULL,
    compliance_portal_id text NOT NULL,
    email public.citext,
    name text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    organization_id text NOT NULL,
    state public.trust_center_access_state NOT NULL,
    electronic_signature_id text,
    identity_id text NOT NULL,
    authenticated_at timestamp with time zone
);


--
-- Name: cp_audits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cp_audits (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    compliance_portal_id text NOT NULL,
    audit_id text NOT NULL,
    visibility public.compliance_portal_visibility NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: cp_document_accesses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cp_document_accesses (
    id text NOT NULL,
    tenant_id text NOT NULL,
    compliance_portal_access_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    compliance_portal_file_id text,
    organization_id text NOT NULL,
    status public.compliance_portal_document_access_status NOT NULL,
    requested_at timestamp with time zone,
    compliance_portal_document_id text,
    compliance_portal_audit_id text,
    CONSTRAINT cp_document_accesses_check CHECK ((((((compliance_portal_document_id IS NOT NULL))::integer + ((compliance_portal_audit_id IS NOT NULL))::integer) + ((compliance_portal_file_id IS NOT NULL))::integer) = 1))
);


--
-- Name: cp_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cp_documents (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    compliance_portal_id text NOT NULL,
    document_id text NOT NULL,
    visibility public.compliance_portal_visibility NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: cp_files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cp_files (
    id text NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    name text NOT NULL,
    category text NOT NULL,
    file_id text NOT NULL,
    compliance_portal_visibility public.compliance_portal_visibility NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    compliance_portal_id text NOT NULL
);


--
-- Name: cp_references; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cp_references (
    id text NOT NULL,
    tenant_id text NOT NULL,
    compliance_portal_id text NOT NULL,
    name text NOT NULL,
    description text,
    website_url text NOT NULL,
    logo_file_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    rank integer NOT NULL,
    organization_id text NOT NULL
);


--
-- Name: cp_third_parties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cp_third_parties (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    compliance_portal_id text NOT NULL,
    third_party_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: custom_domains; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.custom_domains (
    id text NOT NULL,
    tenant_id text NOT NULL,
    domain public.citext NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    organization_id text NOT NULL,
    managed boolean NOT NULL,
    certificate_id text
);


--
-- Name: data; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.data (
    id text NOT NULL,
    tenant_id text NOT NULL,
    name text NOT NULL,
    organization_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    data_classification public.data_classification_type NOT NULL,
    owner_profile_id text NOT NULL
);


--
-- Name: data_third_parties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.data_third_parties (
    datum_id text CONSTRAINT data_vendors_datum_id_not_null NOT NULL,
    third_party_id text CONSTRAINT data_vendors_vendor_id_not_null NOT NULL,
    tenant_id text CONSTRAINT data_vendors_tenant_id_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT data_vendors_created_at_not_null NOT NULL,
    organization_id text CONSTRAINT data_vendors_organization_id_not_null NOT NULL
);


--
-- Name: detected_trackers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.detected_trackers (
    id text NOT NULL,
    tenant_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    tracker_pattern_id text,
    tracker_type public.tracker_type NOT NULL,
    identifier text NOT NULL,
    max_age_seconds integer,
    source public.cookie_source,
    value_size integer,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    initiator_url text,
    initiator_domain public.citext,
    cookie_domain public.citext,
    host_only boolean,
    observed_by public.detected_tracker_observer NOT NULL,
    http_only boolean,
    secure boolean,
    same_site public.cookie_same_site,
    partitioned boolean,
    set_by_url text,
    cookie_banner_scan_id text
);


--
-- Name: device_enrollment_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.device_enrollment_tokens (
    id text NOT NULL,
    tenant_id text NOT NULL,
    device_id text NOT NULL,
    hashed_value bytea NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: device_postures; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.device_postures (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    device_id text NOT NULL,
    check_key text NOT NULL,
    status public.device_posture_status NOT NULL,
    evidence jsonb DEFAULT '{}'::jsonb NOT NULL,
    observed_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    correlation_id text NOT NULL,
    version text NOT NULL,
    agent_version text NOT NULL
);


--
-- Name: devices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.devices (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    state public.device_state DEFAULT 'PENDING'::public.device_state NOT NULL,
    hardware_uuid text,
    serial_number text,
    hostname text,
    platform public.device_platform,
    os_version text,
    agent_version text,
    api_key_hash bytea,
    owner_profile_id text,
    labels jsonb DEFAULT '{}'::jsonb NOT NULL,
    enrolled_at timestamp with time zone,
    last_seen_at timestamp with time zone,
    revoked_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT devices_active_api_key_hash_check CHECK (((state <> 'ACTIVE'::public.device_state) OR (api_key_hash IS NOT NULL))),
    CONSTRAINT devices_active_fields_check CHECK (((state <> 'ACTIVE'::public.device_state) OR ((hardware_uuid IS NOT NULL) AND (hostname IS NOT NULL) AND (platform IS NOT NULL) AND (os_version IS NOT NULL) AND (agent_version IS NOT NULL) AND (enrolled_at IS NOT NULL) AND (last_seen_at IS NOT NULL)))),
    CONSTRAINT devices_deleted_at_check CHECK (((deleted_at IS NULL) OR (state = 'REVOKED'::public.device_state))),
    CONSTRAINT devices_revoked_at_check CHECK ((((state = 'REVOKED'::public.device_state) AND (revoked_at IS NOT NULL)) OR ((state <> 'REVOKED'::public.device_state) AND (revoked_at IS NULL))))
);


--
-- Name: document_comment_threads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_comment_threads (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    document_id text NOT NULL,
    suggestion_id text,
    anchor_start bytea,
    anchor_end bytea,
    anchor_quote text NOT NULL,
    resolved_at timestamp with time zone,
    resolved_by_profile_id text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: document_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_comments (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    thread_id text NOT NULL,
    author_profile_id text NOT NULL,
    content jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: document_default_approvers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_default_approvers (
    document_id text NOT NULL,
    approver_profile_id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: document_suggestions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_suggestions (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    document_id text NOT NULL,
    author_profile_id text NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resolved_at timestamp with time zone,
    resolved_by_profile_id text
);


--
-- Name: document_version_approval_decisions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_version_approval_decisions (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    quorum_id text NOT NULL,
    approver_id text NOT NULL,
    state public.document_version_approval_decision_state NOT NULL,
    comment text,
    electronic_signature_id text,
    decided_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: document_version_approval_quorums; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_version_approval_quorums (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    version_id text NOT NULL,
    status public.document_version_approval_quorum_status NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    file_id text,
    pdf_attempt_count integer NOT NULL,
    pdf_claimed_at timestamp with time zone
);


--
-- Name: document_version_signatures; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_version_signatures (
    id text CONSTRAINT policy_version_signatures_id_not_null NOT NULL,
    tenant_id text CONSTRAINT policy_version_signatures_tenant_id_not_null NOT NULL,
    document_version_id text CONSTRAINT policy_version_signatures_policy_version_id_not_null NOT NULL,
    state public.policy_version_signature_state CONSTRAINT policy_version_signatures_state_not_null NOT NULL,
    signed_at timestamp with time zone,
    requested_at timestamp with time zone CONSTRAINT policy_version_signatures_requested_at_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT policy_version_signatures_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT policy_version_signatures_updated_at_not_null NOT NULL,
    organization_id text NOT NULL,
    signed_by_profile_id text NOT NULL,
    electronic_signature_id text
);


--
-- Name: document_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.document_versions (
    id text CONSTRAINT policy_versions_id_not_null NOT NULL,
    tenant_id text CONSTRAINT policy_versions_tenant_id_not_null NOT NULL,
    document_id text CONSTRAINT policy_versions_policy_id_not_null NOT NULL,
    major integer CONSTRAINT policy_versions_version_number_not_null NOT NULL,
    content text CONSTRAINT policy_versions_content_not_null NOT NULL,
    changelog text CONSTRAINT policy_versions_changelog_not_null NOT NULL,
    published_at timestamp with time zone,
    created_at timestamp with time zone CONSTRAINT policy_versions_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT policy_versions_updated_at_not_null NOT NULL,
    status public.document_version_status,
    title text NOT NULL,
    classification public.document_classification NOT NULL,
    organization_id text NOT NULL,
    minor integer NOT NULL,
    document_type public.document_type NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, COALESCE(title, ''::text))) STORED,
    orientation public.document_version_orientation,
    file_id text,
    pdf_attempt_count integer NOT NULL
);


--
-- Name: documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.documents (
    id text CONSTRAINT policies_id_not_null NOT NULL,
    organization_id text CONSTRAINT policies_organization_id_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT policies_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT policies_updated_at_not_null NOT NULL,
    tenant_id text CONSTRAINT policies_tenant_id_not_null NOT NULL,
    current_published_major integer,
    deleted_at timestamp with time zone,
    archived_at timestamp with time zone,
    status public.document_status NOT NULL,
    current_published_minor integer,
    write_mode public.document_write_mode NOT NULL,
    title text NOT NULL,
    classification public.document_classification NOT NULL,
    document_type public.document_type NOT NULL,
    orientation public.document_version_orientation NOT NULL,
    content text NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, translate(COALESCE(title, ''::text), '-+'::text, '  '::text))) STORED
);


--
-- Name: electronic_signature_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.electronic_signature_events (
    id text NOT NULL,
    tenant_id text NOT NULL,
    electronic_signature_id text NOT NULL,
    event_type public.electronic_signature_event_type NOT NULL,
    event_source public.electronic_signature_event_source NOT NULL,
    actor_email public.citext NOT NULL,
    actor_ip_address text NOT NULL,
    actor_user_agent text NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: electronic_signatures; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.electronic_signatures (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    status public.electronic_signature_status DEFAULT 'PENDING'::public.electronic_signature_status NOT NULL,
    document_type public.electronic_signature_document_type NOT NULL,
    file_id text NOT NULL,
    signer_email public.citext NOT NULL,
    consent_text text NOT NULL,
    signer_full_name text,
    signer_ip_address text,
    signer_user_agent text,
    file_hash text,
    seal text,
    seal_version integer DEFAULT 1 NOT NULL,
    tsa_token bytea,
    signed_at timestamp with time zone,
    certificate_file_id text,
    certificate_processing_started_at timestamp with time zone,
    attempt_count integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 10 NOT NULL,
    last_attempted_at timestamp with time zone,
    last_error text,
    processing_started_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    document_name text,
    email_subject text
);


--
-- Name: email_attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_attachments (
    id text NOT NULL,
    email_id text NOT NULL,
    file_id text NOT NULL,
    filename text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: emails; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.emails (
    id text NOT NULL,
    recipient_email text CONSTRAINT emails_email_address_to_not_null NOT NULL,
    subject text NOT NULL,
    text_body text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    sent_at timestamp with time zone,
    recipient_name text NOT NULL,
    html_body text,
    reply_to text,
    unsubscribe_url text,
    mailing_list_update_id text,
    status text DEFAULT 'PENDING'::text NOT NULL,
    processing_started_at timestamp with time zone,
    attempt_count integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 10 NOT NULL,
    last_attempted_at timestamp with time zone,
    last_error text,
    sender_name text
);


--
-- Name: employee_portal_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.employee_portal_documents (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    employee_portal_id text NOT NULL,
    document_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: employee_portals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.employee_portals (
    id text NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    name text NOT NULL,
    active boolean NOT NULL,
    capabilities jsonb NOT NULL,
    logo_file_id text,
    dark_logo_file_id text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: evidences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.evidences (
    id text NOT NULL,
    task_id text,
    object_key text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    mime_type text,
    size bigint,
    filename text,
    tenant_id text NOT NULL,
    state public.evidence_state,
    type public.evidence_type NOT NULL,
    url text NOT NULL,
    description text,
    reference_id text NOT NULL,
    internal_control_id text NOT NULL,
    evidence_file_id text,
    organization_id text NOT NULL,
    description_status public.evidence_description_status NOT NULL,
    description_processing_started_at timestamp with time zone
);


--
-- Name: export_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.export_jobs (
    id text NOT NULL,
    tenant_id text NOT NULL,
    type public.export_jobs_type NOT NULL,
    arguments jsonb NOT NULL,
    error text,
    status public.export_jobs_status NOT NULL,
    file_id text,
    recipient_email text NOT NULL,
    recipient_name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    organization_id text NOT NULL
);


--
-- Name: files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.files (
    id text NOT NULL,
    tenant_id text NOT NULL,
    bucket_name text NOT NULL,
    mime_type text NOT NULL,
    file_name text NOT NULL,
    file_key uuid NOT NULL,
    file_size bigint NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    organization_id text NOT NULL,
    visibility text NOT NULL
);


--
-- Name: findings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.findings (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    kind public.findings_kind NOT NULL,
    reference_id text NOT NULL,
    description text,
    source text,
    risk_id text,
    identified_on date,
    root_cause text,
    corrective_action text,
    owner_id text,
    due_date date,
    status public.findings_status NOT NULL,
    priority public.findings_priority NOT NULL,
    effectiveness_check text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    title text NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, translate(((COALESCE(reference_id, ''::text) || ' '::text) || COALESCE(title, ''::text)), '-+'::text, '  '::text))) STORED
);


--
-- Name: findings_audits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.findings_audits (
    finding_id text NOT NULL,
    audit_id text NOT NULL,
    reference_id text NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: findings_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.findings_tasks (
    finding_id text NOT NULL,
    task_id text NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: frameworks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.frameworks (
    id text NOT NULL,
    organization_id text NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    tenant_id text NOT NULL,
    reference_id text NOT NULL,
    light_logo_file_id text,
    dark_logo_file_id text
);


--
-- Name: generated_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.generated_documents (
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    data_document_id text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    asset_list_document_id text,
    findings_document_id text,
    obligations_document_id text,
    processing_activities_document_id text,
    data_protection_impact_assessments_document_id text,
    transfer_impact_assessments_document_id text,
    third_parties_document_id text,
    risks_document_id text,
    business_functions_document_id text,
    ai_systems_document_id text
);


--
-- Name: iam_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_invitations (
    id text CONSTRAINT authz_invitations_id_not_null NOT NULL,
    tenant_id text CONSTRAINT authz_invitations_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT authz_invitations_organization_id_not_null NOT NULL,
    expires_at timestamp with time zone CONSTRAINT authz_invitations_expires_at_not_null NOT NULL,
    accepted_at timestamp with time zone,
    created_at timestamp with time zone CONSTRAINT authz_invitations_created_at_not_null NOT NULL,
    user_id text NOT NULL
);


--
-- Name: iam_membership_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_membership_profiles (
    tenant_id text CONSTRAINT iam_identity_profiles_tenant_id_not_null NOT NULL,
    id text CONSTRAINT iam_identity_profiles_id_not_null NOT NULL,
    full_name text CONSTRAINT iam_identity_profiles_full_name_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT iam_identity_profiles_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT iam_identity_profiles_updated_at_not_null NOT NULL,
    identity_id text NOT NULL,
    organization_id text NOT NULL,
    additional_email_addresses public.citext[] NOT NULL,
    kind text,
    contract_start_date date,
    contract_end_date date,
    "position" text,
    state public.membership_state NOT NULL,
    source text DEFAULT 'MANUAL'::text NOT NULL,
    user_name public.citext,
    external_id text,
    nickname text,
    locale text,
    timezone text,
    profile_url text,
    preferred_language text,
    given_name text,
    family_name text,
    formatted_name text,
    middle_name text,
    honorific_prefix text,
    honorific_suffix text,
    employee_number text,
    department text,
    cost_center text,
    enterprise_organization text,
    division text,
    manager_value text,
    activated_at timestamp with time zone,
    deactivated_at timestamp with time zone,
    scim_removed_at timestamp with time zone,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, COALESCE(full_name, ''::text))) STORED
);


--
-- Name: iam_memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_memberships (
    id text CONSTRAINT authz_memberships_id_not_null NOT NULL,
    tenant_id text CONSTRAINT authz_memberships_tenant_id_not_null NOT NULL,
    identity_id text CONSTRAINT authz_memberships_user_id_not_null NOT NULL,
    organization_id text CONSTRAINT authz_memberships_organization_id_not_null NOT NULL,
    role public.authz_role CONSTRAINT authz_memberships_role_not_null NOT NULL,
    created_at timestamp without time zone CONSTRAINT authz_memberships_created_at_not_null NOT NULL,
    updated_at timestamp without time zone CONSTRAINT authz_memberships_updated_at_not_null NOT NULL
);


--
-- Name: iam_oauth2_access_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_oauth2_access_tokens (
    id text NOT NULL,
    hashed_value bytea NOT NULL,
    client_id text,
    identity_id text NOT NULL,
    scopes text[] NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    resources text[]
);


--
-- Name: iam_oauth2_authorization_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_oauth2_authorization_codes (
    id text NOT NULL,
    client_id text NOT NULL,
    identity_id text NOT NULL,
    redirect_uri text NOT NULL,
    scopes text[] NOT NULL,
    code_challenge text,
    code_challenge_method text,
    nonce text,
    auth_time timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    redeemed_at timestamp with time zone,
    access_token_id text,
    hashed_value bytea,
    resources text[]
);


--
-- Name: iam_oauth2_clients; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_oauth2_clients (
    id text NOT NULL,
    tenant_id text,
    organization_id text,
    client_secret_hash bytea,
    client_name text NOT NULL,
    visibility public.oauth2_client_visibility NOT NULL,
    redirect_uris text[] NOT NULL,
    scopes text[] NOT NULL,
    grant_types text[] NOT NULL,
    response_types text[] NOT NULL,
    token_endpoint_auth_method public.oauth2_client_token_endpoint_auth_method NOT NULL,
    logo_uri text,
    client_uri text,
    contacts text[],
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    external_client_id text
);


--
-- Name: iam_oauth2_consents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_oauth2_consents (
    id text NOT NULL,
    identity_id text NOT NULL,
    client_id text NOT NULL,
    scopes text[] NOT NULL,
    redirect_uri text,
    code_challenge text NOT NULL,
    code_challenge_method text NOT NULL,
    nonce text NOT NULL,
    state text NOT NULL,
    approved boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    session_id text NOT NULL,
    device_code_id text,
    resources text[]
);


--
-- Name: iam_oauth2_device_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_oauth2_device_codes (
    id text NOT NULL,
    device_code_hash bytea NOT NULL,
    user_code text NOT NULL,
    client_id text NOT NULL,
    scopes text[] NOT NULL,
    identity_id text,
    status public.oauth2_device_code_status NOT NULL,
    last_polled_at timestamp with time zone,
    poll_interval integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL
);


--
-- Name: iam_oauth2_refresh_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_oauth2_refresh_tokens (
    id text NOT NULL,
    hashed_value bytea NOT NULL,
    client_id text NOT NULL,
    identity_id text NOT NULL,
    scopes text[] NOT NULL,
    access_token_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    resources text[]
);


--
-- Name: iam_oidc_states; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_oidc_states (
    id text NOT NULL,
    provider public.iam_oidc_provider NOT NULL,
    nonce text NOT NULL,
    code_verifier text NOT NULL,
    continue_url text DEFAULT ''::text NOT NULL,
    created_at timestamp without time zone NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    organization_id text
);


--
-- Name: iam_saml_assertions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_saml_assertions (
    id text CONSTRAINT auth_saml_assertions_id_not_null NOT NULL,
    organization_id text CONSTRAINT auth_saml_assertions_organization_id_not_null NOT NULL,
    used_at timestamp without time zone CONSTRAINT auth_saml_assertions_used_at_not_null NOT NULL,
    expires_at timestamp without time zone CONSTRAINT auth_saml_assertions_expires_at_not_null NOT NULL
);


--
-- Name: iam_saml_configurations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_saml_configurations (
    id text CONSTRAINT auth_saml_configurations_id_not_null NOT NULL,
    tenant_id text CONSTRAINT auth_saml_configurations_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT auth_saml_configurations_organization_id_not_null NOT NULL,
    email_domain public.citext CONSTRAINT auth_saml_configurations_email_domain_not_null NOT NULL,
    enforcement_policy public.saml_enforcement_policy CONSTRAINT auth_saml_configurations_enforcement_policy_not_null NOT NULL,
    idp_entity_id text CONSTRAINT auth_saml_configurations_idp_entity_id_not_null NOT NULL,
    idp_sso_url text CONSTRAINT auth_saml_configurations_idp_sso_url_not_null NOT NULL,
    idp_certificate text CONSTRAINT auth_saml_configurations_idp_certificate_not_null NOT NULL,
    idp_metadata_url text,
    attribute_email text DEFAULT 'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress'::text CONSTRAINT auth_saml_configurations_attribute_email_not_null NOT NULL,
    attribute_firstname text DEFAULT 'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname'::text CONSTRAINT auth_saml_configurations_attribute_firstname_not_null NOT NULL,
    attribute_lastname text DEFAULT 'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname'::text CONSTRAINT auth_saml_configurations_attribute_lastname_not_null NOT NULL,
    attribute_role text DEFAULT 'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role'::text CONSTRAINT auth_saml_configurations_attribute_role_not_null NOT NULL,
    auto_signup_enabled boolean DEFAULT false CONSTRAINT auth_saml_configurations_auto_signup_enabled_not_null NOT NULL,
    domain_verification_token text,
    domain_verified_at timestamp without time zone,
    created_at timestamp without time zone CONSTRAINT auth_saml_configurations_created_at_not_null NOT NULL,
    updated_at timestamp without time zone CONSTRAINT auth_saml_configurations_updated_at_not_null NOT NULL
);


--
-- Name: iam_saml_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_saml_requests (
    id text CONSTRAINT auth_saml_requests_id_not_null NOT NULL,
    organization_id text CONSTRAINT auth_saml_requests_organization_id_not_null NOT NULL,
    created_at timestamp without time zone CONSTRAINT auth_saml_requests_created_at_not_null NOT NULL,
    expires_at timestamp without time zone CONSTRAINT auth_saml_requests_expires_at_not_null NOT NULL
);


--
-- Name: iam_scim_bridges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_scim_bridges (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    scim_configuration_id text NOT NULL,
    connector_id text,
    type text NOT NULL,
    state text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    last_synced_at timestamp with time zone,
    next_sync_at timestamp with time zone DEFAULT now(),
    sync_error text,
    consecutive_failures integer NOT NULL,
    total_sync_count integer NOT NULL,
    total_failure_count integer NOT NULL,
    excluded_user_names text[] NOT NULL
);


--
-- Name: iam_scim_configurations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_scim_configurations (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    hashed_token bytea NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: iam_scim_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_scim_events (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    scim_configuration_id text NOT NULL,
    method text NOT NULL,
    path text NOT NULL,
    request_body text,
    response_body text,
    status_code integer NOT NULL,
    error_message text,
    membership_id text,
    ip_address inet NOT NULL,
    created_at timestamp with time zone NOT NULL,
    user_name public.citext NOT NULL
);


--
-- Name: iam_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_sessions (
    id text CONSTRAINT usrmgr_sessions_id_not_null NOT NULL,
    identity_id text CONSTRAINT usrmgr_sessions_user_id_not_null NOT NULL,
    created_at timestamp without time zone CONSTRAINT usrmgr_sessions_created_at_not_null NOT NULL,
    updated_at timestamp without time zone CONSTRAINT usrmgr_sessions_updated_at_not_null NOT NULL,
    expired_at timestamp with time zone CONSTRAINT usrmgr_sessions_expired_at_not_null NOT NULL,
    data jsonb DEFAULT '{}'::jsonb CONSTRAINT usrmgr_sessions_data_not_null NOT NULL,
    expire_reason public.session_expire_reason,
    user_agent text DEFAULT 'SESSION_CREATED_BEFORE_USER_AGENT_COLUMN_ADDED'::text,
    ip_address inet DEFAULT '::1'::inet,
    tenant_id text,
    parent_session_id text,
    auth_method public.session_auth_method CONSTRAINT sessions_auth_method_not_null NOT NULL,
    authenticated_at timestamp without time zone CONSTRAINT sessions_authenticated_at_not_null NOT NULL,
    membership_id text,
    CONSTRAINT session_tenant_check CHECK ((((parent_session_id IS NULL) AND (tenant_id IS NULL)) OR ((parent_session_id IS NOT NULL) AND (tenant_id IS NOT NULL))))
);


--
-- Name: iam_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.iam_tokens (
    id text NOT NULL,
    hashed_value bytea NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.identities (
    id text CONSTRAINT usrmgr_users_id_not_null NOT NULL,
    created_at timestamp without time zone CONSTRAINT usrmgr_users_created_at_not_null NOT NULL,
    updated_at timestamp without time zone CONSTRAINT usrmgr_users_updated_at_not_null NOT NULL,
    email_address public.citext CONSTRAINT usrmgr_users_email_address_not_null NOT NULL,
    hashed_password bytea,
    organization_id text,
    email_address_verified boolean CONSTRAINT users_email_address_verified_not_null NOT NULL,
    saml_subject text,
    full_name text DEFAULT ''::text NOT NULL,
    locale text,
    avatar_file_id text
);


--
-- Name: internal_control_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.internal_control_events (
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    internal_control_id text NOT NULL,
    event_type public.internal_control_event_type NOT NULL,
    name text NOT NULL,
    category text NOT NULL,
    state text NOT NULL,
    internal_control_created_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    implementation_status public.internal_control_implementation_status,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, COALESCE(name, ''::text))) STORED
);


--
-- Name: internal_controls; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.internal_controls (
    id text CONSTRAINT mitigations_id_not_null NOT NULL,
    tenant_id text CONSTRAINT mitigations_tenant_id_not_null NOT NULL,
    name text CONSTRAINT mitigations_name_not_null NOT NULL,
    description text,
    reference_id text CONSTRAINT mitigations_content_ref_not_null NOT NULL,
    category text CONSTRAINT mitigations_category_not_null NOT NULL,
    state public.mitigation_state CONSTRAINT mitigations_state_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT mitigations_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT mitigations_updated_at_not_null NOT NULL,
    organization_id text CONSTRAINT mitigations_organization_id_not_null NOT NULL,
    code text,
    control_type public.internal_control_type,
    nature public.internal_control_nature,
    operating_mode public.internal_control_operating_mode,
    operating_frequency interval,
    operating_event text,
    evidence_cadence interval,
    testing_cadence interval,
    next_evidence_due timestamp with time zone,
    next_test_due timestamp with time zone,
    implementation_status public.internal_control_implementation_status NOT NULL,
    owner_profile_id text,
    reviewer_profile_id text,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, ((((COALESCE(name, ''::text) || ' '::text) || COALESCE(code, ''::text)) || ' '::text) || COALESCE(operating_event, ''::text)))) STORED,
    CONSTRAINT internal_controls_operating_frequency_shape_check CHECK ((((operating_mode IS NULL) AND (operating_frequency IS NULL) AND (operating_event IS NULL)) OR ((operating_mode IS NOT NULL) AND (operating_mode = 'CONTINUOUS'::public.internal_control_operating_mode) AND (operating_frequency IS NULL) AND (operating_event IS NULL)) OR ((operating_mode IS NOT NULL) AND (operating_mode = 'EVENT'::public.internal_control_operating_mode) AND (operating_frequency IS NULL)) OR ((operating_mode IS NOT NULL) AND (operating_mode = 'PERIODIC'::public.internal_control_operating_mode) AND (operating_frequency IS NOT NULL) AND (operating_event IS NULL)))),
    CONSTRAINT internal_controls_owner_reviewer_distinct_check CHECK (((owner_profile_id IS NULL) OR (reviewer_profile_id IS NULL) OR (owner_profile_id <> reviewer_profile_id)))
);


--
-- Name: internal_controls_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.internal_controls_documents (
    internal_control_id text NOT NULL,
    document_id text NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: internal_controls_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.internal_controls_tasks (
    internal_control_id text NOT NULL,
    task_id text NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    reference_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: internal_controls_third_parties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.internal_controls_third_parties (
    internal_control_id text NOT NULL,
    third_party_id text NOT NULL,
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: linear_webhook_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.linear_webhook_events (
    delivery_id text NOT NULL,
    envelope jsonb NOT NULL,
    processing_owner_token text,
    processing_started_at timestamp with time zone,
    processed_at timestamp with time zone,
    attempt_count integer NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    dead_lettered_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: mailing_list_subscribers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.mailing_list_subscribers (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    mailing_list_id text NOT NULL,
    full_name text NOT NULL,
    email public.citext NOT NULL,
    status public.mailing_list_subscriber_status DEFAULT 'PENDING'::public.mailing_list_subscriber_status NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: mailing_list_updates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.mailing_list_updates (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    mailing_list_id text NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    status public.mailing_list_update_status NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: mailing_lists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.mailing_lists (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    reply_to public.citext,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: notification_preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_preferences (
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    recipient_profile_id text NOT NULL,
    notification_type public.notification_type NOT NULL,
    email_delivery public.notification_email_delivery NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: notification_weekly_sends; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_weekly_sends (
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    recipient_profile_id text NOT NULL,
    week_start date NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    recipient_profile_id text NOT NULL,
    type public.notification_type NOT NULL,
    subject_id text NOT NULL,
    attributes jsonb NOT NULL,
    read_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    status public.notification_status NOT NULL,
    email_count integer NOT NULL,
    last_emailed_at timestamp with time zone
);


--
-- Name: obligations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.obligations (
    id text CONSTRAINT compliance_registries_id_not_null NOT NULL,
    tenant_id text CONSTRAINT compliance_registries_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT compliance_registries_organization_id_not_null NOT NULL,
    area text,
    source text,
    requirement text,
    actions_to_be_implemented text,
    regulator text,
    last_review_date date,
    due_date date,
    status public.obligations_status CONSTRAINT compliance_registries_status_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT compliance_registries_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT compliance_registries_updated_at_not_null NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, ((((((((COALESCE(requirement, ''::text) || ' '::text) || COALESCE(area, ''::text)) || ' '::text) || COALESCE(source, ''::text)) || ' '::text) || COALESCE(regulator, ''::text)) || ' '::text) || COALESCE(actions_to_be_implemented, ''::text)))) STORED,
    type public.obligation_type NOT NULL,
    owner_profile_id text NOT NULL
);


--
-- Name: operation_receipts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.operation_receipts (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    operation_key text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: organization_contexts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_contexts (
    organization_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    product text,
    architecture text,
    team text,
    processes text,
    customers text
);


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organizations (
    id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    tenant_id text NOT NULL,
    horizontal_logo_file_id text,
    logo_file_id text,
    legal_name text
);


--
-- Name: probot_identity_binding_challenges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.probot_identity_binding_challenges (
    hashed_token bytea NOT NULL,
    provider text NOT NULL,
    external_tenant_id text NOT NULL,
    external_user_id text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    confirmed_by_identity_id text,
    confirmed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    external_tenant_name text CONSTRAINT probot_identity_binding_challenge_external_tenant_name_not_null NOT NULL,
    external_user_name text NOT NULL
);


--
-- Name: probot_identity_bindings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.probot_identity_bindings (
    id text NOT NULL,
    provider text NOT NULL,
    external_tenant_id text NOT NULL,
    external_user_id text NOT NULL,
    identity_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    external_tenant_name text NOT NULL,
    external_user_name text NOT NULL
);


--
-- Name: processing_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.processing_activities (
    id text CONSTRAINT processing_activity_registries_id_not_null NOT NULL,
    tenant_id text CONSTRAINT processing_activity_registries_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT processing_activity_registries_organization_id_not_null NOT NULL,
    name text CONSTRAINT processing_activity_registries_name_not_null NOT NULL,
    purpose text,
    data_subject_category text,
    personal_data_category text,
    special_or_criminal_data public.processing_activities_special_or_criminal_data CONSTRAINT processing_activity_registrie_special_or_criminal_data_not_null NOT NULL,
    consent_evidence_link text,
    lawful_basis public.processing_activities_lawful_basis CONSTRAINT processing_activity_registries_lawful_basis_not_null NOT NULL,
    recipients text,
    location text,
    international_transfers boolean CONSTRAINT processing_activity_registries_international_transfers_not_null NOT NULL,
    transfer_safeguards public.processing_activities_transfer_safeguards,
    retention_period text,
    security_measures text,
    data_protection_impact_assessment_needed public.processing_activities_data_protection_impact_assessment CONSTRAINT processing_activity_registr_data_protection_impact_ass_not_null NOT NULL,
    transfer_impact_assessment_needed public.processing_activities_transfer_impact_assessment CONSTRAINT processing_activity_registr_transfer_impact_assessment_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT processing_activity_registries_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT processing_activity_registries_updated_at_not_null NOT NULL,
    last_review_date date,
    next_review_date date,
    role public.processing_activity_role NOT NULL,
    dpo_profile_id text
);


--
-- Name: processing_activity_data_protection_impact_assessments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.processing_activity_data_protection_impact_assessments (
    id text CONSTRAINT processing_activity_data_protection_impact_assessme_id_not_null NOT NULL,
    tenant_id text CONSTRAINT processing_activity_data_protection_impact_a_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT processing_activity_data_protection_im_organization_id_not_null NOT NULL,
    processing_activity_id text CONSTRAINT processing_activity_data_protec_processing_activity_id_not_null NOT NULL,
    description text,
    necessity_and_proportionality text,
    potential_risk text,
    mitigations text,
    residual_risk public.processing_activity_dpia_residual_risk,
    created_at timestamp with time zone CONSTRAINT processing_activity_data_protection_impact__created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT processing_activity_data_protection_impact__updated_at_not_null NOT NULL
);


--
-- Name: processing_activity_third_parties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.processing_activity_third_parties (
    processing_activity_id text CONSTRAINT processing_activity_vendors_processing_activity_id_not_null NOT NULL,
    third_party_id text CONSTRAINT processing_activity_vendors_vendor_id_not_null NOT NULL,
    tenant_id text CONSTRAINT processing_activity_vendors_tenant_id_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT processing_activity_vendors_created_at_not_null NOT NULL,
    organization_id text CONSTRAINT processing_activity_vendors_organization_id_not_null NOT NULL
);


--
-- Name: processing_activity_transfer_impact_assessments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.processing_activity_transfer_impact_assessments (
    id text NOT NULL,
    tenant_id text CONSTRAINT processing_activity_transfer_impact_assessme_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT processing_activity_transfer_impact_as_organization_id_not_null NOT NULL,
    processing_activity_id text CONSTRAINT processing_activity_transfer_im_processing_activity_id_not_null NOT NULL,
    data_subjects text,
    legal_mechanism text,
    transfer text,
    local_law_risk text,
    supplementary_measures text,
    created_at timestamp with time zone CONSTRAINT processing_activity_transfer_impact_assessm_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT processing_activity_transfer_impact_assessm_updated_at_not_null NOT NULL
);


--
-- Name: resource_aliases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.resource_aliases (
    tenant_id text CONSTRAINT trust_center_aliases_tenant_id_not_null NOT NULL,
    alias text CONSTRAINT trust_center_aliases_alias_not_null NOT NULL,
    resource_id text CONSTRAINT trust_center_aliases_resource_id_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT trust_center_aliases_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT trust_center_aliases_updated_at_not_null NOT NULL,
    CONSTRAINT trust_center_aliases_alias_check CHECK ((alias ~ '^[a-z0-9]+(-[a-z0-9]+)*$'::text))
);


--
-- Name: rights_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.rights_requests (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    request_type public.rights_request_type NOT NULL,
    request_state public.rights_request_state NOT NULL,
    data_subject text,
    contact text,
    details text,
    deadline date,
    action_taken text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    compliance_portal_id text
);


--
-- Name: risk_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risk_analyses (
    id text CONSTRAINT risk_assessments_id_not_null1 NOT NULL,
    tenant_id text CONSTRAINT risk_assessments_tenant_id_not_null1 NOT NULL,
    organization_id text CONSTRAINT risk_assessments_organization_id_not_null NOT NULL,
    name text CONSTRAINT risk_assessments_name_not_null NOT NULL,
    description jsonb,
    created_at timestamp with time zone CONSTRAINT risk_assessments_created_at_not_null1 NOT NULL,
    updated_at timestamp with time zone CONSTRAINT risk_assessments_updated_at_not_null1 NOT NULL,
    period_start date,
    period_end date,
    matrix_rows integer NOT NULL,
    matrix_cols integer NOT NULL,
    document_id text,
    score_levels jsonb NOT NULL
);


--
-- Name: risk_analysis_boundaries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risk_analysis_boundaries (
    id text CONSTRAINT risk_assessment_boundaries_id_not_null NOT NULL,
    tenant_id text CONSTRAINT risk_assessment_boundaries_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT risk_assessment_boundaries_organization_id_not_null NOT NULL,
    risk_analysis_diagram_id text CONSTRAINT risk_assessment_boundaries_risk_assessment_scope_id_not_null NOT NULL,
    parent_boundary_id text,
    name text CONSTRAINT risk_assessment_boundaries_name_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT risk_assessment_boundaries_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT risk_assessment_boundaries_updated_at_not_null NOT NULL
);


--
-- Name: risk_analysis_diagrams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risk_analysis_diagrams (
    id text CONSTRAINT risk_assessment_scopes_id_not_null NOT NULL,
    tenant_id text CONSTRAINT risk_assessment_scopes_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT risk_assessment_scopes_organization_id_not_null NOT NULL,
    risk_analysis_id text CONSTRAINT risk_assessment_scopes_risk_assessment_id_not_null NOT NULL,
    name text CONSTRAINT risk_assessment_scopes_name_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT risk_assessment_scopes_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT risk_assessment_scopes_updated_at_not_null NOT NULL
);


--
-- Name: risk_analysis_nodes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risk_analysis_nodes (
    id text CONSTRAINT risk_assessment_nodes_id_not_null NOT NULL,
    tenant_id text CONSTRAINT risk_assessment_nodes_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT risk_assessment_nodes_organization_id_not_null NOT NULL,
    risk_analysis_diagram_id text CONSTRAINT risk_assessment_nodes_risk_assessment_scope_id_not_null NOT NULL,
    node_type public.risk_analysis_node_type CONSTRAINT risk_assessment_nodes_node_type_not_null NOT NULL,
    name text CONSTRAINT risk_assessment_nodes_name_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT risk_assessment_nodes_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT risk_assessment_nodes_updated_at_not_null NOT NULL,
    boundary_id text
);


--
-- Name: risk_analysis_processes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risk_analysis_processes (
    id text CONSTRAINT risk_assessment_processes_id_not_null NOT NULL,
    tenant_id text CONSTRAINT risk_assessment_processes_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT risk_assessment_processes_organization_id_not_null NOT NULL,
    risk_analysis_diagram_id text CONSTRAINT risk_assessment_processes_risk_assessment_scope_id_not_null NOT NULL,
    source_node_id text CONSTRAINT risk_assessment_processes_source_node_id_not_null NOT NULL,
    target_node_id text CONSTRAINT risk_assessment_processes_target_node_id_not_null NOT NULL,
    name text CONSTRAINT risk_assessment_processes_name_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT risk_assessment_processes_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT risk_assessment_processes_updated_at_not_null NOT NULL
);


--
-- Name: risk_analysis_scenario_risks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risk_analysis_scenario_risks (
    tenant_id text CONSTRAINT risk_assessment_scenario_risks_tenant_id_not_null NOT NULL,
    risk_analysis_scenario_id text CONSTRAINT risk_assessment_scenario_ri_risk_assessment_scenario_i_not_null NOT NULL,
    risk_id text CONSTRAINT risk_assessment_scenario_risks_risk_id_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT now() CONSTRAINT risk_assessment_scenario_risks_created_at_not_null NOT NULL
);


--
-- Name: risk_analysis_scenario_threats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risk_analysis_scenario_threats (
    tenant_id text CONSTRAINT risk_assessment_scenario_threats_tenant_id_not_null NOT NULL,
    risk_analysis_scenario_id text CONSTRAINT risk_assessment_scenario_th_risk_assessment_scenario_i_not_null NOT NULL,
    risk_analysis_threat_id text CONSTRAINT risk_assessment_scenario_thr_risk_assessment_threat_id_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT now() CONSTRAINT risk_assessment_scenario_threats_created_at_not_null NOT NULL
);


--
-- Name: risk_analysis_scenarios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risk_analysis_scenarios (
    id text CONSTRAINT risk_assessment_scenarios_id_not_null NOT NULL,
    tenant_id text CONSTRAINT risk_assessment_scenarios_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT risk_assessment_scenarios_organization_id_not_null NOT NULL,
    risk_analysis_diagram_id text CONSTRAINT risk_assessment_scenarios_risk_assessment_scope_id_not_null NOT NULL,
    name text CONSTRAINT risk_assessment_scenarios_name_not_null NOT NULL,
    description text,
    created_at timestamp with time zone CONSTRAINT risk_assessment_scenarios_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT risk_assessment_scenarios_updated_at_not_null NOT NULL
);


--
-- Name: risk_analysis_threats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risk_analysis_threats (
    id text CONSTRAINT risk_assessment_threats_id_not_null NOT NULL,
    tenant_id text CONSTRAINT risk_assessment_threats_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT risk_assessment_threats_organization_id_not_null NOT NULL,
    risk_analysis_diagram_id text CONSTRAINT risk_assessment_threats_risk_assessment_scope_id_not_null NOT NULL,
    process_id text CONSTRAINT risk_assessment_threats_process_id_not_null NOT NULL,
    name text CONSTRAINT risk_assessment_threats_name_not_null NOT NULL,
    category text CONSTRAINT risk_assessment_threats_category_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT risk_assessment_threats_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT risk_assessment_threats_updated_at_not_null NOT NULL
);


--
-- Name: risks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risks (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    treatment public.risk_treatment,
    inherent_likelihood integer,
    inherent_impact integer,
    residual_likelihood integer,
    residual_impact integer,
    category text NOT NULL,
    note text,
    inherent_risk_score integer GENERATED ALWAYS AS ((inherent_impact * inherent_likelihood)) STORED,
    residual_risk_score integer GENERATED ALWAYS AS ((residual_impact * residual_likelihood)) STORED,
    owner_profile_id text,
    reference_id text NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, translate(((COALESCE(reference_id, ''::text) || ' '::text) || COALESCE(name, ''::text)), '-+'::text, '  '::text))) STORED
);


--
-- Name: risks_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risks_documents (
    risk_id text CONSTRAINT risks_policies_risk_id_not_null NOT NULL,
    document_id text CONSTRAINT risks_policies_policy_id_not_null NOT NULL,
    tenant_id text CONSTRAINT risks_policies_tenant_id_not_null NOT NULL,
    created_at timestamp without time zone CONSTRAINT risks_policies_created_at_not_null NOT NULL,
    organization_id text NOT NULL
);


--
-- Name: risks_internal_controls; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risks_internal_controls (
    risk_id text CONSTRAINT risks_mitigations_risk_id_not_null NOT NULL,
    internal_control_id text CONSTRAINT risks_mitigations_mitigation_id_not_null NOT NULL,
    tenant_id text CONSTRAINT risks_mitigations_tenant_id_not_null NOT NULL,
    created_at timestamp without time zone CONSTRAINT risks_mitigations_created_at_not_null NOT NULL,
    organization_id text NOT NULL
);


--
-- Name: risks_obligations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.risks_obligations (
    risk_id text NOT NULL,
    obligation_id text NOT NULL,
    tenant_id text NOT NULL,
    created_at timestamp without time zone NOT NULL,
    organization_id text NOT NULL
);


--
-- Name: search_resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.search_resources (
    resource_id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    entity_type integer NOT NULL,
    title text NOT NULL,
    body_vector tsvector NOT NULL,
    tag_vector tsvector NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS ((body_vector || tag_vector)) STORED
);


--
-- Name: slack_delivery_operations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slack_delivery_operations (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    operation_key text NOT NULL,
    operation_kind public.slack_delivery_operation_kind NOT NULL,
    payload jsonb NOT NULL,
    client_msg_id text,
    processing_started_at timestamp with time zone,
    completed_at timestamp with time zone,
    attempt_count integer NOT NULL,
    max_attempts integer NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    dead_lettered_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    processing_owner_token uuid,
    CONSTRAINT slack_delivery_operations_client_msg_id_kind_check CHECK ((((operation_kind = 'POST_MESSAGE'::public.slack_delivery_operation_kind) AND (client_msg_id IS NOT NULL)) OR ((operation_kind = 'ADD_REACTION'::public.slack_delivery_operation_kind) AND (client_msg_id IS NULL))))
);


--
-- Name: slack_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slack_messages (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    body jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    sent_at timestamp with time zone,
    error text,
    message_ts text,
    channel_id text,
    requester_email text,
    type public.slack_message_type NOT NULL,
    metadata jsonb,
    initial_slack_message_id text NOT NULL,
    processing_started_at timestamp with time zone
);


--
-- Name: slackbot_bind_callbacks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slackbot_bind_callbacks (
    team_id text NOT NULL,
    user_id text NOT NULL,
    encrypted_response_url bytea NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: slackbot_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slackbot_events (
    id text NOT NULL,
    event_id text NOT NULL,
    envelope jsonb NOT NULL,
    processing_started_at timestamp with time zone,
    processed_at timestamp with time zone,
    attempt_count integer NOT NULL,
    max_attempts integer NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    dead_lettered_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    processing_owner_token uuid
);


--
-- Name: slackbot_install_state_claims; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slackbot_install_state_claims (
    state_digest bytea NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    processing_token uuid,
    processing_started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: slackbot_installations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slackbot_installations (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    team_id text NOT NULL,
    bot_user_id text NOT NULL,
    encrypted_access_token bytea NOT NULL,
    encrypted_refresh_token bytea,
    access_token_expires_at timestamp with time zone,
    scopes text[] NOT NULL,
    status public.slackbot_installation_status NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    team_name text
);


--
-- Name: slackbot_interactive_claims; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slackbot_interactive_claims (
    interaction_digest bytea NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    processing_token uuid,
    processing_started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: slackbot_interactive_commands; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slackbot_interactive_commands (
    id text NOT NULL,
    tenant_id text,
    organization_id text,
    request_digest bytea NOT NULL,
    encrypted_payload bytea NOT NULL,
    processing_started_at timestamp with time zone,
    processed_at timestamp with time zone,
    attempt_count integer NOT NULL,
    max_attempts integer NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    dead_lettered_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    processing_owner_token uuid,
    CONSTRAINT slackbot_interactive_commands_scope_check CHECK ((((tenant_id IS NULL) AND (organization_id IS NULL)) OR ((tenant_id IS NOT NULL) AND (organization_id IS NOT NULL))))
);


--
-- Name: slackbot_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slackbot_messages (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    message_type text NOT NULL,
    body jsonb NOT NULL,
    metadata jsonb NOT NULL,
    channel_id text,
    message_ts text,
    initial_slackbot_message_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    sent_at timestamp with time zone,
    processing_started_at timestamp with time zone,
    attempt_count integer NOT NULL,
    max_attempts integer NOT NULL,
    last_attempted_at timestamp with time zone,
    next_attempt_at timestamp with time zone,
    last_error text,
    client_msg_id uuid NOT NULL,
    error text
);


--
-- Name: slackbot_processed_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.slackbot_processed_events (
    event_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: statements_of_applicability; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.statements_of_applicability (
    id text CONSTRAINT states_of_applicability_id_not_null NOT NULL,
    tenant_id text CONSTRAINT states_of_applicability_tenant_id_not_null NOT NULL,
    organization_id text CONSTRAINT states_of_applicability_organization_id_not_null NOT NULL,
    name text CONSTRAINT states_of_applicability_name_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT states_of_applicability_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT states_of_applicability_updated_at_not_null NOT NULL,
    document_id text
);


--
-- Name: tag_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tag_assignments (
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    resource_id text NOT NULL,
    tag_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    id text NOT NULL
);


--
-- Name: tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tags (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    name text NOT NULL,
    color text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT tags_color_hex_check CHECK (((color IS NULL) OR (color ~ '^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$'::text)))
);


--
-- Name: task_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_activities (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    task_id text NOT NULL,
    actor_profile_id text,
    activity_type public.task_activity_type NOT NULL,
    field public.task_activity_field,
    old_value text,
    new_value text,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: task_comment_external_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_comment_external_links (
    task_comment_id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    task_id text NOT NULL,
    connector_id text NOT NULL,
    provider text NOT NULL,
    external_id text NOT NULL,
    remote_updated_at timestamp with time zone,
    content_hash text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: task_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_comments (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    task_id text NOT NULL,
    owner_profile_id text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    content jsonb NOT NULL
);


--
-- Name: task_external_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_external_links (
    task_id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    connector_id text NOT NULL,
    provider text NOT NULL,
    external_id text NOT NULL,
    external_identifier text NOT NULL,
    external_url text NOT NULL,
    destination jsonb NOT NULL,
    origin text NOT NULL,
    remote_updated_at timestamp with time zone,
    content_hash text,
    metadata jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: task_sync_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.task_sync_jobs (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    direction text NOT NULL,
    status text NOT NULL,
    payload jsonb NOT NULL,
    error text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    attempt_count integer NOT NULL,
    next_attempt_at timestamp with time zone,
    processing_owner_token text
);


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tasks (
    id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    tenant_id text NOT NULL,
    internal_control_id text,
    state public.task_state,
    version integer DEFAULT 1 NOT NULL,
    time_estimate interval,
    reference_id text NOT NULL,
    organization_id text NOT NULL,
    deadline timestamp with time zone,
    assigned_to_profile_id text,
    rank integer CONSTRAINT tasks_priority_not_null NOT NULL,
    priority public.task_priority CONSTRAINT tasks_priority_not_null1 NOT NULL,
    priority_rank integer GENERATED ALWAYS AS (((
CASE priority
    WHEN 'URGENT'::public.task_priority THEN 1
    WHEN 'HIGH'::public.task_priority THEN 2
    WHEN 'MEDIUM'::public.task_priority THEN 3
    WHEN 'LOW'::public.task_priority THEN 4
    ELSE NULL::integer
END * 1000000) + rank)) STORED,
    content jsonb NOT NULL,
    recurrence interval
);


--
-- Name: tenant_id_counter_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_id_counter_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: third_parties; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.third_parties (
    id text CONSTRAINT vendors_id_not_null NOT NULL,
    name text CONSTRAINT vendors_name_not_null NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    organization_id text CONSTRAINT vendors_organization_id_not_null NOT NULL,
    status_page_url text,
    description text,
    terms_of_service_url text,
    privacy_policy_url text,
    version integer DEFAULT 1 CONSTRAINT vendors_version_not_null NOT NULL,
    tenant_id text CONSTRAINT vendors_tenant_id_not_null NOT NULL,
    service_level_agreement_url text,
    data_processing_agreement_url text,
    security_page_url text,
    trust_page_url text,
    headquarter_address text,
    legal_name text,
    website_url text,
    certifications text[],
    category public.third_party_category DEFAULT 'OTHER'::public.third_party_category CONSTRAINT vendors_category_not_null NOT NULL,
    business_associate_agreement_url text,
    subprocessors_list_url text,
    countries public.country_code[] CONSTRAINT vendors_countries_not_null NOT NULL,
    business_owner_profile_id text,
    security_owner_profile_id text,
    common_third_party_id text,
    first_level boolean DEFAULT true NOT NULL,
    vetting_status public.third_party_vetting_status,
    vetting_website_url text,
    vetting_procedure text,
    vetting_processing_started_at timestamp with time zone,
    vetting_error_message text,
    level integer NOT NULL,
    parent_third_party_id text
);


--
-- Name: third_party_administrators; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.third_party_administrators (
    third_party_id text NOT NULL,
    administrator_profile_id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: third_party_business_associate_agreements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.third_party_business_associate_agreements (
    id text CONSTRAINT vendor_business_associate_agreements_id_not_null NOT NULL,
    organization_id text CONSTRAINT vendor_business_associate_agreements_organization_id_not_null NOT NULL,
    tenant_id text CONSTRAINT vendor_business_associate_agreements_tenant_id_not_null NOT NULL,
    third_party_id text CONSTRAINT vendor_business_associate_agreements_vendor_id_not_null NOT NULL,
    valid_from date,
    valid_until date,
    file_id text CONSTRAINT vendor_business_associate_agreements_file_id_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT vendor_business_associate_agreements_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT vendor_business_associate_agreements_updated_at_not_null NOT NULL
);


--
-- Name: third_party_compliance_reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.third_party_compliance_reports (
    id text CONSTRAINT vendor_compliance_reports_id_not_null NOT NULL,
    tenant_id text CONSTRAINT vendor_compliance_reports_tenant_id_not_null NOT NULL,
    third_party_id text CONSTRAINT vendor_compliance_reports_vendor_id_not_null NOT NULL,
    report_date date CONSTRAINT vendor_compliance_reports_report_date_not_null NOT NULL,
    valid_until date,
    report_name text CONSTRAINT vendor_compliance_reports_report_name_not_null NOT NULL,
    file_key text,
    file_size bigint,
    created_at timestamp without time zone CONSTRAINT vendor_compliance_reports_created_at_not_null NOT NULL,
    updated_at timestamp without time zone CONSTRAINT vendor_compliance_reports_updated_at_not_null NOT NULL,
    report_file_id text,
    organization_id text CONSTRAINT vendor_compliance_reports_organization_id_not_null NOT NULL
);


--
-- Name: third_party_contacts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.third_party_contacts (
    id text CONSTRAINT vendor_contacts_id_not_null NOT NULL,
    tenant_id text CONSTRAINT vendor_contacts_tenant_id_not_null NOT NULL,
    third_party_id text CONSTRAINT vendor_contacts_vendor_id_not_null NOT NULL,
    full_name text,
    email public.citext,
    phone text,
    role text,
    created_at timestamp with time zone CONSTRAINT vendor_contacts_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT vendor_contacts_updated_at_not_null NOT NULL,
    organization_id text CONSTRAINT vendor_contacts_organization_id_not_null NOT NULL
);


--
-- Name: third_party_data_privacy_agreements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.third_party_data_privacy_agreements (
    id text CONSTRAINT vendor_data_privacy_agreements_id_not_null NOT NULL,
    organization_id text CONSTRAINT vendor_data_privacy_agreements_organization_id_not_null NOT NULL,
    tenant_id text CONSTRAINT vendor_data_privacy_agreements_tenant_id_not_null NOT NULL,
    third_party_id text CONSTRAINT vendor_data_privacy_agreements_vendor_id_not_null NOT NULL,
    valid_from date,
    valid_until date,
    file_id text CONSTRAINT vendor_data_privacy_agreements_file_id_not_null NOT NULL,
    created_at timestamp with time zone CONSTRAINT vendor_data_privacy_agreements_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT vendor_data_privacy_agreements_updated_at_not_null NOT NULL
);


--
-- Name: third_party_risk_assessments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.third_party_risk_assessments (
    tenant_id text CONSTRAINT risk_assessments_tenant_id_not_null NOT NULL,
    id text CONSTRAINT risk_assessments_id_not_null NOT NULL,
    third_party_id text CONSTRAINT risk_assessments_vendor_id_not_null NOT NULL,
    expires_at timestamp with time zone CONSTRAINT risk_assessments_expires_at_not_null NOT NULL,
    data_sensitivity public.data_sensitivity CONSTRAINT risk_assessments_data_sensitivity_not_null NOT NULL,
    business_impact public.business_impact CONSTRAINT risk_assessments_business_impact_not_null NOT NULL,
    notes text,
    created_at timestamp with time zone CONSTRAINT risk_assessments_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT risk_assessments_updated_at_not_null NOT NULL,
    organization_id text CONSTRAINT vendor_risk_assessments_organization_id_not_null NOT NULL
);


--
-- Name: third_party_services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.third_party_services (
    id text CONSTRAINT vendor_services_id_not_null NOT NULL,
    tenant_id text CONSTRAINT vendor_services_tenant_id_not_null NOT NULL,
    third_party_id text CONSTRAINT vendor_services_vendor_id_not_null NOT NULL,
    name text CONSTRAINT vendor_services_name_not_null NOT NULL,
    description text,
    created_at timestamp with time zone CONSTRAINT vendor_services_created_at_not_null NOT NULL,
    updated_at timestamp with time zone CONSTRAINT vendor_services_updated_at_not_null NOT NULL,
    organization_id text CONSTRAINT vendor_services_organization_id_not_null NOT NULL
);


--
-- Name: tracker_pattern_daily_hits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tracker_pattern_daily_hits (
    tracker_pattern_id text NOT NULL,
    day date NOT NULL,
    browser_family text NOT NULL,
    presentation text NOT NULL,
    consent_action text NOT NULL,
    category_granted boolean NOT NULL,
    tenant_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    count integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: tracker_pattern_discovery_hits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tracker_pattern_discovery_hits (
    tracker_pattern_id text NOT NULL,
    tenant_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    chrome_hits integer DEFAULT 0 NOT NULL,
    edge_hits integer DEFAULT 0 NOT NULL,
    firefox_hits integer DEFAULT 0 NOT NULL,
    safari_hits integer DEFAULT 0 NOT NULL,
    other_hits integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: tracker_patterns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tracker_patterns (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    cookie_category_id text NOT NULL,
    tracker_type public.tracker_type NOT NULL,
    pattern text NOT NULL,
    match_type public.cookie_pattern_match_type NOT NULL,
    description text NOT NULL,
    excluded boolean NOT NULL,
    max_age_seconds integer,
    source public.cookie_source,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    common_tracker_pattern_id text,
    mapping_requested_at timestamp with time zone,
    source_cookie_banner_id text
);


--
-- Name: tracker_resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tracker_resources (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    cookie_banner_id text NOT NULL,
    cookie_category_id text NOT NULL,
    resource_type public.tracker_resource_type NOT NULL,
    origin text NOT NULL,
    path text NOT NULL,
    display_name text NOT NULL,
    description text NOT NULL,
    excluded boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    cookie_banner_scan_id text
);


--
-- Name: treatment_plan_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.treatment_plan_events (
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    risk_analysis_id text NOT NULL,
    treatment_plan_id text NOT NULL,
    event_type public.treatment_plan_event_type NOT NULL,
    risk_id text NOT NULL,
    owner_profile_id text NOT NULL,
    treatment public.risk_treatment NOT NULL,
    inherent_likelihood integer NOT NULL,
    inherent_impact integer NOT NULL,
    residual_likelihood integer NOT NULL,
    residual_impact integer NOT NULL,
    internal_control_ids text[] NOT NULL,
    category text NOT NULL,
    treatment_plan_created_at timestamp with time zone NOT NULL,
    treatment_plan_updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: treatment_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.treatment_plans (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    risk_id text NOT NULL,
    risk_analysis_id text NOT NULL,
    treatment public.risk_treatment NOT NULL,
    owner_profile_id text NOT NULL,
    inherent_likelihood integer NOT NULL,
    inherent_impact integer NOT NULL,
    inherent_risk_score integer GENERATED ALWAYS AS ((inherent_impact * inherent_likelihood)) STORED,
    residual_likelihood integer NOT NULL,
    residual_impact integer NOT NULL,
    residual_risk_score integer GENERATED ALWAYS AS ((residual_impact * residual_likelihood)) STORED,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: treatment_plans_internal_controls; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.treatment_plans_internal_controls (
    treatment_plan_id text NOT NULL,
    internal_control_id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: webhook_data; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.webhook_data (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    event_type public.webhook_event_type NOT NULL,
    data jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    processed_at timestamp with time zone,
    updated_from jsonb,
    actor jsonb
);


--
-- Name: webhook_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.webhook_events (
    id text NOT NULL,
    tenant_id text NOT NULL,
    webhook_data_id text NOT NULL,
    webhook_subscription_id text NOT NULL,
    status public.webhook_event_status NOT NULL,
    response jsonb,
    created_at timestamp with time zone NOT NULL,
    processing_owner_token text,
    processing_started_at timestamp with time zone,
    attempt_count integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 12 NOT NULL,
    next_attempt_at timestamp with time zone,
    last_error text,
    completed_at timestamp with time zone,
    dead_lettered_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: webhook_subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.webhook_subscriptions (
    id text NOT NULL,
    tenant_id text NOT NULL,
    organization_id text NOT NULL,
    endpoint_url text NOT NULL,
    selected_events public.webhook_event_type[] NOT NULL,
    encrypted_signing_secret bytea NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: access_review_campaign_sources access_review_campaign_source_access_review_campaign_id_acc_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_sources
    ADD CONSTRAINT access_review_campaign_source_access_review_campaign_id_acc_key UNIQUE (access_review_campaign_id, access_review_source_id);


--
-- Name: access_review_campaign_source_fetch_attempts access_review_campaign_source_access_review_campaign_source_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_source_fetch_attempts
    ADD CONSTRAINT access_review_campaign_source_access_review_campaign_source_key UNIQUE (access_review_campaign_source_id, attempt_number);


--
-- Name: access_review_campaign_source_fetch_attempts access_review_campaign_source_fetch_attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_source_fetch_attempts
    ADD CONSTRAINT access_review_campaign_source_fetch_attempts_pkey PRIMARY KEY (id);


--
-- Name: access_review_campaign_sources access_review_campaign_sources_campaign_source_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_sources
    ADD CONSTRAINT access_review_campaign_sources_campaign_source_unique UNIQUE (access_review_campaign_id, access_review_source_id);


--
-- Name: access_review_campaign_sources access_review_campaign_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_sources
    ADD CONSTRAINT access_review_campaign_sources_pkey PRIMARY KEY (id);


--
-- Name: access_review_campaigns access_review_campaigns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaigns
    ADD CONSTRAINT access_review_campaigns_pkey PRIMARY KEY (id);


--
-- Name: access_review_entries access_review_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_entries
    ADD CONSTRAINT access_review_entries_pkey PRIMARY KEY (id);


--
-- Name: access_review_entry_decision_history access_review_entry_decision_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_entry_decision_history
    ADD CONSTRAINT access_review_entry_decision_history_pkey PRIMARY KEY (id);


--
-- Name: access_review_sources access_review_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_sources
    ADD CONSTRAINT access_review_sources_pkey PRIMARY KEY (id);


--
-- Name: agent_execution_anchors agent_execution_anchors_execution_provider_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_execution_anchors
    ADD CONSTRAINT agent_execution_anchors_execution_provider_key UNIQUE (tenant_id, organization_id, agent_execution_id, provider);


--
-- Name: agent_execution_anchors agent_execution_anchors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_execution_anchors
    ADD CONSTRAINT agent_execution_anchors_pkey PRIMARY KEY (id);


--
-- Name: agent_execution_anchors agent_execution_anchors_provider_coordinates_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_execution_anchors
    ADD CONSTRAINT agent_execution_anchors_provider_coordinates_key UNIQUE (tenant_id, organization_id, provider, external_conversation_id, external_message_id);


--
-- Name: agent_executions agent_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_executions
    ADD CONSTRAINT agent_executions_pkey PRIMARY KEY (id);


--
-- Name: agent_inputs agent_inputs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_inputs
    ADD CONSTRAINT agent_inputs_pkey PRIMARY KEY (id);


--
-- Name: ai_systems ai_systems_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_systems
    ADD CONSTRAINT ai_systems_pkey PRIMARY KEY (id);


--
-- Name: applicability_statements applicability_statements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applicability_statements
    ADD CONSTRAINT applicability_statements_pkey PRIMARY KEY (id);


--
-- Name: asana_webhook_events asana_webhook_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.asana_webhook_events
    ADD CONSTRAINT asana_webhook_events_pkey PRIMARY KEY (delivery_id);


--
-- Name: asana_webhook_handshakes asana_webhook_handshakes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.asana_webhook_handshakes
    ADD CONSTRAINT asana_webhook_handshakes_pkey PRIMARY KEY (nonce);


--
-- Name: asset_third_parties asset_vendors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.asset_third_parties
    ADD CONSTRAINT asset_vendors_pkey PRIMARY KEY (asset_id, third_party_id);


--
-- Name: assets assets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assets
    ADD CONSTRAINT assets_pkey PRIMARY KEY (id);


--
-- Name: audit_log_entries audit_log_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_log_entries
    ADD CONSTRAINT audit_log_entries_pkey PRIMARY KEY (id);


--
-- Name: audits_frameworks audits_frameworks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audits_frameworks
    ADD CONSTRAINT audits_frameworks_pkey PRIMARY KEY (audit_id, framework_id);


--
-- Name: audits audits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audits
    ADD CONSTRAINT audits_pkey PRIMARY KEY (id);


--
-- Name: iam_saml_configurations auth_saml_configurations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_saml_configurations
    ADD CONSTRAINT auth_saml_configurations_pkey PRIMARY KEY (id);


--
-- Name: iam_saml_requests auth_saml_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_saml_requests
    ADD CONSTRAINT auth_saml_requests_pkey PRIMARY KEY (id);


--
-- Name: iam_invitations authz_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_invitations
    ADD CONSTRAINT authz_invitations_pkey PRIMARY KEY (id);


--
-- Name: iam_memberships authz_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_memberships
    ADD CONSTRAINT authz_memberships_pkey PRIMARY KEY (id);


--
-- Name: iam_memberships authz_memberships_user_id_organization_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_memberships
    ADD CONSTRAINT authz_memberships_user_id_organization_id_key UNIQUE (identity_id, organization_id);


--
-- Name: bot_delivery_destinations bot_delivery_destinations_org_target_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_delivery_destinations
    ADD CONSTRAINT bot_delivery_destinations_org_target_key UNIQUE (tenant_id, organization_id, provider, target_namespace, target_key);


--
-- Name: bot_delivery_destinations bot_delivery_destinations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_delivery_destinations
    ADD CONSTRAINT bot_delivery_destinations_pkey PRIMARY KEY (id);


--
-- Name: bot_messages bot_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_messages
    ADD CONSTRAINT bot_messages_pkey PRIMARY KEY (id);


--
-- Name: bot_thread_subjects bot_thread_subjects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_thread_subjects
    ADD CONSTRAINT bot_thread_subjects_pkey PRIMARY KEY (id);


--
-- Name: business_function_assets business_function_assets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_function_assets
    ADD CONSTRAINT business_function_assets_pkey PRIMARY KEY (business_function_id, asset_id);


--
-- Name: business_function_third_parties business_function_third_parties_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_function_third_parties
    ADD CONSTRAINT business_function_third_parties_pkey PRIMARY KEY (business_function_id, third_party_id);


--
-- Name: business_functions business_functions_organization_id_reference_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_functions
    ADD CONSTRAINT business_functions_organization_id_reference_id_key UNIQUE (organization_id, reference_id);


--
-- Name: business_functions business_functions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_functions
    ADD CONSTRAINT business_functions_pkey PRIMARY KEY (id);


--
-- Name: cached_certificates cached_certificates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cached_certificates
    ADD CONSTRAINT cached_certificates_pkey PRIMARY KEY (domain);


--
-- Name: certificates certificates_hostname_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.certificates
    ADD CONSTRAINT certificates_hostname_key UNIQUE (hostname);


--
-- Name: certificates certificates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.certificates
    ADD CONSTRAINT certificates_pkey PRIMARY KEY (id);


--
-- Name: cmp_domain_surfaces cmp_domain_surfaces_cmp_domain_id_origin_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cmp_domain_surfaces
    ADD CONSTRAINT cmp_domain_surfaces_cmp_domain_id_origin_key UNIQUE (cmp_domain_id, origin);


--
-- Name: cmp_domain_surfaces cmp_domain_surfaces_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cmp_domain_surfaces
    ADD CONSTRAINT cmp_domain_surfaces_pkey PRIMARY KEY (id);


--
-- Name: cmp_domains cmp_domains_organization_id_domain_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cmp_domains
    ADD CONSTRAINT cmp_domains_organization_id_domain_key UNIQUE (organization_id, domain);


--
-- Name: cmp_domains cmp_domains_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cmp_domains
    ADD CONSTRAINT cmp_domains_pkey PRIMARY KEY (id);


--
-- Name: collaboration_actors collaboration_actors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_actors
    ADD CONSTRAINT collaboration_actors_pkey PRIMARY KEY (document_id, actor_id);


--
-- Name: collaboration_changes collaboration_changes_document_id_revision_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_changes
    ADD CONSTRAINT collaboration_changes_document_id_revision_key UNIQUE (document_id, revision);


--
-- Name: collaboration_changes collaboration_changes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_changes
    ADD CONSTRAINT collaboration_changes_pkey PRIMARY KEY (document_id, change_hash);


--
-- Name: collaboration_documents collaboration_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_documents
    ADD CONSTRAINT collaboration_documents_pkey PRIMARY KEY (document_id);


--
-- Name: common_gvl_snapshots common_gvl_snapshots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_gvl_snapshots
    ADD CONSTRAINT common_gvl_snapshots_pkey PRIMARY KEY (id);


--
-- Name: common_gvl_snapshots common_gvl_snapshots_vendor_list_version_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_gvl_snapshots
    ADD CONSTRAINT common_gvl_snapshots_vendor_list_version_key UNIQUE (vendor_list_version);


--
-- Name: common_gvl_state common_gvl_state_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_gvl_state
    ADD CONSTRAINT common_gvl_state_pkey PRIMARY KEY (singleton);


--
-- Name: common_gvl_vendors common_gvl_vendors_iab_vendor_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_gvl_vendors
    ADD CONSTRAINT common_gvl_vendors_iab_vendor_id_key UNIQUE (iab_vendor_id);


--
-- Name: common_gvl_vendors common_gvl_vendors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_gvl_vendors
    ADD CONSTRAINT common_gvl_vendors_pkey PRIMARY KEY (id);


--
-- Name: common_third_parties common_third_parties_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_third_parties
    ADD CONSTRAINT common_third_parties_pkey PRIMARY KEY (id);


--
-- Name: common_third_party_domains common_third_party_domains_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_third_party_domains
    ADD CONSTRAINT common_third_party_domains_pkey PRIMARY KEY (id);


--
-- Name: common_tracker_patterns common_tracker_patterns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_tracker_patterns
    ADD CONSTRAINT common_tracker_patterns_pkey PRIMARY KEY (id);


--
-- Name: compliance_custom_links compliance_custom_links_compliance_portal_id_rank_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_custom_links
    ADD CONSTRAINT compliance_custom_links_compliance_portal_id_rank_key UNIQUE (compliance_portal_id, rank) DEFERRABLE INITIALLY DEFERRED;


--
-- Name: compliance_custom_links compliance_external_urls_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_custom_links
    ADD CONSTRAINT compliance_external_urls_pkey PRIMARY KEY (id);


--
-- Name: compliance_frameworks compliance_frameworks_compliance_portal_id_framework_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_frameworks
    ADD CONSTRAINT compliance_frameworks_compliance_portal_id_framework_id_key UNIQUE (compliance_portal_id, framework_id);


--
-- Name: compliance_frameworks compliance_frameworks_compliance_portal_id_rank_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_frameworks
    ADD CONSTRAINT compliance_frameworks_compliance_portal_id_rank_key UNIQUE (compliance_portal_id, rank) DEFERRABLE INITIALLY DEFERRED;


--
-- Name: compliance_frameworks compliance_frameworks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_frameworks
    ADD CONSTRAINT compliance_frameworks_pkey PRIMARY KEY (id);


--
-- Name: compliance_portal_commitment_groups compliance_portal_commitment_groups_compliance_portal_id_rank_k; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portal_commitment_groups
    ADD CONSTRAINT compliance_portal_commitment_groups_compliance_portal_id_rank_k UNIQUE (compliance_portal_id, rank) DEFERRABLE INITIALLY DEFERRED;


--
-- Name: compliance_portal_commitment_groups compliance_portal_commitment_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portal_commitment_groups
    ADD CONSTRAINT compliance_portal_commitment_groups_pkey PRIMARY KEY (id);


--
-- Name: compliance_portal_commitments compliance_portal_commitments_group_id_rank_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portal_commitments
    ADD CONSTRAINT compliance_portal_commitments_group_id_rank_key UNIQUE (group_id, rank) DEFERRABLE INITIALLY DEFERRED;


--
-- Name: compliance_portal_commitments compliance_portal_commitments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portal_commitments
    ADD CONSTRAINT compliance_portal_commitments_pkey PRIMARY KEY (id);


--
-- Name: compliance_portals compliance_portals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portals
    ADD CONSTRAINT compliance_portals_pkey PRIMARY KEY (id);


--
-- Name: compliance_portals compliance_portals_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portals
    ADD CONSTRAINT compliance_portals_slug_key UNIQUE (slug);


--
-- Name: obligations compliance_registries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.obligations
    ADD CONSTRAINT compliance_registries_pkey PRIMARY KEY (id);


--
-- Name: connector_accounts connector_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.connector_accounts
    ADD CONSTRAINT connector_accounts_pkey PRIMARY KEY (id);


--
-- Name: connector_install_state_claims connector_install_state_claims_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.connector_install_state_claims
    ADD CONSTRAINT connector_install_state_claims_pkey PRIMARY KEY (state_digest);


--
-- Name: connectors connectors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.connectors
    ADD CONSTRAINT connectors_pkey PRIMARY KEY (id);


--
-- Name: controls_audits controls_audits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_audits
    ADD CONSTRAINT controls_audits_pkey PRIMARY KEY (control_id, audit_id);


--
-- Name: controls controls_framework_ref_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls
    ADD CONSTRAINT controls_framework_ref_unique UNIQUE (framework_id, section_title);


--
-- Name: controls_internal_controls controls_internal_controls_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_internal_controls
    ADD CONSTRAINT controls_internal_controls_pkey PRIMARY KEY (control_id, internal_control_id);


--
-- Name: controls_obligations controls_obligations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_obligations
    ADD CONSTRAINT controls_obligations_pkey PRIMARY KEY (control_id, obligation_id);


--
-- Name: controls controls_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls
    ADD CONSTRAINT controls_pkey PRIMARY KEY (id);


--
-- Name: controls_documents controls_policies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_documents
    ADD CONSTRAINT controls_policies_pkey PRIMARY KEY (control_id, document_id);


--
-- Name: cookie_banner_daily_page_loads cookie_banner_daily_page_loads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_daily_page_loads
    ADD CONSTRAINT cookie_banner_daily_page_loads_pkey PRIMARY KEY (cookie_banner_id, day, browser_family, presentation, consent_action);


--
-- Name: cookie_banner_discovery_stats cookie_banner_discovery_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_discovery_stats
    ADD CONSTRAINT cookie_banner_discovery_stats_pkey PRIMARY KEY (cookie_banner_id);


--
-- Name: cookie_banner_gvl_vendors cookie_banner_gvl_vendors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_gvl_vendors
    ADD CONSTRAINT cookie_banner_gvl_vendors_pkey PRIMARY KEY (cookie_banner_id, iab_vendor_id);


--
-- Name: cookie_banner_scans cookie_banner_scans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_scans
    ADD CONSTRAINT cookie_banner_scans_pkey PRIMARY KEY (id);


--
-- Name: cookie_banner_translations cookie_banner_translations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_translations
    ADD CONSTRAINT cookie_banner_translations_pkey PRIMARY KEY (id);


--
-- Name: cookie_banner_versions cookie_banner_versions_banner_version_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_versions
    ADD CONSTRAINT cookie_banner_versions_banner_version_key UNIQUE (cookie_banner_id, version);


--
-- Name: cookie_banner_versions cookie_banner_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_versions
    ADD CONSTRAINT cookie_banner_versions_pkey PRIMARY KEY (id);


--
-- Name: cookie_banners cookie_banners_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banners
    ADD CONSTRAINT cookie_banners_pkey PRIMARY KEY (id);


--
-- Name: cookie_categories cookie_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_categories
    ADD CONSTRAINT cookie_categories_pkey PRIMARY KEY (id);


--
-- Name: cookie_consent_records cookie_consent_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_consent_records
    ADD CONSTRAINT cookie_consent_records_pkey PRIMARY KEY (id);


--
-- Name: cp_accesses cp_accesses_compliance_portal_id_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_accesses
    ADD CONSTRAINT cp_accesses_compliance_portal_id_email_key UNIQUE (compliance_portal_id, email);


--
-- Name: cp_accesses cp_accesses_identity_id_compliance_portal_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_accesses
    ADD CONSTRAINT cp_accesses_identity_id_compliance_portal_id_key UNIQUE (identity_id, compliance_portal_id);


--
-- Name: cp_accesses cp_accesses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_accesses
    ADD CONSTRAINT cp_accesses_pkey PRIMARY KEY (id);


--
-- Name: cp_audits cp_audits_compliance_portal_id_audit_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_audits
    ADD CONSTRAINT cp_audits_compliance_portal_id_audit_id_key UNIQUE (compliance_portal_id, audit_id);


--
-- Name: cp_audits cp_audits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_audits
    ADD CONSTRAINT cp_audits_pkey PRIMARY KEY (id);


--
-- Name: cp_document_accesses cp_document_accesses_access_id_cp_audit_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_access_id_cp_audit_id_key UNIQUE (compliance_portal_access_id, compliance_portal_audit_id);


--
-- Name: cp_document_accesses cp_document_accesses_access_id_cp_document_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_access_id_cp_document_id_key UNIQUE (compliance_portal_access_id, compliance_portal_document_id);


--
-- Name: cp_document_accesses cp_document_accesses_compliance_portal_file_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_compliance_portal_file_id_key UNIQUE (compliance_portal_access_id, compliance_portal_file_id);


--
-- Name: cp_document_accesses cp_document_accesses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_pkey PRIMARY KEY (id);


--
-- Name: cp_documents cp_documents_compliance_portal_id_document_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_documents
    ADD CONSTRAINT cp_documents_compliance_portal_id_document_id_key UNIQUE (compliance_portal_id, document_id);


--
-- Name: cp_documents cp_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_documents
    ADD CONSTRAINT cp_documents_pkey PRIMARY KEY (id);


--
-- Name: cp_files cp_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_files
    ADD CONSTRAINT cp_files_pkey PRIMARY KEY (id);


--
-- Name: cp_references cp_references_compliance_portal_id_rank_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_references
    ADD CONSTRAINT cp_references_compliance_portal_id_rank_key UNIQUE (compliance_portal_id, rank) DEFERRABLE INITIALLY DEFERRED;


--
-- Name: cp_references cp_references_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_references
    ADD CONSTRAINT cp_references_pkey PRIMARY KEY (id);


--
-- Name: cp_third_parties cp_third_parties_compliance_portal_id_third_party_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_third_parties
    ADD CONSTRAINT cp_third_parties_compliance_portal_id_third_party_id_key UNIQUE (compliance_portal_id, third_party_id);


--
-- Name: cp_third_parties cp_third_parties_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_third_parties
    ADD CONSTRAINT cp_third_parties_pkey PRIMARY KEY (id);


--
-- Name: custom_domains custom_domains_domain_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.custom_domains
    ADD CONSTRAINT custom_domains_domain_key UNIQUE (domain);


--
-- Name: custom_domains custom_domains_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.custom_domains
    ADD CONSTRAINT custom_domains_pkey PRIMARY KEY (id);


--
-- Name: data data_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data
    ADD CONSTRAINT data_pkey PRIMARY KEY (id);


--
-- Name: data_third_parties data_vendors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data_third_parties
    ADD CONSTRAINT data_vendors_pkey PRIMARY KEY (datum_id, third_party_id);


--
-- Name: detected_trackers detected_trackers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.detected_trackers
    ADD CONSTRAINT detected_trackers_pkey PRIMARY KEY (id);


--
-- Name: device_enrollment_tokens device_enrollment_tokens_hashed_value_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_enrollment_tokens
    ADD CONSTRAINT device_enrollment_tokens_hashed_value_unique UNIQUE (hashed_value);


--
-- Name: device_enrollment_tokens device_enrollment_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_enrollment_tokens
    ADD CONSTRAINT device_enrollment_tokens_pkey PRIMARY KEY (id);


--
-- Name: device_postures device_postures_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_postures
    ADD CONSTRAINT device_postures_pkey PRIMARY KEY (id);


--
-- Name: devices devices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.devices
    ADD CONSTRAINT devices_pkey PRIMARY KEY (id);


--
-- Name: document_comment_threads document_comment_threads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comment_threads
    ADD CONSTRAINT document_comment_threads_pkey PRIMARY KEY (id);


--
-- Name: document_comments document_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comments
    ADD CONSTRAINT document_comments_pkey PRIMARY KEY (id);


--
-- Name: document_default_approvers document_default_approvers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_default_approvers
    ADD CONSTRAINT document_default_approvers_pkey PRIMARY KEY (document_id, approver_profile_id);


--
-- Name: document_suggestions document_suggestions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_suggestions
    ADD CONSTRAINT document_suggestions_pkey PRIMARY KEY (id);


--
-- Name: document_version_approval_decisions document_version_approval_decisions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_approval_decisions
    ADD CONSTRAINT document_version_approval_decisions_pkey PRIMARY KEY (id);


--
-- Name: document_version_approval_decisions document_version_approval_decisions_quorum_id_approver_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_approval_decisions
    ADD CONSTRAINT document_version_approval_decisions_quorum_id_approver_id_key UNIQUE (quorum_id, approver_id);


--
-- Name: document_version_approval_quorums document_version_approval_quorums_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_approval_quorums
    ADD CONSTRAINT document_version_approval_quorums_pkey PRIMARY KEY (id);


--
-- Name: document_versions document_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_versions
    ADD CONSTRAINT document_versions_pkey PRIMARY KEY (id);


--
-- Name: electronic_signature_events electronic_signature_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.electronic_signature_events
    ADD CONSTRAINT electronic_signature_events_pkey PRIMARY KEY (id);


--
-- Name: electronic_signatures electronic_signatures_organization_id_signer_email_document_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.electronic_signatures
    ADD CONSTRAINT electronic_signatures_organization_id_signer_email_document_key UNIQUE (organization_id, signer_email, document_type, file_id);


--
-- Name: electronic_signatures electronic_signatures_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.electronic_signatures
    ADD CONSTRAINT electronic_signatures_pkey PRIMARY KEY (id);


--
-- Name: email_attachments email_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_attachments
    ADD CONSTRAINT email_attachments_pkey PRIMARY KEY (id);


--
-- Name: emails emails_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.emails
    ADD CONSTRAINT emails_pkey PRIMARY KEY (id);


--
-- Name: employee_portal_documents employee_portal_documents_employee_portal_id_document_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_portal_documents
    ADD CONSTRAINT employee_portal_documents_employee_portal_id_document_id_key UNIQUE (employee_portal_id, document_id);


--
-- Name: employee_portal_documents employee_portal_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_portal_documents
    ADD CONSTRAINT employee_portal_documents_pkey PRIMARY KEY (id);


--
-- Name: employee_portals employee_portals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_portals
    ADD CONSTRAINT employee_portals_pkey PRIMARY KEY (id);


--
-- Name: evidences evidences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidences
    ADD CONSTRAINT evidences_pkey PRIMARY KEY (id);


--
-- Name: evidences evidences_reference_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidences
    ADD CONSTRAINT evidences_reference_id_key UNIQUE (task_id, reference_id);


--
-- Name: export_jobs export_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.export_jobs
    ADD CONSTRAINT export_jobs_pkey PRIMARY KEY (id);


--
-- Name: files files_file_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_file_key_key UNIQUE (file_key);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- Name: findings_audits findings_audits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings_audits
    ADD CONSTRAINT findings_audits_pkey PRIMARY KEY (finding_id, audit_id);


--
-- Name: findings findings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings
    ADD CONSTRAINT findings_pkey PRIMARY KEY (id);


--
-- Name: findings_tasks findings_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings_tasks
    ADD CONSTRAINT findings_tasks_pkey PRIMARY KEY (finding_id, task_id);


--
-- Name: frameworks frameworks_org_ref_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.frameworks
    ADD CONSTRAINT frameworks_org_ref_unique UNIQUE (organization_id, reference_id);


--
-- Name: frameworks frameworks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.frameworks
    ADD CONSTRAINT frameworks_pkey PRIMARY KEY (id);


--
-- Name: generated_documents generated_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_pkey PRIMARY KEY (organization_id);


--
-- Name: iam_membership_profiles iam_identity_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_membership_profiles
    ADD CONSTRAINT iam_identity_profiles_pkey PRIMARY KEY (id);


--
-- Name: iam_oauth2_access_tokens iam_oauth2_access_tokens_hashed_value_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_access_tokens
    ADD CONSTRAINT iam_oauth2_access_tokens_hashed_value_unique UNIQUE (hashed_value);


--
-- Name: iam_oauth2_access_tokens iam_oauth2_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_access_tokens
    ADD CONSTRAINT iam_oauth2_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: iam_oauth2_authorization_codes iam_oauth2_authorization_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_authorization_codes
    ADD CONSTRAINT iam_oauth2_authorization_codes_pkey PRIMARY KEY (id);


--
-- Name: iam_oauth2_clients iam_oauth2_clients_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_clients
    ADD CONSTRAINT iam_oauth2_clients_pkey PRIMARY KEY (id);


--
-- Name: iam_oauth2_consents iam_oauth2_consents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_consents
    ADD CONSTRAINT iam_oauth2_consents_pkey PRIMARY KEY (id);


--
-- Name: iam_oauth2_device_codes iam_oauth2_device_codes_device_code_hash_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_device_codes
    ADD CONSTRAINT iam_oauth2_device_codes_device_code_hash_unique UNIQUE (device_code_hash);


--
-- Name: iam_oauth2_device_codes iam_oauth2_device_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_device_codes
    ADD CONSTRAINT iam_oauth2_device_codes_pkey PRIMARY KEY (id);


--
-- Name: iam_oauth2_device_codes iam_oauth2_device_codes_user_code_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_device_codes
    ADD CONSTRAINT iam_oauth2_device_codes_user_code_unique UNIQUE (user_code);


--
-- Name: iam_oauth2_refresh_tokens iam_oauth2_refresh_tokens_hashed_value_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_refresh_tokens
    ADD CONSTRAINT iam_oauth2_refresh_tokens_hashed_value_unique UNIQUE (hashed_value);


--
-- Name: iam_oauth2_refresh_tokens iam_oauth2_refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_refresh_tokens
    ADD CONSTRAINT iam_oauth2_refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: iam_oidc_states iam_oidc_states_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oidc_states
    ADD CONSTRAINT iam_oidc_states_pkey PRIMARY KEY (id);


--
-- Name: iam_saml_assertions iam_saml_assertions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_saml_assertions
    ADD CONSTRAINT iam_saml_assertions_pkey PRIMARY KEY (id);


--
-- Name: iam_scim_bridges iam_scim_bridges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_bridges
    ADD CONSTRAINT iam_scim_bridges_pkey PRIMARY KEY (id);


--
-- Name: iam_scim_bridges iam_scim_bridges_scim_configuration_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_bridges
    ADD CONSTRAINT iam_scim_bridges_scim_configuration_unique UNIQUE (scim_configuration_id);


--
-- Name: iam_scim_configurations iam_scim_configurations_organization_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_configurations
    ADD CONSTRAINT iam_scim_configurations_organization_unique UNIQUE (organization_id);


--
-- Name: iam_scim_configurations iam_scim_configurations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_configurations
    ADD CONSTRAINT iam_scim_configurations_pkey PRIMARY KEY (id);


--
-- Name: iam_scim_events iam_scim_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_events
    ADD CONSTRAINT iam_scim_events_pkey PRIMARY KEY (id);


--
-- Name: iam_tokens iam_tokens_hashed_value_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_tokens
    ADD CONSTRAINT iam_tokens_hashed_value_unique UNIQUE (hashed_value);


--
-- Name: iam_tokens iam_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_tokens
    ADD CONSTRAINT iam_tokens_pkey PRIMARY KEY (id);


--
-- Name: internal_controls_documents internal_controls_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_documents
    ADD CONSTRAINT internal_controls_documents_pkey PRIMARY KEY (internal_control_id, document_id);


--
-- Name: internal_controls internal_controls_organization_id_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls
    ADD CONSTRAINT internal_controls_organization_id_code_key UNIQUE (organization_id, code);


--
-- Name: internal_controls internal_controls_organization_id_reference_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls
    ADD CONSTRAINT internal_controls_organization_id_reference_id_key UNIQUE (organization_id, reference_id);


--
-- Name: internal_controls_tasks internal_controls_tasks_internal_control_reference_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_tasks
    ADD CONSTRAINT internal_controls_tasks_internal_control_reference_id_key UNIQUE (internal_control_id, reference_id);


--
-- Name: internal_controls_tasks internal_controls_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_tasks
    ADD CONSTRAINT internal_controls_tasks_pkey PRIMARY KEY (internal_control_id, task_id);


--
-- Name: internal_controls_third_parties internal_controls_third_parties_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_third_parties
    ADD CONSTRAINT internal_controls_third_parties_pkey PRIMARY KEY (internal_control_id, third_party_id);


--
-- Name: linear_webhook_events linear_webhook_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.linear_webhook_events
    ADD CONSTRAINT linear_webhook_events_pkey PRIMARY KEY (delivery_id);


--
-- Name: mailing_list_subscribers mailing_list_subscribers_mailing_list_id_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mailing_list_subscribers
    ADD CONSTRAINT mailing_list_subscribers_mailing_list_id_email_key UNIQUE (mailing_list_id, email);


--
-- Name: mailing_list_subscribers mailing_list_subscribers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mailing_list_subscribers
    ADD CONSTRAINT mailing_list_subscribers_pkey PRIMARY KEY (id);


--
-- Name: mailing_list_updates mailing_list_updates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mailing_list_updates
    ADD CONSTRAINT mailing_list_updates_pkey PRIMARY KEY (id);


--
-- Name: mailing_lists mailing_lists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mailing_lists
    ADD CONSTRAINT mailing_lists_pkey PRIMARY KEY (id);


--
-- Name: internal_controls mitigations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls
    ADD CONSTRAINT mitigations_pkey PRIMARY KEY (id);


--
-- Name: notification_preferences notification_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_preferences
    ADD CONSTRAINT notification_preferences_pkey PRIMARY KEY (recipient_profile_id, notification_type);


--
-- Name: notification_weekly_sends notification_weekly_sends_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_weekly_sends
    ADD CONSTRAINT notification_weekly_sends_pkey PRIMARY KEY (recipient_profile_id, week_start);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: operation_receipts operation_receipts_organization_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operation_receipts
    ADD CONSTRAINT operation_receipts_organization_key_key UNIQUE (organization_id, operation_key);


--
-- Name: operation_receipts operation_receipts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operation_receipts
    ADD CONSTRAINT operation_receipts_pkey PRIMARY KEY (id);


--
-- Name: organization_contexts organization_contexts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_contexts
    ADD CONSTRAINT organization_contexts_pkey PRIMARY KEY (organization_id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: documents policies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT policies_pkey PRIMARY KEY (id);


--
-- Name: document_version_signatures policy_version_signatures_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_signatures
    ADD CONSTRAINT policy_version_signatures_pkey PRIMARY KEY (id);


--
-- Name: probot_identity_binding_challenges probot_identity_binding_challenges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.probot_identity_binding_challenges
    ADD CONSTRAINT probot_identity_binding_challenges_pkey PRIMARY KEY (hashed_token);


--
-- Name: probot_identity_bindings probot_identity_bindings_identity_id_provider_external_tena_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.probot_identity_bindings
    ADD CONSTRAINT probot_identity_bindings_identity_id_provider_external_tena_key UNIQUE (identity_id, provider, external_tenant_id);


--
-- Name: probot_identity_bindings probot_identity_bindings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.probot_identity_bindings
    ADD CONSTRAINT probot_identity_bindings_pkey PRIMARY KEY (id);


--
-- Name: probot_identity_bindings probot_identity_bindings_provider_external_tenant_id_extern_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.probot_identity_bindings
    ADD CONSTRAINT probot_identity_bindings_provider_external_tenant_id_extern_key UNIQUE (provider, external_tenant_id, external_user_id);


--
-- Name: processing_activity_data_protection_impact_assessments processing_activity_data_protection_impact_assessments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activity_data_protection_impact_assessments
    ADD CONSTRAINT processing_activity_data_protection_impact_assessments_pkey PRIMARY KEY (id);


--
-- Name: processing_activities processing_activity_registries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activities
    ADD CONSTRAINT processing_activity_registries_pkey PRIMARY KEY (id);


--
-- Name: processing_activity_transfer_impact_assessments processing_activity_transfer_impact_assessments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activity_transfer_impact_assessments
    ADD CONSTRAINT processing_activity_transfer_impact_assessments_pkey PRIMARY KEY (id);


--
-- Name: processing_activity_third_parties processing_activity_vendors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activity_third_parties
    ADD CONSTRAINT processing_activity_vendors_pkey PRIMARY KEY (processing_activity_id, third_party_id);


--
-- Name: resource_aliases resource_aliases_tenant_id_alias_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_aliases
    ADD CONSTRAINT resource_aliases_tenant_id_alias_key UNIQUE (tenant_id, alias);


--
-- Name: rights_requests rights_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rights_requests
    ADD CONSTRAINT rights_requests_pkey PRIMARY KEY (id);


--
-- Name: risk_analyses risk_analyses_document_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analyses
    ADD CONSTRAINT risk_analyses_document_id_key UNIQUE (document_id);


--
-- Name: risk_analyses risk_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analyses
    ADD CONSTRAINT risk_analyses_pkey PRIMARY KEY (id);


--
-- Name: risk_analysis_boundaries risk_analysis_boundaries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_boundaries
    ADD CONSTRAINT risk_analysis_boundaries_pkey PRIMARY KEY (id);


--
-- Name: risk_analysis_boundaries risk_analysis_boundaries_unique_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_boundaries
    ADD CONSTRAINT risk_analysis_boundaries_unique_name UNIQUE (risk_analysis_diagram_id, name);


--
-- Name: risk_analysis_diagrams risk_analysis_diagrams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_diagrams
    ADD CONSTRAINT risk_analysis_diagrams_pkey PRIMARY KEY (id);


--
-- Name: risk_analysis_nodes risk_analysis_nodes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_nodes
    ADD CONSTRAINT risk_analysis_nodes_pkey PRIMARY KEY (id);


--
-- Name: risk_analysis_nodes risk_analysis_nodes_unique_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_nodes
    ADD CONSTRAINT risk_analysis_nodes_unique_name UNIQUE (risk_analysis_diagram_id, name);


--
-- Name: risk_analysis_processes risk_analysis_processes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_processes
    ADD CONSTRAINT risk_analysis_processes_pkey PRIMARY KEY (id);


--
-- Name: risk_analysis_processes risk_analysis_processes_unique_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_processes
    ADD CONSTRAINT risk_analysis_processes_unique_name UNIQUE (risk_analysis_diagram_id, name);


--
-- Name: risk_analysis_scenario_risks risk_analysis_scenario_risks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_scenario_risks
    ADD CONSTRAINT risk_analysis_scenario_risks_pkey PRIMARY KEY (risk_analysis_scenario_id, risk_id);


--
-- Name: risk_analysis_scenario_threats risk_analysis_scenario_threats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_scenario_threats
    ADD CONSTRAINT risk_analysis_scenario_threats_pkey PRIMARY KEY (risk_analysis_scenario_id, risk_analysis_threat_id);


--
-- Name: risk_analysis_scenarios risk_analysis_scenarios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_scenarios
    ADD CONSTRAINT risk_analysis_scenarios_pkey PRIMARY KEY (id);


--
-- Name: risk_analysis_threats risk_analysis_threats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_threats
    ADD CONSTRAINT risk_analysis_threats_pkey PRIMARY KEY (id);


--
-- Name: risk_analysis_threats risk_analysis_threats_unique_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_threats
    ADD CONSTRAINT risk_analysis_threats_unique_name UNIQUE (risk_analysis_diagram_id, name);


--
-- Name: third_party_risk_assessments risk_assessments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_risk_assessments
    ADD CONSTRAINT risk_assessments_pkey PRIMARY KEY (id);


--
-- Name: risks_internal_controls risks_internal_controls_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks_internal_controls
    ADD CONSTRAINT risks_internal_controls_pkey PRIMARY KEY (risk_id, internal_control_id);


--
-- Name: risks_obligations risks_obligations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks_obligations
    ADD CONSTRAINT risks_obligations_pkey PRIMARY KEY (risk_id, obligation_id);


--
-- Name: risks risks_organization_id_reference_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks
    ADD CONSTRAINT risks_organization_id_reference_id_key UNIQUE (organization_id, reference_id);


--
-- Name: risks risks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks
    ADD CONSTRAINT risks_pkey PRIMARY KEY (id);


--
-- Name: risks_documents risks_policies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks_documents
    ADD CONSTRAINT risks_policies_pkey PRIMARY KEY (risk_id, document_id);


--
-- Name: search_resources search_resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_resources
    ADD CONSTRAINT search_resources_pkey PRIMARY KEY (resource_id);


--
-- Name: slack_delivery_operations slack_delivery_operations_organization_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_delivery_operations
    ADD CONSTRAINT slack_delivery_operations_organization_key_key UNIQUE (organization_id, operation_key);


--
-- Name: slack_delivery_operations slack_delivery_operations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_delivery_operations
    ADD CONSTRAINT slack_delivery_operations_pkey PRIMARY KEY (id);


--
-- Name: slack_messages slack_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_messages
    ADD CONSTRAINT slack_messages_pkey PRIMARY KEY (id);


--
-- Name: slackbot_bind_callbacks slackbot_bind_callbacks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_bind_callbacks
    ADD CONSTRAINT slackbot_bind_callbacks_pkey PRIMARY KEY (team_id, user_id);


--
-- Name: slackbot_events slackbot_events_event_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_events
    ADD CONSTRAINT slackbot_events_event_id_key UNIQUE (event_id);


--
-- Name: slackbot_events slackbot_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_events
    ADD CONSTRAINT slackbot_events_pkey PRIMARY KEY (id);


--
-- Name: slackbot_install_state_claims slackbot_install_state_claims_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_install_state_claims
    ADD CONSTRAINT slackbot_install_state_claims_pkey PRIMARY KEY (state_digest);


--
-- Name: slackbot_installations slackbot_installations_organization_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_installations
    ADD CONSTRAINT slackbot_installations_organization_id_key UNIQUE (organization_id);


--
-- Name: slackbot_installations slackbot_installations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_installations
    ADD CONSTRAINT slackbot_installations_pkey PRIMARY KEY (id);


--
-- Name: slackbot_installations slackbot_installations_team_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_installations
    ADD CONSTRAINT slackbot_installations_team_id_key UNIQUE (team_id);


--
-- Name: slackbot_interactive_claims slackbot_interactive_claims_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_interactive_claims
    ADD CONSTRAINT slackbot_interactive_claims_pkey PRIMARY KEY (interaction_digest);


--
-- Name: slackbot_interactive_commands slackbot_interactive_commands_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_interactive_commands
    ADD CONSTRAINT slackbot_interactive_commands_pkey PRIMARY KEY (id);


--
-- Name: slackbot_interactive_commands slackbot_interactive_commands_request_digest_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_interactive_commands
    ADD CONSTRAINT slackbot_interactive_commands_request_digest_key UNIQUE (request_digest);


--
-- Name: slackbot_messages slackbot_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_messages
    ADD CONSTRAINT slackbot_messages_pkey PRIMARY KEY (id);


--
-- Name: slackbot_processed_events slackbot_processed_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_processed_events
    ADD CONSTRAINT slackbot_processed_events_pkey PRIMARY KEY (event_id);


--
-- Name: statements_of_applicability statements_of_applicability_document_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.statements_of_applicability
    ADD CONSTRAINT statements_of_applicability_document_id_key UNIQUE (document_id);


--
-- Name: applicability_statements states_of_applicability_contr_state_of_applicability_id_con_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applicability_statements
    ADD CONSTRAINT states_of_applicability_contr_state_of_applicability_id_con_key UNIQUE (statement_of_applicability_id, control_id);


--
-- Name: statements_of_applicability states_of_applicability_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.statements_of_applicability
    ADD CONSTRAINT states_of_applicability_pkey PRIMARY KEY (id);


--
-- Name: tag_assignments tag_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tag_assignments
    ADD CONSTRAINT tag_assignments_pkey PRIMARY KEY (id);


--
-- Name: tag_assignments tag_assignments_resource_id_tag_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tag_assignments
    ADD CONSTRAINT tag_assignments_resource_id_tag_id_key UNIQUE (resource_id, tag_id);


--
-- Name: tags tags_organization_id_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_organization_id_name_key UNIQUE (organization_id, name);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- Name: task_activities task_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_activities
    ADD CONSTRAINT task_activities_pkey PRIMARY KEY (id);


--
-- Name: task_comment_external_links task_comment_external_links_external_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comment_external_links
    ADD CONSTRAINT task_comment_external_links_external_key UNIQUE (organization_id, provider, external_id);


--
-- Name: task_comment_external_links task_comment_external_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comment_external_links
    ADD CONSTRAINT task_comment_external_links_pkey PRIMARY KEY (task_comment_id);


--
-- Name: task_comments task_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_pkey PRIMARY KEY (id);


--
-- Name: task_external_links task_external_links_organization_id_provider_external_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_external_links
    ADD CONSTRAINT task_external_links_organization_id_provider_external_id_key UNIQUE (organization_id, provider, external_id);


--
-- Name: task_external_links task_external_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_external_links
    ADD CONSTRAINT task_external_links_pkey PRIMARY KEY (task_id);


--
-- Name: task_sync_jobs task_sync_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_sync_jobs
    ADD CONSTRAINT task_sync_jobs_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_organization_id_state_priority_rank_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_organization_id_state_priority_rank_key UNIQUE (organization_id, state, priority, rank) DEFERRABLE INITIALLY DEFERRED;


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: third_party_administrators third_party_administrators_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_administrators
    ADD CONSTRAINT third_party_administrators_pkey PRIMARY KEY (third_party_id, administrator_profile_id);


--
-- Name: tracker_pattern_daily_hits tracker_pattern_daily_hits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_pattern_daily_hits
    ADD CONSTRAINT tracker_pattern_daily_hits_pkey PRIMARY KEY (tracker_pattern_id, day, browser_family, presentation, consent_action, category_granted);


--
-- Name: tracker_pattern_discovery_hits tracker_pattern_discovery_hits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_pattern_discovery_hits
    ADD CONSTRAINT tracker_pattern_discovery_hits_pkey PRIMARY KEY (tracker_pattern_id);


--
-- Name: tracker_patterns tracker_patterns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_patterns
    ADD CONSTRAINT tracker_patterns_pkey PRIMARY KEY (id);


--
-- Name: tracker_resources tracker_resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_resources
    ADD CONSTRAINT tracker_resources_pkey PRIMARY KEY (id);


--
-- Name: treatment_plans_internal_controls treatment_plans_internal_controls_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans_internal_controls
    ADD CONSTRAINT treatment_plans_internal_controls_pkey PRIMARY KEY (treatment_plan_id, internal_control_id);


--
-- Name: treatment_plans treatment_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans
    ADD CONSTRAINT treatment_plans_pkey PRIMARY KEY (id);


--
-- Name: treatment_plans treatment_plans_risk_analysis_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans
    ADD CONSTRAINT treatment_plans_risk_analysis_unique UNIQUE (risk_id, risk_analysis_id);


--
-- Name: resource_aliases trust_center_aliases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_aliases
    ADD CONSTRAINT trust_center_aliases_pkey PRIMARY KEY (resource_id);


--
-- Name: iam_sessions usrmgr_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_sessions
    ADD CONSTRAINT usrmgr_sessions_pkey PRIMARY KEY (id);


--
-- Name: identities usrmgr_users_email_address_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT usrmgr_users_email_address_key UNIQUE (email_address);


--
-- Name: identities usrmgr_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT usrmgr_users_pkey PRIMARY KEY (id);


--
-- Name: third_party_business_associate_agreements vendor_business_associate_agreements_organization_vendor_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_business_associate_agreements
    ADD CONSTRAINT vendor_business_associate_agreements_organization_vendor_unique UNIQUE (organization_id, third_party_id);


--
-- Name: third_party_business_associate_agreements vendor_business_associate_agreements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_business_associate_agreements
    ADD CONSTRAINT vendor_business_associate_agreements_pkey PRIMARY KEY (id);


--
-- Name: third_party_compliance_reports vendor_compliance_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_compliance_reports
    ADD CONSTRAINT vendor_compliance_reports_pkey PRIMARY KEY (id);


--
-- Name: third_party_contacts vendor_contacts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_contacts
    ADD CONSTRAINT vendor_contacts_pkey PRIMARY KEY (id);


--
-- Name: third_party_data_privacy_agreements vendor_data_privacy_agreements_organization_vendor_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_data_privacy_agreements
    ADD CONSTRAINT vendor_data_privacy_agreements_organization_vendor_unique UNIQUE (organization_id, third_party_id);


--
-- Name: third_party_data_privacy_agreements vendor_data_privacy_agreements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_data_privacy_agreements
    ADD CONSTRAINT vendor_data_privacy_agreements_pkey PRIMARY KEY (id);


--
-- Name: third_party_services vendor_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_services
    ADD CONSTRAINT vendor_services_pkey PRIMARY KEY (id);


--
-- Name: third_parties vendors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_parties
    ADD CONSTRAINT vendors_pkey PRIMARY KEY (id);


--
-- Name: webhook_data webhook_data_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_data
    ADD CONSTRAINT webhook_data_pkey PRIMARY KEY (id);


--
-- Name: webhook_events webhook_events_data_subscription_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_data_subscription_unique UNIQUE (webhook_data_id, webhook_subscription_id);


--
-- Name: webhook_events webhook_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_pkey PRIMARY KEY (id);


--
-- Name: webhook_subscriptions webhook_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_subscriptions
    ADD CONSTRAINT webhook_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: agent_executions_source_session_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX agent_executions_source_session_key ON public.agent_executions USING btree (tenant_id, organization_id, source, session_key) WHERE ((source IS NOT NULL) AND (session_key IS NOT NULL));


--
-- Name: agent_inputs_source_event_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX agent_inputs_source_event_key ON public.agent_inputs USING btree (tenant_id, organization_id, source, source_event_id) WHERE (source_event_id IS NOT NULL);


--
-- Name: bot_messages_event_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX bot_messages_event_key ON public.bot_messages USING btree (tenant_id, organization_id, subject_namespace, subject_key, event_key);


--
-- Name: bot_thread_subjects_provider_coordinates; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX bot_thread_subjects_provider_coordinates ON public.bot_thread_subjects USING btree (tenant_id, organization_id, provider, external_conversation_id, external_message_id);


--
-- Name: bot_thread_subjects_subject; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX bot_thread_subjects_subject ON public.bot_thread_subjects USING btree (tenant_id, organization_id, subject_namespace, subject_key);


--
-- Name: common_third_parties_enrichment_requested_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX common_third_parties_enrichment_requested_at_idx ON public.common_third_parties USING btree (enrichment_requested_at) WHERE (enrichment_requested_at IS NOT NULL);


--
-- Name: common_third_parties_review_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX common_third_parties_review_idx ON public.common_third_parties USING btree (review);


--
-- Name: common_third_parties_slug_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX common_third_parties_slug_key ON public.common_third_parties USING btree (slug);


--
-- Name: common_third_party_domains_party_domain_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX common_third_party_domains_party_domain_key ON public.common_third_party_domains USING btree (common_third_party_id, domain);


--
-- Name: controls_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX controls_search_idx ON public.controls USING gin (search_vector);


--
-- Name: cookie_banner_scans_one_active_per_banner_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX cookie_banner_scans_one_active_per_banner_idx ON public.cookie_banner_scans USING btree (cookie_banner_id) WHERE (state = ANY (ARRAY['PENDING'::public.cookie_banner_scan_state, 'RUNNING'::public.cookie_banner_scan_state, 'PAUSED'::public.cookie_banner_scan_state]));


--
-- Name: device_enrollment_tokens_device_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX device_enrollment_tokens_device_id_idx ON public.device_enrollment_tokens USING btree (device_id);


--
-- Name: device_postures_device_id_check_key_observed_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX device_postures_device_id_check_key_observed_at_idx ON public.device_postures USING btree (device_id, check_key, observed_at DESC);


--
-- Name: device_postures_device_id_correlation_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX device_postures_device_id_correlation_id_created_at_idx ON public.device_postures USING btree (device_id, correlation_id, created_at DESC);


--
-- Name: device_postures_organization_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX device_postures_organization_id_idx ON public.device_postures USING btree (organization_id);


--
-- Name: devices_api_key_hash_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX devices_api_key_hash_idx ON public.devices USING btree (api_key_hash);


--
-- Name: devices_org_hardware_uuid_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX devices_org_hardware_uuid_idx ON public.devices USING btree (organization_id, hardware_uuid) WHERE ((hardware_uuid IS NOT NULL) AND (state <> 'REVOKED'::public.device_state) AND (deleted_at IS NULL));


--
-- Name: devices_owner_profile_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX devices_owner_profile_idx ON public.devices USING btree (owner_profile_id) WHERE (owner_profile_id IS NOT NULL);


--
-- Name: document_comment_threads_document_id_created_at_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX document_comment_threads_document_id_created_at_id_idx ON public.document_comment_threads USING btree (document_id, created_at, id);


--
-- Name: document_comment_threads_open_suggestion_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX document_comment_threads_open_suggestion_key ON public.document_comment_threads USING btree (suggestion_id) WHERE ((resolved_at IS NULL) AND (suggestion_id IS NOT NULL));


--
-- Name: document_comments_thread_id_created_at_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX document_comments_thread_id_created_at_id_idx ON public.document_comments USING btree (thread_id, created_at, id);


--
-- Name: document_one_pending_approval_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX document_one_pending_approval_idx ON public.document_versions USING btree (document_id) WHERE (status = 'PENDING_APPROVAL'::public.document_version_status);


--
-- Name: document_one_pending_quorum_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX document_one_pending_quorum_idx ON public.document_version_approval_quorums USING btree (version_id) WHERE (status = 'PENDING'::public.document_version_approval_quorum_status);


--
-- Name: document_suggestions_document_id_created_at_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX document_suggestions_document_id_created_at_id_idx ON public.document_suggestions USING btree (document_id, created_at, id);


--
-- Name: document_versions_one_published_number_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX document_versions_one_published_number_idx ON public.document_versions USING btree (document_id, major, minor) WHERE (status = 'PUBLISHED'::public.document_version_status);


--
-- Name: document_versions_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX document_versions_search_idx ON public.document_versions USING gin (search_vector);


--
-- Name: documents_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX documents_search_idx ON public.documents USING gin (search_vector);


--
-- Name: emails_sent_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX emails_sent_at_idx ON public.emails USING btree (sent_at) WHERE (sent_at IS NULL);


--
-- Name: findings_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX findings_search_idx ON public.findings USING gin (search_vector);


--
-- Name: iam_membership_profiles_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX iam_membership_profiles_search_idx ON public.iam_membership_profiles USING gin (search_vector);


--
-- Name: iam_oauth2_authorization_codes_hashed_value_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX iam_oauth2_authorization_codes_hashed_value_unique ON public.iam_oauth2_authorization_codes USING btree (hashed_value) WHERE (hashed_value IS NOT NULL);


--
-- Name: iam_oauth2_clients_external_client_id_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX iam_oauth2_clients_external_client_id_unique ON public.iam_oauth2_clients USING btree (external_client_id) WHERE (external_client_id IS NOT NULL);


--
-- Name: iam_scim_events_membership_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX iam_scim_events_membership_id_idx ON public.iam_scim_events USING btree (membership_id);


--
-- Name: iam_sessions_membership_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX iam_sessions_membership_id_idx ON public.iam_sessions USING btree (membership_id);


--
-- Name: iam_sessions_parent_membership_created_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX iam_sessions_parent_membership_created_idx ON public.iam_sessions USING btree (parent_session_id, membership_id, created_at DESC);


--
-- Name: idx_access_review_entries_campaign_source_account_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_access_review_entries_campaign_source_account_key ON public.access_review_entries USING btree (access_review_campaign_source_id, account_key);


--
-- Name: idx_access_review_sources_connector_account; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_access_review_sources_connector_account ON public.access_review_sources USING btree (connector_account_id) WHERE (connector_account_id IS NOT NULL);


--
-- Name: idx_audit_log_entries_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_entries_action ON public.audit_log_entries USING btree (action);


--
-- Name: idx_audit_log_entries_actor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_entries_actor_id ON public.audit_log_entries USING btree (actor_id);


--
-- Name: idx_audit_log_entries_organization_id_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_entries_organization_id_created_at ON public.audit_log_entries USING btree (organization_id, tenant_id, created_at DESC, id DESC);


--
-- Name: idx_auth_saml_assertions_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_saml_assertions_expires_at ON public.iam_saml_assertions USING btree (expires_at);


--
-- Name: idx_auth_saml_assertions_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_saml_assertions_organization_id ON public.iam_saml_assertions USING btree (organization_id);


--
-- Name: idx_auth_saml_configurations_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_saml_configurations_organization_id ON public.iam_saml_configurations USING btree (organization_id);


--
-- Name: idx_auth_saml_configurations_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_saml_configurations_tenant_id ON public.iam_saml_configurations USING btree (tenant_id);


--
-- Name: idx_auth_saml_requests_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_saml_requests_expires_at ON public.iam_saml_requests USING btree (expires_at);


--
-- Name: idx_auth_saml_requests_id_org; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auth_saml_requests_id_org ON public.iam_saml_requests USING btree (id, organization_id);


--
-- Name: idx_certificate_cache_expires; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_certificate_cache_expires ON public.cached_certificates USING btree (expires_at);


--
-- Name: idx_certificates_http_challenge_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_certificates_http_challenge_token ON public.certificates USING btree (http_challenge_token) WHERE (http_challenge_token IS NOT NULL);


--
-- Name: idx_certificates_ssl_expires; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_certificates_ssl_expires ON public.certificates USING btree (ssl_expires_at) WHERE (status = 'ACTIVE'::public.custom_domain_ssl_status);


--
-- Name: idx_common_ip_country_blocks_cidr; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_common_ip_country_blocks_cidr ON public.common_ip_country_blocks USING gist (cidr inet_ops);


--
-- Name: idx_common_ip_location_blocks_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_common_ip_location_blocks_lookup ON public.common_ip_location_blocks USING btree (address_family, ip_start DESC, ip_end);


--
-- Name: idx_common_ip_location_blocks_range; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_common_ip_location_blocks_range ON public.common_ip_location_blocks USING btree (address_family, ip_start, ip_end);


--
-- Name: idx_common_tracker_patterns_enrichment; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_common_tracker_patterns_enrichment ON public.common_tracker_patterns USING btree (enrichment_requested_at) WHERE (enrichment_requested_at IS NOT NULL);


--
-- Name: idx_common_tracker_patterns_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_common_tracker_patterns_unique ON public.common_tracker_patterns USING btree (tracker_type, pattern, COALESCE(max_age_seconds, '-1'::integer));


--
-- Name: idx_connector_accounts_connector_external; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_connector_accounts_connector_external ON public.connector_accounts USING btree (connector_id, external_account_id);


--
-- Name: idx_connector_accounts_name_sync; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_connector_accounts_name_sync ON public.connector_accounts USING btree (created_at) WHERE (name_synced_at IS NULL);


--
-- Name: idx_connectors_organization_id_provider_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_connectors_organization_id_provider_lookup ON public.connectors USING btree (organization_id, provider);


--
-- Name: idx_cookie_banner_translations_unique_language_per_banner; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_cookie_banner_translations_unique_language_per_banner ON public.cookie_banner_translations USING btree (cookie_banner_id, language);


--
-- Name: idx_cookie_banners_unique_active_origin; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_cookie_banners_unique_active_origin ON public.cookie_banners USING btree (tenant_id, origin) WHERE (state = 'ACTIVE'::public.cookie_banner_state);


--
-- Name: idx_cookie_categories_one_necessary_per_banner; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_cookie_categories_one_necessary_per_banner ON public.cookie_categories USING btree (cookie_banner_id) WHERE (kind = 'NECESSARY'::public.cookie_category_kind);


--
-- Name: idx_cookie_categories_one_uncategorised_per_banner; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_cookie_categories_one_uncategorised_per_banner ON public.cookie_categories USING btree (cookie_banner_id) WHERE (kind = 'UNCATEGORISED'::public.cookie_category_kind);


--
-- Name: idx_cookie_categories_unique_posthog_per_banner; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_cookie_categories_unique_posthog_per_banner ON public.cookie_categories USING btree (cookie_banner_id) WHERE (posthog_consent = true);


--
-- Name: idx_cookie_categories_unique_slug_per_banner; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_cookie_categories_unique_slug_per_banner ON public.cookie_categories USING btree (cookie_banner_id, slug);


--
-- Name: idx_custom_domains_domain; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_custom_domains_domain ON public.custom_domains USING btree (domain);


--
-- Name: idx_detected_trackers_cookie_banner_scan_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_detected_trackers_cookie_banner_scan_id ON public.detected_trackers USING btree (cookie_banner_scan_id) WHERE (cookie_banner_scan_id IS NOT NULL);


--
-- Name: idx_detected_trackers_initiator_domain; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_detected_trackers_initiator_domain ON public.detected_trackers USING btree (initiator_domain) WHERE (initiator_domain IS NOT NULL);


--
-- Name: idx_detected_trackers_unique_per_banner_observer; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_detected_trackers_unique_per_banner_observer ON public.detected_trackers USING btree (cookie_banner_id, tracker_type, identifier, COALESCE(cookie_domain, ''::public.citext), observed_by);


--
-- Name: idx_email_attachments_email_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_attachments_email_id ON public.email_attachments USING btree (email_id);


--
-- Name: idx_emails_pending; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_emails_pending ON public.emails USING btree (status, created_at) WHERE (status = 'PENDING'::text);


--
-- Name: idx_employee_portals_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_employee_portals_organization_id ON public.employee_portals USING btree (organization_id);


--
-- Name: idx_esig_events_signature; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_esig_events_signature ON public.electronic_signature_events USING btree (electronic_signature_id, occurred_at);


--
-- Name: idx_fetch_attempts_queued; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fetch_attempts_queued ON public.access_review_campaign_source_fetch_attempts USING btree (created_at) WHERE (status = 'QUEUED'::public.access_review_campaign_source_fetch_status);


--
-- Name: idx_findings_audits_audit_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_findings_audits_audit_id ON public.findings_audits USING btree (audit_id);


--
-- Name: idx_findings_org_reference_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_findings_org_reference_id ON public.findings USING btree (organization_id, reference_id);


--
-- Name: idx_iam_memberships_identity_tenant_org; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_iam_memberships_identity_tenant_org ON public.iam_memberships USING btree (identity_id, tenant_id, organization_id);


--
-- Name: idx_iam_oidc_states_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_iam_oidc_states_expires_at ON public.iam_oidc_states USING btree (expires_at);


--
-- Name: idx_iam_scim_bridges_connector_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_iam_scim_bridges_connector_id ON public.iam_scim_bridges USING btree (connector_id) WHERE (connector_id IS NOT NULL);


--
-- Name: idx_iam_scim_bridges_next_sync; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_iam_scim_bridges_next_sync ON public.iam_scim_bridges USING btree (next_sync_at) WHERE (state = ANY (ARRAY['ACTIVE'::text, 'FAILED'::text, 'SYNCING'::text]));


--
-- Name: idx_iam_scim_configurations_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_iam_scim_configurations_organization_id ON public.iam_scim_configurations USING btree (organization_id);


--
-- Name: idx_iam_scim_configurations_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_iam_scim_configurations_tenant_id ON public.iam_scim_configurations USING btree (tenant_id);


--
-- Name: idx_iam_scim_events_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_iam_scim_events_created_at ON public.iam_scim_events USING btree (created_at DESC);


--
-- Name: idx_iam_scim_events_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_iam_scim_events_organization_id ON public.iam_scim_events USING btree (organization_id);


--
-- Name: idx_iam_scim_events_scim_configuration_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_iam_scim_events_scim_configuration_id ON public.iam_scim_events USING btree (scim_configuration_id);


--
-- Name: idx_iam_scim_events_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_iam_scim_events_tenant_id ON public.iam_scim_events USING btree (tenant_id);


--
-- Name: idx_mitigations_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_mitigations_organization_id ON public.internal_controls USING btree (organization_id);


--
-- Name: idx_profiles_external_id_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_profiles_external_id_organization_id ON public.iam_membership_profiles USING btree (external_id, organization_id) WHERE (external_id IS NOT NULL);


--
-- Name: idx_profiles_identity_id_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_profiles_identity_id_organization_id ON public.iam_membership_profiles USING btree (identity_id, organization_id);


--
-- Name: idx_profiles_user_name_organization_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_profiles_user_name_organization_id ON public.iam_membership_profiles USING btree (user_name, organization_id) WHERE (user_name IS NOT NULL);


--
-- Name: idx_saml_config_domain_org_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saml_config_domain_org_unique ON public.iam_saml_configurations USING btree (organization_id, email_domain);


--
-- Name: idx_saml_config_domain_verification_token_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saml_config_domain_verification_token_unique ON public.iam_saml_configurations USING btree (domain_verification_token);


--
-- Name: idx_tracker_patterns_unique_pattern_per_banner; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tracker_patterns_unique_pattern_per_banner ON public.tracker_patterns USING btree (cookie_banner_id, tracker_type, pattern, COALESCE(max_age_seconds, '-1'::integer));


--
-- Name: idx_tracker_resources_cookie_banner_scan_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tracker_resources_cookie_banner_scan_id ON public.tracker_resources USING btree (cookie_banner_scan_id) WHERE (cookie_banner_scan_id IS NOT NULL);


--
-- Name: idx_tracker_resources_unique_resource_per_banner; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tracker_resources_unique_resource_per_banner ON public.tracker_resources USING btree (cookie_banner_id, resource_type, origin, path);


--
-- Name: idx_users_saml_subject; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_saml_subject ON public.identities USING btree (saml_subject) WHERE (saml_subject IS NOT NULL);


--
-- Name: internal_control_events_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX internal_control_events_search_idx ON public.internal_control_events USING gin (search_vector);


--
-- Name: internal_controls_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX internal_controls_search_idx ON public.internal_controls USING gin (search_vector);


--
-- Name: obligations_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX obligations_search_idx ON public.obligations USING gin (search_vector);


--
-- Name: probot_identity_binding_challenges_expires_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX probot_identity_binding_challenges_expires_at_idx ON public.probot_identity_binding_challenges USING btree (expires_at);


--
-- Name: processing_activity_dpias_processing_activity_id_uniq; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX processing_activity_dpias_processing_activity_id_uniq ON public.processing_activity_data_protection_impact_assessments USING btree (processing_activity_id);


--
-- Name: processing_activity_tias_processing_activity_id_uniq; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX processing_activity_tias_processing_activity_id_uniq ON public.processing_activity_transfer_impact_assessments USING btree (processing_activity_id);


--
-- Name: risks_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX risks_search_idx ON public.risks USING gin (search_vector);


--
-- Name: search_resources_organization_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX search_resources_organization_id_idx ON public.search_resources USING btree (organization_id);


--
-- Name: search_resources_search_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX search_resources_search_idx ON public.search_resources USING gin (search_vector);


--
-- Name: slack_messages_sent_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX slack_messages_sent_at_idx ON public.slack_messages USING btree (sent_at) WHERE ((sent_at IS NULL) AND (error IS NULL));


--
-- Name: slack_messages_source_event_id_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX slack_messages_source_event_id_unique ON public.slack_messages USING btree (tenant_id, ((metadata ->> 'source_event_id'::text))) WHERE ((metadata ->> 'source_event_id'::text) IS NOT NULL);


--
-- Name: slackbot_messages_source_event_id_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX slackbot_messages_source_event_id_unique ON public.slackbot_messages USING btree (tenant_id, ((metadata ->> 'source_event_id'::text))) WHERE ((metadata ->> 'source_event_id'::text) IS NOT NULL);


--
-- Name: statements_of_applicability_name_organization_id_uniq; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX statements_of_applicability_name_organization_id_uniq ON public.statements_of_applicability USING btree (name, organization_id);


--
-- Name: tag_assignments_tag_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX tag_assignments_tag_id_idx ON public.tag_assignments USING btree (tag_id);


--
-- Name: third_parties_common_third_party_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX third_parties_common_third_party_id_idx ON public.third_parties USING btree (common_third_party_id) WHERE (common_third_party_id IS NOT NULL);


--
-- Name: usrmgr_users_organization_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX usrmgr_users_organization_id_idx ON public.identities USING btree (organization_id);


--
-- Name: access_review_campaign_source_fetch_attempts access_review_campaign_source_access_review_campaign_sourc_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_source_fetch_attempts
    ADD CONSTRAINT access_review_campaign_source_access_review_campaign_sourc_fkey FOREIGN KEY (access_review_campaign_source_id) REFERENCES public.access_review_campaign_sources(id) ON DELETE CASCADE;


--
-- Name: access_review_campaign_source_fetch_attempts access_review_campaign_source_fetch_attemp_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_source_fetch_attempts
    ADD CONSTRAINT access_review_campaign_source_fetch_attemp_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: access_review_campaign_sources access_review_campaign_sources_access_review_campaign_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_sources
    ADD CONSTRAINT access_review_campaign_sources_access_review_campaign_id_fkey FOREIGN KEY (access_review_campaign_id) REFERENCES public.access_review_campaigns(id) ON DELETE CASCADE;


--
-- Name: access_review_campaign_sources access_review_campaign_sources_access_review_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_sources
    ADD CONSTRAINT access_review_campaign_sources_access_review_source_id_fkey FOREIGN KEY (access_review_source_id) REFERENCES public.access_review_sources(id) ON DELETE SET NULL;


--
-- Name: access_review_campaign_sources access_review_campaign_sources_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaign_sources
    ADD CONSTRAINT access_review_campaign_sources_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: access_review_campaigns access_review_campaigns_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_campaigns
    ADD CONSTRAINT access_review_campaigns_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: access_review_entries access_review_entries_access_review_campaign_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_entries
    ADD CONSTRAINT access_review_entries_access_review_campaign_id_fkey FOREIGN KEY (access_review_campaign_id) REFERENCES public.access_review_campaigns(id) ON DELETE CASCADE;


--
-- Name: access_review_entries access_review_entries_campaign_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_entries
    ADD CONSTRAINT access_review_entries_campaign_source_id_fkey FOREIGN KEY (access_review_campaign_source_id) REFERENCES public.access_review_campaign_sources(id) ON DELETE CASCADE;


--
-- Name: access_review_entries access_review_entries_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_entries
    ADD CONSTRAINT access_review_entries_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id) ON DELETE SET NULL;


--
-- Name: access_review_entries access_review_entries_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_entries
    ADD CONSTRAINT access_review_entries_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: access_review_entry_decision_history access_review_entry_decision_history_access_review_entry_id_fke; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_entry_decision_history
    ADD CONSTRAINT access_review_entry_decision_history_access_review_entry_id_fke FOREIGN KEY (access_review_entry_id) REFERENCES public.access_review_entries(id) ON DELETE CASCADE;


--
-- Name: access_review_entry_decision_history access_review_entry_decision_history_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_entry_decision_history
    ADD CONSTRAINT access_review_entry_decision_history_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: access_review_sources access_review_sources_connector_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_sources
    ADD CONSTRAINT access_review_sources_connector_account_id_fkey FOREIGN KEY (connector_account_id) REFERENCES public.connector_accounts(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: access_review_sources access_review_sources_connector_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_sources
    ADD CONSTRAINT access_review_sources_connector_id_fkey FOREIGN KEY (connector_id) REFERENCES public.connectors(id);


--
-- Name: access_review_sources access_review_sources_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.access_review_sources
    ADD CONSTRAINT access_review_sources_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: agent_execution_anchors agent_execution_anchors_agent_execution_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_execution_anchors
    ADD CONSTRAINT agent_execution_anchors_agent_execution_id_fkey FOREIGN KEY (agent_execution_id) REFERENCES public.agent_executions(id) ON DELETE CASCADE;


--
-- Name: agent_execution_anchors agent_execution_anchors_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_execution_anchors
    ADD CONSTRAINT agent_execution_anchors_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: agent_executions agent_executions_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_executions
    ADD CONSTRAINT agent_executions_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: agent_inputs agent_inputs_agent_execution_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_inputs
    ADD CONSTRAINT agent_inputs_agent_execution_id_fkey FOREIGN KEY (agent_execution_id) REFERENCES public.agent_executions(id) ON DELETE CASCADE;


--
-- Name: agent_inputs agent_inputs_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_inputs
    ADD CONSTRAINT agent_inputs_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id) ON DELETE SET NULL;


--
-- Name: agent_inputs agent_inputs_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.agent_inputs
    ADD CONSTRAINT agent_inputs_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: ai_systems ai_systems_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_systems
    ADD CONSTRAINT ai_systems_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: ai_systems ai_systems_owner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_systems
    ADD CONSTRAINT ai_systems_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.iam_membership_profiles(id);


--
-- Name: applicability_statements applicability_statements_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applicability_statements
    ADD CONSTRAINT applicability_statements_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: asset_third_parties asset_vendors_asset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.asset_third_parties
    ADD CONSTRAINT asset_vendors_asset_id_fkey FOREIGN KEY (asset_id) REFERENCES public.assets(id) ON DELETE CASCADE;


--
-- Name: asset_third_parties asset_vendors_vendor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.asset_third_parties
    ADD CONSTRAINT asset_vendors_vendor_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON DELETE CASCADE;


--
-- Name: assets assets_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assets
    ADD CONSTRAINT assets_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: assets assets_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.assets
    ADD CONSTRAINT assets_owner_profile_id_fkey FOREIGN KEY (owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: audit_log_entries audit_log_entries_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_log_entries
    ADD CONSTRAINT audit_log_entries_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: audits_frameworks audits_frameworks_audit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audits_frameworks
    ADD CONSTRAINT audits_frameworks_audit_id_fkey FOREIGN KEY (audit_id) REFERENCES public.audits(id) ON DELETE CASCADE;


--
-- Name: audits_frameworks audits_frameworks_framework_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audits_frameworks
    ADD CONSTRAINT audits_frameworks_framework_id_fkey FOREIGN KEY (framework_id) REFERENCES public.frameworks(id) ON DELETE CASCADE;


--
-- Name: audits_frameworks audits_frameworks_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audits_frameworks
    ADD CONSTRAINT audits_frameworks_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: audits audits_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audits
    ADD CONSTRAINT audits_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: audits audits_report_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audits
    ADD CONSTRAINT audits_report_file_id_fkey FOREIGN KEY (report_file_id) REFERENCES public.files(id);


--
-- Name: iam_invitations authz_invitations_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_invitations
    ADD CONSTRAINT authz_invitations_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: iam_memberships authz_memberships_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_memberships
    ADD CONSTRAINT authz_memberships_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: bot_delivery_destinations bot_delivery_destinations_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_delivery_destinations
    ADD CONSTRAINT bot_delivery_destinations_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: bot_messages bot_messages_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_messages
    ADD CONSTRAINT bot_messages_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: bot_thread_subjects bot_thread_subjects_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bot_thread_subjects
    ADD CONSTRAINT bot_thread_subjects_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: business_function_assets business_function_assets_asset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_function_assets
    ADD CONSTRAINT business_function_assets_asset_id_fkey FOREIGN KEY (asset_id) REFERENCES public.assets(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: business_function_assets business_function_assets_business_function_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_function_assets
    ADD CONSTRAINT business_function_assets_business_function_id_fkey FOREIGN KEY (business_function_id) REFERENCES public.business_functions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: business_function_assets business_function_assets_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_function_assets
    ADD CONSTRAINT business_function_assets_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: business_function_third_parties business_function_third_parties_business_function_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_function_third_parties
    ADD CONSTRAINT business_function_third_parties_business_function_id_fkey FOREIGN KEY (business_function_id) REFERENCES public.business_functions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: business_function_third_parties business_function_third_parties_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_function_third_parties
    ADD CONSTRAINT business_function_third_parties_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: business_function_third_parties business_function_third_parties_third_party_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_function_third_parties
    ADD CONSTRAINT business_function_third_parties_third_party_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: business_functions business_functions_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_functions
    ADD CONSTRAINT business_functions_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: business_functions business_functions_owner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_functions
    ADD CONSTRAINT business_functions_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.iam_membership_profiles(id);


--
-- Name: cached_certificates cached_certificates_certificate_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cached_certificates
    ADD CONSTRAINT cached_certificates_certificate_id_fkey FOREIGN KEY (certificate_id) REFERENCES public.certificates(id) ON DELETE CASCADE;


--
-- Name: cmp_domain_surfaces cmp_domain_surfaces_cmp_domain_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cmp_domain_surfaces
    ADD CONSTRAINT cmp_domain_surfaces_cmp_domain_id_fkey FOREIGN KEY (cmp_domain_id) REFERENCES public.cmp_domains(id) ON DELETE CASCADE;


--
-- Name: cmp_domain_surfaces cmp_domain_surfaces_discovered_by_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cmp_domain_surfaces
    ADD CONSTRAINT cmp_domain_surfaces_discovered_by_cookie_banner_id_fkey FOREIGN KEY (discovered_by_cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE SET NULL;


--
-- Name: cmp_domain_surfaces cmp_domain_surfaces_last_cookie_banner_scan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cmp_domain_surfaces
    ADD CONSTRAINT cmp_domain_surfaces_last_cookie_banner_scan_id_fkey FOREIGN KEY (last_cookie_banner_scan_id) REFERENCES public.cookie_banner_scans(id) ON DELETE SET NULL;


--
-- Name: cmp_domain_surfaces cmp_domain_surfaces_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cmp_domain_surfaces
    ADD CONSTRAINT cmp_domain_surfaces_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: cmp_domains cmp_domains_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cmp_domains
    ADD CONSTRAINT cmp_domains_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: collaboration_actors collaboration_actors_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_actors
    ADD CONSTRAINT collaboration_actors_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: collaboration_actors collaboration_actors_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_actors
    ADD CONSTRAINT collaboration_actors_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id);


--
-- Name: collaboration_actors collaboration_actors_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_actors
    ADD CONSTRAINT collaboration_actors_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: collaboration_actors collaboration_actors_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_actors
    ADD CONSTRAINT collaboration_actors_profile_id_fkey FOREIGN KEY (profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: collaboration_changes collaboration_changes_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_changes
    ADD CONSTRAINT collaboration_changes_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: collaboration_changes collaboration_changes_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_changes
    ADD CONSTRAINT collaboration_changes_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: collaboration_documents collaboration_documents_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_documents
    ADD CONSTRAINT collaboration_documents_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: collaboration_documents collaboration_documents_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaboration_documents
    ADD CONSTRAINT collaboration_documents_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: common_gvl_state common_gvl_state_latest_vendor_list_version_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_gvl_state
    ADD CONSTRAINT common_gvl_state_latest_vendor_list_version_fkey FOREIGN KEY (latest_vendor_list_version) REFERENCES public.common_gvl_snapshots(vendor_list_version);


--
-- Name: common_gvl_vendors common_gvl_vendors_vendor_list_version_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_gvl_vendors
    ADD CONSTRAINT common_gvl_vendors_vendor_list_version_fkey FOREIGN KEY (vendor_list_version) REFERENCES public.common_gvl_snapshots(vendor_list_version);


--
-- Name: common_third_parties common_third_parties_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_third_parties
    ADD CONSTRAINT common_third_parties_logo_file_id_fkey FOREIGN KEY (logo_file_id) REFERENCES public.files(id);


--
-- Name: common_third_party_domains common_third_party_domains_common_third_party_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_third_party_domains
    ADD CONSTRAINT common_third_party_domains_common_third_party_id_fkey FOREIGN KEY (common_third_party_id) REFERENCES public.common_third_parties(id) ON DELETE CASCADE;


--
-- Name: common_tracker_patterns common_tracker_patterns_common_third_party_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.common_tracker_patterns
    ADD CONSTRAINT common_tracker_patterns_common_third_party_id_fkey FOREIGN KEY (common_third_party_id) REFERENCES public.common_third_parties(id) ON DELETE SET NULL;


--
-- Name: compliance_custom_links compliance_external_urls_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_custom_links
    ADD CONSTRAINT compliance_external_urls_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_custom_links compliance_external_urls_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_custom_links
    ADD CONSTRAINT compliance_external_urls_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_frameworks compliance_frameworks_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_frameworks
    ADD CONSTRAINT compliance_frameworks_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_frameworks compliance_frameworks_framework_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_frameworks
    ADD CONSTRAINT compliance_frameworks_framework_id_fkey FOREIGN KEY (framework_id) REFERENCES public.frameworks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_frameworks compliance_frameworks_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_frameworks
    ADD CONSTRAINT compliance_frameworks_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_portal_commitment_groups compliance_portal_commitment_groups_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portal_commitment_groups
    ADD CONSTRAINT compliance_portal_commitment_groups_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_portal_commitment_groups compliance_portal_commitment_groups_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portal_commitment_groups
    ADD CONSTRAINT compliance_portal_commitment_groups_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_portal_commitments compliance_portal_commitments_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portal_commitments
    ADD CONSTRAINT compliance_portal_commitments_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_portal_commitments compliance_portal_commitments_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portal_commitments
    ADD CONSTRAINT compliance_portal_commitments_group_id_fkey FOREIGN KEY (group_id) REFERENCES public.compliance_portal_commitment_groups(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_portal_commitments compliance_portal_commitments_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portal_commitments
    ADD CONSTRAINT compliance_portal_commitments_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: compliance_portals compliance_portals_custom_domain_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portals
    ADD CONSTRAINT compliance_portals_custom_domain_id_fkey FOREIGN KEY (custom_domain_id) REFERENCES public.custom_domains(id) ON DELETE SET NULL;


--
-- Name: compliance_portals compliance_portals_dark_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portals
    ADD CONSTRAINT compliance_portals_dark_logo_file_id_fkey FOREIGN KEY (dark_logo_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: compliance_portals compliance_portals_default_domain_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portals
    ADD CONSTRAINT compliance_portals_default_domain_id_fkey FOREIGN KEY (default_domain_id) REFERENCES public.custom_domains(id) ON DELETE SET NULL;


--
-- Name: compliance_portals compliance_portals_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portals
    ADD CONSTRAINT compliance_portals_logo_file_id_fkey FOREIGN KEY (logo_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: compliance_portals compliance_portals_mailing_list_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portals
    ADD CONSTRAINT compliance_portals_mailing_list_id_fkey FOREIGN KEY (mailing_list_id) REFERENCES public.mailing_lists(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: compliance_portals compliance_portals_non_disclosure_agreement_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portals
    ADD CONSTRAINT compliance_portals_non_disclosure_agreement_file_id_fkey FOREIGN KEY (non_disclosure_agreement_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: compliance_portals compliance_portals_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.compliance_portals
    ADD CONSTRAINT compliance_portals_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: connector_accounts connector_accounts_connector_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.connector_accounts
    ADD CONSTRAINT connector_accounts_connector_id_fkey FOREIGN KEY (connector_id) REFERENCES public.connectors(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: connector_accounts connector_accounts_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.connector_accounts
    ADD CONSTRAINT connector_accounts_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: connector_install_state_claims connector_install_state_claims_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.connector_install_state_claims
    ADD CONSTRAINT connector_install_state_claims_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: connectors connectors_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.connectors
    ADD CONSTRAINT connectors_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: controls_audits controls_audits_audit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_audits
    ADD CONSTRAINT controls_audits_audit_id_fkey FOREIGN KEY (audit_id) REFERENCES public.audits(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: controls_audits controls_audits_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_audits
    ADD CONSTRAINT controls_audits_control_id_fkey FOREIGN KEY (control_id) REFERENCES public.controls(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: controls_documents controls_documents_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_documents
    ADD CONSTRAINT controls_documents_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id);


--
-- Name: controls controls_framework_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls
    ADD CONSTRAINT controls_framework_id_fkey FOREIGN KEY (framework_id) REFERENCES public.frameworks(id) ON DELETE CASCADE;


--
-- Name: controls_internal_controls controls_internal_controls_internal_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_internal_controls
    ADD CONSTRAINT controls_internal_controls_internal_control_id_fkey FOREIGN KEY (internal_control_id) REFERENCES public.internal_controls(id) ON DELETE CASCADE;


--
-- Name: controls_internal_controls controls_mesures_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_internal_controls
    ADD CONSTRAINT controls_mesures_control_id_fkey FOREIGN KEY (control_id) REFERENCES public.controls(id) ON DELETE CASCADE;


--
-- Name: controls_obligations controls_obligations_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_obligations
    ADD CONSTRAINT controls_obligations_control_id_fkey FOREIGN KEY (control_id) REFERENCES public.controls(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: controls_obligations controls_obligations_obligation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_obligations
    ADD CONSTRAINT controls_obligations_obligation_id_fkey FOREIGN KEY (obligation_id) REFERENCES public.obligations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: controls_documents controls_policies_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_documents
    ADD CONSTRAINT controls_policies_control_id_fkey FOREIGN KEY (control_id) REFERENCES public.controls(id) ON DELETE CASCADE;


--
-- Name: cookie_banner_daily_page_loads cookie_banner_daily_page_loads_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_daily_page_loads
    ADD CONSTRAINT cookie_banner_daily_page_loads_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: cookie_banner_discovery_stats cookie_banner_discovery_stats_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_discovery_stats
    ADD CONSTRAINT cookie_banner_discovery_stats_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: cookie_banner_gvl_vendors cookie_banner_gvl_vendors_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_gvl_vendors
    ADD CONSTRAINT cookie_banner_gvl_vendors_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: cookie_banner_gvl_vendors cookie_banner_gvl_vendors_iab_vendor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_gvl_vendors
    ADD CONSTRAINT cookie_banner_gvl_vendors_iab_vendor_id_fkey FOREIGN KEY (iab_vendor_id) REFERENCES public.common_gvl_vendors(iab_vendor_id);


--
-- Name: cookie_banner_gvl_vendors cookie_banner_gvl_vendors_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_gvl_vendors
    ADD CONSTRAINT cookie_banner_gvl_vendors_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: cookie_banner_scans cookie_banner_scans_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_scans
    ADD CONSTRAINT cookie_banner_scans_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: cookie_banner_scans cookie_banner_scans_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_scans
    ADD CONSTRAINT cookie_banner_scans_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: cookie_banner_translations cookie_banner_translations_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_translations
    ADD CONSTRAINT cookie_banner_translations_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: cookie_banner_translations cookie_banner_translations_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_translations
    ADD CONSTRAINT cookie_banner_translations_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cookie_banner_versions cookie_banner_versions_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_versions
    ADD CONSTRAINT cookie_banner_versions_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: cookie_banner_versions cookie_banner_versions_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banner_versions
    ADD CONSTRAINT cookie_banner_versions_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: cookie_banners cookie_banners_cmp_domain_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banners
    ADD CONSTRAINT cookie_banners_cmp_domain_id_fkey FOREIGN KEY (cmp_domain_id) REFERENCES public.cmp_domains(id) ON DELETE SET NULL;


--
-- Name: cookie_banners cookie_banners_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_banners
    ADD CONSTRAINT cookie_banners_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: cookie_categories cookie_categories_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_categories
    ADD CONSTRAINT cookie_categories_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: cookie_categories cookie_categories_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_categories
    ADD CONSTRAINT cookie_categories_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: cookie_consent_records cookie_consent_records_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_consent_records
    ADD CONSTRAINT cookie_consent_records_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: cookie_consent_records cookie_consent_records_cookie_banner_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_consent_records
    ADD CONSTRAINT cookie_consent_records_cookie_banner_version_id_fkey FOREIGN KEY (cookie_banner_version_id) REFERENCES public.cookie_banner_versions(id);


--
-- Name: cookie_consent_records cookie_consent_records_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookie_consent_records
    ADD CONSTRAINT cookie_consent_records_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: cp_accesses cp_accesses_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_accesses
    ADD CONSTRAINT cp_accesses_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON DELETE CASCADE;


--
-- Name: cp_accesses cp_accesses_electronic_signature_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_accesses
    ADD CONSTRAINT cp_accesses_electronic_signature_id_fkey FOREIGN KEY (electronic_signature_id) REFERENCES public.electronic_signatures(id);


--
-- Name: cp_accesses cp_accesses_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_accesses
    ADD CONSTRAINT cp_accesses_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id);


--
-- Name: cp_audits cp_audits_audit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_audits
    ADD CONSTRAINT cp_audits_audit_id_fkey FOREIGN KEY (audit_id) REFERENCES public.audits(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_audits cp_audits_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_audits
    ADD CONSTRAINT cp_audits_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_audits cp_audits_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_audits
    ADD CONSTRAINT cp_audits_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_document_accesses cp_document_accesses_compliance_portal_access_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_compliance_portal_access_id_fkey FOREIGN KEY (compliance_portal_access_id) REFERENCES public.cp_accesses(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_document_accesses cp_document_accesses_compliance_portal_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_compliance_portal_file_id_fkey FOREIGN KEY (compliance_portal_file_id) REFERENCES public.cp_files(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_document_accesses cp_document_accesses_cp_audit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_cp_audit_id_fkey FOREIGN KEY (compliance_portal_audit_id) REFERENCES public.cp_audits(id) ON DELETE CASCADE;


--
-- Name: cp_document_accesses cp_document_accesses_cp_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_document_accesses
    ADD CONSTRAINT cp_document_accesses_cp_document_id_fkey FOREIGN KEY (compliance_portal_document_id) REFERENCES public.cp_documents(id) ON DELETE CASCADE;


--
-- Name: cp_documents cp_documents_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_documents
    ADD CONSTRAINT cp_documents_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_documents cp_documents_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_documents
    ADD CONSTRAINT cp_documents_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_documents cp_documents_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_documents
    ADD CONSTRAINT cp_documents_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_files cp_files_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_files
    ADD CONSTRAINT cp_files_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_files cp_files_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_files
    ADD CONSTRAINT cp_files_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: cp_files cp_files_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_files
    ADD CONSTRAINT cp_files_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: cp_references cp_references_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_references
    ADD CONSTRAINT cp_references_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_references cp_references_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_references
    ADD CONSTRAINT cp_references_logo_file_id_fkey FOREIGN KEY (logo_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: cp_third_parties cp_third_parties_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_third_parties
    ADD CONSTRAINT cp_third_parties_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_third_parties cp_third_parties_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_third_parties
    ADD CONSTRAINT cp_third_parties_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: cp_third_parties cp_third_parties_third_party_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cp_third_parties
    ADD CONSTRAINT cp_third_parties_third_party_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: custom_domains custom_domains_certificate_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.custom_domains
    ADD CONSTRAINT custom_domains_certificate_id_fkey FOREIGN KEY (certificate_id) REFERENCES public.certificates(id) ON DELETE SET NULL;


--
-- Name: data data_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data
    ADD CONSTRAINT data_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: data data_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data
    ADD CONSTRAINT data_owner_profile_id_fkey FOREIGN KEY (owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: data_third_parties data_vendors_datum_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data_third_parties
    ADD CONSTRAINT data_vendors_datum_id_fkey FOREIGN KEY (datum_id) REFERENCES public.data(id) ON DELETE CASCADE;


--
-- Name: data_third_parties data_vendors_vendor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.data_third_parties
    ADD CONSTRAINT data_vendors_vendor_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON DELETE CASCADE;


--
-- Name: detected_trackers detected_trackers_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.detected_trackers
    ADD CONSTRAINT detected_trackers_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: detected_trackers detected_trackers_cookie_banner_scan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.detected_trackers
    ADD CONSTRAINT detected_trackers_cookie_banner_scan_id_fkey FOREIGN KEY (cookie_banner_scan_id) REFERENCES public.cookie_banner_scans(id) ON DELETE SET NULL;


--
-- Name: detected_trackers detected_trackers_tracker_pattern_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.detected_trackers
    ADD CONSTRAINT detected_trackers_tracker_pattern_id_fkey FOREIGN KEY (tracker_pattern_id) REFERENCES public.tracker_patterns(id) ON DELETE CASCADE;


--
-- Name: device_enrollment_tokens device_enrollment_tokens_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_enrollment_tokens
    ADD CONSTRAINT device_enrollment_tokens_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: device_postures device_postures_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_postures
    ADD CONSTRAINT device_postures_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: device_postures device_postures_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_postures
    ADD CONSTRAINT device_postures_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: devices devices_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.devices
    ADD CONSTRAINT devices_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: devices devices_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.devices
    ADD CONSTRAINT devices_owner_profile_id_fkey FOREIGN KEY (owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: document_comment_threads document_comment_threads_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comment_threads
    ADD CONSTRAINT document_comment_threads_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document_comment_threads document_comment_threads_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comment_threads
    ADD CONSTRAINT document_comment_threads_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document_comment_threads document_comment_threads_resolved_by_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comment_threads
    ADD CONSTRAINT document_comment_threads_resolved_by_profile_id_fkey FOREIGN KEY (resolved_by_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: document_comment_threads document_comment_threads_suggestion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comment_threads
    ADD CONSTRAINT document_comment_threads_suggestion_id_fkey FOREIGN KEY (suggestion_id) REFERENCES public.document_suggestions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document_comments document_comments_author_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comments
    ADD CONSTRAINT document_comments_author_profile_id_fkey FOREIGN KEY (author_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: document_comments document_comments_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comments
    ADD CONSTRAINT document_comments_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document_comments document_comments_thread_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_comments
    ADD CONSTRAINT document_comments_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES public.document_comment_threads(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document_default_approvers document_default_approvers_approver_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_default_approvers
    ADD CONSTRAINT document_default_approvers_approver_profile_id_fkey FOREIGN KEY (approver_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document_default_approvers document_default_approvers_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_default_approvers
    ADD CONSTRAINT document_default_approvers_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document_suggestions document_suggestions_author_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_suggestions
    ADD CONSTRAINT document_suggestions_author_profile_id_fkey FOREIGN KEY (author_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: document_suggestions document_suggestions_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_suggestions
    ADD CONSTRAINT document_suggestions_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document_suggestions document_suggestions_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_suggestions
    ADD CONSTRAINT document_suggestions_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: document_suggestions document_suggestions_resolved_by_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_suggestions
    ADD CONSTRAINT document_suggestions_resolved_by_profile_id_fkey FOREIGN KEY (resolved_by_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: document_version_approval_decisions document_version_approval_decision_electronic_signature_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_approval_decisions
    ADD CONSTRAINT document_version_approval_decision_electronic_signature_id_fkey FOREIGN KEY (electronic_signature_id) REFERENCES public.electronic_signatures(id) ON DELETE RESTRICT;


--
-- Name: document_version_approval_decisions document_version_approval_decisions_approver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_approval_decisions
    ADD CONSTRAINT document_version_approval_decisions_approver_id_fkey FOREIGN KEY (approver_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: document_version_approval_decisions document_version_approval_decisions_quorum_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_approval_decisions
    ADD CONSTRAINT document_version_approval_decisions_quorum_id_fkey FOREIGN KEY (quorum_id) REFERENCES public.document_version_approval_quorums(id) ON DELETE CASCADE;


--
-- Name: document_version_approval_quorums document_version_approval_quorums_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_approval_quorums
    ADD CONSTRAINT document_version_approval_quorums_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.files(id);


--
-- Name: document_version_approval_quorums document_version_approval_quorums_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_approval_quorums
    ADD CONSTRAINT document_version_approval_quorums_version_id_fkey FOREIGN KEY (version_id) REFERENCES public.document_versions(id) ON DELETE CASCADE;


--
-- Name: document_version_signatures document_version_signatures_document_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_signatures
    ADD CONSTRAINT document_version_signatures_document_version_id_fkey FOREIGN KEY (document_version_id) REFERENCES public.document_versions(id) ON DELETE CASCADE;


--
-- Name: document_version_signatures document_version_signatures_electronic_signature_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_signatures
    ADD CONSTRAINT document_version_signatures_electronic_signature_id_fkey FOREIGN KEY (electronic_signature_id) REFERENCES public.electronic_signatures(id) ON DELETE RESTRICT;


--
-- Name: document_version_signatures document_version_signatures_signed_by_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_version_signatures
    ADD CONSTRAINT document_version_signatures_signed_by_profile_id_fkey FOREIGN KEY (signed_by_profile_id) REFERENCES public.iam_membership_profiles(id);


--
-- Name: document_versions document_versions_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_versions
    ADD CONSTRAINT document_versions_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: document_versions document_versions_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.document_versions
    ADD CONSTRAINT document_versions_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.files(id);


--
-- Name: documents documents_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: electronic_signature_events electronic_signature_events_electronic_signature_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.electronic_signature_events
    ADD CONSTRAINT electronic_signature_events_electronic_signature_id_fkey FOREIGN KEY (electronic_signature_id) REFERENCES public.electronic_signatures(id) ON DELETE CASCADE;


--
-- Name: electronic_signatures electronic_signatures_certificate_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.electronic_signatures
    ADD CONSTRAINT electronic_signatures_certificate_file_id_fkey FOREIGN KEY (certificate_file_id) REFERENCES public.files(id);


--
-- Name: electronic_signatures electronic_signatures_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.electronic_signatures
    ADD CONSTRAINT electronic_signatures_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.files(id);


--
-- Name: electronic_signatures electronic_signatures_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.electronic_signatures
    ADD CONSTRAINT electronic_signatures_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: email_attachments email_attachments_email_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_attachments
    ADD CONSTRAINT email_attachments_email_id_fkey FOREIGN KEY (email_id) REFERENCES public.emails(id) ON DELETE CASCADE;


--
-- Name: email_attachments email_attachments_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_attachments
    ADD CONSTRAINT email_attachments_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.files(id);


--
-- Name: emails emails_mailing_list_update_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.emails
    ADD CONSTRAINT emails_mailing_list_update_id_fkey FOREIGN KEY (mailing_list_update_id) REFERENCES public.mailing_list_updates(id) ON DELETE SET NULL;


--
-- Name: employee_portal_documents employee_portal_documents_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_portal_documents
    ADD CONSTRAINT employee_portal_documents_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: employee_portal_documents employee_portal_documents_employee_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_portal_documents
    ADD CONSTRAINT employee_portal_documents_employee_portal_id_fkey FOREIGN KEY (employee_portal_id) REFERENCES public.employee_portals(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: employee_portal_documents employee_portal_documents_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_portal_documents
    ADD CONSTRAINT employee_portal_documents_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: employee_portals employee_portals_dark_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_portals
    ADD CONSTRAINT employee_portals_dark_logo_file_id_fkey FOREIGN KEY (dark_logo_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: employee_portals employee_portals_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_portals
    ADD CONSTRAINT employee_portals_logo_file_id_fkey FOREIGN KEY (logo_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: employee_portals employee_portals_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.employee_portals
    ADD CONSTRAINT employee_portals_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: evidences evidences_internal_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidences
    ADD CONSTRAINT evidences_internal_control_id_fkey FOREIGN KEY (internal_control_id) REFERENCES public.internal_controls(id) ON DELETE CASCADE;


--
-- Name: evidences evidences_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidences
    ADD CONSTRAINT evidences_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: export_jobs export_jobs_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.export_jobs
    ADD CONSTRAINT export_jobs_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: findings_audits findings_audits_audit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings_audits
    ADD CONSTRAINT findings_audits_audit_id_fkey FOREIGN KEY (audit_id) REFERENCES public.audits(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: findings_audits findings_audits_finding_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings_audits
    ADD CONSTRAINT findings_audits_finding_id_fkey FOREIGN KEY (finding_id) REFERENCES public.findings(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: findings_tasks findings_tasks_finding_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings_tasks
    ADD CONSTRAINT findings_tasks_finding_id_fkey FOREIGN KEY (finding_id) REFERENCES public.findings(id) ON DELETE CASCADE;


--
-- Name: findings_tasks findings_tasks_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings_tasks
    ADD CONSTRAINT findings_tasks_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: iam_saml_assertions fk_auth_saml_assertions_organization; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_saml_assertions
    ADD CONSTRAINT fk_auth_saml_assertions_organization FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: iam_saml_configurations fk_auth_saml_configurations_organization; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_saml_configurations
    ADD CONSTRAINT fk_auth_saml_configurations_organization FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: iam_saml_requests fk_auth_saml_requests_organization; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_saml_requests
    ADD CONSTRAINT fk_auth_saml_requests_organization FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: controls_internal_controls fk_controls_mitigations_control_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.controls_internal_controls
    ADD CONSTRAINT fk_controls_mitigations_control_id FOREIGN KEY (control_id) REFERENCES public.controls(id) ON DELETE CASCADE;


--
-- Name: evidences fk_evidence_file; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.evidences
    ADD CONSTRAINT fk_evidence_file FOREIGN KEY (evidence_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: findings fk_findings_organization; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings
    ADD CONSTRAINT fk_findings_organization FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: findings fk_findings_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings
    ADD CONSTRAINT fk_findings_owner FOREIGN KEY (owner_id) REFERENCES public.iam_membership_profiles(id);


--
-- Name: findings fk_findings_risk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.findings
    ADD CONSTRAINT fk_findings_risk FOREIGN KEY (risk_id) REFERENCES public.risks(id) ON DELETE RESTRICT;


--
-- Name: internal_controls fk_mitigations_organization_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls
    ADD CONSTRAINT fk_mitigations_organization_id FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: slack_messages fk_slack_messages_initial_slack_message_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_messages
    ADD CONSTRAINT fk_slack_messages_initial_slack_message_id FOREIGN KEY (initial_slack_message_id) REFERENCES public.slack_messages(id) ON DELETE CASCADE;


--
-- Name: slack_messages fk_slack_messages_organization_id; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_messages
    ADD CONSTRAINT fk_slack_messages_organization_id FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: frameworks frameworks_dark_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.frameworks
    ADD CONSTRAINT frameworks_dark_logo_file_id_fkey FOREIGN KEY (dark_logo_file_id) REFERENCES public.files(id) ON DELETE SET NULL;


--
-- Name: frameworks frameworks_light_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.frameworks
    ADD CONSTRAINT frameworks_light_logo_file_id_fkey FOREIGN KEY (light_logo_file_id) REFERENCES public.files(id) ON DELETE SET NULL;


--
-- Name: frameworks frameworks_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.frameworks
    ADD CONSTRAINT frameworks_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: generated_documents generated_documents_ai_systems_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_ai_systems_document_id_fkey FOREIGN KEY (ai_systems_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_asset_list_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_asset_list_document_id_fkey FOREIGN KEY (asset_list_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_business_functions_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_business_functions_document_id_fkey FOREIGN KEY (business_functions_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_data_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_data_document_id_fkey FOREIGN KEY (data_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_data_protection_impact_assessments_doc_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_data_protection_impact_assessments_doc_fkey FOREIGN KEY (data_protection_impact_assessments_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_findings_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_findings_document_id_fkey FOREIGN KEY (findings_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_obligations_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_obligations_document_id_fkey FOREIGN KEY (obligations_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: generated_documents generated_documents_processing_activities_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_processing_activities_document_id_fkey FOREIGN KEY (processing_activities_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_risks_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_risks_document_id_fkey FOREIGN KEY (risks_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_transfer_impact_assessments_document_i_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_transfer_impact_assessments_document_i_fkey FOREIGN KEY (transfer_impact_assessments_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: generated_documents generated_documents_vendors_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generated_documents
    ADD CONSTRAINT generated_documents_vendors_document_id_fkey FOREIGN KEY (third_parties_document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: iam_invitations iam_invitations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_invitations
    ADD CONSTRAINT iam_invitations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.iam_membership_profiles(id) ON DELETE CASCADE;


--
-- Name: iam_membership_profiles iam_membership_profiles_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_membership_profiles
    ADD CONSTRAINT iam_membership_profiles_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id);


--
-- Name: iam_membership_profiles iam_membership_profiles_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_membership_profiles
    ADD CONSTRAINT iam_membership_profiles_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: iam_oauth2_access_tokens iam_oauth2_access_tokens_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_access_tokens
    ADD CONSTRAINT iam_oauth2_access_tokens_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.iam_oauth2_clients(id);


--
-- Name: iam_oauth2_authorization_codes iam_oauth2_authorization_codes_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_authorization_codes
    ADD CONSTRAINT iam_oauth2_authorization_codes_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.iam_oauth2_clients(id);


--
-- Name: iam_oauth2_consents iam_oauth2_consents_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_consents
    ADD CONSTRAINT iam_oauth2_consents_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.iam_oauth2_clients(id);


--
-- Name: iam_oauth2_consents iam_oauth2_consents_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_consents
    ADD CONSTRAINT iam_oauth2_consents_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.iam_sessions(id);


--
-- Name: iam_oauth2_device_codes iam_oauth2_device_codes_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_device_codes
    ADD CONSTRAINT iam_oauth2_device_codes_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.iam_oauth2_clients(id);


--
-- Name: iam_oauth2_refresh_tokens iam_oauth2_refresh_tokens_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_oauth2_refresh_tokens
    ADD CONSTRAINT iam_oauth2_refresh_tokens_client_id_fkey FOREIGN KEY (client_id) REFERENCES public.iam_oauth2_clients(id);


--
-- Name: iam_scim_bridges iam_scim_bridges_connector_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_bridges
    ADD CONSTRAINT iam_scim_bridges_connector_id_fkey FOREIGN KEY (connector_id) REFERENCES public.connectors(id);


--
-- Name: iam_scim_bridges iam_scim_bridges_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_bridges
    ADD CONSTRAINT iam_scim_bridges_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: iam_scim_bridges iam_scim_bridges_scim_configuration_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_bridges
    ADD CONSTRAINT iam_scim_bridges_scim_configuration_id_fkey FOREIGN KEY (scim_configuration_id) REFERENCES public.iam_scim_configurations(id) ON DELETE CASCADE;


--
-- Name: iam_scim_configurations iam_scim_configurations_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_configurations
    ADD CONSTRAINT iam_scim_configurations_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: iam_scim_events iam_scim_events_membership_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_events
    ADD CONSTRAINT iam_scim_events_membership_id_fkey FOREIGN KEY (membership_id) REFERENCES public.iam_memberships(id) ON DELETE SET NULL;


--
-- Name: iam_scim_events iam_scim_events_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_events
    ADD CONSTRAINT iam_scim_events_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: iam_scim_events iam_scim_events_scim_configuration_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_scim_events
    ADD CONSTRAINT iam_scim_events_scim_configuration_id_fkey FOREIGN KEY (scim_configuration_id) REFERENCES public.iam_scim_configurations(id) ON DELETE CASCADE;


--
-- Name: identities identities_avatar_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_avatar_file_id_fkey FOREIGN KEY (avatar_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: internal_control_events internal_control_events_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_control_events
    ADD CONSTRAINT internal_control_events_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: internal_controls_documents internal_controls_documents_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_documents
    ADD CONSTRAINT internal_controls_documents_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id);


--
-- Name: internal_controls_documents internal_controls_documents_internal_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_documents
    ADD CONSTRAINT internal_controls_documents_internal_control_id_fkey FOREIGN KEY (internal_control_id) REFERENCES public.internal_controls(id) ON DELETE CASCADE;


--
-- Name: internal_controls internal_controls_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls
    ADD CONSTRAINT internal_controls_owner_profile_id_fkey FOREIGN KEY (owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: internal_controls internal_controls_reviewer_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls
    ADD CONSTRAINT internal_controls_reviewer_profile_id_fkey FOREIGN KEY (reviewer_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: internal_controls_tasks internal_controls_tasks_internal_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_tasks
    ADD CONSTRAINT internal_controls_tasks_internal_control_id_fkey FOREIGN KEY (internal_control_id) REFERENCES public.internal_controls(id) ON DELETE CASCADE;


--
-- Name: internal_controls_tasks internal_controls_tasks_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_tasks
    ADD CONSTRAINT internal_controls_tasks_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: internal_controls_third_parties internal_controls_third_parties_internal_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_third_parties
    ADD CONSTRAINT internal_controls_third_parties_internal_control_id_fkey FOREIGN KEY (internal_control_id) REFERENCES public.internal_controls(id) ON DELETE CASCADE;


--
-- Name: internal_controls_third_parties internal_controls_third_parties_third_party_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.internal_controls_third_parties
    ADD CONSTRAINT internal_controls_third_parties_third_party_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON DELETE CASCADE;


--
-- Name: mailing_list_subscribers mailing_list_subscribers_mailing_list_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mailing_list_subscribers
    ADD CONSTRAINT mailing_list_subscribers_mailing_list_id_fkey FOREIGN KEY (mailing_list_id) REFERENCES public.mailing_lists(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: mailing_list_subscribers mailing_list_subscribers_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mailing_list_subscribers
    ADD CONSTRAINT mailing_list_subscribers_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: mailing_list_updates mailing_list_updates_mailing_list_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mailing_list_updates
    ADD CONSTRAINT mailing_list_updates_mailing_list_id_fkey FOREIGN KEY (mailing_list_id) REFERENCES public.mailing_lists(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: mailing_list_updates mailing_list_updates_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mailing_list_updates
    ADD CONSTRAINT mailing_list_updates_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: mailing_lists mailing_lists_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mailing_lists
    ADD CONSTRAINT mailing_lists_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: notification_preferences notification_preferences_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_preferences
    ADD CONSTRAINT notification_preferences_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: notification_preferences notification_preferences_recipient_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_preferences
    ADD CONSTRAINT notification_preferences_recipient_profile_id_fkey FOREIGN KEY (recipient_profile_id) REFERENCES public.iam_membership_profiles(id) ON DELETE CASCADE;


--
-- Name: notification_weekly_sends notification_weekly_sends_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_weekly_sends
    ADD CONSTRAINT notification_weekly_sends_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: notification_weekly_sends notification_weekly_sends_recipient_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_weekly_sends
    ADD CONSTRAINT notification_weekly_sends_recipient_profile_id_fkey FOREIGN KEY (recipient_profile_id) REFERENCES public.iam_membership_profiles(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_recipient_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_recipient_profile_id_fkey FOREIGN KEY (recipient_profile_id) REFERENCES public.iam_membership_profiles(id) ON DELETE CASCADE;


--
-- Name: obligations obligations_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.obligations
    ADD CONSTRAINT obligations_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: obligations obligations_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.obligations
    ADD CONSTRAINT obligations_owner_profile_id_fkey FOREIGN KEY (owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: operation_receipts operation_receipts_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.operation_receipts
    ADD CONSTRAINT operation_receipts_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: organization_contexts organization_contexts_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_contexts
    ADD CONSTRAINT organization_contexts_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: organizations organizations_horizontal_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_horizontal_logo_file_id_fkey FOREIGN KEY (horizontal_logo_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: organizations organizations_logo_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_logo_file_id_fkey FOREIGN KEY (logo_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: probot_identity_binding_challenges probot_identity_binding_challenge_confirmed_by_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.probot_identity_binding_challenges
    ADD CONSTRAINT probot_identity_binding_challenge_confirmed_by_identity_id_fkey FOREIGN KEY (confirmed_by_identity_id) REFERENCES public.identities(id) ON DELETE CASCADE;


--
-- Name: probot_identity_bindings probot_identity_bindings_identity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.probot_identity_bindings
    ADD CONSTRAINT probot_identity_bindings_identity_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id) ON DELETE CASCADE;


--
-- Name: processing_activities processing_activities_dpo_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activities
    ADD CONSTRAINT processing_activities_dpo_profile_id_fkey FOREIGN KEY (dpo_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: processing_activities processing_activities_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activities
    ADD CONSTRAINT processing_activities_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: processing_activity_data_protection_impact_assessments processing_activity_dpia_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activity_data_protection_impact_assessments
    ADD CONSTRAINT processing_activity_dpia_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: processing_activity_data_protection_impact_assessments processing_activity_dpia_processing_activity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activity_data_protection_impact_assessments
    ADD CONSTRAINT processing_activity_dpia_processing_activity_id_fkey FOREIGN KEY (processing_activity_id) REFERENCES public.processing_activities(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: processing_activity_transfer_impact_assessments processing_activity_tia_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activity_transfer_impact_assessments
    ADD CONSTRAINT processing_activity_tia_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: processing_activity_transfer_impact_assessments processing_activity_tia_processing_activity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activity_transfer_impact_assessments
    ADD CONSTRAINT processing_activity_tia_processing_activity_id_fkey FOREIGN KEY (processing_activity_id) REFERENCES public.processing_activities(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: processing_activity_third_parties processing_activity_vendors_processing_activity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activity_third_parties
    ADD CONSTRAINT processing_activity_vendors_processing_activity_id_fkey FOREIGN KEY (processing_activity_id) REFERENCES public.processing_activities(id) ON DELETE CASCADE;


--
-- Name: processing_activity_third_parties processing_activity_vendors_vendor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.processing_activity_third_parties
    ADD CONSTRAINT processing_activity_vendors_vendor_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON DELETE CASCADE;


--
-- Name: rights_requests rights_requests_compliance_portal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rights_requests
    ADD CONSTRAINT rights_requests_compliance_portal_id_fkey FOREIGN KEY (compliance_portal_id) REFERENCES public.compliance_portals(id) ON DELETE SET NULL;


--
-- Name: rights_requests rights_requests_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rights_requests
    ADD CONSTRAINT rights_requests_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: risk_analyses risk_analyses_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analyses
    ADD CONSTRAINT risk_analyses_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: risk_analyses risk_analyses_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analyses
    ADD CONSTRAINT risk_analyses_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_boundaries risk_analysis_boundaries_parent_boundary_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_boundaries
    ADD CONSTRAINT risk_analysis_boundaries_parent_boundary_id_fkey FOREIGN KEY (parent_boundary_id) REFERENCES public.risk_analysis_boundaries(id) ON DELETE SET NULL;


--
-- Name: risk_analysis_boundaries risk_analysis_boundaries_risk_analysis_diagram_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_boundaries
    ADD CONSTRAINT risk_analysis_boundaries_risk_analysis_diagram_id_fkey FOREIGN KEY (risk_analysis_diagram_id) REFERENCES public.risk_analysis_diagrams(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_diagrams risk_analysis_diagrams_risk_analysis_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_diagrams
    ADD CONSTRAINT risk_analysis_diagrams_risk_analysis_id_fkey FOREIGN KEY (risk_analysis_id) REFERENCES public.risk_analyses(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_nodes risk_analysis_nodes_boundary_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_nodes
    ADD CONSTRAINT risk_analysis_nodes_boundary_id_fkey FOREIGN KEY (boundary_id) REFERENCES public.risk_analysis_boundaries(id) ON DELETE SET NULL;


--
-- Name: risk_analysis_nodes risk_analysis_nodes_risk_analysis_diagram_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_nodes
    ADD CONSTRAINT risk_analysis_nodes_risk_analysis_diagram_id_fkey FOREIGN KEY (risk_analysis_diagram_id) REFERENCES public.risk_analysis_diagrams(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_processes risk_analysis_processes_risk_analysis_diagram_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_processes
    ADD CONSTRAINT risk_analysis_processes_risk_analysis_diagram_id_fkey FOREIGN KEY (risk_analysis_diagram_id) REFERENCES public.risk_analysis_diagrams(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_processes risk_analysis_processes_source_node_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_processes
    ADD CONSTRAINT risk_analysis_processes_source_node_id_fkey FOREIGN KEY (source_node_id) REFERENCES public.risk_analysis_nodes(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_processes risk_analysis_processes_target_node_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_processes
    ADD CONSTRAINT risk_analysis_processes_target_node_id_fkey FOREIGN KEY (target_node_id) REFERENCES public.risk_analysis_nodes(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_scenario_risks risk_analysis_scenario_risks_risk_analysis_scenario_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_scenario_risks
    ADD CONSTRAINT risk_analysis_scenario_risks_risk_analysis_scenario_id_fkey FOREIGN KEY (risk_analysis_scenario_id) REFERENCES public.risk_analysis_scenarios(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_scenario_risks risk_analysis_scenario_risks_risk_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_scenario_risks
    ADD CONSTRAINT risk_analysis_scenario_risks_risk_id_fkey FOREIGN KEY (risk_id) REFERENCES public.risks(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: risk_analysis_scenario_threats risk_analysis_scenario_threats_risk_analysis_scenario_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_scenario_threats
    ADD CONSTRAINT risk_analysis_scenario_threats_risk_analysis_scenario_id_fkey FOREIGN KEY (risk_analysis_scenario_id) REFERENCES public.risk_analysis_scenarios(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_scenario_threats risk_analysis_scenario_threats_risk_analysis_threat_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_scenario_threats
    ADD CONSTRAINT risk_analysis_scenario_threats_risk_analysis_threat_id_fkey FOREIGN KEY (risk_analysis_threat_id) REFERENCES public.risk_analysis_threats(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_scenarios risk_analysis_scenarios_risk_analysis_diagram_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_scenarios
    ADD CONSTRAINT risk_analysis_scenarios_risk_analysis_diagram_id_fkey FOREIGN KEY (risk_analysis_diagram_id) REFERENCES public.risk_analysis_diagrams(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_threats risk_analysis_threats_process_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_threats
    ADD CONSTRAINT risk_analysis_threats_process_id_fkey FOREIGN KEY (process_id) REFERENCES public.risk_analysis_processes(id) ON DELETE CASCADE;


--
-- Name: risk_analysis_threats risk_analysis_threats_risk_analysis_diagram_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risk_analysis_threats
    ADD CONSTRAINT risk_analysis_threats_risk_analysis_diagram_id_fkey FOREIGN KEY (risk_analysis_diagram_id) REFERENCES public.risk_analysis_diagrams(id) ON DELETE CASCADE;


--
-- Name: third_party_risk_assessments risk_assessments_vendor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_risk_assessments
    ADD CONSTRAINT risk_assessments_vendor_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON DELETE CASCADE;


--
-- Name: risks_documents risks_documents_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks_documents
    ADD CONSTRAINT risks_documents_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id);


--
-- Name: risks_internal_controls risks_internal_controls_internal_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks_internal_controls
    ADD CONSTRAINT risks_internal_controls_internal_control_id_fkey FOREIGN KEY (internal_control_id) REFERENCES public.internal_controls(id) ON DELETE CASCADE;


--
-- Name: risks_internal_controls risks_mitigations_risk_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks_internal_controls
    ADD CONSTRAINT risks_mitigations_risk_id_fkey FOREIGN KEY (risk_id) REFERENCES public.risks(id);


--
-- Name: risks_obligations risks_obligations_obligation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks_obligations
    ADD CONSTRAINT risks_obligations_obligation_id_fkey FOREIGN KEY (obligation_id) REFERENCES public.obligations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: risks_obligations risks_obligations_risk_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks_obligations
    ADD CONSTRAINT risks_obligations_risk_id_fkey FOREIGN KEY (risk_id) REFERENCES public.risks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: risks risks_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks
    ADD CONSTRAINT risks_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: risks risks_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks
    ADD CONSTRAINT risks_owner_profile_id_fkey FOREIGN KEY (owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: risks_documents risks_policies_risk_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.risks_documents
    ADD CONSTRAINT risks_policies_risk_id_fkey FOREIGN KEY (risk_id) REFERENCES public.risks(id);


--
-- Name: search_resources search_resources_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_resources
    ADD CONSTRAINT search_resources_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: iam_sessions sessions_membership_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_sessions
    ADD CONSTRAINT sessions_membership_id_fkey FOREIGN KEY (membership_id) REFERENCES public.iam_memberships(id) ON DELETE CASCADE;


--
-- Name: iam_sessions sessions_parent_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_sessions
    ADD CONSTRAINT sessions_parent_session_id_fkey FOREIGN KEY (parent_session_id) REFERENCES public.iam_sessions(id);


--
-- Name: slack_delivery_operations slack_delivery_operations_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slack_delivery_operations
    ADD CONSTRAINT slack_delivery_operations_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: slackbot_install_state_claims slackbot_install_state_claims_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_install_state_claims
    ADD CONSTRAINT slackbot_install_state_claims_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: slackbot_installations slackbot_installations_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_installations
    ADD CONSTRAINT slackbot_installations_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: slackbot_interactive_claims slackbot_interactive_claims_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_interactive_claims
    ADD CONSTRAINT slackbot_interactive_claims_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: slackbot_interactive_commands slackbot_interactive_commands_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_interactive_commands
    ADD CONSTRAINT slackbot_interactive_commands_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: slackbot_messages slackbot_messages_initial_slackbot_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_messages
    ADD CONSTRAINT slackbot_messages_initial_slackbot_message_id_fkey FOREIGN KEY (initial_slackbot_message_id) REFERENCES public.slackbot_messages(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: slackbot_messages slackbot_messages_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.slackbot_messages
    ADD CONSTRAINT slackbot_messages_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: statements_of_applicability statements_of_applicability_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.statements_of_applicability
    ADD CONSTRAINT statements_of_applicability_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE SET NULL;


--
-- Name: applicability_statements states_of_applicability_controls_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applicability_statements
    ADD CONSTRAINT states_of_applicability_controls_control_id_fkey FOREIGN KEY (control_id) REFERENCES public.controls(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: applicability_statements states_of_applicability_controls_state_of_applicability_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.applicability_statements
    ADD CONSTRAINT states_of_applicability_controls_state_of_applicability_id_fkey FOREIGN KEY (statement_of_applicability_id) REFERENCES public.statements_of_applicability(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: statements_of_applicability states_of_applicability_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.statements_of_applicability
    ADD CONSTRAINT states_of_applicability_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tag_assignments tag_assignments_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tag_assignments
    ADD CONSTRAINT tag_assignments_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tag_assignments tag_assignments_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tag_assignments
    ADD CONSTRAINT tag_assignments_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tags tags_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: task_activities task_activities_actor_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_activities
    ADD CONSTRAINT task_activities_actor_profile_id_fkey FOREIGN KEY (actor_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: task_activities task_activities_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_activities
    ADD CONSTRAINT task_activities_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: task_activities task_activities_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_activities
    ADD CONSTRAINT task_activities_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: task_comment_external_links task_comment_external_links_connector_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comment_external_links
    ADD CONSTRAINT task_comment_external_links_connector_id_fkey FOREIGN KEY (connector_id) REFERENCES public.connectors(id) ON DELETE CASCADE;


--
-- Name: task_comment_external_links task_comment_external_links_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comment_external_links
    ADD CONSTRAINT task_comment_external_links_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: task_comment_external_links task_comment_external_links_task_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comment_external_links
    ADD CONSTRAINT task_comment_external_links_task_comment_id_fkey FOREIGN KEY (task_comment_id) REFERENCES public.task_comments(id) ON DELETE CASCADE;


--
-- Name: task_comment_external_links task_comment_external_links_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comment_external_links
    ADD CONSTRAINT task_comment_external_links_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_comments task_comments_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: task_comments task_comments_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_owner_profile_id_fkey FOREIGN KEY (owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: task_comments task_comments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: task_external_links task_external_links_connector_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_external_links
    ADD CONSTRAINT task_external_links_connector_id_fkey FOREIGN KEY (connector_id) REFERENCES public.connectors(id) ON DELETE CASCADE;


--
-- Name: task_external_links task_external_links_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_external_links
    ADD CONSTRAINT task_external_links_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: task_external_links task_external_links_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_external_links
    ADD CONSTRAINT task_external_links_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_sync_jobs task_sync_jobs_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.task_sync_jobs
    ADD CONSTRAINT task_sync_jobs_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_assigned_to_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_assigned_to_profile_id_fkey FOREIGN KEY (assigned_to_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: tasks tasks_internal_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_internal_control_id_fkey FOREIGN KEY (internal_control_id) REFERENCES public.internal_controls(id) ON DELETE SET NULL;


--
-- Name: tasks tasks_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: third_parties third_parties_common_third_party_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_parties
    ADD CONSTRAINT third_parties_common_third_party_id_fkey FOREIGN KEY (common_third_party_id) REFERENCES public.common_third_parties(id) ON DELETE SET NULL;


--
-- Name: third_parties third_parties_parent_third_party_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_parties
    ADD CONSTRAINT third_parties_parent_third_party_id_fkey FOREIGN KEY (parent_third_party_id) REFERENCES public.third_parties(id) ON DELETE CASCADE;


--
-- Name: third_party_administrators third_party_administrators_administrator_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_administrators
    ADD CONSTRAINT third_party_administrators_administrator_profile_id_fkey FOREIGN KEY (administrator_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: third_party_administrators third_party_administrators_third_party_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_administrators
    ADD CONSTRAINT third_party_administrators_third_party_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tracker_pattern_daily_hits tracker_pattern_daily_hits_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_pattern_daily_hits
    ADD CONSTRAINT tracker_pattern_daily_hits_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: tracker_pattern_daily_hits tracker_pattern_daily_hits_tracker_pattern_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_pattern_daily_hits
    ADD CONSTRAINT tracker_pattern_daily_hits_tracker_pattern_id_fkey FOREIGN KEY (tracker_pattern_id) REFERENCES public.tracker_patterns(id) ON DELETE CASCADE;


--
-- Name: tracker_pattern_discovery_hits tracker_pattern_discovery_hits_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_pattern_discovery_hits
    ADD CONSTRAINT tracker_pattern_discovery_hits_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: tracker_pattern_discovery_hits tracker_pattern_discovery_hits_tracker_pattern_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_pattern_discovery_hits
    ADD CONSTRAINT tracker_pattern_discovery_hits_tracker_pattern_id_fkey FOREIGN KEY (tracker_pattern_id) REFERENCES public.tracker_patterns(id) ON DELETE CASCADE;


--
-- Name: tracker_patterns tracker_patterns_common_tracker_pattern_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_patterns
    ADD CONSTRAINT tracker_patterns_common_tracker_pattern_id_fkey FOREIGN KEY (common_tracker_pattern_id) REFERENCES public.common_tracker_patterns(id) ON DELETE SET NULL;


--
-- Name: tracker_patterns tracker_patterns_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_patterns
    ADD CONSTRAINT tracker_patterns_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: tracker_patterns tracker_patterns_cookie_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_patterns
    ADD CONSTRAINT tracker_patterns_cookie_category_id_fkey FOREIGN KEY (cookie_category_id) REFERENCES public.cookie_categories(id) ON DELETE CASCADE;


--
-- Name: tracker_patterns tracker_patterns_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_patterns
    ADD CONSTRAINT tracker_patterns_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: tracker_patterns tracker_patterns_source_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_patterns
    ADD CONSTRAINT tracker_patterns_source_cookie_banner_id_fkey FOREIGN KEY (source_cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE SET NULL;


--
-- Name: tracker_resources tracker_resources_cookie_banner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_resources
    ADD CONSTRAINT tracker_resources_cookie_banner_id_fkey FOREIGN KEY (cookie_banner_id) REFERENCES public.cookie_banners(id) ON DELETE CASCADE;


--
-- Name: tracker_resources tracker_resources_cookie_banner_scan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_resources
    ADD CONSTRAINT tracker_resources_cookie_banner_scan_id_fkey FOREIGN KEY (cookie_banner_scan_id) REFERENCES public.cookie_banner_scans(id) ON DELETE SET NULL;


--
-- Name: tracker_resources tracker_resources_cookie_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_resources
    ADD CONSTRAINT tracker_resources_cookie_category_id_fkey FOREIGN KEY (cookie_category_id) REFERENCES public.cookie_categories(id) ON DELETE CASCADE;


--
-- Name: tracker_resources tracker_resources_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tracker_resources
    ADD CONSTRAINT tracker_resources_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: treatment_plan_events treatment_plan_events_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plan_events
    ADD CONSTRAINT treatment_plan_events_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: treatment_plan_events treatment_plan_events_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plan_events
    ADD CONSTRAINT treatment_plan_events_owner_profile_id_fkey FOREIGN KEY (owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: treatment_plan_events treatment_plan_events_risk_analysis_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plan_events
    ADD CONSTRAINT treatment_plan_events_risk_analysis_id_fkey FOREIGN KEY (risk_analysis_id) REFERENCES public.risk_analyses(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: treatment_plan_events treatment_plan_events_risk_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plan_events
    ADD CONSTRAINT treatment_plan_events_risk_id_fkey FOREIGN KEY (risk_id) REFERENCES public.risks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: treatment_plans_internal_controls treatment_plans_internal_controls_internal_control_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans_internal_controls
    ADD CONSTRAINT treatment_plans_internal_controls_internal_control_id_fkey FOREIGN KEY (internal_control_id) REFERENCES public.internal_controls(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: treatment_plans_internal_controls treatment_plans_internal_controls_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans_internal_controls
    ADD CONSTRAINT treatment_plans_internal_controls_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: treatment_plans_internal_controls treatment_plans_internal_controls_treatment_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans_internal_controls
    ADD CONSTRAINT treatment_plans_internal_controls_treatment_plan_id_fkey FOREIGN KEY (treatment_plan_id) REFERENCES public.treatment_plans(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: treatment_plans treatment_plans_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans
    ADD CONSTRAINT treatment_plans_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: treatment_plans treatment_plans_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans
    ADD CONSTRAINT treatment_plans_owner_profile_id_fkey FOREIGN KEY (owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: treatment_plans treatment_plans_risk_analysis_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans
    ADD CONSTRAINT treatment_plans_risk_analysis_id_fkey FOREIGN KEY (risk_analysis_id) REFERENCES public.risk_analyses(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: treatment_plans treatment_plans_risk_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treatment_plans
    ADD CONSTRAINT treatment_plans_risk_id_fkey FOREIGN KEY (risk_id) REFERENCES public.risks(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: iam_sessions usrmgr_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.iam_sessions
    ADD CONSTRAINT usrmgr_sessions_user_id_fkey FOREIGN KEY (identity_id) REFERENCES public.identities(id);


--
-- Name: third_party_business_associate_agreements vendor_business_associate_agreements_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_business_associate_agreements
    ADD CONSTRAINT vendor_business_associate_agreements_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: third_party_business_associate_agreements vendor_business_associate_agreements_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_business_associate_agreements
    ADD CONSTRAINT vendor_business_associate_agreements_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: third_party_compliance_reports vendor_compliance_reports_report_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_compliance_reports
    ADD CONSTRAINT vendor_compliance_reports_report_file_id_fkey FOREIGN KEY (report_file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: third_party_compliance_reports vendor_compliance_reports_vendor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_compliance_reports
    ADD CONSTRAINT vendor_compliance_reports_vendor_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON DELETE CASCADE;


--
-- Name: third_party_contacts vendor_contacts_vendor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_contacts
    ADD CONSTRAINT vendor_contacts_vendor_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: third_party_data_privacy_agreements vendor_data_privacy_agreements_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_data_privacy_agreements
    ADD CONSTRAINT vendor_data_privacy_agreements_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.files(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: third_party_data_privacy_agreements vendor_data_privacy_agreements_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_data_privacy_agreements
    ADD CONSTRAINT vendor_data_privacy_agreements_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: third_party_services vendor_services_vendor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_party_services
    ADD CONSTRAINT vendor_services_vendor_id_fkey FOREIGN KEY (third_party_id) REFERENCES public.third_parties(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: third_parties vendors_business_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_parties
    ADD CONSTRAINT vendors_business_owner_profile_id_fkey FOREIGN KEY (business_owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: third_parties vendors_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_parties
    ADD CONSTRAINT vendors_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: third_parties vendors_security_owner_profile_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.third_parties
    ADD CONSTRAINT vendors_security_owner_profile_id_fkey FOREIGN KEY (security_owner_profile_id) REFERENCES public.iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: webhook_data webhook_data_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_data
    ADD CONSTRAINT webhook_data_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: webhook_events webhook_events_webhook_data_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_webhook_data_id_fkey FOREIGN KEY (webhook_data_id) REFERENCES public.webhook_data(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: webhook_events webhook_events_webhook_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_webhook_subscription_id_fkey FOREIGN KEY (webhook_subscription_id) REFERENCES public.webhook_subscriptions(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: webhook_subscriptions webhook_subscriptions_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhook_subscriptions
    ADD CONSTRAINT webhook_subscriptions_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON UPDATE CASCADE ON DELETE CASCADE;

-- Reference rows required on an empty database.

INSERT INTO public.iam_oauth2_clients (id, tenant_id, organization_id, client_secret_hash, client_name, visibility, redirect_uris, scopes, grant_types, response_types, token_endpoint_auth_method, logo_uri, client_uri, contacts, created_at, updated_at, external_client_id) VALUES ('AAAAAAAAAAAASwAAAAAAAAAAcHJiYXVk', NULL, NULL, NULL, 'Auditor Mode', 'public', '{}', '{openid,profile,email,offline_access,v1:control:read,v1:iam:read,v1:org:read,v1:control}', '{urn:ietf:params:oauth:grant-type:device_code,refresh_token}', '{code}', 'none', NULL, NULL, NULL, NOW(), NOW(), NULL);
INSERT INTO public.iam_oauth2_clients (id, tenant_id, organization_id, client_secret_hash, client_name, visibility, redirect_uris, scopes, grant_types, response_types, token_endpoint_auth_method, logo_uri, client_uri, contacts, created_at, updated_at, external_client_id) VALUES ('AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp', NULL, NULL, NULL, 'Probo CLI', 'public', '{}', '{openid,profile,email,offline_access,v1:access-review,v1:agent,v1:ai-system,v1:asset,v1:audit,v1:business-function,v1:common-third-party,v1:compliance-page,v1:connector,v1:control,v1:datum,v1:document,v1:iam,v1:itam,v1:org,v1:privacy,v1:risk,v1:slack-connection,v1:task,v1:third-party,v1:webhook,v1:tag,v1:employee-portal}', '{urn:ietf:params:oauth:grant-type:device_code,refresh_token}', '{code}', 'none', NULL, NULL, NULL, NOW(), NOW(), NULL);

SELECT pg_catalog.set_config('search_path', 'public', true);
