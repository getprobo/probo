#!/usr/bin/env bash
# Copyright (c) 2026 Probo Inc <hello@probo.com>.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

# GraphQL documents are intentional single-quoted literals (no expansion).
# shellcheck disable=SC2016

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
  local context="$1"
  shift
  local resp
  resp=$($PRB api "$@")
  check_error "$resp" "$context"
  echo "$resp"
}

curl -sf -o /dev/null "$BASE_URL/healthz" \
  || {
    echo "ERROR: API at $BASE_URL is not available" >&2
    exit 1
  }

echo "==> Bootstrapping user and organization..."
vars=$(jo input="$(
  jo \
    email="$EMAIL" \
    password="$PASSWORD" \
    fullName="$FULL_NAME"
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

vars=$(jo input="$(
  jo \
    organizationId="$ORG_ID" \
    continue="$BASE_URL"
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
vars=$(jo input="$(
  jo \
    name=seed \
    expiresAt="$EXPIRES_AT" \
    scopes="$(
      jo -a \
        v1:access-review \
        v1:agent \
        v1:ai-system \
        v1:asset \
        v1:audit \
        v1:business-function \
        v1:compliance-page \
        v1:connector \
        v1:control \
        v1:datum \
        v1:document \
        v1:iam \
        v1:itam \
        v1:org \
        v1:privacy \
        v1:resource-alias \
        v1:risk \
        v1:task \
        v1:third-party \
        v1:webhook
    )"
)")
resp=$(gql_connect '
  mutation($input: CreateOAuth2AccessTokenInput!) {
    createOAuth2AccessToken(input: $input) {
      token
    }
  }
' "$vars")
check_error "$resp" "createOAuth2AccessToken"
PROBO_TOKEN=$(echo "$resp" | jq -r '.data.createOAuth2AccessToken.token')
export PROBO_TOKEN
export PROBO_HOST="$BASE_URL"
echo "  Created OAuth access token"

echo ""
echo "==> Seeding data..."

PRB="./bin/prb"

echo "  Creating people..."

PROFILE_IDS=()

create_person() {
  local full_name="$1"
  local position="$2"
  local email="$3"

  local vars
  vars=$(jo input="$(
    jo \
      organizationId="$ORG_ID" \
      emailAddress="$email" \
      fullName="$full_name" \
      role=EMPLOYEE \
      kind=EMPLOYEE \
      additionalEmailAddresses="$(jo -a </dev/null)" \
      position="$position"
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

  local profile_id
  profile_id=$(echo "$resp" | jq -r '.data.createUser.profileEdge.node.id // empty')
  if [ -z "$profile_id" ]; then
    echo "ERROR (createPerson: $full_name): no profile id in response" >&2
    exit 1
  fi
  PROFILE_IDS+=("$profile_id")
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
  ' -f input="$(
    jo \
      organizationId="$ORG_ID" \
      name="$name" \
      description="$desc"
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
    --description "$desc" >/dev/null
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

  $PRB risk create \
    --org "$ORG_ID" \
    --name "$name" \
    --category "$category" >/dev/null
}

create_risk \
  "Unauthorized access to production systems" \
  SECURITY
create_risk \
  "Sensitive data exfiltration by insider threat" \
  SECURITY
create_risk \
  "Phishing campaign targeting employees" \
  SECURITY
create_risk \
  "Ransomware via supply chain compromise" \
  SECURITY
create_risk \
  "Credential stuffing on customer login portal" \
  SECURITY
create_risk \
  "Stolen developer laptop with source code" \
  SECURITY
create_risk \
  "API key leaked in public repository" \
  SECURITY
create_risk \
  "Social engineering of support staff" \
  SECURITY
create_risk \
  "Brute force attack on admin panel" \
  SECURITY
create_risk \
  "Man-in-the-middle attack on internal network" \
  SECURITY
create_risk \
  "Malicious browser extension on corporate devices" \
  SECURITY
create_risk \
  "Unauthorized physical access to server room" \
  SECURITY

create_risk \
  "Third-party SaaS data breach" \
  OPERATIONAL
create_risk \
  "Cloud region outage causing service disruption" \
  OPERATIONAL
create_risk \
  "Database corruption from failed migration" \
  OPERATIONAL
create_risk \
  "Loss of key personnel with critical knowledge" \
  OPERATIONAL
create_risk \
  "Backup restoration failure during disaster recovery" \
  OPERATIONAL
create_risk \
  "DNS hijacking of company domain" \
  OPERATIONAL
create_risk \
  "CDN provider outage affecting content delivery" \
  OPERATIONAL
create_risk \
  "Email service compromise leaking internal comms" \
  OPERATIONAL

create_risk \
  "Non-compliance with GDPR data subject rights" \
  COMPLIANCE
create_risk \
  "Failure to meet SOC 2 audit requirements" \
  COMPLIANCE
create_risk \
  "Breach notification deadline missed" \
  COMPLIANCE
create_risk \
  "Inadequate data processing agreements with third parties" \
  COMPLIANCE
create_risk \
  "Employee data retained beyond legal period" \
  COMPLIANCE
create_risk \
  "Cross-border data transfer without safeguards" \
  COMPLIANCE
create_risk \
  "Cookie consent mechanism non-compliant" \
  COMPLIANCE
create_risk \
  "Incomplete records of processing activities" \
  COMPLIANCE

create_risk \
  "Cloud infrastructure misconfiguration exposing data" \
  TECHNICAL
create_risk \
  "Loss of encryption keys for production database" \
  TECHNICAL
create_risk \
  "Denial of service attack on customer-facing APIs" \
  TECHNICAL
create_risk \
  "TLS certificate expiration causing service outage" \
  TECHNICAL
create_risk \
  "Container escape vulnerability in production cluster" \
  TECHNICAL
create_risk \
  "Dependency with known CVE deployed to production" \
  TECHNICAL
create_risk \
  "Logging pipeline failure hiding security incidents" \
  TECHNICAL

echo "    35 risks created"

echo "  Creating third parties..."

create_third_party() {
  local name="$1"
  local description="$2"

  local resp
  resp=$(prb_api "createThirdParty: $name" '
    mutation($input: CreateThirdPartyInput!) {
      createThirdParty(input: $input) {
        thirdPartyEdge {
          node { id }
        }
      }
    }
  ' -f input="$(
    jo \
      organizationId="$ORG_ID" \
      name="$name" \
      description="$description"
  )")
  local id
  id=$(echo "$resp" | jq -r '.data.createThirdParty.thirdPartyEdge.node.id // empty')
  if [ -z "$id" ]; then
    echo "ERROR (createThirdParty: $name): no third party id in response" >&2
    exit 1
  fi
}

create_third_party "Amazon Web Services" \
  "Cloud infrastructure and compute"
create_third_party "Google Cloud Platform" \
  "BigQuery analytics and AI services"
create_third_party "Google Workspace" \
  "Email, calendar, and productivity suite"
create_third_party "Microsoft 365" \
  "Office productivity and collaboration"
create_third_party "Datadog" \
  "Application monitoring and observability"
create_third_party "PagerDuty" \
  "Incident management and on-call scheduling"
create_third_party "Slack" \
  "Team communication and messaging"
create_third_party "GitHub" \
  "Source code management and CI/CD"
create_third_party "Stripe" \
  "Payment processing and billing"
create_third_party "Salesforce" \
  "Customer relationship management"
create_third_party "HubSpot" \
  "Marketing automation and CRM"
create_third_party "Notion" \
  "Documentation and knowledge management"
create_third_party "1Password" \
  "Enterprise password management"
create_third_party "Okta" \
  "Identity and access management"
create_third_party "CrowdStrike" \
  "Endpoint protection and threat intelligence"
create_third_party "Vanta" \
  "Compliance automation and monitoring"
create_third_party "Jira" \
  "Project management and issue tracking"
create_third_party "Cloudflare" \
  "CDN, DNS, and DDoS protection"
create_third_party "Twilio SendGrid" \
  "Transactional email delivery"
create_third_party "Snowflake" \
  "Cloud data warehouse"

echo "    20 third parties created"

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
  ' -f input="$(
    jo \
      organizationId="$ORG_ID" \
      name="$name" \
      category="$category"
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

echo "  Creating devices..."

AGENT_API="$BASE_URL/api/agent/v1"

create_device() {
  local owner_id="$1"

  local input
  if [ -n "$owner_id" ]; then
    input=$(jo organizationId="$ORG_ID" ownerId="$owner_id")
  else
    input=$(jo organizationId="$ORG_ID")
  fi

  local resp
  resp=$(prb_api "createDevice" '
    mutation($input: CreateDeviceInput!) {
      createDevice(input: $input) {
        device { id }
        enrollmentToken
      }
    }
  ' -f input="$input")

  local device_id token
  device_id=$(echo "$resp" | jq -r '.data.createDevice.device.id // empty')
  token=$(echo "$resp" | jq -r '.data.createDevice.enrollmentToken // empty')
  if [ -z "$device_id" ] || [ -z "$token" ]; then
    echo "ERROR (createDevice): missing device id or enrollment token in response" >&2
    exit 1
  fi

  echo "$device_id $token"
}

agent_enroll() {
  local token="$1"

  local resp api_key
  resp=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$(jq -n --arg t "$token" '{token: $t}')" \
    "$AGENT_API/enroll")
  api_key=$(echo "$resp" | jq -r '.api_key // empty')
  if [ -z "$api_key" ]; then
    echo "ERROR (agent_enroll): no api_key in response: $resp" >&2
    exit 1
  fi

  echo "$api_key"
}

agent_heartbeat() {
  local api_key="$1"
  local hardware_uuid="$2"
  local hostname="$3"
  local platform="$4"
  local os_version="$5"
  local agent_version="$6"
  local serial="${7:-}"

  # Build with jq (not jo) so numeric-looking values like os_version "14.5"
  # stay JSON strings; jo would coerce them to numbers and fail decoding.
  local body
  body=$(jq -n \
    --arg hw "$hardware_uuid" \
    --arg sn "$serial" \
    --arg hn "$hostname" \
    --arg pf "$platform" \
    --arg ov "$os_version" \
    --arg av "$agent_version" \
    '{hardware_uuid: $hw, hostname: $hn, platform: $pf, os_version: $ov, agent_version: $av}
      + (if $sn == "" then {} else {serial_number: $sn} end)')

  local resp code
  resp=$(curl -s -w '\n%{http_code}' -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $api_key" \
    -d "$body" \
    "$AGENT_API/heartbeat")
  code=$(echo "$resp" | tail -n1)
  if [ "$code" != "200" ]; then
    echo "ERROR (agent_heartbeat $hostname): HTTP $code" >&2
    echo "  request: $body" >&2
    echo "  response: $(echo "$resp" | sed '$d')" >&2
    exit 1
  fi
}

# posture_evidence <platform> <check_key> <status> <os_version>
# Emits platform-shaped evidence JSON that ParseDevicePostureValue can turn
# into a non-UNKNOWN value for PASS/FAIL, or UNKNOWN/None for the rest.
posture_evidence() {
  local platform="$1"
  local check_key="$2"
  local status="$3"
  local os_version="$4"

  case "$platform:$check_key" in
    DARWIN:DISK_ENCRYPTION)
      case "$status" in
        PASS) jq -nc '{raw:"FileVault is On."}' ;;
        FAIL) jq -nc '{raw:"FileVault is Off."}' ;;
        *) jq -nc '{note:"fdesetup unavailable"}' ;;
      esac
      ;;
    LINUX:DISK_ENCRYPTION)
      case "$status" in
        PASS) jq -nc '{crypttab_present:true,crypttab_lines:["nvme0n1p3_crypt UUID=6c2f none luks,discard"]}' ;;
        FAIL) jq -nc '{crypttab_present:false,lsblk:"nvme0n1 disk\nnvme0n1p2 part ext4 /"}' ;;
        *) jq -nc '{crypttab_present:false,lsblk_error:"lsblk not found"}' ;;
      esac
      ;;
    WINDOWS:DISK_ENCRYPTION)
      case "$status" in
        PASS) jq -nc '{backend:"Get-BitLockerVolume",volumes:{ "C:":"On"}}' ;;
        FAIL) jq -nc '{backend:"Get-BitLockerVolume",volumes:{ "C:":"Off"},encryption_percentage:100}' ;;
        *) jq -nc '{note:"Get-BitLockerVolume not available"}' ;;
      esac
      ;;
    DARWIN:SCREEN_LOCK)
      case "$status" in
        PASS) jq -nc '{backend:"sysadminctl",mode:"seconds",seconds:900,raw:"screenLock delay is 900 seconds"}' ;;
        FAIL) jq -nc '{backend:"sysadminctl",mode:"off",raw:"screenLock is off"}' ;;
        *) jq -nc '{backend:"sysadminctl",error:"sysadminctl failed"}' ;;
      esac
      ;;
    LINUX:SCREEN_LOCK)
      case "$status" in
        PASS) jq -nc '{backend:"gnome",schema:"org.gnome.desktop.screensaver",lock_enabled:"true"}' ;;
        FAIL) jq -nc '{backend:"gnome",schema:"org.gnome.desktop.screensaver",lock_enabled:"false"}' ;;
        *) jq -nc '{backend:"gnome",error:"gsettings failed"}' ;;
      esac
      ;;
    WINDOWS:SCREEN_LOCK)
      case "$status" in
        PASS) jq -nc '{backend:"hkey_users",users:{"S-1-5-21-1004336348-1177238915-682003330-1001":"1"}}' ;;
        FAIL) jq -nc '{backend:"hkey_users",users:{"S-1-5-21-1004336348-1177238915-682003330-1001":"0"}}' ;;
        *) jq -nc '{backend:"hkey_users",users:{},note:"no interactive user hives loaded"}' ;;
      esac
      ;;
    DARWIN:FIREWALL_ENABLED)
      case "$status" in
        PASS) jq -nc '{backend:"defaults",global_state:"1"}' ;;
        FAIL) jq -nc '{backend:"defaults",global_state:"0"}' ;;
        *) jq -nc '{note:"no known firewall tool found"}' ;;
      esac
      ;;
    LINUX:FIREWALL_ENABLED)
      case "$status" in
        PASS) jq -nc '{backend:"ufw",raw:"Status: active"}' ;;
        FAIL) jq -nc '{backend:"ufw",raw:"Status: inactive"}' ;;
        *) jq -nc '{note:"no known firewall tool found"}' ;;
      esac
      ;;
    WINDOWS:FIREWALL_ENABLED)
      case "$status" in
        PASS) jq -nc '{backend:"Get-NetFirewallProfile",raw:"Domain=True;Private=True;Public=True",profiles:{Domain:"True",Private:"True",Public:"True"}}' ;;
        FAIL) jq -nc '{backend:"Get-NetFirewallProfile",raw:"Domain=True;Private=True;Public=False",profiles:{Domain:"True",Private:"True",Public:"False"}}' ;;
        *) jq -nc '{backend:"netsh",state_lines:[]}' ;;
      esac
      ;;
    DARWIN:TIME_SYNC)
      case "$status" in
        PASS) jq -nc '{raw:"Network Time: On"}' ;;
        FAIL) jq -nc '{raw:"Network Time: Off"}' ;;
        *) jq -nc '{note:"systemsetup unavailable"}' ;;
      esac
      ;;
    LINUX:TIME_SYNC)
      case "$status" in
        PASS) jq -nc '{raw:"Timezone=Europe/Paris\nLocalRTC=no\nCanNTP=yes\nNTP=yes\nNTPSynchronized=yes"}' ;;
        FAIL) jq -nc '{raw:"Timezone=Europe/Paris\nLocalRTC=no\nCanNTP=yes\nNTP=yes\nNTPSynchronized=no"}' ;;
        *) jq -nc '{note:"timedatectl not installed"}' ;;
      esac
      ;;
    WINDOWS:TIME_SYNC)
      case "$status" in
        PASS) jq -nc '{backend:"w32time",w32time_status:"Running",w32time_type:"NTP",ntp_server:"time.windows.com,0x8"}' ;;
        FAIL) jq -nc '{backend:"w32time",w32time_status:"Stopped",w32time_type:"NTP"}' ;;
        *) jq -nc '{error:"exit status 0x80070426",stderr:""}' ;;
      esac
      ;;
    DARWIN:OS_VERSION)
      case "$status" in
        UNKNOWN | NOT_APPLICABLE) jq -nc '{error:"sw_vers failed"}' ;;
        *) jq -nc --arg v "$os_version" '{product_version:$v,build_version:"24E248"}' ;;
      esac
      ;;
    LINUX:OS_VERSION)
      case "$status" in
        UNKNOWN | NOT_APPLICABLE) jq -nc '{error:"os-release unreadable"}' ;;
        *) jq -nc --arg v "$os_version" '{pretty_name:$v,version_id:$v,id:"linux"}' ;;
      esac
      ;;
    WINDOWS:OS_VERSION)
      case "$status" in
        UNKNOWN | NOT_APPLICABLE) jq -nc '{error:"wmic failed"}' ;;
        *) jq -nc --arg v "$os_version" '{caption:$v}' ;;
      esac
      ;;
    DARWIN:AUTO_UPDATE)
      case "$status" in
        PASS) jq -nc '{backend:"defaults",AutomaticCheckEnabled:{source:"system",value:"1",enabled:true},AutomaticDownload:{source:"default",enabled:true}}' ;;
        FAIL) jq -nc '{backend:"defaults",disabled_keys:["AutomaticDownload"]}' ;;
        *) jq -nc '{backend:"defaults",indeterminate_keys:["ConfigDataInstall"]}' ;;
      esac
      ;;
    LINUX:AUTO_UPDATE)
      case "$status" in
        PASS) jq -nc '{backend:"unattended-upgrades",raw:"APT::Periodic::Update-Package-Lists \"1\";\nAPT::Periodic::Unattended-Upgrade \"1\";\n"}' ;;
        FAIL) jq -nc '{backend:"unattended-upgrades",raw:"APT::Periodic::Update-Package-Lists \"0\";\nAPT::Periodic::Unattended-Upgrade \"0\";\n"}' ;;
        *) jq -nc '{note:"unattended-upgrades not installed"}' ;;
      esac
      ;;
    WINDOWS:AUTO_UPDATE)
      case "$status" in
        PASS) jq -nc '{no_auto_update:"0",au_options:"4"}' ;;
        FAIL) jq -nc '{no_auto_update:"",au_options:"2"}' ;;
        *) jq -nc '{no_auto_update:"",au_options:"",wuauserv:""}' ;;
      esac
      ;;
    DARWIN:PASSWORD_POLICY)
      case "$status" in
        PASS) jq -nc '{raw_truncated:"<dict><key>policyCategoryPasswordContent</key><array/></dict>"}' ;;
        FAIL) jq -nc '{raw_truncated:"There are no account policies for all users."}' ;;
        *) jq -nc '{error:"pwpolicy failed"}' ;;
      esac
      ;;
    LINUX:PASSWORD_POLICY)
      case "$status" in
        PASS) jq -nc '{pass_min_len:"12",pass_max_days:"90",pass_min_len_value:12}' ;;
        FAIL) jq -nc '{pass_min_len:"",pass_max_days:"99999",parse_error:"PASS_MIN_LEN not set"}' ;;
        *) jq -nc '{error:"login.defs unreadable"}' ;;
      esac
      ;;
    WINDOWS:PASSWORD_POLICY)
      case "$status" in
        PASS) jq -nc '{min_password_length:8}' ;;
        FAIL) jq -nc '{min_password_length:0}' ;;
        *) jq -nc '{error:"MinPasswordLength unavailable"}' ;;
      esac
      ;;
    # REMOTE_LOGIN is inverted: PASS means remote access is Off / denied.
    DARWIN:REMOTE_LOGIN)
      case "$status" in
        PASS) jq -nc '{raw:"Remote Login: Off"}' ;;
        FAIL) jq -nc '{raw:"Remote Login: On"}' ;;
        *) jq -nc '{error:"systemsetup failed"}' ;;
      esac
      ;;
    LINUX:REMOTE_LOGIN)
      case "$status" in
        PASS) jq -nc '{is_active:"inactive"}' ;;
        FAIL) jq -nc '{is_active:"active"}' ;;
        *) jq -nc '{is_active:""}' ;;
      esac
      ;;
    WINDOWS:REMOTE_LOGIN)
      case "$status" in
        PASS) jq -nc '{fdeny_ts_connections:"1"}' ;;
        FAIL) jq -nc '{fdeny_ts_connections:"0"}' ;;
        *) jq -nc '{error:"registry read failed"}' ;;
      esac
      ;;
    DARWIN:MALWARE_PROTECTION)
      case "$status" in
        PASS) jq -nc '{engine:"XProtect",version:"5260"}' ;;
        FAIL) jq -nc '{engine:"XProtect",note:"XProtect.meta.plist not found in expected locations"}' ;;
        *) jq -nc '{note:"XProtect check skipped"}' ;;
      esac
      ;;
    LINUX:MALWARE_PROTECTION)
      case "$status" in
        PASS) jq -nc '{active:["ClamAV"],installed:[]}' ;;
        FAIL) jq -nc '{active:[],installed:["ClamAV"]}' ;;
        *) jq -nc '{active:[],installed:[]}' ;;
      esac
      ;;
    WINDOWS:MALWARE_PROTECTION)
      case "$status" in
        PASS) jq -nc '{antivirus_enabled:true,real_time_protection:true,am_service_enabled:true}' ;;
        FAIL) jq -nc '{antivirus_enabled:false,real_time_protection:false,am_service_enabled:false}' ;;
        *) jq -nc '{note:"defender status unavailable"}' ;;
      esac
      ;;
    *)
      jq -nc '{}'
      ;;
  esac
}

