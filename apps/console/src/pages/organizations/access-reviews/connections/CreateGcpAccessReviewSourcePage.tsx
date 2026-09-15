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
  IconListStack,
  IconSquareBehindSquare2,
  Input,
  PageHeader,
  useToast,
} from "@probo/ui";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { type ChangeEvent, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link, useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { CreateGcpAccessReviewSourcePageCreateMutation } from "#/__generated__/core/CreateGcpAccessReviewSourcePageCreateMutation.graphql";
import type { CreateGcpAccessReviewSourcePageCreateSourcesMutation } from "#/__generated__/core/CreateGcpAccessReviewSourcePageCreateSourcesMutation.graphql";
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
import { createAccessReviewSourceMutation, prependCreatedGcpSourceEdges, prependCreatedSourceEdge } from "../dialogs/accessReviewSourceMutations";

import {
  parseGcpTerraformConnectors,
} from "./_lib/parseGcpTerraformConnectors";

export const createGcpAccessReviewSourcePageQuery = graphql`
  query CreateGcpAccessReviewSourcePageQuery($organizationId: ID!) {
    gcpConnectorSetup(organizationId: $organizationId) {
      issuer
      audience
      subject
      terraformSnippet
      terraformBulkSnippet
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

const createGcpAccessReviewSourcesMutation = graphql`
  mutation CreateGcpAccessReviewSourcePageCreateSourcesMutation(
    $input: CreateGcpAccessReviewSourcesInput!
  ) {
    createGcpAccessReviewSources(input: $input) {
      accessReviewSourceEdges {
        node {
          id
          name
          connectorId
          createdAt
          ...AccessReviewSourceListItem_source
        }
      }
      failures {
        index
        projectId
        reason
      }
    }
  }
`;

type ConnectMode = "single" | "bulk";
type ConnectOneResult = "connected" | "disconnected" | "failed";

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
  const [mode, setMode] = useState<ConnectMode>("single");
  const [providerResource, setProviderResource] = useState("");
  const [serviceAccountEmail, setServiceAccountEmail] = useState("");
  const [terraformJson, setTerraformJson] = useState("");
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
  const [createGcpAccessReviewSources] = useMutation<
    CreateGcpAccessReviewSourcePageCreateSourcesMutation
  >(createGcpAccessReviewSourcesMutation);

  const parsedTerraform = useMemo(
    () => parseGcpTerraformConnectors(terraformJson),
    [terraformJson],
  );

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
  const terraformTouched = terraformJson.trim() !== "";
  const terraformError = (() => {
    if (!terraformTouched) {
      return undefined;
    }

    if (parsedTerraform.ok) {
      if (parsedTerraform.connectors.length > 100) {
        return t("createGcpAccessReviewSourcePage.errors.tooMany");
      }

      return undefined;
    }

    if (parsedTerraform.error === "invalidEntry") {
      return t("createGcpAccessReviewSourcePage.errors.terraformEntry", {
        projectId: parsedTerraform.projectId,
      });
    }

    return t(`createGcpAccessReviewSourcePage.errors.${parsedTerraform.error}`);
  })();
  const bulkConnectors = parsedTerraform.ok ? parsedTerraform.connectors : [];
  const tooManyProjects = bulkConnectors.length > 100;
  const formValid = mode === "single"
    ? providerValid && emailValid
    : parsedTerraform.ok && !tooManyProjects;

  const connectOne = async (
    workloadIdentityProvider: string,
    serviceAccountEmailValue: string,
    sourceName: string,
    notifyErrors: boolean,
  ): Promise<ConnectOneResult> => {
    const createErrorToast = notifyErrors
      ? t("createGcpAccessReviewSourcePage.errors.create")
      : false;
    const sourceErrorToast = notifyErrors
      ? t("createGcpAccessReviewSourcePage.errors.source")
      : false;
    const deleteErrorToast = notifyErrors
      ? t("createGcpAccessReviewSourcePage.errors.delete")
      : false;

    try {
      const created = await createWorkloadIdentityConnector(
        {
          variables: {
            input: {
              organizationId,
              provider: "GCP",
              gcpWorkloadIdentityProvider: workloadIdentityProvider,
              gcpServiceAccountEmail: serviceAccountEmailValue,
            },
          },
        },
        { errorToast: createErrorToast },
      );
      const { id: connectorId, connectionStatus }
        = created.createWorkloadIdentityConnector.connector;

      const discardConnector = () =>
        deleteConnector(
          { variables: { input: { connectorId } } },
          { errorToast: deleteErrorToast },
        );

      if (connectionStatus !== "CONNECTED") {
        await discardConnector();
        return "disconnected";
      }

      try {
        await createAccessReviewSource(
          {
            variables: {
              input: {
                organizationId,
                connectorId,
                name: sourceName,
                csvData: null,
              },
            },
            updater: (store) => {
              if (connectionId) {
                prependCreatedSourceEdge(store, connectionId);
              }
            },
          },
          { errorToast: sourceErrorToast },
        );
      } catch {
        await discardConnector();
        return "failed";
      }

      return "connected";
    } catch {
      return "failed";
    }
  };

  const onSubmitSingle = async () => {
    if (!formValid) {
      return;
    }

    setIsCreating(true);

    try {
      const result = await connectOne(
        providerResource.trim(),
        serviceAccountEmail.trim(),
        gcpAccessReviewSourceName(gcpDriver.displayName, providerResource),
        true,
      );

      if (result === "disconnected") {
        toast({
          title: t("createGcpAccessReviewSourcePage.messages.error"),
          description: t(
            "createGcpAccessReviewSourcePage.errors.disconnected",
          ),
          variant: "error",
        });
        return;
      }

      if (result !== "connected") {
        return;
      }

      toast({
        title: t("createGcpAccessReviewSourcePage.messages.success"),
        description: t("createGcpAccessReviewSourcePage.messages.created"),
        variant: "success",
      });
      void navigate(`/organizations/${organizationId}/access-reviews/connections`);
    } finally {
      setIsCreating(false);
    }
  };

  const onSubmitBulk = async () => {
    if (!parsedTerraform.ok || parsedTerraform.connectors.length > 100) {
      return;
    }

    const connectors = parsedTerraform.connectors;
    setIsCreating(true);

    try {
      const result = await createGcpAccessReviewSources(
        {
          variables: {
            input: {
              organizationId,
              projects: connectors.map(connector => ({
                projectId: connector.projectId,
                gcpWorkloadIdentityProvider: connector.workloadIdentityProvider,
                gcpServiceAccountEmail: connector.serviceAccountEmail,
              })),
            },
          },
          updater: (store) => {
            if (connectionId) {
              prependCreatedGcpSourceEdges(store, connectionId);
            }
          },
        },
        { errorToast: t("createGcpAccessReviewSourcePage.errors.bulkFailed") },
      );

      const createdCount
        = result.createGcpAccessReviewSources.accessReviewSourceEdges.length;
      const failed = result.createGcpAccessReviewSources.failures.map((failure) => {
        if (failure.projectId) {
          return failure.projectId;
        }

        const connector = connectors[failure.index];
        return connector?.projectId ?? String(failure.index + 1);
      });

      if (createdCount === connectors.length) {
        toast({
          title: t("createGcpAccessReviewSourcePage.messages.success"),
          description: t("createGcpAccessReviewSourcePage.messages.createdCount", {
            count: createdCount,
          }),
          variant: "success",
        });
        void navigate(`/organizations/${organizationId}/access-reviews/connections`);
        return;
      }

      toast({
        title: t("createGcpAccessReviewSourcePage.messages.error"),
        description: createdCount === 0
          ? t("createGcpAccessReviewSourcePage.errors.bulkFailed")
          : t("createGcpAccessReviewSourcePage.messages.partialCreated", {
              created: createdCount,
              total: connectors.length,
              failed: failed.join(", "),
            }),
        variant: "error",
      });
    } finally {
      setIsCreating(false);
    }
  };

  const onSubmit = () => {
    if (mode === "bulk") {
      void onSubmitBulk();
      return;
    }

    void onSubmitSingle();
  };

  const terraformSnippetToCopy = mode === "bulk"
    ? gcpConnectorSetup.terraformBulkSnippet
    : gcpConnectorSetup.terraformSnippet;

  const installActions: ActionSplitButtonAction[] = [];
  if (terraformSnippetToCopy) {
    installActions.push({
      id: "terraform",
      label: t("createGcpAccessReviewSourcePage.actions.installViaTerraform"),
      onSelect: () =>
        copyValue(
          terraformSnippetToCopy,
          "createGcpAccessReviewSourcePage.messages.copiedTerraform",
        ),
    });
  }

  const submitLabel = (() => {
    if (isCreating) {
      return t("createGcpAccessReviewSourcePage.actions.connecting");
    }

    if (mode === "bulk" && bulkConnectors.length > 0) {
      return t("createGcpAccessReviewSourcePage.actions.connectCount", {
        count: bulkConnectors.length,
      });
    }

    return t("createGcpAccessReviewSourcePage.actions.connect");
  })();

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("createGcpAccessReviewSourcePage.title")}
        description={t("createGcpAccessReviewSourcePage.description")}
      >
        <div className="flex items-center gap-2">
          {installActions.length > 0 && (
            <ActionSplitButton
              actions={installActions}
              chooseAnotherMethodLabel={t(
                "createGcpAccessReviewSourcePage.actions.chooseAnotherInstallMethod",
              )}
            />
          )}
          <Tooltip>
            <TooltipTrigger
              render={(
                <Button
                  type="button"
                  variant={mode === "bulk" ? "primary" : "secondary"}
                  icon={IconListStack}
                  aria-pressed={mode === "bulk"}
                  aria-label={t("createGcpAccessReviewSourcePage.modes.bulk")}
                  onClick={() => setMode(mode === "bulk" ? "single" : "bulk")}
                />
              )}
            />
            <TooltipPopup>
              {t("createGcpAccessReviewSourcePage.modes.bulk")}
            </TooltipPopup>
          </Tooltip>
        </div>
      </PageHeader>

      <Card padded>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            onSubmit();
          }}
          className="space-y-4"
        >
          {mode === "single"
            ? (
                <>
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
                </>
              )
            : (
                <>
                  <Field
                    type="textarea"
                    name="terraformOutput"
                    label={t(
                      "createGcpAccessReviewSourcePage.fields.terraformOutput",
                    )}
                    help={t(
                      "createGcpAccessReviewSourcePage.fields.terraformOutputHelp",
                    )}
                    value={terraformJson}
                    onValueChange={setTerraformJson}
                    required
                    rows={12}
                    placeholder={t(
                      "createGcpAccessReviewSourcePage.fields.terraformOutputPlaceholder",
                    )}
                    error={terraformError}
                  />
                  {bulkConnectors.length > 0 && (
                    <ul className="text-txt-secondary text-sm list-disc pl-5">
                      {bulkConnectors.map(connector => (
                        <li key={connector.projectId}>{connector.projectId}</li>
                      ))}
                    </ul>
                  )}
                </>
              )}
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
                {submitLabel}
              </Button>
            </div>
          </div>
        </form>
      </Card>
    </div>
  );
}
