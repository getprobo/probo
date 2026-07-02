// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconPlusLarge,
  IconTrashCan,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  TrButton,
} from "@probo/ui";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { RiskScenariosPageLinkMutation } from "#/__generated__/core/RiskScenariosPageLinkMutation.graphql";
import type { RiskScenariosPageQuery } from "#/__generated__/core/RiskScenariosPageQuery.graphql";
import type { RiskScenariosPageUnlinkMutation } from "#/__generated__/core/RiskScenariosPageUnlinkMutation.graphql";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkScenarioDialog } from "../_components/LinkScenarioDialog";

export const riskScenariosPageQuery = graphql`
  query RiskScenariosPageQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        id
        scenarios(first: 100)
          @connection(key: "RiskScenariosPage_scenarios", filters: []) {
          __id
          edges {
            node {
              id
              name
              description
              scope { riskAssessmentId }
            }
          }
        }
      }
    }
  }
`;

const linkMutation = graphql`
  mutation RiskScenariosPageLinkMutation(
    $input: LinkRiskAssessmentScenarioRiskInput!
    $connections: [ID!]!
  ) {
    linkRiskAssessmentScenarioRisk(input: $input) {
      riskAssessmentScenarioEdge @appendEdge(connections: $connections) {
        node {
          id
          name
          description
          scope { riskAssessmentId }
        }
      }
    }
  }
`;

const unlinkMutation = graphql`
  mutation RiskScenariosPageUnlinkMutation(
    $input: UnlinkRiskAssessmentScenarioRiskInput!
    $connections: [ID!]!
  ) {
    unlinkRiskAssessmentScenarioRisk(input: $input) {
      deletedRiskAssessmentScenarioId @deleteEdge(connections: $connections)
    }
  }
`;

interface RiskScenariosPageProps {
  queryRef: PreloadedQuery<RiskScenariosPageQuery>;
}

export default function RiskScenariosPage(props: RiskScenariosPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<RiskScenariosPageQuery>(riskScenariosPageQuery, props.queryRef);
  if (data.node?.__typename !== "Risk") {
    throw new Error("Risk not found");
  }
  const risk = data.node;
  const scenarios = risk.scenarios.edges.map(e => e.node);
  const connectionId = risk.scenarios.__id;
  const riskId = risk.id;

  const incrementOptions = {
    id: riskId,
    node: "scenarios(first:0)",
  };

  const [linkScenario, isLinking] = useMutationWithIncrement<RiskScenariosPageLinkMutation>(
    linkMutation,
    {
      ...incrementOptions,
      value: 1,
    },
  );

  const [unlinkScenario, isUnlinking] = useMutationWithIncrement<RiskScenariosPageUnlinkMutation>(
    unlinkMutation,
    {
      ...incrementOptions,
      value: -1,
    },
  );

  const isLoading = isLinking || isUnlinking;

  const onLink = (scenarioId: string) => {
    linkScenario({
      variables: {
        input: {
          riskAssessmentScenarioId: scenarioId,
          riskId,
        },
        connections: [connectionId],
      },
    });
  };

  const onUnlink = (scenarioId: string) => {
    unlinkScenario({
      variables: {
        input: {
          riskAssessmentScenarioId: scenarioId,
          riskId,
        },
        connections: [connectionId],
      },
    });
  };

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>{__("Scenario")}</Th>
          <Th>{__("Description")}</Th>
          <Th className="w-12" />
        </Tr>
      </Thead>
      <Tbody>
        {scenarios.length === 0 && (
          <Tr>
            <Td colSpan={3} className="text-center text-txt-secondary">
              {__("No scenarios linked to this risk yet.")}
            </Td>
          </Tr>
        )}
        {scenarios.map(scenario => (
          <Tr key={scenario.id} to={`/organizations/${organizationId}/risk-assessments/${scenario.scope?.riskAssessmentId}`}>
            <Td className="font-medium">{scenario.name}</Td>
            <Td className="text-txt-secondary">
              {scenario.description || "—"}
            </Td>
            <Td noLink width={50} className="text-end">
              <Button
                variant="secondary"
                icon={IconTrashCan}
                onClick={() => onUnlink(scenario.id)}
                disabled={isLoading}
              >
                {__("Unlink")}
              </Button>
            </Td>
          </Tr>
        ))}
        <LinkScenarioDialog
          connectionId={connectionId}
          disabled={isLoading}
          linkedScenarios={scenarios}
          onLink={onLink}
          onUnlink={onUnlink}
        >
          <TrButton colspan={3} icon={IconPlusLarge}>
            {__("Link Scenario")}
          </TrButton>
        </LinkScenarioDialog>
      </Tbody>
    </Table>
  );
}
