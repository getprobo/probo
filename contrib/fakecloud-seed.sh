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

# Seed a fresh fakecloud with one organization of three accounts and a
# ProboAudit role in each. Run it, do not source it:
#
#   ./contrib/fakecloud-seed.sh http://localhost:8080/federation/<organization-id>
#
# The issuer is the URL on the AWS connect page. The subject defaults to
# its last path segment. fakecloud keeps this state in memory, so run it
# again after make stack-down.

set -euo pipefail

ENDPOINT="${FAKECLOUD_ENDPOINT:-http://127.0.0.1:4566}"
MANAGEMENT_ACCOUNT="123456789012"
ROLE_NAME="ProboAudit"
AUDIENCE="sts.amazonaws.com"
THUMBPRINT="0000000000000000000000000000000000000000"

usage() {
	printf 'usage: %s <issuer-url> [subject]\n' "$0" >&2
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
	usage
	exit 1
fi

ISSUER="${1%/}"
if [[ ${ISSUER} != http://* && ${ISSUER} != https://* ]]; then
	printf 'issuer must be an absolute http or https URL\n' >&2
	exit 1
fi

ISSUER_HOST="${ISSUER#http://}"
ISSUER_HOST="${ISSUER_HOST#https://}"
SUBJECT="${2:-${ISSUER_HOST##*/}}"
PROVIDER_ARN="arn:aws:iam::${MANAGEMENT_ACCOUNT}:oidc-provider/${ISSUER_HOST}"

export AWS_ACCESS_KEY_ID="test"
export AWS_SECRET_ACCESS_KEY="test"
export AWS_SESSION_TOKEN=""
export AWS_EC2_METADATA_DISABLED="true"
export AWS_ENDPOINT_URL="${ENDPOINT}"
export AWS_DEFAULT_REGION="us-east-1"
export AWS_PAGER=""
export AWS_CONFIG_FILE="/dev/null"
export AWS_SHARED_CREDENTIALS_FILE="/dev/null"
unset AWS_PROFILE || true

aws() {
	command aws --endpoint-url "${ENDPOINT}" "$@"
}

use_management_credentials() {
	export AWS_ACCESS_KEY_ID="test"
	export AWS_SECRET_ACCESS_KEY="test"
	export AWS_SESSION_TOKEN=""
}

trust_policy() {
	cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"Federated": "${PROVIDER_ARN}"},
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "${ISSUER_HOST}:aud": "${AUDIENCE}",
        "${ISSUER_HOST}:sub": "${SUBJECT}"
      }
    }
  }]
}
EOF
}

create_audit_role() {
	aws iam create-role \
		--role-name "${ROLE_NAME}" \
		--assume-role-policy-document "$(trust_policy)" \
		>/dev/null
}

wait_for_account() {
	local request_id="$1"
	local attempt state

	for attempt in $(seq 1 30); do
		state="$(aws organizations describe-create-account-status \
			--create-account-request-id "${request_id}" \
			--query 'CreateAccountStatus.State' \
			--output text)"
		if [[ ${state} == "SUCCEEDED" ]]; then
			aws organizations describe-create-account-status \
				--create-account-request-id "${request_id}" \
				--query 'CreateAccountStatus.AccountId' \
				--output text
			return
		fi
		if [[ ${state} == "FAILED" ]]; then
			printf 'creating account failed for request %s\n' "${request_id}" >&2
			exit 1
		fi
		sleep 1
	done

	printf 'timed out waiting for account request %s\n' "${request_id}" >&2
	exit 1
}

create_member() {
	local name="$1"
	local email="$2"
	local request_id account_id creds

	request_id="$(aws organizations create-account \
		--email "${email}" \
		--account-name "${name}" \
		--query 'CreateAccountStatus.Id' \
		--output text)"
	account_id="$(wait_for_account "${request_id}")"

	creds="$(curl -fsS -X POST "${ENDPOINT}/_fakecloud/iam/create-admin" \
		-H 'content-type: application/json' \
		-d "$(printf '{"accountId":"%s","userName":"admin"}' "${account_id}")")"
	export AWS_ACCESS_KEY_ID="$(printf '%s' "${creds}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["accessKeyId"])')"
	export AWS_SECRET_ACCESS_KEY="$(printf '%s' "${creds}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["secretAccessKey"])')"
	export AWS_SESSION_TOKEN=""

	create_audit_role
	use_management_credentials
	printf '%s\n' "arn:aws:iam::${account_id}:role/${ROLE_NAME}"
}

use_management_credentials

aws organizations create-organization --feature-set ALL >/dev/null

aws iam create-open-id-connect-provider \
	--url "${ISSUER}" \
	--client-id-list "${AUDIENCE}" \
	--thumbprint-list "${THUMBPRINT}" \
	>/dev/null

create_audit_role

MEMBER_ONE="$(create_member "Member One" "member-one@example.com")"
MEMBER_TWO="$(create_member "Member Two" "member-two@example.com")"

printf '\nPaste this role ARN into the connector:\n'
printf '  arn:aws:iam::%s:role/%s\n' "${MANAGEMENT_ACCOUNT}" "${ROLE_NAME}"
printf '\nDiscovered accounts also have:\n'
printf '  %s\n' "${MEMBER_ONE}" "${MEMBER_TWO}"
