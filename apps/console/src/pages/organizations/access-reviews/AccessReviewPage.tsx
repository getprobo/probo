import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  IconPlusLarge,
  PageHeader,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link } from "react-router";

import type { AccessReviewPageQuery } from "#/__generated__/core/AccessReviewPageQuery.graphql";
import type { AccessReviewPageSourcesFragment$key } from "#/__generated__/core/AccessReviewPageSourcesFragment.graphql";
import type { AccessReviewPageSourcesPaginationQuery } from "#/__generated__/core/AccessReviewPageSourcesPaginationQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AccessSourceRow } from "./_components/AccessSourceRow";

export const accessReviewPageQuery = graphql`
  query AccessReviewPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateSource: permission(action: "core:access-source:create")
        ...AccessReviewPageSourcesFragment
      }
    }
  }
`;

export const sourcesPaginatedFragment = graphql`
  fragment AccessReviewPageSourcesFragment on Organization
  @refetchable(queryName: "AccessReviewPageSourcesPaginationQuery")
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
    ) @connection(key: "AccessReviewPage_accessSources") {
      __id
      edges {
        node {
          id
          name
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

export default function AccessReviewPage({
  queryRef,
}: {
  queryRef: PreloadedQuery<AccessReviewPageQuery>;
}) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  usePageTitle(__("Access Reviews"));

  const { organization } = usePreloadedQuery(accessReviewPageQuery, queryRef);

  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  return (
    <AccessReviewContent
      organization={organization}
      organizationId={organizationId}
      canCreateSource={organization.canCreateSource}
    />
  );
}

function AccessReviewContent({
  organization,
  organizationId,
  canCreateSource,
}: {
  organization: AccessReviewPageSourcesFragment$key;
  organizationId: string;
  canCreateSource: boolean;
}) {
  const { __ } = useTranslate();

  const {
    data: { accessSources },
    loadNext: loadNextSources,
    hasNext: hasNextSources,
    isLoadingNext: isLoadingNextSources,
  } = usePaginationFragment<
    AccessReviewPageSourcesPaginationQuery,
    AccessReviewPageSourcesFragment$key
  >(sourcesPaginatedFragment, organization);

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Access Reviews")}
        description={__(
          "Review and manage user access across your organization's systems and applications.",
        )}
      />

      {/* Access Sources Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-medium">{__("Access Sources")}</h2>
          {canCreateSource && (
            <Button icon={IconPlusLarge} asChild>
              <Link to={`/organizations/${organizationId}/access-reviews/sources/new`}>
                {__("Add source")}
              </Link>
            </Button>
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
                      />
                    ))}
                  </Tbody>
                </Table>

                {hasNextSources && (
                  <div className="p-4 border-t">
                    <Button
                      variant="secondary"
                      onClick={() => loadNextSources(50)}
                      disabled={isLoadingNextSources}
                    >
                      {isLoadingNextSources
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
    </div>
  );
}
