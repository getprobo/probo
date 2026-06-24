// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { formatError, getTreatment, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  SeverityBadge,
  Td,
  Tr,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { graphql, useFragment, useMutation } from "react-relay";

import type { RiskRow_risk$key } from "#/__generated__/core/RiskRow_risk.graphql";
import type { RiskRowDeleteMutation } from "#/__generated__/core/RiskRowDeleteMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { FormRiskDialog } from "./FormRiskDialog";

const riskRowFragment = graphql`
  fragment RiskRow_risk on Risk {
    id
    name
    category
    treatment
    owner {
      id
      fullName
    }
    inherentRiskScore
    residualRiskScore
    canUpdate: permission(action: "core:risk:update")
    canDelete: permission(action: "core:risk:delete")
    ...FormRiskDialog_risk
  }
`;

const deleteRiskMutation = graphql`
  mutation RiskRowDeleteMutation(
    $input: DeleteRiskInput!
    $connections: [ID!]!
  ) {
    deleteRisk(input: $input) {
      deletedRiskId @deleteEdge(connections: $connections)
    }
  }
`;

interface RiskRowProps {
  riskKey: RiskRow_risk$key;
  connectionId: string;
  hasAnyAction: boolean;
}

export function RiskRow(props: RiskRowProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const risk = useFragment(riskRowFragment, props.riskKey);
  const [deleteRisk] = useMutation<RiskRowDeleteMutation>(deleteRiskMutation);
  const confirm = useConfirm();
  const { toast } = useToast();
  const formDialogRef = useDialogRef();

  const onDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteRisk({
            variables: {
              input: { riskId: risk.id },
              connections: [props.connectionId],
            },
            onCompleted() {
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete risk"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
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

  const riskUrl = `/organizations/${organizationId}/risks/${risk.id}/overview`;

  return (
    <>
      <FormRiskDialog
        ref={formDialogRef}
        risk={risk}
        connection={props.connectionId}
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
