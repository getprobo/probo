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

import { getCompliancePortalVisibilityOptions } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import { Badge, Button, Field, IconArrowLink, IconPencil, IconTrashCan, Option, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalFileListItem_compliancePortalFragment$key } from "#/__generated__/core/CompliancePortalFileListItem_compliancePortalFragment.graphql";
import type { CompliancePortalFileListItem_fileFragment$data, CompliancePortalFileListItem_fileFragment$key } from "#/__generated__/core/CompliancePortalFileListItem_fileFragment.graphql";
import type { CompliancePortalFileListItemMutation } from "#/__generated__/core/CompliancePortalFileListItemMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { CompliancePortalAliasField } from "../../_components/CompliancePortalAliasField";

const compliancePortalFragment = graphql`
  fragment CompliancePortalFileListItem_compliancePortalFragment on CompliancePortal {
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const fileFragment = graphql`
  fragment CompliancePortalFileListItem_fileFragment on CompliancePortalFile {
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

const updateCompliancePortalFileMutation = graphql`
  mutation CompliancePortalFileListItemMutation($input: UpdateCompliancePortalFileInput!) {
    updateCompliancePortalFile(input: $input) {
      compliancePortalFile {
        ...CompliancePortalFileListItem_fileFragment
      }
    }
  }
`;

export function CompliancePortalFileListItem(props: {
  compliancePortalFragmentRef: CompliancePortalFileListItem_compliancePortalFragment$key;
  fileFragmentRef: CompliancePortalFileListItem_fileFragment$key;
  onEdit: (file: CompliancePortalFileListItem_fileFragment$data) => void;
  onDelete: (id: string) => void;
}) {
  const { compliancePortalFragmentRef, fileFragmentRef, onEdit, onDelete } = props;

  const { t, i18n } = useTranslation("organizations/compliance-portals");
  const visibilityOptions = getCompliancePortalVisibilityOptions(t);

  const compliancePortal = useFragment<CompliancePortalFileListItem_compliancePortalFragment$key>(
    compliancePortalFragment,
    compliancePortalFragmentRef,
  );
  const file = useFragment<CompliancePortalFileListItem_fileFragment$key>(fileFragment, fileFragmentRef);

  const [updateFile, isUpdating] = useMutation<CompliancePortalFileListItemMutation>(
    updateCompliancePortalFileMutation,
    {
      successMessage: t("fileListItem.messages.updated"),
      errorToast: t("fileListItem.errors.update"),
    },
  );

  const handleValueChange = useCallback(
    async (value: string) => {
      const stringValue = typeof value === "string" ? value : "";
      const typedValue = stringValue as "NONE" | "RESTRICTED" | "PUBLIC";
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
      <Td>{dateFormat(i18n.language, file.createdAt)}</Td>
      <Td noLink>
        <CompliancePortalAliasField
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
          disabled={isUpdating || !compliancePortal.canUpdate}
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
            title={t("fileListItem.actions.download")}
          />
          {file.canUpdate && (
            <Button
              variant="secondary"
              icon={IconPencil}
              onClick={() => onEdit(file)}
              disabled={isUpdating}
              title={t("fileListItem.actions.edit")}
            />
          )}
          {file.canDelete && (
            <Button
              variant="danger"
              icon={IconTrashCan}
              onClick={() => onDelete(file.id)}
              disabled={isUpdating}
              title={t("fileListItem.actions.delete")}
            />
          )}
        </div>
      </Td>
    </Tr>
  );
}
