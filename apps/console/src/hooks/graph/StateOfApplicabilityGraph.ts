import { graphql } from "relay-runtime";
import type { StateOfApplicabilityGraphPaginatedQuery } from "/__generated__/core/StateOfApplicabilityGraphPaginatedQuery.graphql";
import type { StateOfApplicabilityGraphPaginatedFragment$key } from "/__generated__/core/StateOfApplicabilityGraphPaginatedFragment.graphql";
import {
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { useConfirm, useToast } from "@probo/ui";
import type { StateOfApplicabilityGraphDeleteMutation } from "/__generated__/core/StateOfApplicabilityGraphDeleteMutation.graphql";
import {
  promisifyMutation,
  sprintf,
  formatError,
  type GraphQLError,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";

export const paginatedStateOfApplicabilityQuery = graphql`
  query StateOfApplicabilityGraphPaginatedQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateStateOfApplicability: permission(action: "core:state-of-applicability:create")
        ...StateOfApplicabilityGraphPaginatedFragment
      }
    }
  }
`;

export const paginatedStateOfApplicabilityFragment = graphql`
  fragment StateOfApplicabilityGraphPaginatedFragment on Organization
  @refetchable(queryName: "StateOfApplicabilityListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "StateOfApplicabilityOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    filter: { type: "StateOfApplicabilityFilter", defaultValue: { snapshotId: null } }
  ) {
    statesOfApplicability(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "StateOfApplicabilityGraphPaginatedQuery_statesOfApplicability") {
      __id
      edges {
        node {
          id
          name
          sourceId
          snapshotId
          createdAt
          updatedAt

          canDelete: permission(action: "core:state-of-applicability:delete")

          controlsInfo: controls(first: 0) {
            totalCount
          }
        }
      }
    }
  }
`;

export function useStateOfApplicabilityQuery(
  queryRef: PreloadedQuery<StateOfApplicabilityGraphPaginatedQuery>,
) {
  const data = usePreloadedQuery(paginatedStateOfApplicabilityQuery, queryRef);
  const pagination = usePaginationFragment(
    paginatedStateOfApplicabilityFragment,
    data.organization as StateOfApplicabilityGraphPaginatedFragment$key,
  );
  const statesOfApplicability = pagination.data.statesOfApplicability?.edges.map((edge) => edge.node);
  return {
    ...pagination,
    statesOfApplicability,
    connectionId: pagination.data.statesOfApplicability.__id,
  };
}

export const deleteStateOfApplicabilityMutation = graphql`
  mutation StateOfApplicabilityGraphDeleteMutation(
    $input: DeleteStateOfApplicabilityInput!
    $connections: [ID!]!
  ) {
    deleteStateOfApplicability(input: $input) {
      deletedStateOfApplicabilityId @deleteEdge(connections: $connections)
    }
  }
`;

export const StateOfApplicabilityConnectionKey = "StateOfApplicabilityGraphPaginatedQuery_statesOfApplicability";

export const useDeleteStateOfApplicability = (
  stateOfApplicability: { id?: string; name?: string },
  connectionId: string,
  onSuccess?: () => void,
) => {
  const [mutate] = useMutation<StateOfApplicabilityGraphDeleteMutation>(deleteStateOfApplicabilityMutation);
  const confirm = useConfirm();
  const { toast } = useToast();
  const { __ } = useTranslate();

  return () => {
    if (!stateOfApplicability.id || !stateOfApplicability.name) {
      return alert(__("Failed to delete state of applicability: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: {
            input: {
              stateOfApplicabilityId: stateOfApplicability.id!,
            },
            connections: [connectionId],
          },
        })
          .then(() => {
            onSuccess?.();
          })
          .catch((error) => {
            toast({
              title: __("Error"),
              description: formatError(
                __("Failed to delete state of applicability"),
                error as GraphQLError,
              ),
              variant: "error",
            });
          }),
      {
        message: sprintf(
          __(
            'This will permanently delete "%s". This action cannot be undone.',
          ),
          stateOfApplicability.name,
        ),
      },
    );
  };
};

export const createStateOfApplicabilityMutation = graphql`
  mutation StateOfApplicabilityGraphCreateMutation(
    $input: CreateStateOfApplicabilityInput!
    $connections: [ID!]!
  ) {
    createStateOfApplicability(input: $input) {
      stateOfApplicabilityEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          sourceId
          snapshotId
          createdAt
          updatedAt
        }
      }
    }
  }
`;

export const updateStateOfApplicabilityMutation = graphql`
  mutation StateOfApplicabilityGraphUpdateMutation(
    $input: UpdateStateOfApplicabilityInput!
  ) {
    updateStateOfApplicability(input: $input) {
      stateOfApplicability {
        id
        name
        sourceId
        snapshotId
        createdAt
        updatedAt
        owner {
          id
          fullName
        }
      }
    }
  }
`;

export const stateOfApplicabilityNodeQuery = graphql`
  query StateOfApplicabilityGraphNodeQuery($stateOfApplicabilityId: ID!) {
    node(id: $stateOfApplicabilityId) {
      ... on StateOfApplicability {
        id
        name
        sourceId
        snapshotId
        createdAt
        updatedAt
        canUpdate: permission(action: "core:state-of-applicability:update")
        canDelete: permission(action: "core:state-of-applicability:delete")
        canExport: permission(action: "core:state-of-applicability:export")
        organization {
          id
        }
        owner {
          id
          fullName
        }
        ...StateOfApplicabilityControlsTabFragment
      }
    }
  }
`;

export const stateOfApplicabilityForEditQuery = graphql`
  query StateOfApplicabilityGraphForEditQuery($stateOfApplicabilityId: ID!) {
    node(id: $stateOfApplicabilityId) {
      ... on StateOfApplicability {
        id
        name
        controls(first: 1000, orderBy: { field: SECTION_TITLE, direction: ASC }) {
          edges {
            node {
              id
              sectionTitle
              name
              framework {
                id
                name
              }
            }
          }
        }
      }
    }
  }
`;
