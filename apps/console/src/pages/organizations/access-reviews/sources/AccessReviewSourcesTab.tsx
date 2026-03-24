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
} from "@probo/ui";
import { graphql, usePaginationFragment } from "react-relay";
import { Link, useOutletContext } from "react-router";

import type { AccessReviewSourcesTabFragment$key } from "#/__generated__/core/AccessReviewSourcesTabFragment.graphql";
import type { AccessReviewSourcesTabPaginationQuery } from "#/__generated__/core/AccessReviewSourcesTabPaginationQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AccessSourceRow } from "../_components/AccessSourceRow";

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

export default function AccessReviewSourcesTab() {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { organizationRef, canCreateSource } = useOutletContext<{
    organizationRef: AccessReviewSourcesTabFragment$key;
    canCreateSource: boolean;
  }>();

  const {
    data: { accessSources },
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    AccessReviewSourcesTabPaginationQuery,
    AccessReviewSourcesTabFragment$key
  >(sourcesFragment, organizationRef);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
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
