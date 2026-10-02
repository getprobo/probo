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
  Button,
  DropdownItem,
  IconArrowLink,
  IconTrashCan,
  IconWarning,
  Input,
  Option,
  Select,
  ThirdPartyLogo,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useLazyLoadQuery, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessReviewSourceListItem_source$data, AccessReviewSourceListItem_source$key } from "#/__generated__/core/AccessReviewSourceListItem_source.graphql";
import type { AccessReviewSourceListItemCapturedOrganization_source$key } from "#/__generated__/core/AccessReviewSourceListItemCapturedOrganization_source.graphql";
import type { AccessReviewSourceListItemConfigureMutation } from "#/__generated__/core/AccessReviewSourceListItemConfigureMutation.graphql";
import type { AccessReviewSourceListItemDeleteMutation } from "#/__generated__/core/AccessReviewSourceListItemDeleteMutation.graphql";
import type { AccessReviewSourceListItemManualOrgInput_source$key } from "#/__generated__/core/AccessReviewSourceListItemManualOrgInput_source.graphql";
import type { AccessReviewSourceListItemOrganizations_source$key } from "#/__generated__/core/AccessReviewSourceListItemOrganizations_source.graphql";
import type { AccessReviewSourceListItemOrganizationsEmpty_source$key } from "#/__generated__/core/AccessReviewSourceListItemOrganizationsEmpty_source.graphql";
import type { AccessReviewSourceListItemOrganizationsUnavailable_source$key } from "#/__generated__/core/AccessReviewSourceListItemOrganizationsUnavailable_source.graphql";
import type { AccessReviewSourceListItemOrgsQuery$data, AccessReviewSourceListItemOrgsQuery } from "#/__generated__/core/AccessReviewSourceListItemOrgsQuery.graphql";
import type { AccessReviewSourceListItemSourceConnectionIssue_connector$key } from "#/__generated__/core/AccessReviewSourceListItemSourceConnectionIssue_connector.graphql";
import type { AccessReviewSourceConnectionStatus, AccessReviewSourceListItemStatus_source$key } from "#/__generated__/core/AccessReviewSourceListItemStatus_source.graphql";
import {
  type ConnectionIssueKey,
  connectionSignalFrom,
  presentConnection,
} from "#/pages/organizations/_lib/connectorStatus";
import { ConnectorDocumentationLink } from "#/pages/organizations/settings/integrations/_components/ConnectorDocumentationLink";
import { buildConnectorInitiateURL } from "#/pages/organizations/settings/integrations/_lib/connectorSettings";

import { accessReviewSourceSection } from "../connections/_components/variants";

function canReconnectConnector(
  connector: { canReconnect: boolean } | null | undefined,
): boolean {
  return connector?.canReconnect ?? false;
}

// A source with no connector is NOT_APPLICABLE. That is not a credential failure.
function sourceConnectorIssue(
  connectionStatus: AccessReviewSourceConnectionStatus,
  canReconnect: boolean,
  organizationsUnavailable: boolean,
): ConnectionIssueKey | null {
  if (connectionStatus === "NOT_APPLICABLE") {
    return null;
  }

  const signal = connectionSignalFrom({
    connectionStatus,
    canReconnect,
    providerOrganizations: {
      status: organizationsUnavailable ? "UNAVAILABLE" : "NOT_APPLICABLE",
    },
  });
  if (signal == null) {
    return null;
  }

  return presentConnection(signal).issue;
}

const fragment = graphql`
  fragment AccessReviewSourceListItem_source on AccessReviewSource
    @argumentDefinitions(
      deferConnectionStatus: { type: "Boolean!", defaultValue: true }
    ) {
    id
    name
    connectorId
    connector {
      provider
      protocol
      canReconnect
      oauth2Scopes
      ...AccessReviewSourceListItemSourceConnectionIssue_connector
    }
    selectedOrganization
    needsConfiguration
    createdAt
    canDelete: permission(action: "access-review:source:delete")
    ...AccessReviewSourceListItemStatus_source
      @defer(if: $deferConnectionStatus, label: "$defer$AccessReviewSourceListItemStatus")
  }
`;

const statusFragment = graphql`
  fragment AccessReviewSourceListItemStatus_source on AccessReviewSource {
    connectionStatus
  }
`;

