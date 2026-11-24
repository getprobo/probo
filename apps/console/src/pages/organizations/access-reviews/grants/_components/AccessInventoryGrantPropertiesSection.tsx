// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { dateTimeFormat } from "@probo/i18n";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessInventoryGrantPropertiesSection_grant$key } from "#/__generated__/core/AccessInventoryGrantPropertiesSection_grant.graphql";

import { accessReviewSourceLabel } from "../../_lib/sourceLabel";

import { accessInventoryGrantPropertiesSection } from "./variants";

const accessInventoryGrantPropertiesSectionFragment = graphql`
  fragment AccessInventoryGrantPropertiesSection_grant on AccessInventoryGrant {
    entitlementDisplayName
    entitlementKey
    presence
    consecutiveMissingSyncs
    firstSeenAt
    lastSeenAt
    providerChangedAt
    principal {
      email
      fullName
      principalKey
    }
    resource {
      displayName
      kind
      resourceKey
    }
    source {
      name
      connectorAccount {
        name
      }
    }
  }
`;

interface AccessInventoryGrantPropertiesSectionProps {
  grantKey: AccessInventoryGrantPropertiesSection_grant$key;
}

export function AccessInventoryGrantPropertiesSection({
  grantKey,
}: AccessInventoryGrantPropertiesSectionProps) {
  const { i18n, t } = useTranslation("organizations/access-reviews");
  const grant = useFragment(
    accessInventoryGrantPropertiesSectionFragment,
    grantKey,
  );
  const { root, intro, fields, stack } = accessInventoryGrantPropertiesSection();

  const fullName = grant.principal.fullName.trim();
  const email = grant.principal.email.trim();
  const principalTitle = fullName || email || t("grants.values.unknown");
  const entitlement = grant.entitlementDisplayName.trim()
    || grant.entitlementKey
    || t("grants.values.unknown");
  const resourceName = grant.resource.displayName.trim()
    || grant.resource.resourceKey
    || t("grants.values.unknown");
  const sourceLabel = accessReviewSourceLabel({
    name: grant.source.name,
    accountName: grant.source.connectorAccount?.name,
  });
  const disappeared = grant.presence === "DISAPPEARED";

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("grantDetail.properties.title")}
        </Heading>
      </div>
      <Card variant="soft" size={2}>
        <div className={fields()}>
          <Field label={t("grantDetail.fields.principal")}>
            <div className={stack()}>
              <Text size={2} highContrast>
                {principalTitle}
              </Text>
              {fullName !== "" && email !== "" && (
                <Text size={1} color="faint">
                  {email}
                </Text>
              )}
              <Text size={1} color="faint">
                {grant.principal.principalKey}
              </Text>
            </div>
          </Field>
          <Field label={t("grantDetail.fields.resource")}>
            <div className={stack()}>
              <Text size={2} highContrast>
                {resourceName}
              </Text>
              {grant.resource.kind.trim() !== "" && (
                <Text size={1} color="faint">
                  {grant.resource.kind}
                </Text>
              )}
              <Text size={1} color="faint">
                {grant.resource.resourceKey}
              </Text>
            </div>
          </Field>
          <Field label={t("grantDetail.fields.entitlement")}>
            <div className={stack()}>
              <Text size={2} highContrast>
                {entitlement}
              </Text>
              <Text size={1} color="faint">
                {grant.entitlementKey}
              </Text>
            </div>
          </Field>
          <Field label={t("grantDetail.fields.source")}>
            <Text size={2} highContrast>
              {sourceLabel}
            </Text>
          </Field>
          <Field label={t("grantDetail.fields.presence")}>
            <Text size={2} highContrast>
              {t(
                disappeared
                  ? "grants.presence.disappeared"
                  : "grants.presence.present",
              )}
            </Text>
          </Field>
          <Field label={t("grantDetail.fields.missingSyncs")}>
            <Text size={2} highContrast>
              {grant.consecutiveMissingSyncs}
            </Text>
          </Field>
          <Field label={t("grantDetail.fields.firstSeen")}>
            <Text size={2} highContrast>
              {dateTimeFormat(i18n.language, grant.firstSeenAt)}
            </Text>
          </Field>
          <Field label={t("grantDetail.fields.lastSeen")}>
            <Text size={2} highContrast>
              {dateTimeFormat(i18n.language, grant.lastSeenAt)}
            </Text>
          </Field>
          <Field label={t("grantDetail.fields.providerChanged")}>
            {grant.providerChangedAt == null
              ? (
                  <Text size={2} color="faint">
                    {t("grantDetail.values.empty")}
                  </Text>
                )
              : (
                  <Text size={2} highContrast>
                    {dateTimeFormat(i18n.language, grant.providerChangedAt)}
                  </Text>
                )}
          </Field>
        </div>
      </Card>
    </section>
  );
}
