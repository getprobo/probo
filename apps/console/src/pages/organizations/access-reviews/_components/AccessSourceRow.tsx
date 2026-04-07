// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatDate, formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  DropdownItem,
  IconTrashCan,
  Input,
  Option,
  Select,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Suspense, useState } from "react";
import { useFragment, useLazyLoadQuery, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessSourceRowConfigureMutation } from "#/__generated__/core/AccessSourceRowConfigureMutation.graphql";
import type { AccessSourceRowDeleteMutation } from "#/__generated__/core/AccessSourceRowDeleteMutation.graphql";
import type { AccessSourceRowFragment$key } from "#/__generated__/core/AccessSourceRowFragment.graphql";
import type { AccessSourceRowOrgsQuery } from "#/__generated__/core/AccessSourceRowOrgsQuery.graphql";

const fragment = graphql`
  fragment AccessSourceRowFragment on AccessSource {
    id
    name
    connectorId
    connector {
      provider
    }
    oauth2Scopes
    connectionStatus
    selectedOrganization
    needsConfiguration
    createdAt
    canDelete: permission(action: "core:access-source:delete")
  }
`;

export const deleteAccessSourceMutation = graphql`
  mutation AccessSourceRowDeleteMutation(
    $input: DeleteAccessSourceInput!
    $connections: [ID!]!
  ) {
    deleteAccessSource(input: $input) {
      deletedAccessSourceId @deleteEdge(connections: $connections)
    }
  }
`;

const configureMutation = graphql`
  mutation AccessSourceRowConfigureMutation(
    $input: ConfigureAccessSourceInput!
  ) {
    configureAccessSource(input: $input) {
      accessSource {
        id
        selectedOrganization
        needsConfiguration
      }
    }
  }
`;

const orgsQuery = graphql`
  query AccessSourceRowOrgsQuery($accessSourceId: ID!) {
    node(id: $accessSourceId) @required(action: THROW) {
      ... on AccessSource {
        providerOrganizations {
          slug
          displayName
        }
      }
    }
  }
`;

type Props = {
  fKey: AccessSourceRowFragment$key;
  connectionId: string;
  organizationId: string;
};

function sourceLabel(connectorProvider: string | null | undefined): string {
  if (!connectorProvider) {
    return "CSV";
  }

  switch (connectorProvider) {
    case "GOOGLE_WORKSPACE":
      return "Google Workspace";
    case "LINEAR":
      return "Linear";
    case "SLACK":
      return "Slack";
    default:
      return connectorProvider;
  }
}

export function AccessSourceRow({ fKey, connectionId, organizationId }: Props) {
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const { toast } = useToast();

  const accessSource = useFragment(fragment, fKey);

  const [deleteAccessSource] = useMutation<AccessSourceRowDeleteMutation>(deleteAccessSourceMutation);
  const [configure] = useMutation<AccessSourceRowConfigureMutation>(configureMutation);

  const handleDelete = () => {
    confirm(
      () => {
        deleteAccessSource({
          variables: {
            input: { accessSourceId: accessSource.id },
            connections: [connectionId],
          },
          onCompleted: (_response, errors) => {
            if (errors?.length) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete access source"),
                  errors as GraphQLError[],
                ),
                variant: "error",
              });
            }
          },
          onError: (error) => {
            toast({
              title: __("Error"),
              description: formatError(
                __("Failed to delete access source"),
                error as GraphQLError,
              ),
              variant: "error",
            });
          },
        });
      },
      {
        message: sprintf(
          __("This will permanently delete \"%s\". This action cannot be undone."),
          accessSource.name,
        ),
      },
    );
  };

  const handleOrgChange = (slug: string) => {
    configure({
      variables: {
        input: {
          accessSourceId: accessSource.id,
          organizationSlug: slug,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to configure source"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Organization updated."),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to configure source"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleReconnect = () => {
    const provider = accessSource.connector?.provider;
    if (!provider || !accessSource.connectorId) return;

    const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseURL);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", provider);
    url.searchParams.append("connector_id", accessSource.connectorId);
    for (const scope of accessSource.oauth2Scopes) {
      url.searchParams.append("scope", scope);
    }
    url.searchParams.append(
      "continue",
      `/organizations/${organizationId}/access-reviews/sources`,
    );
    window.location.href = url.toString();
  };

  const showOrgSelector = accessSource.needsConfiguration || accessSource.selectedOrganization;

  return (
    <Tr>
      <Td>{accessSource.name}</Td>
      <Td>
        <Badge variant="neutral" size="sm">
          {sourceLabel(accessSource.connector?.provider ?? null)}
        </Badge>
      </Td>
      <Td>
        {accessSource.connectionStatus === "CONNECTED" && (
          <Badge variant="success" size="sm">{__("Connected")}</Badge>
        )}
        {accessSource.connectionStatus === "DISCONNECTED" && (
          <div className="flex items-center gap-2">
            <Badge variant="danger" size="sm">{__("Disconnected")}</Badge>
            <Button variant="secondary" onClick={handleReconnect}>
              {__("Reconnect")}
            </Button>
          </div>
        )}
      </Td>
      <Td>
        {showOrgSelector && (
          <Suspense
            fallback={
              <Select variant="editor" disabled placeholder={__("Loading...")} />
            }
          >
            <InlineOrgSelect
              accessSourceId={accessSource.id}
              selectedOrganization={accessSource.selectedOrganization ?? ""}
              onSelect={handleOrgChange}
            />
          </Suspense>
        )}
      </Td>
      <Td>
        <time dateTime={accessSource.createdAt}>
          {formatDate(accessSource.createdAt)}
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
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}

function InlineOrgSelect({
  accessSourceId,
  selectedOrganization,
  onSelect,
}: {
  accessSourceId: string;
  selectedOrganization: string;
  onSelect: (slug: string) => void;
}) {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<AccessSourceRowOrgsQuery>(
    orgsQuery,
    { accessSourceId },
    { fetchPolicy: "store-or-network" },
  );

  const orgs = data.node.providerOrganizations ?? [];

  if (orgs.length === 0) {
    return (
      <ManualOrgInput
        selectedOrganization={selectedOrganization}
        onSubmit={onSelect}
      />
    );
  }

  return (
    <Select
      variant="editor"
      placeholder={__("Select organization")}
      value={selectedOrganization}
      onValueChange={onSelect}
    >
      {orgs.map(org => (
        <Option key={org.slug} value={org.slug}>
          {org.displayName}
        </Option>
      ))}
    </Select>
  );
}

function ManualOrgInput({
  selectedOrganization,
  onSubmit,
}: {
  selectedOrganization: string;
  onSubmit: (slug: string) => void;
}) {
  const { __ } = useTranslate();
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
      placeholder={__("org-slug")}
      value={value}
      onChange={e => setValue(e.target.value)}
      onBlur={handleBlur}
      onKeyDown={handleKeyDown}
      className="max-w-40"
    />
  );
}
