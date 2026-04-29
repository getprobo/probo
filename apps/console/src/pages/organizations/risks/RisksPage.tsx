// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { getTreatment, sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconPageTextLine,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  IconUpload,
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
import type { PreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";

import type { RiskGraphFragment$data } from "#/__generated__/core/RiskGraphFragment.graphql";
import type { RiskGraphListQuery } from "#/__generated__/core/RiskGraphListQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useDeleteRiskMutation, useRisksQuery } from "#/hooks/graph/RiskGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { PublishRiskListDialog } from "./dialogs/PublishRiskListDialog";
import FormRiskDialog from "./FormRiskDialog";

type Props = {
  queryRef: PreloadedQuery<RiskGraphListQuery>;
};

export default function RisksPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const {
    data: { canCreateRisk, canPublishRisk, risksDocument },
    connectionId,
    risks,
    ...pagination
  } = useRisksQuery(props.queryRef);

  const refetch = ({
    order,
  }: {
    order: { direction: string; field: string };
  }) => {
    pagination.refetch(
      {
        order: {
          direction: order.direction as "ASC" | "DESC",
          field: order.field as
          | "NAME"
          | "CATEGORY"
          | "TREATMENT"
          | "INHERENT_RISK_SCORE"
          | "RESIDUAL_RISK_SCORE"
          | "OWNER_FULL_NAME"
          | "CREATED_AT",
        },
      },
      { fetchPolicy: "network-only" },
    );
  };

  usePageTitle(__("Risks"));

  const hasAnyAction
    = risks.some(({ canDelete, canUpdate }) => canUpdate || canDelete);

  const defaultApproverIds
    = risksDocument?.defaultApprovers?.map(a => a.id) ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Risks")}
        description={__(
          "Risks are potential threats to your organization. Manage them by identifying, assessing, and implementing mitigation measures.",
        )}
      >
        <div className="flex gap-2">
          {risksDocument && (
            <Button
              variant="secondary"
              icon={IconPageTextLine}
              onClick={() => void navigate(
                `/organizations/${organizationId}/documents/${risksDocument.id}`,
              )}
            >
              {__("Document")}
            </Button>
          )}
          {canPublishRisk && (
            <PublishRiskListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={documentId => void navigate(
                `/organizations/${organizationId}/documents/${documentId}`,
              )}
            >
              <Button variant="secondary" icon={IconUpload}>
                {__("Publish")}
              </Button>
            </PublishRiskListDialog>
          )}
          {canCreateRisk && (
            <FormRiskDialog
              connection={connectionId}
              onSuccess={() => {
                pagination.refetch({});
              }}
              trigger={<Button icon={IconPlusLarge}>{__("New Risk")}</Button>}
            />
          )}
        </div>
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
      <SortableTable {...pagination} refetch={refetch}>
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
            <SortableTh field="OWNER_FULL_NAME">{__("Owner")}</SortableTh>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {risks?.map(risk => (
            <RiskRow
              risk={risk}
              key={risk.id}
              connectionId={connectionId}
              organizationId={organizationId}
              hasAnyAction={hasAnyAction}
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
  hasAnyAction: boolean;
};

function RiskRow(props: RowProps) {
  const { __ } = useTranslate();
  const { risk, connectionId, organizationId } = props;
  const [deleteRisk] = useDeleteRiskMutation();
  const confirm = useConfirm();
  const onDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteRisk({
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
            "This will permanently delete the risk \"%s\". This action cannot be undone.",
          ),
          risk.name,
        ),
      },
    );
  };
  const formDialogRef = useDialogRef();

  const riskUrl = `/organizations/${organizationId}/risks/${risk.id}/overview`;

  return (
    <>
      <FormRiskDialog
        ref={formDialogRef}
        risk={risk}
        connection={connectionId}
      />
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
        <Td>{risk.owner?.fullName || __("Unassigned")}</Td>
        {props.hasAnyAction && (
          <Td noLink className="text-end">
            <ActionDropdown>
              {risk.canUpdate && (
                <DropdownItem
                  icon={IconPencil}
                  onClick={() => formDialogRef.current?.open()}
                >
                  {__("Edit")}
                </DropdownItem>
              )}

              {risk.canDelete && (
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={onDelete}
                >
                  {__("Delete")}
                </DropdownItem>
              )}
            </ActionDropdown>
          </Td>
        )}
      </Tr>
    </>
  );
}
