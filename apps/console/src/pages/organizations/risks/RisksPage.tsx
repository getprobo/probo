import {
  ActionDropdown,
  Button,
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
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import FormRiskDialog from "./FormRiskDialog";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { usePageTitle } from "@probo/hooks";
import { getTreatment, sprintf } from "@probo/helpers";
import type { NodeOf } from "/types";
import { useDeleteRiskMutation, useRisksQuery } from "/hooks/graph/RiskGraph";
import { SortableTable, SortableTh } from "/components/SortableTable";
import type { PreloadedQuery } from "react-relay";
import type { RiskGraphListQuery } from "/hooks/graph/__generated__/RiskGraphListQuery.graphql";
import type { RiskGraphFragment$data } from "/hooks/graph/__generated__/RiskGraphFragment.graphql";
import { useParams } from "react-router";
import { SnapshotBanner } from "/components/SnapshotBanner";

type Props = {
  queryRef: PreloadedQuery<RiskGraphListQuery>;
};

export default function RisksPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const { connectionId, risks, refetch } = useRisksQuery(props.queryRef);

  usePageTitle(__("Risks"));

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <PageHeader
        title={__("Risks")}
        description={__(
          "Risks are potential threats to your organization. Manage them by identifying, assessing, and implementing mitigation measures."
        )}
      >
        {!isSnapshotMode && (
          <FormRiskDialog
            connection={connectionId}
            trigger={<Button icon={IconPlusLarge}>{__("New Risk")}</Button>}
          />
        )}
      </PageHeader>

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
            <RiskRow
              risk={risk}
              key={risk.id}
              connectionId={connectionId}
              organizationId={organizationId}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

type RowProps = {
  risk: NodeOf<RiskGraphFragment$data["risks"]>;
  connectionId: string;
  organizationId: string;
};

function RiskRow(props: RowProps) {
  const { __ } = useTranslate();
  const { risk, connectionId, organizationId } = props;
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const [deleteRisk] = useDeleteRiskMutation();
  const confirm = useConfirm();
  const onDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteRisk({
            variables: {
              input: { riskId: risk.id },
              connections: [connectionId],
            },
            onCompleted: () => resolve(),
          });
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the risk "%s". This action cannot be undone.'
          ),
          risk.name
        ),
      }
    );
  };
  const formDialogRef = useDialogRef();

  const riskUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/risks/${risk.id}/overview`
    : `/organizations/${organizationId}/risks/${risk.id}/overview`;

  return (
    <>
      {!isSnapshotMode && (
        <FormRiskDialog
          ref={formDialogRef}
          risk={risk}
          connection={connectionId}
        />
      )}
      <Tr to={riskUrl}>
        <Td>{risk.name}</Td>
        <Td>{risk.category}</Td>
        <Td>{getTreatment(__, risk.treatment)}</Td>
        <Td>
          <SeverityBadge score={risk.inherentRiskScore} />
        </Td>
        <Td>
          <SeverityBadge score={risk.residualRiskScore} />
        </Td>
        <Td noLink className="text-end">
          {!isSnapshotMode && (
            <ActionDropdown>
              <DropdownItem
                icon={IconPencil}
                onClick={() => formDialogRef.current?.open()}
              >
                {__("Edit")}
              </DropdownItem>

              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                onClick={onDelete}
              >
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </Td>
      </Tr>
    </>
  );
}
