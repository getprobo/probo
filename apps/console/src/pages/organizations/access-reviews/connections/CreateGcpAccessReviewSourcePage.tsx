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
  PageHeader,
  useToast,
} from "@probo/ui";
import { type ChangeEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link, useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { CreateGcpAccessReviewSourcePageCreateMutation } from "#/__generated__/core/CreateGcpAccessReviewSourcePageCreateMutation.graphql";
import type { CreateGcpAccessReviewSourcePageDeleteMutation } from "#/__generated__/core/CreateGcpAccessReviewSourcePageDeleteMutation.graphql";
import type { CreateGcpAccessReviewSourcePageQuery } from "#/__generated__/core/CreateGcpAccessReviewSourcePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import {
  ActionSplitButton,
  type ActionSplitButtonAction,
} from "../_components/ActionSplitButton";
import { ConnectorDocumentationLink } from "../dialogs/_components/ConnectorDocumentationLink";
import {
  gcpAccessReviewSourceName,
  isGCPServiceAccountEmail,
  isGCPWorkloadIdentityProvider,
} from "../dialogs/_lib/connectorSettings";
import { createAccessReviewSourceMutation, prependCreatedSourceEdge } from "../dialogs/accessReviewSourceMutations";

export const createGcpAccessReviewSourcePageQuery = graphql`
  query CreateGcpAccessReviewSourcePageQuery($organizationId: ID!) {
    gcpConnectorSetup(organizationId: $organizationId) {
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
  mutation CreateGcpAccessReviewSourcePageCreateMutation(
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
  mutation CreateGcpAccessReviewSourcePageDeleteMutation(
    $input: DeleteConnectorInput!
  ) {
    deleteConnector(input: $input) {
      deletedConnectorId
    }
  }
`;

interface CreateGcpAccessReviewSourcePageProps {
  queryRef: PreloadedQuery<CreateGcpAccessReviewSourcePageQuery>;
}