# agent_postures <api_key> <platform> <os_version> <CHECK_KEY:STATUS>...
agent_postures() {
  local api_key="$1"
  local platform="$2"
  local os_version="$3"
  shift 3

  local now
  now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  local results='[]'
  local pair check_key status evidence
  for pair in "$@"; do
    check_key="${pair%%:*}"
    status="${pair#*:}"
    evidence=$(posture_evidence "$platform" "$check_key" "$status" "$os_version")
    results=$(jq -nc \
      --argjson results "$results" \
      --arg check_key "$check_key" \
      --arg status "$status" \
      --arg observed_at "$now" \
      --argjson evidence "$evidence" \
      '$results + [{
        check_key: $check_key,
        status: $status,
        observed_at: $observed_at,
        evidence: $evidence
      }]')
  done

  local body
  body=$(jq -nc --argjson results "$results" '{results: $results}')

  local code
  code=$(curl -s -o /dev/null -w '%{http_code}' -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $api_key" \
    -d "$body" \
    "$AGENT_API/postures")
  if [ "$code" != "204" ] && [ "$code" != "200" ]; then
    echo "ERROR (agent_postures): HTTP $code" >&2
    echo "  request: $body" >&2
    exit 1
  fi
}

revoke_device() {
  local device_id="$1"

  prb_api "revokeDevice" '
    mutation($input: RevokeDeviceInput!) {
      revokeDevice(input: $input) {
        device { id }
      }
    }
  ' -f input="$(jo deviceId="$device_id")" >/dev/null
}