export const deleteAccessReviewSourceMutation = graphql`
  mutation AccessReviewSourceListItemDeleteMutation(
    $input: DeleteAccessReviewSourceInput!
    $connections: [ID!]!
  ) {
    deleteAccessReviewSource(input: $input) {
      deletedAccessReviewSourceId @deleteEdge(connections: $connections)
    }
  }
`;

const configureMutation = graphql`
  mutation AccessReviewSourceListItemConfigureMutation(
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

const organizationsFragment = graphql`
  fragment AccessReviewSourceListItemOrganizations_source on AccessReviewSource {
    id
    selectedOrganization
    providerOrganizations {
      status
      nodes {
        slug
        displayName
      }
    }
    connector {
      ...AccessReviewSourceListItemSourceConnectionIssue_connector
    }
    ...AccessReviewSourceListItemCapturedOrganization_source
    ...AccessReviewSourceListItemOrganizationsEmpty_source
    ...AccessReviewSourceListItemOrganizationsUnavailable_source
  }
`;

const connectionIssueFragment = graphql`
  fragment AccessReviewSourceListItemSourceConnectionIssue_connector on Connector {
    displayName
    ...ConnectorDocumentationLink_connector
  }
`;

const organizationsEmptyFragment = graphql`
  fragment AccessReviewSourceListItemOrganizationsEmpty_source on AccessReviewSource {
    selectedOrganization
    connector {
      displayName
    }
    providerOrganizations {
      remediationUrl
    }
    ...AccessReviewSourceListItemManualOrgInput_source
  }
`;

const organizationsUnavailableFragment = graphql`
  fragment AccessReviewSourceListItemOrganizationsUnavailable_source on AccessReviewSource {
    connector {
      ...AccessReviewSourceListItemSourceConnectionIssue_connector
    }
  }
`;

const capturedOrganizationFragment = graphql`
  fragment AccessReviewSourceListItemCapturedOrganization_source on AccessReviewSource {
    selectedOrganization
  }
`;

const manualOrgInputFragment = graphql`
  fragment AccessReviewSourceListItemManualOrgInput_source on AccessReviewSource {
    selectedOrganization
  }
`;

const orgsQuery = graphql`
  query AccessReviewSourceListItemOrgsQuery($accessReviewSourceId: ID!) {
    node(id: $accessReviewSourceId) @required(action: THROW) {
      ... on AccessReviewSource {
        ...AccessReviewSourceListItemOrganizations_source
      }
    }
  }