export function CreateGcpAccessReviewSourcePage({
  queryRef,
}: CreateGcpAccessReviewSourcePageProps) {
  const { t } = useTranslation("organizations/access-reviews");
  const { toast } = useToast();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const [providerResource, setProviderResource] = useState("");
  const [serviceAccountEmail, setServiceAccountEmail] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  usePageTitle(t("createGcpAccessReviewSourcePage.pageTitle"));

  const { organization, gcpConnectorSetup, accessReviewDrivers }
    = usePreloadedQuery<CreateGcpAccessReviewSourcePageQuery>(
      createGcpAccessReviewSourcePageQuery,
      queryRef,
    );
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const gcpDriver = accessReviewDrivers.find(
    driver => driver.provider === "GCP",
  );
  if (!gcpDriver) {
    throw new Error("GCP access review driver not found");
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organization.id,
    "AccessReviewConnectionsPage_accessReviewSources",
  );

  const [createWorkloadIdentityConnector] = useMutation<
    CreateGcpAccessReviewSourcePageCreateMutation
  >(createWorkloadIdentityConnectorMutation);
  const [deleteConnector] = useMutation<
    CreateGcpAccessReviewSourcePageDeleteMutation
  >(deleteConnectorMutation);
  const [createAccessReviewSource] = useMutation<
    accessReviewSourceMutationsCreateMutation
  >(createAccessReviewSourceMutation);

  if (!organization.canCreateSource) {
    return (
      <Card padded>
        <p className="text-txt-secondary text-sm">
          {t("createGcpAccessReviewSourcePage.permissionDenied")}
        </p>
      </Card>
    );
  }

  const copyValue = (value: string, successKey: string) => {
    const onCopyFailure = () =>
      toast({
        title: t("createGcpAccessReviewSourcePage.messages.copyFailed"),
        description: t("createGcpAccessReviewSourcePage.errors.copy"),
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
            title: t("createGcpAccessReviewSourcePage.messages.copied"),
            description: t(successKey),
            variant: "success",
          }),
        onCopyFailure,
      );
    } catch {
      onCopyFailure();
    }
  };

  const providerValid = isGCPWorkloadIdentityProvider(providerResource);
  const providerInvalid = providerResource.trim() !== "" && !providerValid;
  const emailValid = isGCPServiceAccountEmail(serviceAccountEmail);
  const emailInvalid = serviceAccountEmail.trim() !== "" && !emailValid;
  const formValid = providerValid && emailValid;

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
              provider: "GCP",
              gcpWorkloadIdentityProvider: providerResource.trim(),
              gcpServiceAccountEmail: serviceAccountEmail.trim(),
            },
          },
        },
        { errorToast: t("createGcpAccessReviewSourcePage.errors.create") },
      );
      const { id: connectorId, connectionStatus }
        = created.createWorkloadIdentityConnector.connector;

      const discardConnector = () =>
        deleteConnector(
          { variables: { input: { connectorId } } },
          { errorToast: t("createGcpAccessReviewSourcePage.errors.delete") },
        );

      if (connectionStatus !== "CONNECTED") {
        toast({
          title: t("createGcpAccessReviewSourcePage.messages.error"),
          description: t(
            "createGcpAccessReviewSourcePage.errors.disconnected",
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
                name: gcpAccessReviewSourceName(
                  gcpDriver.displayName,
                  providerResource,
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
          { errorToast: t("createGcpAccessReviewSourcePage.errors.source") },
        );
      } catch {
        await discardConnector();
        return;
      }

      toast({
        title: t("createGcpAccessReviewSourcePage.messages.success"),
        description: t("createGcpAccessReviewSourcePage.messages.created"),
        variant: "success",
      });
      void navigate(`/organizations/${organizationId}/access-reviews/connections`);
    } catch {
      return;
    } finally {
      setIsCreating(false);
    }
  };

  const installActions: ActionSplitButtonAction[] = [];
  if (gcpConnectorSetup.terraformSnippet) {
    installActions.push({
      id: "terraform",
      label: t("createGcpAccessReviewSourcePage.actions.installViaTerraform"),
      onSelect: () =>
        copyValue(
          gcpConnectorSetup.terraformSnippet,
          "createGcpAccessReviewSourcePage.messages.copiedTerraform",
        ),
    });
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("createGcpAccessReviewSourcePage.title")}
        description={t("createGcpAccessReviewSourcePage.description")}
      >
        {installActions.length > 0 && (
          <ActionSplitButton
            actions={installActions}
            chooseAnotherMethodLabel={t(
              "createGcpAccessReviewSourcePage.actions.chooseAnotherInstallMethod",
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
            name="workloadIdentityProvider"
            label={t(
              "createGcpAccessReviewSourcePage.fields.workloadIdentityProvider",
            )}
            value={providerResource}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              setProviderResource(e.target.value)}
            required
            placeholder={t(
              "createGcpAccessReviewSourcePage.fields.workloadIdentityProviderPlaceholder",
            )}
            error={
              providerInvalid
                ? t("createGcpAccessReviewSourcePage.errors.workloadIdentityProvider")
                : undefined
            }
          />
          <Field
            name="serviceAccountEmail"
            label={t(
              "createGcpAccessReviewSourcePage.fields.serviceAccountEmail",
            )}
            value={serviceAccountEmail}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              setServiceAccountEmail(e.target.value)}
            required
            placeholder={t(
              "createGcpAccessReviewSourcePage.fields.serviceAccountEmailPlaceholder",
            )}
            error={
              emailInvalid
                ? t("createGcpAccessReviewSourcePage.errors.serviceAccountEmail")
                : undefined
            }
          />
          {(
            [
              {
                name: "issuer",
                value: gcpConnectorSetup.issuer,
                successKey:
                  "createGcpAccessReviewSourcePage.messages.copiedIssuer",
              },
              {
                name: "audience",
                value: gcpConnectorSetup.audience,
                successKey:
                  "createGcpAccessReviewSourcePage.messages.copiedAudience",
              },
              {
                name: "subject",
                value: gcpConnectorSetup.subject,
                successKey:
                  "createGcpAccessReviewSourcePage.messages.copiedSubject",
              },
            ] as const
          ).map(row => (
            <Field
              key={row.name}
              name={row.name}
              label={t(`createGcpAccessReviewSourcePage.fields.${row.name}`)}
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
                  aria-label={t("createGcpAccessReviewSourcePage.actions.copy")}
                />
              </div>
            </Field>
          ))}

          <div className="flex items-center justify-between gap-2">
            <ConnectorDocumentationLink
              url={gcpDriver.documentationUrl}
              variant="button"
            />
            <div className="flex items-center justify-end gap-2">
              <Button variant="secondary" asChild>
                <Link to={`/organizations/${organizationId}/access-reviews/connections`}>
                  {t("createGcpAccessReviewSourcePage.actions.back")}
                </Link>
              </Button>
              <Button disabled={!formValid || isCreating} type="submit">
                {t("createGcpAccessReviewSourcePage.actions.connect")}
              </Button>
            </div>
          </div>
        </form>
      </Card>
    </div>
  );
}
