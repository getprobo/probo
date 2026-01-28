import { formatDate, getTrustCenterVisibilityOptions } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Field, IconArrowLink, IconPencil, IconTrashCan, Option, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageFileListItem_compliancePageFragment$key } from "#/__generated__/core/CompliancePageFileListItem_compliancePageFragment.graphql";
import type { CompliancePageFileListItem_fileFragment$data, CompliancePageFileListItem_fileFragment$key } from "#/__generated__/core/CompliancePageFileListItem_fileFragment.graphql";
import type { CompliancePageFileListItemMutation } from "#/__generated__/core/CompliancePageFileListItemMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const compliancePageFragment = graphql`
  fragment CompliancePageFileListItem_compliancePageFragment on TrustCenter {
    canUpdate: permission(action: "core:trust-center:update")
  }
`;

const fileFragment = graphql`
  fragment CompliancePageFileListItem_fileFragment on TrustCenterFile {
    id
    name
    category
    fileUrl
    trustCenterVisibility
    createdAt
    canUpdate: permission(action: "core:trust-center-file:update")
    canDelete: permission(action: "core:trust-center-file:delete")
  }
`;

const updateCompliancePageFileMutation = graphql`
  mutation CompliancePageFileListItemMutation($input: UpdateTrustCenterFileInput!) {
    updateTrustCenterFile(input: $input) {
      trustCenterFile {
        ...CompliancePageFileListItem_fileFragment
      }
    }
  }
`;

export function CompliancePageFileListItem(props: {
  compliancePageFragmentRef: CompliancePageFileListItem_compliancePageFragment$key;
  fileFragmentRef: CompliancePageFileListItem_fileFragment$key;
  onEdit: (file: CompliancePageFileListItem_fileFragment$data) => void;
  onDelete: (id: string) => void;
}) {
  const { compliancePageFragmentRef, fileFragmentRef, onEdit, onDelete } = props;

  const { __ } = useTranslate();
  const visibilityOptions = getTrustCenterVisibilityOptions(__);

  const compliancePage = useFragment<CompliancePageFileListItem_compliancePageFragment$key>(
    compliancePageFragment,
    compliancePageFragmentRef,
  );
  const file = useFragment<CompliancePageFileListItem_fileFragment$key>(fileFragment, fileFragmentRef);

  const [updateFile, isUpdating] = useMutationWithToasts<CompliancePageFileListItemMutation>(
    updateCompliancePageFileMutation,
    {
      successMessage: "File updated successfully",
      errorMessage: "Failed to update file",
    },
  );

  const handleValueChange = useCallback(
    async (value: string) => {
      const stringValue = typeof value === "string" ? value : "";
      const typedValue = stringValue as "NONE" | "PRIVATE" | "PUBLIC";
      await updateFile({
        variables: {
          input: {
            id: file.id,
            trustCenterVisibility: typedValue,
          },
        },
      });
    },
    [file.id, updateFile],
  );

  return (
    <Tr>
      <Td>
        <div className="flex gap-4 items-center">{file.name}</div>
      </Td>
      <Td>{file.category}</Td>
      <Td>{formatDate(file.createdAt)}</Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={file.trustCenterVisibility}
          onValueChange={value => void handleValueChange(value)}
          disabled={isUpdating || !compliancePage.canUpdate}
          className="w-[105px]"
        >
          {visibilityOptions.map(option => (
            <Option key={option.value} value={option.value}>
              <div className="flex items-center justify-between w-full">
                <Badge variant={option.variant}>{option.label}</Badge>
              </div>
            </Option>
          ))}
        </Field>
      </Td>
      <Td noLink width={120}>
        <div className="flex gap-2">
          <Button
            variant="secondary"
            icon={IconArrowLink}
            onClick={() =>
              window.open(file.fileUrl, "_blank", "noopener,noreferrer")}
            title={__("Download")}
          />
          {file.canUpdate && (
            <Button
              variant="secondary"
              icon={IconPencil}
              onClick={() => onEdit(file)}
              disabled={isUpdating}
              title={__("Edit")}
            />
          )}
          {file.canDelete && (
            <Button
              variant="danger"
              icon={IconTrashCan}
              onClick={() => onDelete(file.id)}
              disabled={isUpdating}
              title={__("Delete")}
            />
          )}
        </div>
      </Td>
    </Tr>
  );
}
