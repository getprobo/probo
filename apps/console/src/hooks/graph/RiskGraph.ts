import { graphql } from "relay-runtime";
import { useTranslate } from "@probo/i18n";
import { useMutationWithToasts } from "../useMutationWithToasts.ts";
import type { RiskGraphDeleteMutation } from "/__generated__/core/RiskGraphDeleteMutation.graphql.ts";
import {
  usePreloadedQuery,
  usePaginationFragment,
  type PreloadedQuery,
} from "react-relay";
import type { RiskGraphListQuery } from "/__generated__/core/RiskGraphListQuery.graphql.ts";
import type { RiskGraphFragment$key } from "/__generated__/core/RiskGraphFragment.graphql.ts";

const deleteRiskMutation = graphql`
  mutation RiskGraphDeleteMutation(
    $input: DeleteRiskInput!
    $connections: [ID!]!
  ) {
    deleteRisk(input: $input) {
      deletedRiskId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteRiskMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<RiskGraphDeleteMutation>(deleteRiskMutation, {
    successMessage: __("Risk deleted successfully."),
    errorMessage: __("Failed to delete risk"),
  });
}

export const risksQuery = graphql`
  query RiskGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    organization: node(id: $organizationId) {
      id
      ...RiskGraphFragment @arguments(snapshotId: $snapshotId)
    }
  }
`;

const risksFragment = graphql`
  fragment RiskGraphFragment on Organization
  @refetchable(queryName: "RisksListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "RiskOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    risks(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: { snapshotId: $snapshotId }
    ) @connection(key: "RisksListQuery_risks", filters: ["filter"]) {
      __id
      edges {
        node {
          id
          snapshotId
          name
          category
          treatment
          owner {
            id
            fullName
          }
          inherentLikelihood
          inherentImpact
          residualLikelihood
          residualImpact
          inherentRiskScore
          residualRiskScore
          ...useRiskFormFragment
        }
      }
    }
  }
`;

export const RisksConnectionKey = "RisksListQuery_risks";

export function useRisksQuery(queryRef: PreloadedQuery<RiskGraphListQuery>) {
  const data = usePreloadedQuery(risksQuery, queryRef);
  const pagination = usePaginationFragment(
    risksFragment,
    data.organization as RiskGraphFragment$key,
  );
  const risks = pagination.data?.risks?.edges.map((edge) => edge.node);

  return {
    ...pagination,
    risks,
    connectionId: pagination.data.risks.__id,
  };
}

export const riskNodeQuery = graphql`
  query RiskGraphNodeQuery($riskId: ID!) {
    node(id: $riskId) {
      ... on Risk {
        id
        snapshotId
        name
        description
        treatment
        owner {
          id
          fullName
        }
        note
        inherentRiskScore
        residualRiskScore
        measuresInfo: measures(first: 0) {
          totalCount
        }
        documentsInfo: documents(first: 0) {
          totalCount
        }
        controlsInfo: controls(first: 0) {
          totalCount
        }
        obligationsInfo: obligations(first: 0) {
          totalCount
        }
        ...useRiskFormFragment
        ...RiskOverviewTabFragment
        ...RiskMeasuresTabFragment
        ...RiskDocumentsTabFragment
        ...RiskControlsTabFragment
        ...RiskObligationsTabFragment
      }
    }
  }
`;
