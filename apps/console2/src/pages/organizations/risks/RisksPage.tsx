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
import { graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import type { RisksPageQuery as RisksPageQueryType } from "./__generated__/RisksPageQuery.graphql";
import { useOrganizationId } from "../../../hooks/useOrganizationId";
import { useState } from "react";
import { usePageTitle } from "@probo/hooks";
import { getTreatment, sprintf } from "@probo/helpers";
import type { ItemOf } from "../../../types";
import { useDeleteRiskMutation } from "../../../graph/RiskGraph";

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
          __id
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

export default function RisksPage() {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<RisksPageQueryType>(risksQuery, {
    organizationId: useOrganizationId(),
  });
  const risks = data.organization?.risks?.edges.map((edge) => edge.node);
  const [editedRisk, setEditedRisk] = useState<ItemOf<typeof risks> | null>(
    null
  );
  const organizationId = useOrganizationId();

  const [deleteRisk] = useDeleteRiskMutation();
  const connectionId = data.organization!.risks!.__id;

  const onDelete = (riskId: string) => {
    risks;
    deleteRisk({
      variables: {
        input: { riskId },
        connections: [],
      },
    });
  };

  usePageTitle(__("Risks"));

  return (
    <div className="space-y-6">
      <PageHeader title={__("Risks")}>
        <FormRiskDialog
          connection={connectionId}
          trigger={<Button icon={IconPlusLarge}>{__("New Risk")}</Button>}
        />
      </PageHeader>
      {editedRisk && (
        <FormRiskDialog
          open
          risk={editedRisk}
          connection={connectionId}
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
              <Td>{getTreatment(__, risk.treatment)}</Td>
              <Td>
                <RiskBadge
                  level={risk.inherentLikelihood * risk.inherentImpact}
                />
              </Td>
              <Td>
                <RiskBadge
                  level={risk.residualLikelihood * risk.residualImpact}
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