# seed_device <owner_id> <hostname> <platform> <os_version> <serial> <CHECK_KEY:STATUS>...
# Creates a device, enrolls it, sends a heartbeat (activating it) and posts
# posture results. Prints the device id.
seed_device() {
  local owner_id="$1"
  local hostname="$2"
  local platform="$3"
  local os_version="$4"
  local serial="$5"
  shift 5

  local out device_id token
  out=$(create_device "$owner_id")
  device_id="${out%% *}"
  token="${out#* }"

  local api_key
  api_key=$(agent_enroll "$token")

  local hardware_uuid
  hardware_uuid="hw-$(echo "$hostname" | tr '[:upper:]' '[:lower:]')"

  agent_heartbeat "$api_key" "$hardware_uuid" "$hostname" "$platform" "$os_version" "1.0.0" "$serial"
  agent_postures "$api_key" "$platform" "$os_version" "$@"

  echo "$device_id"
}

# 6 active devices across platforms, each owned by a seeded person with a
# varied posture mix so the UI shows every badge state.
seed_device "${PROFILE_IDS[0]}" "jane-macbook-pro" "DARWIN" "14.5" "C02XY1Z2JGH7" \
  DISK_ENCRYPTION:PASS SCREEN_LOCK:PASS FIREWALL_ENABLED:PASS TIME_SYNC:PASS \
  OS_VERSION:PASS AUTO_UPDATE:PASS PASSWORD_POLICY:PASS REMOTE_LOGIN:PASS \
  MALWARE_PROTECTION:PASS >/dev/null

