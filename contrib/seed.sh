#!/usr/bin/env bash
# Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
#
# Permission to use, copy, modify, and/or distribute this software for any
# purpose with or without fee is hereby granted, provided that the above
# copyright notice and this permission notice appear in all copies.
#
# THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
# REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
# AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
# INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
# LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
# OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
# PERFORMANCE OF THIS SOFTWARE.

set -euo pipefail

BASE_URL="${PROBO_SEED_URL:-http://localhost:8080}"
CONNECT_API="$BASE_URL/api/connect/v1/graphql"
COOKIE_JAR=$(mktemp)
trap 'rm -f "$COOKIE_JAR"' EXIT

EMAIL="seed@dev.probo.test"
PASSWORD="seed@dev.probo.test"
FULL_NAME="Seed Admin"
ORG_NAME="Acme Corp"

# Helpers

gql_connect() {
  local query="$1"
  local variables="${2:-"{}"}"

  local payload
  payload=$(jq -n --arg q "$query" --argjson v "$variables" \
    '{query: $q, variables: $v}')

  curl -s -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
    -H "Content-Type: application/json" \
    -d "$payload" \
    "$CONNECT_API"
}

check_error() {
  local response="$1"
  local context="$2"
  local errors
  errors=$(echo "$response" | jq -r '.errors[0].message // empty')
  if [ -n "$errors" ]; then
    echo "ERROR ($context): $errors" >&2
    exit 1
  fi
}

prb_api() {
  local context="$1"; shift
  local resp
  resp=$($PRB api "$@")
  check_error "$resp" "$context"
  echo "$resp"
}

