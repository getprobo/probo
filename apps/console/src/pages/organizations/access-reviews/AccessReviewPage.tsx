import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Field,
  IconPlusLarge,
  Option,
  PageHeader,
  Select,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import {
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link } from "react-router";

import type { AccessReviewPageInitMutation } from "#/__generated__/core/AccessReviewPageInitMutation.graphql";
import type { AccessReviewPageQuery } from "#/__generated__/core/AccessReviewPageQuery.graphql";
import type { AccessReviewPageSourcesFragment$key } from "#/__generated__/core/AccessReviewPageSourcesFragment.graphql";
import type { AccessReviewPageSourcesPaginationQuery } from "#/__generated__/core/AccessReviewPageSourcesPaginationQuery.graphql";
import type { AccessReviewPageUpdateIdentitySourceMutation } from "#/__generated__/core/AccessReviewPageUpdateIdentitySourceMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AccessSourceRow } from "./_components/AccessSourceRow";

export const accessReviewPageQuery = graphql`
  query AccessReviewPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        accessReview {
          id
          identitySource {
            id
          }
          canCreateSource: permission(action: "core:access-source:create")
          canUpdateAccessReview: permission(action: "core:access-review:update")
          ...AccessReviewPageSourcesFragment
        }
      }
    }
  }
`;

const initMutation = graphql`
  mutation AccessReviewPageInitMutation($input: CreateAccessReviewInput!) {
    createAccessReview(input: $input) {
      accessReview {
        id
      }
    }
  }
`;

export const sourcesPaginatedFragment = graphql`
  fragment AccessReviewPageSourcesFragment on AccessReview
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

const updateIdentitySourceMutation = graphql`
  mutation AccessReviewPageUpdateIdentitySourceMutation(
    $input: UpdateAccessReviewInput!
  ) {
    updateAccessReview(input: $input) {
      accessReview {
        id
        identitySource {
          id
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

  const [createAccessReview, isCreating] = useMutation<AccessReviewPageInitMutation>(initMutation);

  if (!organization.accessReview) {
    return (
      <div className="space-y-6">
        <PageHeader
          title={__("Access Reviews")}
          description={__(
            "Review and manage user access across your organization's systems and applications.",
          )}
        />
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("Get started with Access Reviews")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__(
                "Set up Access Reviews to periodically verify user access to your systems.",
              )}
            </p>
            <Button
              icon={IconPlusLarge}
              disabled={isCreating}
              onClick={() => {
                createAccessReview({
                  variables: {
                    input: { organizationId },
                  },
                  updater: (store) => {
                    store.invalidateStore();
                  },
                });
              }}
            >
              {__("Enable Access Reviews")}
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <AccessReviewContent
      accessReview={organization.accessReview}
      organizationId={organizationId}
      accessReviewId={organization.accessReview.id}
      identitySourceId={organization.accessReview.identitySource?.id ?? null}
      canCreateSource={organization.accessReview.canCreateSource}
      canUpdateAccessReview={organization.accessReview.canUpdateAccessReview}
    />
  );
}

function AccessReviewContent({
  accessReview,
  organizationId,
  accessReviewId,
  identitySourceId,
  canCreateSource,
  canUpdateAccessReview,
}: {
  accessReview: AccessReviewPageSourcesFragment$key;
  organizationId: string;
  accessReviewId: string;
  identitySourceId: string | null;
  canCreateSource: boolean;
  canUpdateAccessReview: boolean;
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
  >(sourcesPaginatedFragment, accessReview);

  const [updateIdentitySource, isUpdatingIdentitySource]
    = useMutation<AccessReviewPageUpdateIdentitySourceMutation>(
      updateIdentitySourceMutation,
    );

  const NONE_VALUE = "__none__";

  const handleIdentitySourceChange = (value: string) => {
    updateIdentitySource({
      variables: {
        input: {
          accessReviewId,
          identitySourceId: value === NONE_VALUE ? null : value,
        },
      },
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Access Reviews")}
        description={__(
          "Review and manage user access across your organization's systems and applications.",
        )}
      />

      {/* Identity Provider Section */}
      {canUpdateAccessReview && accessSources && accessSources.edges.length > 0 && (
        <Card padded>
          <div className="space-y-2">
            <h2 className="text-base font-medium">{__("Identity Provider")}</h2>
            <p className="text-sm text-txt-tertiary">
              {__(
                "Select the access source that serves as your organization's identity provider. This source will be used to cross-reference user identities during access reviews.",
              )}
            </p>
            <div className="max-w-sm pt-2">
              <Field label={__("Identity source")}>
                <Select
                  value={identitySourceId ?? NONE_VALUE}
                  onValueChange={handleIdentitySourceChange}
                  disabled={isUpdatingIdentitySource}
                >
                  <Option value={NONE_VALUE}>{__("None")}</Option>
                  {accessSources.edges.map(edge => (
                    <Option key={edge.node.id} value={edge.node.id}>
                      {edge.node.name}
                    </Option>
                  ))}
                </Select>
              </Field>
            </div>
          </div>
        </Card>
      )}

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
