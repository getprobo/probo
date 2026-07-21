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

import { formatDate, getCompliancePageVisibilityOptions } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Field, IconArrowLink, IconPencil, IconTrashCan, Option, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageFileListItem_compliancePageFragment$key } from "#/__generated__/core/CompliancePageFileListItem_compliancePageFragment.graphql";
import type { CompliancePageFileListItem_fileFragment$data, CompliancePageFileListItem_fileFragment$key } from "#/__generated__/core/CompliancePageFileListItem_fileFragment.graphql";
import type { CompliancePageFileListItemMutation } from "#/__generated__/core/CompliancePageFileListItemMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { CompliancePageAliasField } from "../../_components/CompliancePageAliasField";

const compliancePageFragment = graphql`
  fragment CompliancePageFileListItem_compliancePageFragment on CompliancePortal {
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const fileFragment = graphql`
  fragment CompliancePageFileListItem_fileFragment on CompliancePortalFile {
    id
    name
    alias
    canSetAlias: permission(action: "resourcealias:alias:set")
    canRemoveAlias: permission(action: "resourcealias:alias:remove")
    category
    file {
      downloadUrl
    }
    compliancePortalVisibility
    createdAt
    canUpdate: permission(action: "compliance-portal:portal-file:update")
    canDelete: permission(action: "compliance-portal:portal-file:delete")
  }
`;

const updateCompliancePageFileMutation = graphql`
  mutation CompliancePageFileListItemMutation($input: UpdateCompliancePortalFileInput!) {
    updateCompliancePortalFile(input: $input) {
      compliancePortalFile {
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
  const visibilityOptions = getCompliancePageVisibilityOptions(__);

  const compliancePage = useFragment<CompliancePageFileListItem_compliancePageFragment$key>(
    compliancePageFragment,
    compliancePageFragmentRef,
  );
  const file = useFragment<CompliancePageFileListItem_fileFragment$key>(fileFragment, fileFragmentRef);

  const [updateFile, isUpdating] = useMutation<CompliancePageFileListItemMutation>(
    updateCompliancePageFileMutation,
    {
      successMessage: "File updated successfully",
      errorToast: "Failed to update file",
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
            compliancePortalVisibility: typedValue,
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
      <Td noLink>
        <CompliancePageAliasField
          resourceId={file.id}
          alias={file.alias}
          canSetAlias={file.canSetAlias}
          canRemoveAlias={file.canRemoveAlias}
        />
      </Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={file.compliancePortalVisibility}
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
              window.open(file.file?.downloadUrl, "_blank", "noopener,noreferrer")}
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
