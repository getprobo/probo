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

import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Trans, useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SAMLConfigurationListItem_samlConfiguration$key } from "#/__generated__/iam/SAMLConfigurationListItem_samlConfiguration.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";

import {
  samlConfigurationCardTone,
  SAMLConfigurationStatusIcon,
  samlConfigurationStatusKey,
  showsSamlLoginUrl,
} from "../_lib/samlConfigurationCardTone";
import { samlConfigurationListItem } from "../variants";

import { DeleteSAMLConfigurationDialog } from "./DeleteSAMLConfigurationDialog";
import { SAMLConfigurationDnsRecord } from "./SAMLConfigurationDnsRecord";
import { SAMLConfigurationSsoUrl } from "./SAMLConfigurationSsoUrl";

const samlConfigurationListItemFragment = graphql`
  fragment SAMLConfigurationListItem_samlConfiguration on SAMLConfiguration {
    id
    emailDomain
    enforcementPolicy
    autoSignupEnabled
    domainVerificationToken
    domainVerifiedAt
    testLoginUrl
    canGet: permission(action: "iam:saml-configuration:get")
    canDelete: permission(action: "iam:saml-configuration:delete")
    ...DeleteSAMLConfigurationDialog_samlConfiguration
  }
`;

interface SAMLConfigurationListItemProps {
  samlConfigurationKey: SAMLConfigurationListItem_samlConfiguration$key;
  onDeleted: () => void;
}

export function SAMLConfigurationListItem({
  samlConfigurationKey,
  onDeleted,
}: SAMLConfigurationListItemProps) {
  const { t } = useTranslation();
  const { actions, body, callout } = samlConfigurationListItem();
  const config = useFragment(samlConfigurationListItemFragment, samlConfigurationKey);
  const tone = samlConfigurationCardTone(config.domainVerifiedAt, config.enforcementPolicy);
  const statusKey = samlConfigurationStatusKey(
    config.domainVerifiedAt,
    config.enforcementPolicy,
  );
  const showLoginUrl = showsSamlLoginUrl(config.domainVerifiedAt, config.enforcementPolicy);
  const domainVerificationToken = config.domainVerificationToken;
  const hasActions = config.canGet || config.canDelete;

  return (
    <TonedCard
      tone={tone}
      icon={(
        <SAMLConfigurationStatusIcon
          domainVerifiedAt={config.domainVerifiedAt}
          enforcementPolicy={config.enforcementPolicy}
        />
      )}
      lead={(
        <Text size={3} weight="medium" color={tone === "sand" ? "neutral" : tone}>
          {t(`samlConfigurationList.title.${statusKey}`)}
        </Text>
      )}
      control={hasActions
        ? (
            <div className={actions()}>
              {config.canGet && (
                <ButtonLink
                  to={config.id}
                  size={1}
                  variant="surface"
                  color="neutral"
                >
                  {t("samlConfigurationList.actions.open")}
                </ButtonLink>
              )}
              {config.canDelete && (
                <DeleteSAMLConfigurationDialog
                  samlConfigurationKey={config}
                  onDeleted={onDeleted}
                />
              )}
            </div>
          )
        : undefined}
    >
      <div className={body()}>
        <Text size={3} weight="medium" highContrast>
          {config.emailDomain}
        </Text>
        {config.domainVerifiedAt != null && (
          <>
            <Text size={2} color="neutral">
              <Trans
                i18nKey={`samlConfigurationList.enforcement.${config.enforcementPolicy.toLowerCase()}`}
                components={{
                  policy: <Text size={2} weight="medium" highContrast color="current" />,
                }}
              />
            </Text>
            <Text size={2} color="neutral">
              <Trans
                i18nKey={
                  config.autoSignupEnabled
                    ? "samlConfigurationList.signup.enabled"
                    : "samlConfigurationList.signup.disabled"
                }
                components={{
                  policy: <Text size={2} weight="medium" highContrast color="current" />,
                }}
              />
            </Text>
          </>
        )}
        {config.domainVerifiedAt == null && (
          <div className={callout()}>
            <Text size={2} color="neutral">
              <Trans
                i18nKey="samlConfigurationList.pending.description"
                components={{
                  type: <Text size={2} weight="medium" highContrast color="current" />,
                }}
              />
            </Text>
            {domainVerificationToken != null && (
              <SAMLConfigurationDnsRecord
                domainVerificationToken={domainVerificationToken}
              />
            )}
          </div>
        )}
        {showLoginUrl && (
          <SAMLConfigurationSsoUrl testLoginUrl={config.testLoginUrl} />
        )}
      </div>
    </TonedCard>
  );
}
