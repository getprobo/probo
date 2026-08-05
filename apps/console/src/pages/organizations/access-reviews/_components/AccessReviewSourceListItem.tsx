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
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useLazyLoadQuery, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessReviewSourceListItem_source$key } from "#/__generated__/core/AccessReviewSourceListItem_source.graphql";
import type { AccessReviewSourceListItemCapturedOrganization_source$key } from "#/__generated__/core/AccessReviewSourceListItemCapturedOrganization_source.graphql";
import type { AccessReviewSourceListItemConfigureMutation } from "#/__generated__/core/AccessReviewSourceListItemConfigureMutation.graphql";
import type { AccessReviewSourceListItemDeleteMutation } from "#/__generated__/core/AccessReviewSourceListItemDeleteMutation.graphql";
import type { AccessReviewSourceListItemManualOrgInput_source$key } from "#/__generated__/core/AccessReviewSourceListItemManualOrgInput_source.graphql";
import type { AccessReviewSourceListItemOrganizations_source$key } from "#/__generated__/core/AccessReviewSourceListItemOrganizations_source.graphql";
import type { AccessReviewSourceListItemOrganizationsEmpty_source$key } from "#/__generated__/core/AccessReviewSourceListItemOrganizationsEmpty_source.graphql";
import type { AccessReviewSourceListItemOrganizationsUnavailable_source$key } from "#/__generated__/core/AccessReviewSourceListItemOrganizationsUnavailable_source.graphql";
import type { AccessReviewSourceListItemOrgsQuery } from "#/__generated__/core/AccessReviewSourceListItemOrgsQuery.graphql";

import { accessReviewSourceSection } from "../sources/_components/variants";

const fragment = graphql`
  fragment AccessReviewSourceListItem_source on AccessReviewSource {
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
    ...AccessReviewSourceListItemCapturedOrganization_source
    ...AccessReviewSourceListItemOrganizationsEmpty_source
    ...AccessReviewSourceListItemOrganizationsUnavailable_source
  }
`;

const organizationsEmptyFragment = graphql`
  fragment AccessReviewSourceListItemOrganizationsEmpty_source on AccessReviewSource {
    selectedOrganization
    connector {
      provider
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
      provider
      oauth2Scopes
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
    case "GITHUB":
      return t("accessReviewSourceRow.sources.github");
    case "CLOUDFLARE":
      return t("accessReviewSourceRow.sources.cloudflare");
    default:
      return connectorProvider;
  }
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
  const { item, content, trailing } = accessReviewSourceSection();

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
    return url.toString();
  };
  const reconnectUrl = buildReconnectUrl();

  const showOrgSelector
    = accessSource.needsConfiguration || accessSource.selectedOrganization;
  const canReconnect = (accessSource.connector?.oauth2Scopes.length ?? 0) > 0;
  const showReconnect
    = canReconnect
      && (accessSource.connectionStatus === "DISCONNECTED"
        || accessSource.connectionStatus === "RECONNECT_REQUIRED");
  const showStandaloneIssue
    = (accessSource.connectionStatus === "DISCONNECTED"
      || accessSource.connectionStatus === "RECONNECT_REQUIRED")
    && !showOrgSelector;

  return (
    <li className={item()}>
      {accessSource.connector?.provider && (
        <ThirdPartyLogo
          thirdParty={accessSource.connector.provider}
          tint
          className="size-6 shrink-0"
        />
      )}
      <div className={content()}>
        <span className="truncate text-sm font-medium text-txt-primary">
          {accessSource.name}
        </span>
        <time
          dateTime={accessSource.createdAt}
          className="text-xs text-txt-tertiary"
        >
          {dateTimeFormat(i18n.language, accessSource.createdAt)}
        </time>
      </div>

      <div className={trailing()}>
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
              onSelect={handleOrgChange}
              reconnectUrl={reconnectUrl}
            />
          </Suspense>
        )}
        {showStandaloneIssue && (
          <SourceConnectionIssue
            provider={sourceLabel(accessSource.connector?.provider, t)}
            reconnectUrl={showReconnect ? reconnectUrl : null}
          />
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
    </li>
  );
}

function InlineOrgSelect({
  accessReviewSourceId,
  onSelect,
  reconnectUrl,
}: {
  accessReviewSourceId: string;
  reconnectUrl: string | null;
  onSelect: (slug: string) => void;
}) {
  const { t } = useTranslation();
  const data = useLazyLoadQuery<AccessReviewSourceListItemOrgsQuery>(
    orgsQuery,
    { accessReviewSourceId },
    { fetchPolicy: "store-or-network" },
  );

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
    case "NOT_APPLICABLE":
      return <CapturedOrganization sourceKey={source} />;
    case "EMPTY":
      return (
        <ProviderOrganizationsEmpty sourceKey={source} onSubmit={onSelect} />
      );
    // UNAVAILABLE, plus the impossible case of a node that is not an
    // AccessReviewSource: either way the list could not be read.
    default:
      return (
        <ProviderOrganizationsUnavailable
          sourceKey={source}
          reconnectUrl={reconnectUrl}
        />
      );
  }
}

function SourceConnectionIssue({
  provider,
  reconnectUrl,
}: {
  provider: string;
  reconnectUrl: string | null;
}) {
  const { t } = useTranslation();
  const {
    issue,
    issueIcon,
    issueContent,
    issueTitle,
    issueDescription,
  } = accessReviewSourceSection();

  return (
    <div className={issue()}>
      <IconWarning size={16} className={issueIcon()} />
      <div className={issueContent()}>
        <p className={issueTitle()}>
          {reconnectUrl
            ? t("accessReviewSourceRow.organizations.unavailable.reconnectTitle", {
                provider,
              })
            : t(
                "accessReviewSourceRow.organizations.unavailable.credentialsTitle",
                { provider },
              )}
        </p>
        <p className={issueDescription()}>
          {reconnectUrl
            ? t(
                "accessReviewSourceRow.organizations.unavailable.reconnectDescription",
              )
            : t(
                "accessReviewSourceRow.organizations.unavailable.credentialsDescription",
              )}
        </p>
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
  const providerLabel = sourceLabel(source.connector?.provider ?? null, t);
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
  reconnectUrl,
}: {
  sourceKey: AccessReviewSourceListItemOrganizationsUnavailable_source$key;
  reconnectUrl: string | null;
}) {
  const { t } = useTranslation();
  const source = useFragment(organizationsUnavailableFragment, sourceKey);
  const providerLabel = sourceLabel(source.connector?.provider ?? null, t);
  const canReconnect = (source.connector?.oauth2Scopes.length ?? 0) > 0;

  return (
    <SourceConnectionIssue
      provider={providerLabel}
      reconnectUrl={canReconnect ? reconnectUrl : null}
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
