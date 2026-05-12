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

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  IconPlusLarge,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
  useToast,
} from "@probo/ui";
import { useEffect, useMemo, useRef } from "react";
import type { PreloadedQuery } from "react-relay";
import { graphql, useFragment, useMutation, usePaginationFragment, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";

import type { AccessReviewSourcesTabFragment$key } from "#/__generated__/core/AccessReviewSourcesTabFragment.graphql";
import type { AccessReviewSourcesTabPaginationQuery } from "#/__generated__/core/AccessReviewSourcesTabPaginationQuery.graphql";
import type { AccessReviewSourcesTabQuery } from "#/__generated__/core/AccessReviewSourcesTabQuery.graphql";
import type { accessSourceMutationsCreateMutation } from "#/__generated__/core/accessSourceMutationsCreateMutation.graphql";
import type { AddAccessSourceDialogConnectorProviderInfoFragment$key } from "#/__generated__/core/AddAccessSourceDialogConnectorProviderInfoFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AccessSourceRow } from "../_components/AccessSourceRow";
import { createAccessSourceMutation } from "../dialogs/accessSourceMutations";
import { AddAccessSourceDialog, addAccessSourceDialogConnectorProviderInfoFragment } from "../dialogs/AddAccessSourceDialog";

export const accessReviewSourcesTabQuery = graphql`
  query AccessReviewSourcesTabQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateSource: permission(action: "core:access-source:create")
        connectorProviderInfos {
          ...AddAccessSourceDialogConnectorProviderInfoFragment
        }
        ...AccessReviewSourcesTabFragment
      }
    }
  }
`;

const sourcesFragment = graphql`
  fragment AccessReviewSourcesTabFragment on Organization
  @refetchable(queryName: "AccessReviewSourcesTabPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "AccessSourceOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    accessSources(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "AccessReviewSourcesTab_accessSources") {
      __id
      edges {
        node {
          id
          connectorId
          connector {
            provider
          }
          ...AccessSourceRowFragment
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<AccessReviewSourcesTabQuery>;
};

export default function AccessReviewSourcesTab({ queryRef }: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const [searchParams, setSearchParams] = useSearchParams();
  const processedConnectorIdRef = useRef<string | null>(null);

  const { organization } = usePreloadedQuery(accessReviewSourcesTabQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  const connectorProviderInfos = useFragment<AddAccessSourceDialogConnectorProviderInfoFragment$key>(
    addAccessSourceDialogConnectorProviderInfoFragment,
    organization.connectorProviderInfos,
  );

  const {
    data: { accessSources },
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    AccessReviewSourcesTabPaginationQuery,
    AccessReviewSourcesTabFragment$key
  >(sourcesFragment, organization);

  const existingSourceProviders = useMemo(
    () =>
      accessSources.edges
        .map(edge => edge.node.connector?.provider)
        .filter((p): p is NonNullable<typeof p> => p != null),
    [accessSources.edges],
  );

  const [createAccessSource, isCreatingSource]
    = useMutation<accessSourceMutationsCreateMutation>(
      createAccessSourceMutation,
    );

  // Handle OAuth callback: after the provider redirects back with connector_id,
  // automatically create the access source for that connector.
  const callbackConnectorId = searchParams.get("connector_id");
  const callbackProvider = searchParams.get("provider");
  const hasSourceForCallback = !!callbackConnectorId
    && accessSources?.edges.some(edge => edge.node.connectorId === callbackConnectorId);

  useEffect(() => {
    if (!callbackConnectorId) return;

    if (hasSourceForCallback) {
      setSearchParams((params) => {
        params.delete("connector_id");
        params.delete("provider");
        return params;
      }, { replace: true });
      return;
    }

    if (processedConnectorIdRef.current === callbackConnectorId || isCreatingSource) {
      return;
    }
    processedConnectorIdRef.current = callbackConnectorId;

    const providerInfo = callbackProvider
      ? connectorProviderInfos.find(p => p.provider === callbackProvider)
      : null;
    const sourceName = providerInfo?.displayName ?? callbackProvider ?? "Source";

    createAccessSource({
      variables: {
        input: {
          organizationId,
          connectorId: callbackConnectorId,
          name: sourceName,
          csvData: null,
        },
        connections: [accessSources.__id],
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          processedConnectorIdRef.current = null;
          setSearchParams((params) => {
            params.delete("connector_id");
            params.delete("provider");
            return params;
          }, { replace: true });
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to create access source"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Access source created successfully."),
          variant: "success",
        });
        setSearchParams((params) => {
          params.delete("connector_id");
          params.delete("provider");
          return params;
        }, { replace: true });
      },
      onError(error) {
        processedConnectorIdRef.current = null;
        setSearchParams((params) => {
          params.delete("connector_id");
          params.delete("provider");
          return params;
        }, { replace: true });
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create access source"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  }, [
    __,
    callbackConnectorId,
    callbackProvider,
    connectorProviderInfos,
    createAccessSource,
    hasSourceForCallback,
    isCreatingSource,
    organizationId,
    accessSources.__id,
    setSearchParams,
    toast,
  ]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
        {organization.canCreateSource && (
          <AddAccessSourceDialog
            organizationId={organizationId}
            connectionId={accessSources.__id}
            providerInfos={connectorProviderInfos}
            existingSourceProviders={existingSourceProviders}
          >
            <Button icon={IconPlusLarge}>
              {__("Add source")}
            </Button>
          </AddAccessSourceDialog>
        )}
      </div>

      {accessSources && accessSources.edges.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Name")}</Th>
                    <Th>{__("Source")}</Th>
                    <Th>{__("Status")}</Th>
                    <Th>{__("Organization")}</Th>
                    <Th>{__("Created at")}</Th>
                    <Th className="w-12"></Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {accessSources.edges.map(edge => (
                    <AccessSourceRow
                      key={edge.node.id}
                      fKey={edge.node}
                      connectionId={accessSources.__id}
                      organizationId={organizationId}
                    />
                  ))}
                </Tbody>
              </Table>

              {hasNext && (
                <div className="p-4 border-t">
                  <Button
                    variant="secondary"
                    onClick={() => loadNext(50)}
                    disabled={isLoadingNext}
                  >
                    {isLoadingNext
                      ? __("Loading...")
                      : __("Load more")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-8">
                <p className="text-txt-tertiary">
                  {__("No access sources configured yet. Add your first source to start reviewing access.")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
