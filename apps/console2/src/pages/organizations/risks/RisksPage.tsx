import {
  Button,
  PageHeader,
  Td,
  Tr,
  Th,
  Thead,
  Tbody,
  Table,
  RiskBadge,
  SeverityBadge,
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
  IconPencil,
  ConfirmDialog,
  RisksChart,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { IconPlusLarge } from "@probo/ui";
import FormRiskDialog from "./FormRiskDialog";
import { ConnectionHandler, graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import type {
  RisksPageQuery as RisksPageQueryType,
  RiskTreatment,
} from "./__generated__/RisksPageQuery.graphql";
import { useOrganizationId } from "../../../hooks/useOrganizationId";
import { useMemo, useState } from "react";
import { usePageTitle } from "@probo/hooks";
import type { RisksPageDeleteMutation } from "./__generated__/RisksPageDeleteMutation.graphql";
import { useMutationWithToasts } from "../../../hooks/useMutationWithToasts";
import { sprintf } from "@probo/helpers";
import type { ItemOf } from "../../../types";

const risksQuery = graphql`
  query RisksPageQuery(
    $organizationId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
  ) {
    organization: node(id: $organizationId) {
      ... on Organization {
        risks(first: $first, after: $after, last: $last, before: $before)
          @connection(key: "RisksPage_risks") {
          edges {
            node {
              id
              name
              category
              treatment
              inherentLikelihood
              inherentImpact
              residualLikelihood
              residualImpact
              inherentSeverity
              ...useRiskFormFragment
            }
          }
        }
      }
    }
  }
`;

const deleteRiskMutation = graphql`
  mutation RisksPageDeleteMutation(
    $input: DeleteRiskInput!
    $connections: [ID!]!
  ) {
    deleteRisk(input: $input) {
      deletedRiskId @deleteEdge(connections: $connections)
    }
  }
`;

export default function RisksPage() {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<RisksPageQueryType>(risksQuery, {
    organizationId: useOrganizationId(),
  });
  const risks = data.organization?.risks?.edges.map((edge) => edge.node);
  const [editedRisk, setEditedRisk] = useState<ItemOf<typeof risks> | null>(
    null
  );
  const treatmentLabels = useMemo(() => {
    return {
      MITIGATED: __("Mitigate"),
      ACCEPTED: __("Accept"),
      TRANSFERRED: __("Transfer"),
      AVOIDED: __("Avoid"),
    } as Record<RiskTreatment, string>;
  }, [__]);
  const organizationId = useOrganizationId();

  const [deleteRisk] =
    useMutationWithToasts<RisksPageDeleteMutation>(deleteRiskMutation);

  const onDelete = (riskId: string) => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      "RisksPage_risks"
    );
    deleteRisk({
      variables: {
        input: { riskId },
        connections: [connectionId],
      },
      successMessage: __("Risk deleted successfully."),
      errorMessage: __("Failed to delete risk. Please try again."),
    });
  };

  usePageTitle(__("Risks"));

  return (
    <div className="space-y-6">
      <PageHeader title={__("Risks")}>
        <FormRiskDialog
          trigger={<Button icon={IconPlusLarge}>{__("New Risk")}</Button>}
        />
      </PageHeader>
      {editedRisk && (
        <FormRiskDialog
          risk={editedRisk}
          onSuccess={() => setEditedRisk(null)}
        />
      )}

      <div className="grid grid-cols-2 gap-4">
        <RisksChart
          organizationId={organizationId}
          type="inherent"
          risks={risks}
        />
        <RisksChart
          organizationId={organizationId}
          type="residual"
          risks={risks}
        />
      </div>
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Risk name")}</Th>
            <Th>{__("Category")}</Th>
            <Th>{__("Treatment")}</Th>
            <Th>{__("Initial Risk")}</Th>
            <Th>{__("Residual Risk")}</Th>
            <Th>{__("Severity")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {risks?.map((risk) => (
            <Tr
              key={risk.id}
              to={`/organizations/${organizationId}/risks/${risk.id}`}
            >
              <Td>{risk.name}</Td>
              <Td>{risk.category}</Td>
              <Td>{treatmentLabels[risk.treatment]}</Td>
              <Td>
                <RiskBadge
                  score={risk.inherentLikelihood * risk.inherentImpact}
                />
              </Td>
              <Td>
                <RiskBadge
                  score={risk.residualLikelihood * risk.residualImpact}
                />
              </Td>
              <Td>
                <SeverityBadge severity={risk.inherentSeverity} />
              </Td>
              <Td noLink>
                <ActionDropdown>
                  <DropdownItem
                    icon={IconPencil}
                    onClick={() => setEditedRisk(risk)}
                  >
                    {__("Edit")}
                  </DropdownItem>
                  <ConfirmDialog
                    message={sprintf(
                      __(
                        'This will permanently delete the risk "%s". This action cannot be undone.'
                      ),
                      risk.name
                    )}
                    onConfirm={() => onDelete(risk.id)}
                  >
                    <DropdownItem variant="danger" icon={IconTrashCan}>
                      {__("Delete")}
                    </DropdownItem>
                  </ConfirmDialog>
                </ActionDropdown>
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </div>
  );
}