seed_device "${PROFILE_IDS[1]}" "marcus-thinkpad" "LINUX" "Ubuntu 24.04" "PF3ABCDE" \
  DISK_ENCRYPTION:PASS SCREEN_LOCK:PASS FIREWALL_ENABLED:FAIL TIME_SYNC:PASS \
  OS_VERSION:PASS AUTO_UPDATE:UNKNOWN PASSWORD_POLICY:PASS REMOTE_LOGIN:FAIL \
  MALWARE_PROTECTION:NOT_APPLICABLE >/dev/null

seed_device "${PROFILE_IDS[4]}" "emily-macbook-air" "DARWIN" "14.4" "C02AB3C4JGH8" \
  DISK_ENCRYPTION:PASS SCREEN_LOCK:FAIL FIREWALL_ENABLED:PASS TIME_SYNC:PASS \
  OS_VERSION:PASS AUTO_UPDATE:PASS PASSWORD_POLICY:FAIL REMOTE_LOGIN:PASS \
  MALWARE_PROTECTION:PASS >/dev/null

seed_device "${PROFILE_IDS[7]}" "alex-devbox" "LINUX" "Debian 12" "PF9ZYXWV" \
  DISK_ENCRYPTION:FAIL SCREEN_LOCK:PASS FIREWALL_ENABLED:PASS TIME_SYNC:PASS \
  OS_VERSION:UNKNOWN AUTO_UPDATE:PASS PASSWORD_POLICY:PASS REMOTE_LOGIN:PASS \
  MALWARE_PROTECTION:NOT_APPLICABLE >/dev/null

