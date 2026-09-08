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

import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Card,
  Field,
  IconSquareBehindSquare2,
  Input,
  Option,
  PageHeader,
  Select,
  useToast,
} from "@probo/ui";
import { type ChangeEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link, useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type {
  AzureEnvironment,
  CreateAzureAccessReviewSourcePageCreateMutation,
} from "#/__generated__/core/CreateAzureAccessReviewSourcePageCreateMutation.graphql";
import type { CreateAzureAccessReviewSourcePageDeleteMutation } from "#/__generated__/core/CreateAzureAccessReviewSourcePageDeleteMutation.graphql";
import type { CreateAzureAccessReviewSourcePageQuery } from "#/__generated__/core/CreateAzureAccessReviewSourcePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import {
  ActionSplitButton,
  type ActionSplitButtonAction,
} from "../_components/ActionSplitButton";
import { ConnectorDocumentationLink } from "../dialogs/_components/ConnectorDocumentationLink";
import {
  azureAccessReviewSourceName,
  isAzureGUID,
} from "../dialogs/_lib/connectorSettings";
import { createAccessReviewSourceMutation, prependCreatedSourceEdge } from "../dialogs/accessReviewSourceMutations";

const azureEnvironments = [
  "AZURE_PUBLIC",
  "AZURE_GOVERNMENT",
  "AZURE_GOVERNMENT_DOD",
  "AZURE_CHINA",
] as const satisfies ReadonlyArray<AzureEnvironment>;

export const createAzureAccessReviewSourcePageQuery = graphql`
  query CreateAzureAccessReviewSourcePageQuery($organizationId: ID!) {
    azureConnectorSetup(organizationId: $organizationId) {
      issuer
      audience
      subject
      terraformSnippet
    }
    accessReviewDrivers {
      provider
      displayName
      documentationUrl
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateSource: permission(action: "access-review:source:create")
      }
    }
  }
`;

const createWorkloadIdentityConnectorMutation = graphql`
  mutation CreateAzureAccessReviewSourcePageCreateMutation(
    $input: CreateWorkloadIdentityConnectorInput!
  ) {
    createWorkloadIdentityConnector(input: $input) {
      connector {
        id
        provider
        connectionStatus
      }
    }
  }
`;

const deleteConnectorMutation = graphql`
  mutation CreateAzureAccessReviewSourcePageDeleteMutation(
    $input: DeleteConnectorInput!
  ) {
    deleteConnector(input: $input) {
      deletedConnectorId
    }
  }
`;

interface CreateAzureAccessReviewSourcePageProps {
  queryRef: PreloadedQuery<CreateAzureAccessReviewSourcePageQuery>;
}

