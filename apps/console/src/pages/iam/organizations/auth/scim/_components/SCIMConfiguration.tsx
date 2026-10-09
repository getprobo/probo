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

import { ArrowsClockwiseIcon, CopyIcon, PlugsConnectedIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SCIMConfiguration_organization$key } from "#/__generated__/iam/SCIMConfiguration_organization.graphql";
import type { SCIMConfiguration_regenerateMutation } from "#/__generated__/iam/SCIMConfiguration_regenerateMutation.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useMutation } from "#/lib/relay/useMutation";

import { scimConfiguration } from "../variants";

import { DeleteSCIMConfigurationDialog } from "./DeleteSCIMConfigurationDialog";

const fragment = graphql`
  fragment SCIMConfiguration_organization on Organization {
    id
    scimConfiguration {
      id
      endpointUrl
      createdAt
      canUpdate: permission(action: "iam:scim-configuration:update")
      canDelete: permission(action: "iam:scim-configuration:delete")
      ...DeleteSCIMConfigurationDialog_scimConfiguration
    }
  }
`;

const regenerateMutation = graphql`
  mutation SCIMConfiguration_regenerateMutation(
    $input: RegenerateSCIMTokenInput!
  ) {
    regenerateSCIMToken(input: $input) {
      scimConfiguration {
        id
        endpointUrl
      }
      token
    }
  }
`;

export interface SCIMConfigurationProps {
  organizationKey: SCIMConfiguration_organization$key;
  initialToken?: string | null;
  onDeleted?: () => void;
}

export function SCIMConfiguration({
  organizationKey,
  initialToken,
  onDeleted,
}: SCIMConfigurationProps) {
  const { t, i18n } = useTranslation();
  const toast = useToast();
  const organization = useFragment(fragment, organizationKey);
  const configuration = organization.scimConfiguration;
  const [token, setToken] = useState<string | null>(null);
  const visibleToken = token ?? initialToken ?? null;
  const {
    root,
    fields,
    field,
    fieldValue,
    code,
    actions,
  } = scimConfiguration();

  const [regenerateSCIMToken, isRegenerating]
    = useMutation<SCIMConfiguration_regenerateMutation>(
      regenerateMutation,
      {
        successMessage: t("scimConfiguration.messages.regenerated"),
        errorToast: t("scimConfiguration.errors.regenerate"),
      },
    );

  if (configuration == null) {
    return null;
  }

  const {
    id: configurationId,
    endpointUrl,
    createdAt,
    canUpdate,
    canDelete,
  } = configuration;

  async function copyToClipboard(text: string, description: string) {
    try {
      await navigator.clipboard.writeText(text);
      toast.add({
        title: t("scimConfiguration.messages.copied"),
        description,
        type: "success",
      });
    } catch {
      toast.add({
        title: t("scimConfiguration.errors.copy"),
        description,
        type: "error",
      });
    }
  }

  function handleRegenerate() {
    void regenerateSCIMToken({
      variables: {
        input: {
          organizationId: organization.id,
          scimConfigurationId: configurationId,
        },
      },
    }).then((response) => {
      const payload = response.regenerateSCIMToken;
      if (payload == null || payload.token === "") {
        return;
      }
      setToken(payload.token);
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  function handleDeleted() {
    setToken(null);
    onDeleted?.();
  }

  return (
    <TonedCard
      tone="green"
      icon={<PlugsConnectedIcon />}
      lead={(
        <Text size={3} weight="medium" color="green">
          {t("scimPage.setup.manual.title")}
        </Text>
      )}
    >
      <div className={root()}>
        <Text size={2} color="neutral">
          {t("scimConfiguration.configuredOn", {
            date: dateTimeFormat(i18n.language, createdAt),
          })}
        </Text>
        <div className={fields()}>
          <div className={field()}>
            <Text size={1} color="faint">
              {t("scimConfiguration.fields.endpointUrl")}
            </Text>
            <div className={fieldValue()}>
              <Code size={1} className={code()}>
                {endpointUrl}
              </Code>
              <IconButton
                size={1}
                variant="soft"
                color="neutral"
                aria-label={t("scimConfiguration.actions.copyEndpointUrl")}
                onClick={() => {
                  void copyToClipboard(
                    endpointUrl,
                    t("scimConfiguration.fields.endpointUrl"),
                  );
                }}
              >
                <CopyIcon />
              </IconButton>
            </div>
          </div>
          {visibleToken != null && (
            <>
              <div className={field()}>
                <Text size={1} color="faint">
                  {t("scimConfiguration.fields.bearerToken")}
                </Text>
                <div className={fieldValue()}>
                  <Code size={1} className={code()}>
                    {visibleToken}
                  </Code>
                  <IconButton
                    size={1}
                    variant="soft"
                    color="neutral"
                    aria-label={t("scimConfiguration.actions.copyBearerToken")}
                    onClick={() => {
                      void copyToClipboard(
                        visibleToken,
                        t("scimConfiguration.fields.bearerToken"),
                      );
                    }}
                  >
                    <CopyIcon />
                  </IconButton>
                </div>
              </div>
              <Callout color="amber">
                {t("scimConfiguration.tokenWarning")}
              </Callout>
            </>
          )}
        </div>
        {(canUpdate || canDelete) && (
          <div className={actions()}>
            {canUpdate && (
              <Button
                variant="soft"
                color="neutral"
                loading={isRegenerating}
                iconStart={<ArrowsClockwiseIcon />}
                onClick={handleRegenerate}
              >
                {t("scimConfiguration.actions.regenerateToken")}
              </Button>
            )}
            {canDelete && (
              <DeleteSCIMConfigurationDialog
                scimConfigurationKey={configuration}
                onDeleted={handleDeleted}
              />
            )}
          </div>
        )}
      </div>
    </TonedCard>
  );
}
