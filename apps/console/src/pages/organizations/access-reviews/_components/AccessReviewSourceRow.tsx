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

import { formatError } from "@probo/helpers";
import { dateTimeFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  DropdownItem,
  IconArrowLink,
  IconTrashCan,
  IconWarning,
  Input,
  Option,
  Select,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useLazyLoadQuery, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessReviewSourceRowConfigureMutation } from "#/__generated__/core/AccessReviewSourceRowConfigureMutation.graphql";
import type { AccessReviewSourceRowDeleteMutation } from "#/__generated__/core/AccessReviewSourceRowDeleteMutation.graphql";
import type { AccessReviewSourceRowFragment$key } from "#/__generated__/core/AccessReviewSourceRowFragment.graphql";
import type { AccessReviewSourceRowOrgsQuery } from "#/__generated__/core/AccessReviewSourceRowOrgsQuery.graphql";

const fragment = graphql`
  fragment AccessReviewSourceRowFragment on AccessReviewSource {
    id
    name
    connectorId
    connector {
      provider
      oauth2Scopes
    }
    connectionStatus
    selectedOrganization
    needsConfiguration
    createdAt
    canDelete: permission(action: "access-review:source:delete")
  }
`;

export const deleteAccessReviewSourceMutation = graphql`
  mutation AccessReviewSourceRowDeleteMutation(
    $input: DeleteAccessReviewSourceInput!
    $connections: [ID!]!
  ) {
    deleteAccessReviewSource(input: $input) {
      deletedAccessReviewSourceId @deleteEdge(connections: $connections)
    }
  }
`;

const configureMutation = graphql`
  mutation AccessReviewSourceRowConfigureMutation(
    $input: ConfigureAccessReviewSourceInput!
  ) {
    configureAccessReviewSource(input: $input) {
      accessReviewSource {
        id
        selectedOrganization
        needsConfiguration
      }
    }
  }
`;

const orgsQuery = graphql`
  query AccessReviewSourceRowOrgsQuery($accessReviewSourceId: ID!) {
    node(id: $accessReviewSourceId) @required(action: THROW) {
      ... on AccessReviewSource {
        providerOrganizations {
          status
          remediationUrl
          nodes {
            slug
            displayName
          }
        }
      }
    }
  }
`;

type Props = {
  fKey: AccessReviewSourceRowFragment$key;
  connectionId: string;
  organizationId: string;
};

function sourceLabel(
  connectorProvider: string | null | undefined,
  t: (key: string) => string,
): string {
  if (!connectorProvider) {
    return t("accessReviewSourceRow.sources.csv");
  }

  switch (connectorProvider) {
    case "GOOGLE_WORKSPACE":
      return t("accessReviewSourceRow.sources.googleWorkspace");
    case "MICROSOFT_365":
      return t("accessReviewSourceRow.sources.microsoft365");
    case "LINEAR":
      return t("accessReviewSourceRow.sources.linear");
    case "SLACK":
      return t("accessReviewSourceRow.sources.slack");
    case "METABASE":
      return t("accessReviewSourceRow.sources.metabase");
    case "SIGNOZ":
      return t("accessReviewSourceRow.sources.signoz");
    case "CURSOR":
      return t("accessReviewSourceRow.sources.cursor");
    default:
      return connectorProvider;
  }
}