echo "==> Bootstrapping user and organization..."
vars=$(jo input="$(jo \
  email="$EMAIL" \
  password="$PASSWORD" \
  fullName="$FULL_NAME" \
)")
resp=$(gql_connect '
  mutation($input: SignUpInput!) {
    signUp(input: $input) {
      identity { id }
    }
  }
' "$vars")
check_error "$resp" "signUp"
echo "  Created user $EMAIL"

vars=$(jo input="$(jo name="$ORG_NAME")")
resp=$(gql_connect '
  mutation($input: CreateOrganizationInput!) {
    createOrganization(input: $input) {
      organization { id }
    }
  }
' "$vars")
check_error "$resp" "createOrganization"
ORG_ID=$(echo "$resp" | jq -r '.data.createOrganization.organization.id')
echo "  Created organization $ORG_NAME ($ORG_ID)"

vars=$(jo input="$(jo \
  organizationId="$ORG_ID" \
  continue="$BASE_URL" \
)")
resp=$(gql_connect '
  mutation($input: AssumeOrganizationSessionInput!) {
    assumeOrganizationSession(input: $input) {
      result {
        ... on OrganizationSessionCreated {
          session { id }
        }
      }
    }
  }
' "$vars")
check_error "$resp" "assumeOrganizationSession"
echo "  Assumed organization session"

EXPIRES_AT=$(date -u -v+1y +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null \
  || date -u -d "+1 year" +"%Y-%m-%dT%H:%M:%SZ")
vars=$(jo input="$(jo \
  name=seed \
  expiresAt="$EXPIRES_AT" \
)")
resp=$(gql_connect '
  mutation($input: CreatePersonalAPIKeyInput!) {
    createPersonalAPIKey(input: $input) {
      token
    }
  }
' "$vars")
check_error "$resp" "createPersonalAPIKey"
PROBO_TOKEN=$(echo "$resp" | jq -r '.data.createPersonalAPIKey.token')
export PROBO_TOKEN
export PROBO_HOST="$BASE_URL"
echo "  Created API key"

echo ""
echo "==> Seeding data..."

PRB="./bin/prb"

echo "  Creating people..."

create_person() {
  local full_name="$1"
  local position="$2"
  local email="$3"

  local vars
  vars=$(jo input="$(jo \
    organizationId="$ORG_ID" \
    emailAddress="$email" \
    fullName="$full_name" \
    role=EMPLOYEE \
    kind=EMPLOYEE \
    additionalEmailAddresses="$(jo -a)" \
    position="$position" \
  )")
  resp=$(gql_connect '
    mutation($input: CreateUserInput!) {
      createUser(input: $input) {
        profileEdge {
          node { id }
        }
      }
    }
  ' "$vars")
  check_error "$resp" "createPerson: $full_name"
}

create_person "Jane Cooper" \
  "Chief Information Security Officer" \
  "jane.cooper@dev.probo.test"
create_person "Marcus Chen" \
  "Security Engineer" \
  "marcus.chen@dev.probo.test"
create_person "Sofia Rodriguez" \
  "Compliance Manager" \
  "sofia.rodriguez@dev.probo.test"
create_person "David Kim" \
  "IT Administrator" \
  "david.kim@dev.probo.test"
create_person "Emily Nakamura" \
  "VP Engineering" \
  "emily.nakamura@dev.probo.test"
create_person "James O'Brien" \
  "Head of People" \
  "james.obrien@dev.probo.test"
create_person "Priya Patel" \
  "Data Protection Officer" \
  "priya.patel@dev.probo.test"
create_person "Alex Thompson" \
  "DevOps Lead" \
  "alex.thompson@dev.probo.test"

echo "    8 people created"

echo "  Creating frameworks and controls..."

create_framework() {
  local name="$1"
  local desc="$2"

  local resp
  resp=$(prb_api "createFramework: $name" '
    mutation($input: CreateFrameworkInput!) {
      createFramework(input: $input) {
        frameworkEdge {
          node { id }
        }
      }
    }
  ' -f input="$(jo \
    organizationId="$ORG_ID" \
    name="$name" \
    description="$desc" \
  )")
  local id
  id=$(echo "$resp" | jq -r '.data.createFramework.frameworkEdge.node.id // empty')
  if [ -z "$id" ]; then
    echo "ERROR (createFramework: $name): no framework id in response" >&2
    exit 1
  fi
  echo "$id"
}

create_control() {
  local framework_id="$1"
  local section="$2"
  local name="$3"
  local desc="$4"

  $PRB control create \
    --framework "$framework_id" \
    --section-title "$section" \
    --name "$name" \
    --description "$desc" > /dev/null
}

# ISO 27001:2022
ISO_ID=$(create_framework \
  "ISO 27001:2022" \
  "International standard for information security management systems")

create_control "$ISO_ID" \
  "A.5.1" \
  "Policies for information security" \
  "Management direction for information security shall be established"
create_control "$ISO_ID" \
  "A.5.2" \
  "Information security roles and responsibilities" \
  "Information security roles and responsibilities shall be defined and allocated"
create_control "$ISO_ID" \
  "A.5.3" \
  "Segregation of duties" \
  "Conflicting duties and areas of responsibility shall be segregated"
create_control "$ISO_ID" \
  "A.5.10" \
  "Acceptable use of information" \
  "Rules for the acceptable use of information shall be identified and documented"
create_control "$ISO_ID" \
  "A.5.23" \
  "Information security for cloud services" \
  "Processes for acquisition, use, management and exit from cloud services shall be established"
create_control "$ISO_ID" \
  "A.5.34" \
  "Privacy and protection of PII" \
  "The organization shall identify and meet requirements regarding privacy and protection of PII"
create_control "$ISO_ID" \
  "A.6.1" \
  "Screening" \
  "Background verification checks on all candidates for employment shall be carried out"
create_control "$ISO_ID" \
  "A.6.2" \
  "Terms and conditions of employment" \
  "Employment contractual agreements shall state personnel and organization responsibilities for information security"
create_control "$ISO_ID" \
  "A.6.3" \
  "Information security awareness training" \
  "Personnel and relevant interested parties shall receive appropriate information security awareness education and training"
create_control "$ISO_ID" \
  "A.7.1" \
  "Physical security perimeters" \
  "Security perimeters shall be defined and used to protect areas that contain information and information processing facilities"
create_control "$ISO_ID" \
  "A.7.2" \
  "Physical entry" \
  "Secure areas shall be protected by appropriate entry controls and access points"
create_control "$ISO_ID" \
  "A.8.1" \
  "User endpoint devices" \
  "Information stored on, processed by or accessible via user endpoint devices shall be protected"
create_control "$ISO_ID" \
  "A.8.2" \
  "Privileged access rights" \
  "The allocation and use of privileged access rights shall be restricted and managed"
create_control "$ISO_ID" \
  "A.8.3" \
  "Information access restriction" \
  "Access to information and other associated assets shall be restricted"
create_control "$ISO_ID" \
  "A.8.5" \
  "Secure authentication" \
  "Secure authentication technologies and procedures shall be established and implemented"
create_control "$ISO_ID" \
  "A.8.9" \
  "Configuration management" \
  "Configurations, including security configurations, of hardware, software, services and networks shall be established and managed"
create_control "$ISO_ID" \
  "A.8.15" \
  "Logging" \
  "Logs that record activities, exceptions, faults and other relevant events shall be produced and stored"
create_control "$ISO_ID" \
  "A.8.16" \
  "Monitoring activities" \
  "Networks, systems and applications shall be monitored for anomalous behavior"

echo "    ISO 27001:2022 — 18 controls"

# SOC 2
SOC2_ID=$(create_framework \
  "SOC 2" \
  "Service Organization Control 2 trust services criteria")

create_control "$SOC2_ID" \
  "CC1.1" \
  "Integrity and Ethical Values" \
  "The entity demonstrates a commitment to integrity and ethical values"
create_control "$SOC2_ID" \
  "CC1.2" \
  "Board Independence" \
  "The board of directors demonstrates independence from management"
create_control "$SOC2_ID" \
  "CC1.3" \
  "Management Structure" \
  "Management establishes structures, reporting lines, and authorities"
create_control "$SOC2_ID" \
  "CC2.1" \
  "Internal Communication" \
  "The entity obtains or generates and uses relevant information to support internal control"
create_control "$SOC2_ID" \
  "CC3.1" \
  "Objective Specification" \
  "The entity specifies objectives with sufficient clarity"
create_control "$SOC2_ID" \
  "CC3.2" \
  "Risk Identification" \
  "The entity identifies risks to the achievement of its objectives"
create_control "$SOC2_ID" \
  "CC3.3" \
  "Fraud Risk Assessment" \
  "The entity considers the potential for fraud in assessing risks"
create_control "$SOC2_ID" \
  "CC4.1" \
  "Monitoring Activities" \
  "The entity selects, develops, and performs ongoing evaluations"
create_control "$SOC2_ID" \
  "CC5.1" \
  "Control Selection" \
  "The entity selects and develops control activities that mitigate risks"
create_control "$SOC2_ID" \
  "CC5.2" \
  "Technology General Controls" \
  "The entity selects and develops general control activities over technology"
create_control "$SOC2_ID" \
  "CC6.1" \
  "Logical and Physical Access" \
  "The entity implements logical access security over protected assets"
create_control "$SOC2_ID" \
  "CC6.2" \
  "User Registration and Authorization" \
  "The entity registers and authorizes new internal and external users"
create_control "$SOC2_ID" \
  "CC6.3" \
  "Role-Based Access" \
  "The entity authorizes access based on authorization credentials and role"
create_control "$SOC2_ID" \
  "CC7.1" \
  "Infrastructure and Software Monitoring" \
  "The entity uses detection and monitoring procedures to identify changes to configurations"
create_control "$SOC2_ID" \
  "CC7.2" \
  "Change Management" \
  "The entity monitors system components for anomalies indicative of malicious acts"

echo "    SOC 2 — 15 controls"

# GDPR
GDPR_ID=$(create_framework \
  "GDPR" \
  "General Data Protection Regulation")

create_control "$GDPR_ID" \
  "Art.5" \
  "Principles relating to processing" \
  "Personal data shall be processed lawfully, fairly and in a transparent manner"
create_control "$GDPR_ID" \
  "Art.6" \
  "Lawfulness of processing" \
  "Processing shall be lawful only if at least one legal basis applies"
create_control "$GDPR_ID" \
  "Art.7" \
  "Conditions for consent" \
  "Where processing is based on consent, the controller shall be able to demonstrate consent"
create_control "$GDPR_ID" \
  "Art.13" \
  "Information to be provided" \
  "Information to be provided where personal data are collected from the data subject"
create_control "$GDPR_ID" \
  "Art.15" \
  "Right of access" \
  "The data subject shall have the right to obtain confirmation of processing"
create_control "$GDPR_ID" \
  "Art.17" \
  "Right to erasure" \
  "The data subject shall have the right to obtain erasure of personal data"
create_control "$GDPR_ID" \
  "Art.25" \
  "Data protection by design" \
  "The controller shall implement appropriate technical and organizational measures"
create_control "$GDPR_ID" \
  "Art.30" \
  "Records of processing activities" \
  "Each controller shall maintain a record of processing activities"
create_control "$GDPR_ID" \
  "Art.32" \
  "Security of processing" \
  "The controller shall implement appropriate technical and organizational measures to ensure security"
create_control "$GDPR_ID" \
  "Art.33" \
  "Notification of personal data breach" \
  "The controller shall notify the supervisory authority within 72 hours"

echo "    GDPR — 10 controls"

echo "  Creating risks..."

create_risk() {
  local name="$1"
  local category="$2"
  local treatment="$3"
  local likelihood="$4"
  local impact="$5"

  $PRB risk create \
    --org "$ORG_ID" \
    --name "$name" \
    --category "$category" \
    --treatment "$treatment" \
    --inherent-likelihood "$likelihood" \
    --inherent-impact "$impact" > /dev/null
}

create_risk \
  "Unauthorized access to production systems" \
  SECURITY MITIGATED 4 5
create_risk \
  "Sensitive data exfiltration by insider threat" \
  SECURITY MITIGATED 2 5
create_risk \
  "Phishing campaign targeting employees" \
  SECURITY MITIGATED 4 3
create_risk \
  "Ransomware via supply chain compromise" \
  SECURITY MITIGATED 2 5
create_risk \
  "Credential stuffing on customer login portal" \
  SECURITY MITIGATED 3 4
create_risk \
  "Stolen developer laptop with source code" \
  SECURITY MITIGATED 3 3
create_risk \
  "API key leaked in public repository" \
  SECURITY MITIGATED 3 4
create_risk \
  "Social engineering of support staff" \
  SECURITY MITIGATED 3 3
create_risk \
  "Brute force attack on admin panel" \
  SECURITY MITIGATED 2 4
create_risk \
  "Man-in-the-middle attack on internal network" \
  SECURITY MITIGATED 1 4
create_risk \
  "Malicious browser extension on corporate devices" \
  SECURITY MITIGATED 2 3
create_risk \
  "Unauthorized physical access to server room" \
  SECURITY MITIGATED 1 5

create_risk \
  "Third-party SaaS vendor data breach" \
  OPERATIONAL TRANSFERRED 3 4
create_risk \
  "Cloud region outage causing service disruption" \
  OPERATIONAL ACCEPTED 2 4
create_risk \
  "Database corruption from failed migration" \
  OPERATIONAL MITIGATED 2 5
create_risk \
  "Loss of key personnel with critical knowledge" \
  OPERATIONAL MITIGATED 3 3
create_risk \
  "Backup restoration failure during disaster recovery" \
  OPERATIONAL MITIGATED 2 5
create_risk \
  "DNS hijacking of company domain" \
  OPERATIONAL MITIGATED 1 5
create_risk \
  "CDN provider outage affecting content delivery" \
  OPERATIONAL ACCEPTED 2 3
create_risk \
  "Email service compromise leaking internal comms" \
  OPERATIONAL MITIGATED 2 4

create_risk \
  "Non-compliance with GDPR data subject rights" \
  COMPLIANCE MITIGATED 2 4
create_risk \
  "Failure to meet SOC 2 audit requirements" \
  COMPLIANCE MITIGATED 2 4
create_risk \
  "Breach notification deadline missed" \
  COMPLIANCE MITIGATED 1 5
create_risk \
  "Inadequate data processing agreements with vendors" \
  COMPLIANCE MITIGATED 3 3
create_risk \
  "Employee data retained beyond legal period" \
  COMPLIANCE MITIGATED 2 3
create_risk \
  "Cross-border data transfer without safeguards" \
  COMPLIANCE MITIGATED 2 4
create_risk \
  "Cookie consent mechanism non-compliant" \
  COMPLIANCE MITIGATED 3 2
create_risk \
  "Incomplete records of processing activities" \
  COMPLIANCE MITIGATED 3 3

create_risk \
  "Cloud infrastructure misconfiguration exposing data" \
  TECHNICAL MITIGATED 3 4
create_risk \
  "Loss of encryption keys for production database" \
  TECHNICAL MITIGATED 1 5
create_risk \
  "Denial of service attack on customer-facing APIs" \
  TECHNICAL MITIGATED 3 3
create_risk \
  "TLS certificate expiration causing service outage" \
  TECHNICAL MITIGATED 2 3
create_risk \
  "Container escape vulnerability in production cluster" \
  TECHNICAL MITIGATED 1 5
create_risk \
  "Dependency with known CVE deployed to production" \
  TECHNICAL MITIGATED 3 4
create_risk \
  "Logging pipeline failure hiding security incidents" \
  TECHNICAL MITIGATED 2 4

echo "    35 risks created"

echo "  Creating vendors..."

create_vendor() {
  local name="$1"
  local description="$2"

  local resp
  resp=$(prb_api "createVendor: $name" '
    mutation($input: CreateVendorInput!) {
      createVendor(input: $input) {
        vendorEdge {
          node { id }
        }
      }
    }
  ' -f input="$(jo \
      organizationId="$ORG_ID" \
      name="$name" \
      description="$description" \
    )")
  local id
  id=$(echo "$resp" | jq -r '.data.createVendor.vendorEdge.node.id // empty')
  if [ -z "$id" ]; then
    echo "ERROR (createVendor: $name): no vendor id in response" >&2
    exit 1
  fi
}

create_vendor "Amazon Web Services" \
  "Cloud infrastructure and compute"
create_vendor "Google Cloud Platform" \
  "BigQuery analytics and AI services"
create_vendor "Google Workspace" \
  "Email, calendar, and productivity suite"
create_vendor "Microsoft 365" \
  "Office productivity and collaboration"
create_vendor "Datadog" \
  "Application monitoring and observability"
create_vendor "PagerDuty" \
  "Incident management and on-call scheduling"
create_vendor "Slack" \
  "Team communication and messaging"
create_vendor "GitHub" \
  "Source code management and CI/CD"
create_vendor "Stripe" \
  "Payment processing and billing"
create_vendor "Salesforce" \
  "Customer relationship management"
create_vendor "HubSpot" \
  "Marketing automation and CRM"
create_vendor "Notion" \
  "Documentation and knowledge management"
create_vendor "1Password" \
  "Enterprise password management"
create_vendor "Okta" \
  "Identity and access management"
create_vendor "CrowdStrike" \
  "Endpoint protection and threat intelligence"
create_vendor "Vanta" \
  "Compliance automation and monitoring"
create_vendor "Jira" \
  "Project management and issue tracking"
create_vendor "Cloudflare" \
  "CDN, DNS, and DDoS protection"
create_vendor "Twilio SendGrid" \
  "Transactional email delivery"
create_vendor "Snowflake" \
  "Cloud data warehouse"

echo "    20 vendors created"

echo "  Creating measures..."

create_measure() {
  local name="$1"
  local category="$2"

  local resp
  resp=$(prb_api "createMeasure: $name" '
    mutation($input: CreateMeasureInput!) {
      createMeasure(input: $input) {
        measureEdge {
          node { id }
        }
      }
    }
  ' -f input="$(jo \
      organizationId="$ORG_ID" \
      name="$name" \
      category="$category" \
    )")
  local id
  id=$(echo "$resp" | jq -r '.data.createMeasure.measureEdge.node.id // empty')
  if [ -z "$id" ]; then
    echo "ERROR (createMeasure: $name): no measure id in response" >&2
    exit 1
  fi
}

create_measure "Information Security Policy" \
  "POLICY"
create_measure "Access Control Policy" \
  "POLICY"
create_measure "Incident Response Plan" \
  "POLICY"
create_measure "Data Classification Policy" \
  "POLICY"
create_measure "Acceptable Use Policy" \
  "POLICY"

create_measure "Multi-Factor Authentication" \
  "TECHNICAL"
create_measure "Endpoint Detection and Response" \
  "TECHNICAL"
create_measure "Full Disk Encryption" \
  "TECHNICAL"
create_measure "Automated Vulnerability Scanning" \
  "TECHNICAL"
create_measure "Network Segmentation and Firewall Rules" \
  "TECHNICAL"
create_measure "TLS Encryption in Transit" \
  "TECHNICAL"

create_measure "Annual Security Awareness Training" \
  "ORGANIZATIONAL"
create_measure "Quarterly Access Reviews" \
  "ORGANIZATIONAL"
create_measure "Background Checks for New Hires" \
  "ORGANIZATIONAL"
create_measure "Tabletop Disaster Recovery Exercises" \
  "ORGANIZATIONAL"

echo "    15 measures created"

echo ""
echo "Seed complete!"
echo "  Email:    $EMAIL"
echo "  Password: $PASSWORD"
echo "  Org:      $ORG_NAME"
echo ""
echo "  Created:"
echo "    3 frameworks, 43 controls"
echo "    35 risks"
echo "    20 vendors"
echo "    15 measures"
echo "    8 people"
echo ""
echo "  To use the CLI:"
echo "    export PROBO_HOST=$BASE_URL"
echo "    export PROBO_TOKEN=$PROBO_TOKEN"
