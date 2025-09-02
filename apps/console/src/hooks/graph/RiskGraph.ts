import { ConnectionHandler, graphql } from "relay-runtime";
import { useTranslate } from "@probo/i18n";
import { useMutationWithToasts } from "../useMutationWithToasts.ts";
import type { RiskGraphDeleteMutation } from "./__generated__/RiskGraphDeleteMutation.graphql.ts";
import {
  usePreloadedQuery,
  useRefetchableFragment,
  type PreloadedQuery,
} from "react-relay";
import type { RiskGraphListQuery } from "./__generated__/RiskGraphListQuery.graphql.ts";
import type { RiskGraphFragment$key } from "./__generated__/RiskGraphFragment.graphql.ts";

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
    errorMessage: __("Failed to delete risk. Please try again."),
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
    order: { type: "RiskOrder", defaultValue: null }
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
    ) @connection(key: "RisksListQuery_risks") {
      edges {
        node {
          id
          snapshotId
          name
          category
          treatment
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
  const [dataFragment, refetch] = useRefetchableFragment(
    risksFragment,
    data.organization as RiskGraphFragment$key,
  );
  const risks = dataFragment?.risks?.edges.map((edge) => edge.node);
  return {
    risks,
    refetch,
    connectionId: ConnectionHandler.getConnectionID(
      data.organization.id,
      RisksConnectionKey,
    ),
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
        ...useRiskFormFragment
        ...RiskOverviewTabFragment
        ...RiskMeasuresTabFragment
        ...RiskDocumentsTabFragment
        ...RiskControlsTabFragment
      }
    }
  }
`;
