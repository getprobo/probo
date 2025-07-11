import { graphql } from "relay-runtime";
import type { PeopleGraphQuery } from "./__generated__/PeopleGraphQuery.graphql";
import {
  useLazyLoadQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { useMemo } from "react";
import type { PeopleGraphPaginatedQuery } from "./__generated__/PeopleGraphPaginatedQuery.graphql";
import type { PeopleGraphPaginatedFragment$key } from "./__generated__/PeopleGraphPaginatedFragment.graphql";
import { useConfirm } from "@probo/ui";
import type { PeopleGraphDeleteMutation } from "./__generated__/PeopleGraphDeleteMutation.graphql";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";

const peopleQuery = graphql`
  query PeopleGraphQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        peoples(first: 100, orderBy: { direction: ASC, field: CREATED_AT }) {
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
export function usePeople(organizationId: string) {
  const data = useLazyLoadQuery<PeopleGraphQuery>(peopleQuery, {
    organizationId: organizationId,
  });
  return useMemo(() => {
    return data.organization?.peoples?.edges.map((edge) => edge.node) ?? [];
  }, [data]);
}

export const paginatedPeopleQuery = graphql`
  query PeopleGraphPaginatedQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
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
    order: { type: "PeopleOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    peoples(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "PeopleGraphPaginatedQuery_peoples") {
      __id
      edges {
        node {
          id
          fullName
          primaryEmailAddress
          kind
          additionalEmailAddresses
        }
      }
    }
  }
`;

export function usePeopleQuery(
  queryRef: PreloadedQuery<PeopleGraphPaginatedQuery>
) {
  const data = usePreloadedQuery(paginatedPeopleQuery, queryRef);
  const pagination = usePaginationFragment(
    paginatedPeopleFragment,
    data.organization as PeopleGraphPaginatedFragment$key
  );
  const people = pagination.data.peoples?.edges.map((edge) => edge.node);
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
  connectionId: string
) => {
  const [mutate] = useMutation<PeopleGraphDeleteMutation>(deletePeopleMutation);
  const confirm = useConfirm();
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
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete "%s". This action cannot be undone.'
          ),
          people.fullName
        ),
      }
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
        kind
        additionalEmailAddresses
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
        kind
        additionalEmailAddresses
      }
    }
  }
`;
