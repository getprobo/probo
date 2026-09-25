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

import { CopyIcon, PencilSimpleIcon } from "@phosphor-icons/react";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Trans, useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SAMLConfigurationListItem_samlConfiguration$key } from "#/__generated__/iam/SAMLConfigurationListItem_samlConfiguration.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";

import {
  samlConfigurationCardTone,
  SAMLConfigurationStatusIcon,
  showsSamlLoginUrl,
} from "../_lib/samlConfigurationCardTone";
import { samlConfigurationListItem } from "../variants";

import { DeleteSAMLConfigurationDialog } from "./DeleteSAMLConfigurationDialog";

const samlConfigurationListItemFragment = graphql`
  fragment SAMLConfigurationListItem_samlConfiguration on SAMLConfiguration {
    id
    emailDomain
    enforcementPolicy
    domainVerificationToken
    domainVerifiedAt
    testLoginUrl
    canUpdate: permission(action: "iam:saml-configuration:update")
    canDelete: permission(action: "iam:saml-configuration:delete")
    ...DeleteSAMLConfigurationDialog_samlConfiguration
  }
`;

interface SAMLConfigurationListItemProps {
  samlConfigurationKey: SAMLConfigurationListItem_samlConfiguration$key;
  onEdit: (id: string) => void;
  onVerifyDomain: (dnsVerificationToken: string) => void;
  onDeleted: () => void;
}

export function SAMLConfigurationListItem({
  samlConfigurationKey,
  onEdit,
  onVerifyDomain,
  onDeleted,
}: SAMLConfigurationListItemProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const { actions, body, callout, url, urlRow, urlValue } = samlConfigurationListItem();
  const config = useFragment(samlConfigurationListItemFragment, samlConfigurationKey);
  const tone = samlConfigurationCardTone(config.domainVerifiedAt, config.enforcementPolicy);
  const showLoginUrl = showsSamlLoginUrl(config.domainVerifiedAt, config.enforcementPolicy);
  const domainVerificationToken = config.domainVerificationToken;
  const domainVerified = config.domainVerifiedAt != null;
  const hasActions = (domainVerified && config.canUpdate) || config.canDelete;

  async function handleCopyUrl() {
    try {
      await navigator.clipboard.writeText(config.testLoginUrl);
      toast({
        title: t("samlConfigurationList.messages.copied"),
        description: t("samlConfigurationList.fields.ssoUrl"),
        variant: "success",
      });
    } catch {
      toast({
        title: t("samlConfigurationList.errors.copy"),
        description: t("samlConfigurationList.fields.ssoUrl"),
        variant: "error",
      });
    }
  }

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
        <Text size={3} weight="medium" highContrast>
          {config.emailDomain}
        </Text>
      )}
      control={hasActions
        ? (
            <div className={actions()}>
              {domainVerified && config.canUpdate && (
                <IconButton
                  size={1}
                  variant="surface"
                  color="neutral"
                  aria-label={t("samlConfigurationList.actions.edit")}
                  onClick={() => onEdit(config.id)}
                >
                  <PencilSimpleIcon />
                </IconButton>
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
        <Text size={2} color="neutral">
          <Trans
            i18nKey={`samlConfigurationList.enforcement.${config.enforcementPolicy.toLowerCase()}`}
            components={{
              policy: <Text size={2} weight="medium" highContrast color="current" />,
            }}
          />
        </Text>
        {config.domainVerifiedAt == null && (
          <div className={callout()}>
            <Text size={2} color="neutral">
              {t("samlConfigurationList.pending.description")}
            </Text>
            {config.canUpdate && domainVerificationToken != null && (
              <Button
                variant="solid"
                size={2}
                onClick={() => onVerifyDomain(domainVerificationToken)}
              >
                {t("samlConfigurationList.actions.verifyDomain")}
              </Button>
            )}
          </div>
        )}
        {showLoginUrl && (
          <div className={url()}>
            <Text size={1} color="faint">
              {t("samlConfigurationList.fields.ssoUrl")}
            </Text>
            <div className={urlRow()}>
              <Code variant="ghost" size={1} className={urlValue()}>
                {config.testLoginUrl}
              </Code>
              <IconButton
                size={1}
                variant="ghost"
                color="neutral"
                aria-label={t("samlConfigurationList.actions.copyUrl")}
                onClick={() => {
                  void handleCopyUrl();
                }}
              >
                <CopyIcon />
              </IconButton>
            </div>
          </div>
        )}
      </div>
    </TonedCard>
  );
}