export function AccessReviewSourceRow({ fKey, connectionId, organizationId }: Props) {
  const { i18n, t } = useTranslation();
  const confirm = useConfirm();
  const { toast } = useToast();

  const accessSource = useFragment(fragment, fKey);

  const [deleteAccessReviewSource] = useMutation<AccessReviewSourceRowDeleteMutation>(deleteAccessReviewSourceMutation);
  const [configure] = useMutation<AccessReviewSourceRowConfigureMutation>(configureMutation);

  const handleDelete = () => {
    confirm(
      () => {
        deleteAccessReviewSource({
          variables: {
            input: { accessReviewSourceId: accessSource.id },
            connections: [connectionId],
          },
          onCompleted: (_response, errors) => {
            if (errors?.length) {
              toast({
                title: t("accessReviewSourceRow.messages.error"),
                description: formatError(
                  t("accessReviewSourceRow.errors.delete"),
                  errors,
                ),
                variant: "error",
              });
            }
          },
          onError: (error) => {
            toast({
              title: t("accessReviewSourceRow.messages.error"),
              description: formatError(
                t("accessReviewSourceRow.errors.delete"),
                error,
              ),
              variant: "error",
            });
          },
        });
      },
      {
        message: t("accessReviewSourceRow.deleteConfirmation", {
          name: accessSource.name,
        }),
      },
    );
  };

  const handleOrgChange = (slug: string) => {
    configure({
      variables: {
        input: {
          accessReviewSourceId: accessSource.id,
          organizationSlug: slug,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("accessReviewSourceRow.messages.error"),
            description: formatError(
              t("accessReviewSourceRow.errors.configure"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("accessReviewSourceRow.messages.success"),
          description: t("accessReviewSourceRow.messages.organizationUpdated"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: t("accessReviewSourceRow.messages.error"),
          description: formatError(
            t("accessReviewSourceRow.errors.configure"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleReconnect = () => {
    const connector = accessSource.connector;
    if (!connector || !accessSource.connectorId) return;

    const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseURL);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", connector.provider);
    url.searchParams.append("connector_id", accessSource.connectorId);
    for (const scope of connector.oauth2Scopes) {
      url.searchParams.append("scope", scope);
    }
    url.searchParams.append(
      "continue",
      `/organizations/${organizationId}/access-reviews/sources`,
    );
    window.location.href = url.toString();
  };

  const showOrgSelector = accessSource.needsConfiguration || accessSource.selectedOrganization;
  const providerLabel = sourceLabel(accessSource.connector?.provider ?? null, t);
  const canReconnect = (accessSource.connector?.oauth2Scopes.length ?? 0) > 0;
  const showReconnect
    = canReconnect
      && (accessSource.connectionStatus === "DISCONNECTED"
        || accessSource.connectionStatus === "RECONNECT_REQUIRED");

  return (
    <Tr>
      <Td>{accessSource.name}</Td>
      <Td>
        <Badge variant="neutral" size="sm">
          {providerLabel}
        </Badge>
      </Td>
      <Td>
        {accessSource.connectionStatus === "CONNECTED" && (
          <Badge variant="success" size="sm">
            {t("accessReviewSourceRow.status.connected")}
          </Badge>
        )}
        {accessSource.connectionStatus === "RECONNECT_REQUIRED" && (
          <div className="flex items-center gap-2">
            <Badge variant="warning" size="sm">
              {t("accessReviewSourceRow.status.reconnectRequired")}
            </Badge>
            {showReconnect && (
              <Button variant="secondary" onClick={handleReconnect}>
                {t("accessReviewSourceRow.actions.reconnect")}
              </Button>
            )}
          </div>
        )}
        {accessSource.connectionStatus === "DISCONNECTED" && (
          <div className="flex items-center gap-2">
            <Badge variant="danger" size="sm">
              {canReconnect
                ? t("accessReviewSourceRow.status.disconnected")
                : t("accessReviewSourceRow.status.invalidCredentials")}
            </Badge>
            {showReconnect && (
              <Button variant="secondary" onClick={handleReconnect}>
                {t("accessReviewSourceRow.actions.reconnect")}
              </Button>
            )}
          </div>
        )}
      </Td>
      <Td>
        {showOrgSelector && (
          <Suspense
            fallback={(
              <Select
                variant="editor"
                disabled
                placeholder={t("accessReviewSourceRow.loading")}
              />
            )}
          >
            <InlineOrgSelect
              accessReviewSourceId={accessSource.id}
              selectedOrganization={accessSource.selectedOrganization ?? ""}
              providerLabel={providerLabel}
              canReconnect={canReconnect}
              onSelect={handleOrgChange}
              onReconnect={handleReconnect}
            />
          </Suspense>
        )}
      </Td>
      <Td>
        <time dateTime={accessSource.createdAt}>
          {dateTimeFormat(i18n.language, accessSource.createdAt)}
        </time>
      </Td>
      {accessSource.canDelete && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            <DropdownItem
              icon={IconTrashCan}
              variant="danger"
              onSelect={(e) => {
                e.preventDefault();
                e.stopPropagation();
                handleDelete();
              }}
            >
              {t("accessReviewSourceRow.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}

function InlineOrgSelect({
  accessReviewSourceId,
  selectedOrganization,
  providerLabel,
  canReconnect,
  onSelect,
  onReconnect,
}: {
  accessReviewSourceId: string;
  selectedOrganization: string;
  providerLabel: string;
  canReconnect: boolean;
  onSelect: (slug: string) => void;
  onReconnect: () => void;
}) {
  const { t } = useTranslation();
  const data = useLazyLoadQuery<AccessReviewSourceRowOrgsQuery>(
    orgsQuery,
    { accessReviewSourceId },
    { fetchPolicy: "store-or-network" },
  );

  const providerOrganizations = data.node.providerOrganizations;

  switch (providerOrganizations?.status) {
    case "AVAILABLE":
      return (
        <Select
          variant="editor"
          placeholder={t("accessReviewSourceRow.selectOrganization")}
          value={selectedOrganization}
          onValueChange={onSelect}
        >
          {providerOrganizations.nodes.map(org => (
            <Option key={org.slug} value={org.slug}>
              {org.displayName}
            </Option>
          ))}
        </Select>
      );
    // The provider has no picker: its organization was captured during the
    // OAuth callback. An empty list is expected here, so do not warn about a
    // connection that is perfectly healthy — but do not offer an input either.
    // No provider in this branch implements SetOrganizationSettings, so
    // submitting one could only ever return "does not support organization
    // configuration". Show what the callback captured, read-only; changing it
    // means reconnecting.
    case "NOT_APPLICABLE":
      return <CapturedOrganization selectedOrganization={selectedOrganization} />;
    case "EMPTY":
      return (
        <ProviderOrganizationsEmpty
          selectedOrganization={selectedOrganization}
          providerLabel={providerLabel}
          remediationUrl={providerOrganizations.remediationUrl}
          onSubmit={onSelect}
        />
      );
    // UNAVAILABLE, plus the impossible case of a node that is not an
    // AccessReviewSource: either way the list could not be read.
    default:
      return (
        <ProviderOrganizationsUnavailable
          providerLabel={providerLabel}
          canReconnect={canReconnect}
          onReconnect={onReconnect}
        />
      );
  }
}

// The provider answered, with nothing. Typing a slug does not fix the usual
// causes (unapproved app, personal account), so the explanation leads and the
// manual input stays available underneath it.
function ProviderOrganizationsEmpty({
  selectedOrganization,
  providerLabel,
  remediationUrl,
  onSubmit,
}: {
  selectedOrganization: string;
  providerLabel: string;
  remediationUrl: string | null | undefined;
  onSubmit: (slug: string) => void;
}) {
  const { t } = useTranslation();

  return (
    <div className="flex max-w-80 flex-col gap-2">
      <div className="flex items-start gap-2 text-xs text-txt-tertiary">
        <IconWarning size={14} className="mt-0.5 shrink-0 text-txt-warning" />
        <div className="space-y-1">
          <p className="font-medium text-txt-primary">
            {t("accessReviewSourceRow.organizations.empty.title", {
              provider: providerLabel,
            })}
          </p>
          <p>{t("accessReviewSourceRow.organizations.empty.description")}</p>
          {remediationUrl && (
            <a
              href={remediationUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 underline hover:no-underline"
            >
              {t("accessReviewSourceRow.organizations.empty.remediation")}
              <IconArrowLink size={12} />
            </a>
          )}
        </div>
      </div>
      <div className="space-y-1">
        <p className="text-xs text-txt-tertiary">
          {t("accessReviewSourceRow.organizations.empty.manualLabel")}
        </p>
        <ManualOrgInput
          selectedOrganization={selectedOrganization}
          onSubmit={onSubmit}
        />
      </div>
    </div>
  );
}

// The provider call failed. A free text box would hide that, so the only
// affordance offered is reconnecting the connector.
function ProviderOrganizationsUnavailable({
  providerLabel,
  canReconnect,
  onReconnect,
}: {
  providerLabel: string;
  canReconnect: boolean;
  onReconnect: () => void;
}) {
  const { t } = useTranslation();

  return (
    <div className="flex max-w-80 items-start gap-2 text-xs text-txt-tertiary">
      <IconWarning size={14} className="mt-0.5 shrink-0 text-txt-danger" />
      <div className="space-y-1">
        <p className="font-medium text-txt-primary">
          {t("accessReviewSourceRow.organizations.unavailable.title")}
        </p>
        <p>
          {t("accessReviewSourceRow.organizations.unavailable.description", {
            provider: providerLabel,
          })}
        </p>
        {canReconnect && (
          <Button variant="secondary" onClick={onReconnect}>
            {t("accessReviewSourceRow.actions.reconnect")}
          </Button>
        )}
      </div>
    </div>
  );
}

// The organization a 2-auto provider captured during its OAuth callback
// (PagerDuty's subdomain, Vercel's team, Datadog's domain, Zendesk's
// subdomain). It is not user-editable: the value came from the handshake, and
// the only way to change it is to reconnect against a different one.
//
// No empty state: these providers have NeedsPicker false, so an unset
// organization leaves needsConfiguration false and the parent renders no
// selector at all.
function CapturedOrganization({
  selectedOrganization,
}: {
  selectedOrganization: string;
}) {
  return <span className="text-sm text-txt-primary">{selectedOrganization}</span>;
}

function ManualOrgInput({
  selectedOrganization,
  onSubmit,
}: {
  selectedOrganization: string;
  onSubmit: (slug: string) => void;
}) {
  const { t } = useTranslation();
  const [value, setValue] = useState(selectedOrganization);

  const handleBlur = () => {
    const trimmed = value.trim();
    if (trimmed && trimmed !== selectedOrganization) {
      onSubmit(trimmed);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleBlur();
    }
  };

  return (
    <Input
      placeholder={t("accessReviewSourceRow.organizationSlugPlaceholder")}
      value={value}
      onChange={e => setValue(e.target.value)}
      onBlur={handleBlur}
      onKeyDown={handleKeyDown}
      className="max-w-40"
    />
  );
}
