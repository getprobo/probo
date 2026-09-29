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

import { PlugsConnectedIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { GoogleLogo } from "@probo/ui/src/v2/GoogleLogo/GoogleLogo";
import { MicrosoftLogo } from "@probo/ui/src/v2/MicrosoftLogo/MicrosoftLogo";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SCIMSetupCards_createMutation } from "#/__generated__/iam/SCIMSetupCards_createMutation.graphql";
import type { SCIMSetupCards_organization$key } from "#/__generated__/iam/SCIMSetupCards_organization.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { initiateScimConnectorUrl } from "../_lib/initiateScimConnector";
import { scimPage, scimSetupCard } from "../variants";

import { SCIMSetupCard } from "./SCIMSetupCard";

const scimSetupCardsFragment = graphql`
  fragment SCIMSetupCards_organization on Organization {
    id
    canCreateSCIMConfiguration: permission(
      action: "iam:scim-configuration:create"
    )
    scimBridgeTypes {
      type
      oauth2Scopes
    }
  }
`;

const createSCIMConfigurationMutation = graphql`
  mutation SCIMSetupCards_createMutation($input: CreateSCIMConfigurationInput!) {
    createSCIMConfiguration(input: $input) {
      scimConfiguration {
        organization {
          id
          ...SCIMConfiguration_organization
          scimConfiguration {
            id
            ...SCIMEventList_scimConfiguration
            ...SCIMProviderCard_scimConfiguration
          }
        }
      }
      token
    }
  }
`;

export interface SCIMSetupCardsProps {
  organizationKey: SCIMSetupCards_organization$key;
  onManualCreated: (token: string) => void;
}

export function SCIMSetupCards({
  organizationKey,
  onManualCreated,
}: SCIMSetupCardsProps) {
  const organization = useFragment(scimSetupCardsFragment, organizationKey);
  const { t } = useTranslation();
  const { grid } = scimPage();
  const canCreate = organization.canCreateSCIMConfiguration;

  const [createSCIMConfiguration, isCreating]
    = useMutation<SCIMSetupCards_createMutation>(
      createSCIMConfigurationMutation,
      {
        successMessage: t("scimConfiguration.messages.copyToken"),
        errorToast: t("scimConfiguration.errors.create"),
      },
    );

  const googleWorkspaceScopes
    = organization.scimBridgeTypes.find(info => info.type === "GOOGLE_WORKSPACE")
      ?.oauth2Scopes ?? [];
  const microsoft365Scopes
    = organization.scimBridgeTypes.find(info => info.type === "MICROSOFT_365")
      ?.oauth2Scopes ?? [];

  function handleConnect(provider: "GOOGLE_WORKSPACE" | "MICROSOFT_365") {
    const scopes = provider === "GOOGLE_WORKSPACE"
      ? googleWorkspaceScopes
      : microsoft365Scopes;
    window.location.assign(initiateScimConnectorUrl({
      organizationId: organization.id,
      provider,
      scopes,
    }));
  }

  async function handleManualCreate() {
    const response = await createSCIMConfiguration({
      variables: {
        input: {
          organizationId: organization.id,
        },
      },
    });
    const payload = response.createSCIMConfiguration;
    if (payload == null || payload.token === "") {
      return;
    }
    onManualCreated(payload.token);
  }

  return (
    <div className={grid()}>
      <SCIMSetupCard
        icon={<GoogleLogo />}
        title={t("scimPage.setup.google.title")}
        description={t("scimPage.setup.google.description")}
        action={connectAction({
          canCreate,
          label: t("scimPage.setup.google.action"),
          onClick: () => {
            handleConnect("GOOGLE_WORKSPACE");
          },
        })}
      />
      <SCIMSetupCard
        icon={<MicrosoftLogo />}
        title={t("scimPage.setup.microsoft.title")}
        description={t("scimPage.setup.microsoft.description")}
        action={connectAction({
          canCreate,
          label: t("scimPage.setup.microsoft.action"),
          onClick: () => {
            handleConnect("MICROSOFT_365");
          },
        })}
      />
      <SCIMSetupCard
        icon={<PlugsConnectedIcon />}
        title={t("scimPage.setup.manual.title")}
        description={t("scimPage.setup.manual.description")}
        action={connectAction({
          canCreate,
          label: t("scimPage.setup.manual.action"),
          loading: isCreating,
          onClick: () => {
            void handleManualCreate();
          },
        })}
      />
    </div>
  );
}

function connectAction({
  canCreate,
  label,
  loading,
  onClick,
}: {
  canCreate: boolean;
  label: string;
  loading?: boolean;
  onClick: () => void;
}): ReactNode {
  if (!canCreate) {
    return undefined;
  }

  const { action } = scimSetupCard();

  return (
    <Button
      size={2}
      variant="solid"
      color="neutral"
      highContrast
      className={action()}
      loading={loading}
      onClick={onClick}
    >
      {label}
    </Button>
  );
}
