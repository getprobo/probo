import {
  formatError,
  type GraphQLError,
  promisifyMutation,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm, useToast } from "@probo/ui";
import { useMemo } from "react";
import {
  type PreloadedQuery,
  useLazyLoadQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { PeopleGraphDeleteMutation } from "/__generated__/core/PeopleGraphDeleteMutation.graphql";
import type { PeopleGraphPaginatedFragment$key } from "/__generated__/core/PeopleGraphPaginatedFragment.graphql";
import type { PeopleGraphPaginatedQuery } from "/__generated__/core/PeopleGraphPaginatedQuery.graphql";
import type { PeopleGraphQuery } from "/__generated__/core/PeopleGraphQuery.graphql";

export const peopleQuery = graphql`
  query PeopleGraphQuery($organizationId: ID!, $filter: PeopleFilter) {
    organization: node(id: $organizationId) {
      ... on Organization {
        peoples(
          first: 1000
          orderBy: { direction: ASC, field: FULL_NAME }
          filter: $filter
        ) {
          edges {
            node {
              id
              fullName
              primaryEmailAddress
            }
          }
        }
      }
    }
  }
`;

/**
 * Return a list of people (used for people selectors)
 */
export function usePeople(
  organizationId: string,
  { excludeContractEnded }: { excludeContractEnded?: boolean } = {},
) {
  const data = useLazyLoadQuery<PeopleGraphQuery>(
    peopleQuery,
    {
      organizationId: organizationId,
      filter: excludeContractEnded ? { excludeContractEnded: true } : null,
    },
    { fetchPolicy: "network-only" },
  );
  return useMemo(() => {
    return data.organization?.peoples?.edges.map(edge => edge.node) ?? [];
  }, [data]);
}

export const paginatedPeopleQuery = graphql`
  query PeopleGraphPaginatedQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...PeopleGraphPaginatedFragment
      }
    }
  }
`;

export const paginatedPeopleFragment = graphql`
  fragment PeopleGraphPaginatedFragment on Organization
  @refetchable(queryName: "PeopleListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "PeopleOrder"
      defaultValue: { direction: ASC, field: FULL_NAME }
    }
    filter: { type: "PeopleFilter", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    canCreatePeople: permission(action: "core:people:create")
    peoples(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "PeopleGraphPaginatedQuery_peoples") {
      __id
      edges {
        node {
          id
          fullName
          primaryEmailAddress
          kind
          position
          additionalEmailAddresses
          contractStartDate
          contractEndDate
          canDelete: permission(action: "core:people:delete")
          canUpdate: permission(action: "core:people:update")
        }
      }
    }
  }
`;

export function usePeopleQuery(
  queryRef: PreloadedQuery<PeopleGraphPaginatedQuery>,
) {
  const data = usePreloadedQuery(paginatedPeopleQuery, queryRef);
  const pagination = usePaginationFragment(
    paginatedPeopleFragment,
    data.organization as PeopleGraphPaginatedFragment$key,
  );
  const people = pagination.data.peoples?.edges.map(edge => edge.node);
  return {
    ...pagination,
    people,
    connectionId: pagination.data.peoples.__id,
  };
}

export const deletePeopleMutation = graphql`
  mutation PeopleGraphDeleteMutation(
    $input: DeletePeopleInput!
    $connections: [ID!]!
  ) {
    deletePeople(input: $input) {
      deletedPeopleId @deleteEdge(connections: $connections)
    }
  }
`;

export const PeopleConnectionKey = "PeopleGraphPaginatedQuery_peoples";

export const useDeletePeople = (
  people: { id?: string; fullName?: string },
  connectionId: string,
) => {
  const [mutate] = useMutation<PeopleGraphDeleteMutation>(deletePeopleMutation);
  const confirm = useConfirm();
  const { toast } = useToast();
  const { __ } = useTranslate();

  return () => {
    if (!people.id || !people.fullName) {
      return alert(__("Failed to delete people: missing id or fullName"));
    }
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: {
            input: {
              peopleId: people.id!,
            },
            connections: [connectionId],
          },
        }).catch((error) => {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to delete people"),
              error as GraphQLError,
            ),
            variant: "error",
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete \"%s\". This action cannot be undone.",
          ),
          people.fullName,
        ),
      },
    );
  };
};

export const peopleNodeQuery = graphql`
  query PeopleGraphNodeQuery($peopleId: ID!) {
    node(id: $peopleId) {
      ... on People {
        id
        fullName
        primaryEmailAddress
        position
        kind
        additionalEmailAddresses
        contractStartDate
        contractEndDate
        canUpdate: permission(action: "core:people:update")
        canDelete: permission(action: "core:people:delete")
      }
    }
  }
`;

export const updatePeopleMutation = graphql`
  mutation PeopleGraphUpdateMutation($input: UpdatePeopleInput!) {
    updatePeople(input: $input) {
      people {
        id
        fullName
        primaryEmailAddress
        position
        kind
        additionalEmailAddresses
        contractStartDate
        contractEndDate
      }
    }
  }
`;