export function CreateAzureAccessReviewSourcePage({
  queryRef,
}: CreateAzureAccessReviewSourcePageProps) {
  const { t } = useTranslation("organizations/access-reviews");
  const { toast } = useToast();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const [tenantId, setTenantId] = useState("");
  const [clientId, setClientId] = useState("");
  const [subscriptionId, setSubscriptionId] = useState("");
  const [environment, setEnvironment] = useState<AzureEnvironment>("AZURE_PUBLIC");
  const [isCreating, setIsCreating] = useState(false);

  usePageTitle(t("createAzureAccessReviewSourcePage.pageTitle"));

  const { organization, azureConnectorSetup, accessReviewDrivers }
    = usePreloadedQuery<CreateAzureAccessReviewSourcePageQuery>(
      createAzureAccessReviewSourcePageQuery,
      queryRef,
    );
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const azureDriver = accessReviewDrivers.find(
    driver => driver.provider === "AZURE",
  );
  if (!azureDriver) {
    throw new Error("Azure access review driver not found");
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organization.id,
    "AccessReviewConnectionsPage_accessReviewSources",
  );

  const [createWorkloadIdentityConnector] = useMutation<
    CreateAzureAccessReviewSourcePageCreateMutation
  >(createWorkloadIdentityConnectorMutation);
  const [deleteConnector] = useMutation<
    CreateAzureAccessReviewSourcePageDeleteMutation
  >(deleteConnectorMutation);
  const [createAccessReviewSource] = useMutation<
    accessReviewSourceMutationsCreateMutation
  >(createAccessReviewSourceMutation);

  if (!organization.canCreateSource) {
    return (
      <Card padded>
        <p className="text-txt-secondary text-sm">
          {t("createAzureAccessReviewSourcePage.permissionDenied")}
        </p>
      </Card>
    );
  }

  const copyValue = (value: string, successKey: string) => {
    const onCopyFailure = () =>
      toast({
        title: t("createAzureAccessReviewSourcePage.messages.copyFailed"),
        description: t("createAzureAccessReviewSourcePage.errors.copy"),
        variant: "error",
      });

    if (!navigator.clipboard?.writeText) {
      onCopyFailure();
      return;
    }

    try {
      navigator.clipboard.writeText(value).then(
        () =>
          toast({
            title: t("createAzureAccessReviewSourcePage.messages.copied"),
            description: t(successKey),
            variant: "success",
          }),
        onCopyFailure,
      );
    } catch {
      onCopyFailure();
    }
  };

  const tenantValid = isAzureGUID(tenantId);
  const tenantInvalid = tenantId.trim() !== "" && !tenantValid;
  const clientValid = isAzureGUID(clientId);
  const clientInvalid = clientId.trim() !== "" && !clientValid;
  const subscriptionValid = isAzureGUID(subscriptionId);
  const subscriptionInvalid = subscriptionId.trim() !== "" && !subscriptionValid;
  const formValid = tenantValid && clientValid && subscriptionValid;

  const onSubmit = async () => {
    if (!formValid) {
      return;
    }

    setIsCreating(true);

    try {
      const created = await createWorkloadIdentityConnector(
        {
          variables: {
            input: {
              organizationId,
              provider: "AZURE",
              azureTenantId: tenantId.trim(),
              azureClientId: clientId.trim(),
              azureSubscriptionId: subscriptionId.trim(),
              azureEnvironment: environment,
            },
          },
        },
        { errorToast: t("createAzureAccessReviewSourcePage.errors.create") },
      );
      const { id: connectorId, connectionStatus }
        = created.createWorkloadIdentityConnector.connector;

      const discardConnector = () =>
        deleteConnector(
          { variables: { input: { connectorId } } },
          { errorToast: t("createAzureAccessReviewSourcePage.errors.delete") },
        );

      if (connectionStatus !== "CONNECTED") {
        toast({
          title: t("createAzureAccessReviewSourcePage.messages.error"),
          description: t(
            "createAzureAccessReviewSourcePage.errors.disconnected",
          ),
          variant: "error",
        });
        await discardConnector();
        return;
      }

      try {
        await createAccessReviewSource(
          {
            variables: {
              input: {
                organizationId,
                connectorId,
                name: azureAccessReviewSourceName(
                  azureDriver.displayName,
                  subscriptionId,
                ),
                csvData: null,
              },
            },
            updater: (store) => {
              if (connectionId) {
                prependCreatedSourceEdge(store, connectionId);
              }
            },
          },
          { errorToast: t("createAzureAccessReviewSourcePage.errors.source") },
        );
      } catch {
        await discardConnector();
        return;
      }

      toast({
        title: t("createAzureAccessReviewSourcePage.messages.success"),
        description: t("createAzureAccessReviewSourcePage.messages.created"),
        variant: "success",
      });
      void navigate(
        [
          "",
          "organizations",
          organizationId,
          "access-reviews",
          "connections",
        ].join("/"),
      );
    } catch {
      return;
    } finally {
      setIsCreating(false);
    }
  };

  const installActions: ActionSplitButtonAction[] = [];
  if (azureConnectorSetup.terraformSnippet) {
    installActions.push({
      id: "terraform",
      label: t("createAzureAccessReviewSourcePage.actions.installViaTerraform"),
      onSelect: () =>
        copyValue(
          azureConnectorSetup.terraformSnippet,
          "createAzureAccessReviewSourcePage.messages.copiedTerraform",
        ),
    });
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("createAzureAccessReviewSourcePage.title")}
        description={t("createAzureAccessReviewSourcePage.description")}
      >
        {installActions.length > 0 && (
          <ActionSplitButton
            actions={installActions}
            chooseAnotherMethodLabel={t(
              "createAzureAccessReviewSourcePage.actions.chooseAnotherInstallMethod",
            )}
          />
        )}
      </PageHeader>

      <Card padded>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void onSubmit();
          }}
          className="space-y-4"
        >
          <Field
            name="tenantId"
            label={t("createAzureAccessReviewSourcePage.fields.tenantId")}
            value={tenantId}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              setTenantId(e.target.value)}
            required
            placeholder={t(
              "createAzureAccessReviewSourcePage.fields.tenantIdPlaceholder",
            )}
            error={
              tenantInvalid
                ? t("createAzureAccessReviewSourcePage.errors.tenantId")
                : undefined
            }
          />
          <Field
            name="clientId"
            label={t("createAzureAccessReviewSourcePage.fields.clientId")}
            value={clientId}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              setClientId(e.target.value)}
            required
            placeholder={t(
              "createAzureAccessReviewSourcePage.fields.clientIdPlaceholder",
            )}
            error={
              clientInvalid
                ? t("createAzureAccessReviewSourcePage.errors.clientId")
                : undefined
            }
          />
          <Field
            name="subscriptionId"
            label={t("createAzureAccessReviewSourcePage.fields.subscriptionId")}
            value={subscriptionId}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              setSubscriptionId(e.target.value)}
            required
            placeholder={t(
              "createAzureAccessReviewSourcePage.fields.subscriptionIdPlaceholder",
            )}
            error={
              subscriptionInvalid
                ? t("createAzureAccessReviewSourcePage.errors.subscriptionId")
                : undefined
            }
          />
          <Field
            name="environment"
            label={t("createAzureAccessReviewSourcePage.fields.environment")}
            help={t("createAzureAccessReviewSourcePage.fields.environmentGccNote")}
          >
            <Select
              id="environment"
              value={environment}
              onValueChange={(value) => {
                if ((azureEnvironments as ReadonlyArray<string>).includes(value)) {
                  setEnvironment(value as AzureEnvironment);
                }
              }}
            >
              {azureEnvironments.map(value => (
                <Option key={value} value={value}>
                  {t(`createAzureAccessReviewSourcePage.environments.${value}`)}
                </Option>
              ))}
            </Select>
          </Field>
          <p className="text-txt-secondary text-sm">
            {t("createAzureAccessReviewSourcePage.propagationHint")}
          </p>
          {(
            [
              {
                name: "issuer",
                value: azureConnectorSetup.issuer,
                successKey:
                  "createAzureAccessReviewSourcePage.messages.copiedIssuer",
              },
              {
                name: "audience",
                value: azureConnectorSetup.audience,
                successKey:
                  "createAzureAccessReviewSourcePage.messages.copiedAudience",
              },
              {
                name: "subject",
                value: azureConnectorSetup.subject,
                successKey:
                  "createAzureAccessReviewSourcePage.messages.copiedSubject",
              },
            ] as const
          ).map(row => (
            <Field
              key={row.name}
              name={row.name}
              label={t(`createAzureAccessReviewSourcePage.fields.${row.name}`)}
            >
              <div className="flex items-center gap-2">
                <div className="min-w-0 flex-1">
                  <Input
                    id={row.name}
                    name={row.name}
                    value={row.value}
                    readOnly
                    disabled
                  />
                </div>
                <Button
                  type="button"
                  variant="secondary"
                  icon={IconSquareBehindSquare2}
                  onClick={() => copyValue(row.value, row.successKey)}
                  aria-label={t("createAzureAccessReviewSourcePage.actions.copy")}
                />
              </div>
            </Field>
          ))}

          <div className="flex items-center justify-between gap-2">
            <ConnectorDocumentationLink
              url={azureDriver.documentationUrl}
              variant="button"
            />
            <div className="flex items-center justify-end gap-2">
              <Button variant="secondary" asChild>
                <Link
                  to={[
                    "",
                    "organizations",
                    organizationId,
                    "access-reviews",
                    "connections",
                  ].join("/")}
                >
                  {t("createAzureAccessReviewSourcePage.actions.back")}
                </Link>
              </Button>
              <Button disabled={!formValid || isCreating} type="submit">
                {t("createAzureAccessReviewSourcePage.actions.connect")}
              </Button>
            </div>
          </div>
        </form>
      </Card>
    </div>
  );
}
