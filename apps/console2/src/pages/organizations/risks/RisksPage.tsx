import {
  ActionDropdown,
  Button,
  ConfirmDialog,
  DropdownItem,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  RisksChart,
  SeverityBadge,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import FormRiskDialog from "./FormRiskDialog";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useState } from "react";
import { usePageTitle } from "@probo/hooks";
import { getTreatment, sprintf } from "@probo/helpers";
import type { ItemOf } from "/types";
import { useDeleteRiskMutation, useRisksQuery } from "/hooks/graph/RiskGraph";
import { SortableTable, SortableTh } from "/components/SortableTable";
import type { PreloadedQuery } from "react-relay";
import type { RiskGraphListQuery } from "/hooks/graph/__generated__/RiskGraphListQuery.graphql";

type Props = {
  queryRef: PreloadedQuery<RiskGraphListQuery>;
};

export default function RisksPage(props: Props) {
  const { __ } = useTranslate();
  const [editedRisk, setEditedRisk] = useState<ItemOf<typeof risks> | null>(
    null
  );
  const organizationId = useOrganizationId();

  const [deleteRisk] = useDeleteRiskMutation();
  const { connectionId, risks, refetch } = useRisksQuery(props.queryRef);

  const onDelete = (riskId: string) => {
    deleteRisk({
      variables: {
        input: { riskId },
        connections: [connectionId],
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
      <SortableTable refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Risk name")}</SortableTh>
            <SortableTh field="CATEGORY">{__("Category")}</SortableTh>
            <SortableTh field="TREATMENT">{__("Treatment")}</SortableTh>
            <SortableTh field="INHERENT_RISK_SCORE">
              {__("Initial Risk")}
            </SortableTh>
            <SortableTh field="RESIDUAL_RISK_SCORE">
              {__("Residual Risk")}
            </SortableTh>
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
                <SeverityBadge score={risk.inherentRiskScore} />
              </Td>
              <Td>
                <SeverityBadge score={risk.residualRiskScore} />
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
      </SortableTable>
    </div>
  );
}