seed_device "${PROFILE_IDS[3]}" "david-surface" "WINDOWS" "Windows 11 23H2" "5CD1234ABC" \
  DISK_ENCRYPTION:PASS SCREEN_LOCK:PASS FIREWALL_ENABLED:PASS TIME_SYNC:FAIL \
  OS_VERSION:PASS AUTO_UPDATE:PASS PASSWORD_POLICY:PASS REMOTE_LOGIN:PASS \
  MALWARE_PROTECTION:PASS >/dev/null

seed_device "${PROFILE_IDS[2]}" "sofia-latitude" "WINDOWS" "Windows 11 22H2" "5CD9876ZYX" \
  DISK_ENCRYPTION:PASS SCREEN_LOCK:PASS FIREWALL_ENABLED:FAIL TIME_SYNC:PASS \
  OS_VERSION:FAIL AUTO_UPDATE:FAIL PASSWORD_POLICY:PASS REMOTE_LOGIN:PASS \
  MALWARE_PROTECTION:UNKNOWN >/dev/null

# 1 pending device: created and assigned, but never enrolled/activated.
create_device "${PROFILE_IDS[6]}" >/dev/null

# 1 revoked device: fully activated, then revoked.
revoked_id=$(seed_device "${PROFILE_IDS[5]}" "james-old-macbook" "DARWIN" "12.7" "C02OLD1JGH9" \
  DISK_ENCRYPTION:PASS SCREEN_LOCK:PASS FIREWALL_ENABLED:PASS TIME_SYNC:PASS \
  OS_VERSION:FAIL AUTO_UPDATE:FAIL PASSWORD_POLICY:PASS REMOTE_LOGIN:PASS \
  MALWARE_PROTECTION:PASS)
revoke_device "$revoked_id"

echo "    8 devices created (6 active, 1 pending, 1 revoked)"

echo ""
echo "Seed complete!"
echo "  Email:    $EMAIL"
echo "  Password: $PASSWORD"
echo "  Org:      $ORG_NAME"
echo ""
echo "  Created:"
echo "    3 frameworks, 43 controls"
echo "    35 risks"
echo "    20 third parties"
echo "    15 measures"
echo "    8 people"
echo "    8 devices (6 active, 1 pending, 1 revoked)"
echo ""
echo "  To use the CLI:"
echo "    export PROBO_HOST=$BASE_URL"
echo "    export PROBO_TOKEN=$PROBO_TOKEN"
