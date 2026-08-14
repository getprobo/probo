// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { promisifyMutation } from "@probo/helpers";
import {
  ActionDropdown,
  Badge,
  DropdownItem,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { ProcessingActivityGraphDeleteMutation } from "#/__generated__/core/ProcessingActivityGraphDeleteMutation.graphql";
import type { ProcessingActivityListItem_processingActivity$key } from "#/__generated__/core/ProcessingActivityListItem_processingActivity.graphql";
import { getLawfulBasisLabel } from "#/components/form/ProcessingActivityEnumOptions";
import { deleteProcessingActivityMutation } from "#/hooks/graph/ProcessingActivityGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const processingActivityListItemFragment = graphql`
  fragment ProcessingActivityListItem_processingActivity on ProcessingActivity {
    id
    name
    purpose
    dataSubjectCategory
    lawfulBasis
    location
    internationalTransfers
    canDelete: permission(action: "core:processing-activity:delete")
  }
`;

interface ProcessingActivityListItemProps {
  processingActivityKey: ProcessingActivityListItem_processingActivity$key;
  connectionId: string;
  hasAnyAction: boolean;
}

export function ProcessingActivityListItem({
  processingActivityKey,
  connectionId,
  hasAnyAction,
}: ProcessingActivityListItemProps) {
  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const processingActivity = useFragment(
    processingActivityListItemFragment,
    processingActivityKey,
  );
  const [deleteActivity] = useMutation<ProcessingActivityGraphDeleteMutation>(
    deleteProcessingActivityMutation,
  );
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteActivity)({
          variables: {
            input: {
              processingActivityId: processingActivity.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: t("processingActivitiesPage.deleteConfirmation", {
          name: processingActivity.name,
        }),
      },
    );
  };

  const activityUrl
    = `/organizations/${organizationId}/privacy/processing-activities/${processingActivity.id}`;

  return (
    <Tr to={activityUrl}>
      <Td>
        <span className="font-semibold">{processingActivity.name}</span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary">
          {processingActivity.purpose || "-"}
        </span>
      </Td>
      <Td>{processingActivity.dataSubjectCategory || "-"}</Td>
      <Td>{getLawfulBasisLabel(processingActivity.lawfulBasis, t)}</Td>
      <Td>{processingActivity.location || "-"}</Td>
      <Td>
        <Badge
          variant={
            processingActivity.internationalTransfers ? "warning" : "success"
          }
        >
          {processingActivity.internationalTransfers
            ? t("processingActivitiesPage.answers.yes")
            : t("processingActivitiesPage.answers.no")}
        </Badge>
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {processingActivity.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
              >
                {t("processingActivitiesPage.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