`;

type Props = {
  sourceKey: AccessReviewSourceListItem_source$key;
  connectionId: string;
  organizationId: string;
};

// A source with no connector is the manual CSV one; every other name comes from
// the provider registry, where provider names are already maintained.
function sourceLabel(
  connector: { displayName: string } | null | undefined,
  t: (key: string) => string,
): string {
  return connector?.displayName ?? t("accessReviewSourceRow.sources.csv");
}

export function AccessReviewSourceListItem({
  sourceKey,
  connectionId,
  organizationId,
}: Props) {
  const { i18n, t } = useTranslation();
  const confirm = useConfirm();
  const { toast } = useToast();

  const accessSource = useFragment(fragment, sourceKey);

  const [deleteAccessReviewSource]
    = useMutation<AccessReviewSourceListItemDeleteMutation>(
      deleteAccessReviewSourceMutation,
    );
  const [configure]
    = useMutation<AccessReviewSourceListItemConfigureMutation>(configureMutation);

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

  const buildReconnectUrl = () => {
    const connector = accessSource.connector;
    if (!connector || !accessSource.connectorId) return null;

    return buildConnectorInitiateURL(
      organizationId,
      connector.provider,
      connector.protocol,
      {
        connectorId: accessSource.connectorId,
        oauth2Scopes: connector.oauth2Scopes,
      },
    );
  };
  const reconnectUrl = buildReconnectUrl();

  const showOrgSelector
    = accessSource.needsConfiguration || accessSource.selectedOrganization;
  const canReconnect = canReconnectConnector(accessSource.connector);

  return (
    <Card variant="soft" size={2} className="flex h-full min-w-0 flex-col gap-3">
      <div className="flex items-center gap-4">
        {accessSource.connector?.provider && (
          <ThirdPartyLogo
            thirdParty={accessSource.connector.provider}
            className="size-8 shrink-0"
          />
        )}
        <div className="flex min-w-0 flex-1 flex-col gap-0.5">
          <Heading level={3} size={3} weight="medium" highContrast className="min-w-0 truncate">
            {accessSource.name}
          </Heading>
          <Text size={1} color="faint">
            <time dateTime={accessSource.createdAt}>
              {dateTimeFormat(i18n.language, accessSource.createdAt)}
            </time>
          </Text>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-end gap-2">
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
            <DeferredInlineOrgSelect
              sourceKey={accessSource}
              accessReviewSourceId={accessSource.id}
              onSelect={handleOrgChange}
              canReconnect={canReconnect}
              reconnectUrl={canReconnect ? reconnectUrl : null}
            />
          </Suspense>
        )}
        {!showOrgSelector && (
          <Suspense fallback={null}>
            <DeferredStandaloneIssue
              sourceKey={accessSource}
              canReconnect={canReconnect}
              reconnectUrl={canReconnect ? reconnectUrl : null}
            />
          </Suspense>
        )}
        {accessSource.canDelete && (
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
        )}
      </div>
    </Card>
  );
}

function DeferredInlineOrgSelect({
  sourceKey,
  accessReviewSourceId,
  onSelect,
  canReconnect,
  reconnectUrl,
}: {
  sourceKey: AccessReviewSourceListItem_source$data;
  accessReviewSourceId: string;
  canReconnect: boolean;
  reconnectUrl: string | null;
  onSelect: (slug: string) => void;
}) {
  const data = useLazyLoadQuery<AccessReviewSourceListItemOrgsQuery>(
    orgsQuery,
    { accessReviewSourceId },
    { fetchPolicy: "store-or-network" },
  );
  const status = useFragment(
    statusFragment,
    sourceKey as unknown as AccessReviewSourceListItemStatus_source$key,
  );

  return (
    <InlineOrgSelect
      data={data}
      onSelect={onSelect}
      connectionStatus={status.connectionStatus}
      canReconnect={canReconnect}
      reconnectUrl={reconnectUrl}
    />
  );
}

function DeferredStandaloneIssue({
  sourceKey,
  canReconnect,
  reconnectUrl,
}: {
  sourceKey: AccessReviewSourceListItem_source$data;
  canReconnect: boolean;
  reconnectUrl: string | null;
}) {
  const status = useFragment(
    statusFragment,
    sourceKey as unknown as AccessReviewSourceListItemStatus_source$key,
  );
  const hasConnectionIssue
    = status.connectionStatus !== "CONNECTED"
      && status.connectionStatus !== "NOT_APPLICABLE";
  if (!hasConnectionIssue) {
    return null;
  }

  const issue = sourceConnectorIssue(status.connectionStatus, canReconnect, false);
  if (issue == null) {
    return null;
  }

  return (
    <SourceConnectionIssue
      connectorKey={sourceKey.connector}
      issueKey={issue}
      reconnectUrl={reconnectUrl}
    />
  );
}

function InlineOrgSelect({
  data,
  onSelect,
  connectionStatus,
  canReconnect,
  reconnectUrl,
}: {
  data: AccessReviewSourceListItemOrgsQuery$data;
  connectionStatus: AccessReviewSourceConnectionStatus;
  canReconnect: boolean;
  reconnectUrl: string | null;
  onSelect: (slug: string) => void;
}) {
  const { t } = useTranslation();
  const source
    = useFragment<AccessReviewSourceListItemOrganizations_source$key>(
      organizationsFragment,
      data.node,
    );
  const providerOrganizations = source.providerOrganizations;

  switch (providerOrganizations?.status) {
    case "AVAILABLE":
      return (
        <Select
          variant="editor"
          placeholder={t("accessReviewSourceRow.selectOrganization")}
          value={source.selectedOrganization ?? ""}
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
    // A captured organization is still worth showing, but not instead of a
    // refusal: this branch owns the whole trailing cell, so the parent's
    // standalone issue is suppressed here and a broken source would otherwise
    // render as nothing but an organization name.
    case "NOT_APPLICABLE": {
      const issue = sourceConnectorIssue(connectionStatus, canReconnect, false);
      if (issue != null) {
        return (
          <SourceConnectionIssue
            connectorKey={source.connector}
            issueKey={issue}
            reconnectUrl={reconnectUrl}
          />
        );
      }

      return <CapturedOrganization sourceKey={source} />;
    }
    case "EMPTY":
      return (
        <ProviderOrganizationsEmpty sourceKey={source} onSubmit={onSelect} />
      );
    // UNAVAILABLE, plus the impossible case of a node that is not an
    // AccessReviewSource: either way the list could not be read.
    default: {
      const issue = sourceConnectorIssue(connectionStatus, canReconnect, true);
      if (issue == null) {
        return null;
      }

      return (
        <ProviderOrganizationsUnavailable
          sourceKey={source}
          issueKey={issue}
          reconnectUrl={reconnectUrl}
        />
      );
    }
  }
}

function SourceConnectionIssue({
  connectorKey,
  issueKey,
  reconnectUrl,
}: {
  connectorKey: AccessReviewSourceListItemSourceConnectionIssue_connector$key | null | undefined;
  issueKey: ConnectionIssueKey;
  reconnectUrl: string | null;
}) {
  const { t } = useTranslation();
  const connector = useFragment(connectionIssueFragment, connectorKey ?? null);
  const provider = connector?.displayName ?? t("accessReviewSourceRow.sources.csv");
  const {
    issue,
    issueIcon,
    issueContent,
    issueTitle,
    issueDescription,
  } = accessReviewSourceSection();

  const unavailable = "accessReviewSourceRow.organizations.unavailable";

  return (
    <div className={issue()}>
      <IconWarning size={16} className={issueIcon()} />
      <div className={issueContent()}>
        <p className={issueTitle()}>
          {t(`${unavailable}.${issueKey}Title`, { provider })}
        </p>
        <p className={issueDescription()}>
          {t(`${unavailable}.${issueKey}Description`, { provider })}
        </p>
        {issueKey !== "reconnect" && (
          <ConnectorDocumentationLink connectorKey={connector} />
        )}
      </div>
      {reconnectUrl && (
        <Button variant="primary" asChild>
          <a href={reconnectUrl}>
            {t("accessReviewSourceRow.actions.reconnect")}
          </a>
        </Button>
      )}
    </div>
  );
}

// The provider answered, with nothing. Typing a slug does not fix the usual
// causes (unapproved app, personal account), so the explanation leads and the
// manual input stays available underneath it.
function ProviderOrganizationsEmpty({
  sourceKey,
  onSubmit,
}: {
  sourceKey: AccessReviewSourceListItemOrganizationsEmpty_source$key;
  onSubmit: (slug: string) => void;
}) {
  const { t } = useTranslation();
  const source = useFragment(organizationsEmptyFragment, sourceKey);
  const providerLabel = sourceLabel(source.connector, t);
  const remediationUrl = source.providerOrganizations.remediationUrl;

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
        <ManualOrgInput sourceKey={source} onSubmit={onSubmit} />
      </div>
    </div>
  );
}

// The provider call failed. A free text box would hide that, so the only
// affordance offered is reconnecting the connector.
function ProviderOrganizationsUnavailable({
  sourceKey,
  issueKey,
  reconnectUrl,
}: {
  sourceKey: AccessReviewSourceListItemOrganizationsUnavailable_source$key;
  issueKey: ConnectionIssueKey;
  reconnectUrl: string | null;
}) {
  const source = useFragment(organizationsUnavailableFragment, sourceKey);

  return (
    <SourceConnectionIssue
      connectorKey={source.connector}
      issueKey={issueKey}
      reconnectUrl={reconnectUrl}
    />
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
  sourceKey,
}: {
  sourceKey: AccessReviewSourceListItemCapturedOrganization_source$key;
}) {
  const source = useFragment(capturedOrganizationFragment, sourceKey);
  return (
    <span className="text-sm text-txt-primary">
      {source.selectedOrganization}
    </span>
  );
}

function ManualOrgInput({
  sourceKey,
  onSubmit,
}: {
  sourceKey: AccessReviewSourceListItemManualOrgInput_source$key;
  onSubmit: (slug: string) => void;
}) {
  const { t } = useTranslation();
  const source = useFragment(manualOrgInputFragment, sourceKey);
  const selectedOrganization = source.selectedOrganization ?? "";
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
