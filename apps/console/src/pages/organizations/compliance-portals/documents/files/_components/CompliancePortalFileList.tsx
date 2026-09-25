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

import { Table, Tbody, Td, Th, Thead, Tr, useDialogRef } from "@probo/ui";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { type DataID, useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalFileList_compliancePortalFragment$key } from "#/__generated__/core/CompliancePortalFileList_compliancePortalFragment.graphql";
import type { CompliancePortalFileListItem_fileFragment$data } from "#/__generated__/core/CompliancePortalFileListItem_fileFragment.graphql";

import { CompliancePortalFileListItem } from "./CompliancePortalFileListItem";
import { DeleteCompliancePortalFileDialog } from "./DeleteCompliancePortalFileDialog";
import { EditCompliancePortalFileDialog } from "./EditCompliancePortalFileDialog";

const compliancePortalFragment = graphql`
  fragment CompliancePortalFileList_compliancePortalFragment on CompliancePortal {
    ...CompliancePortalFileListItem_compliancePortalFragment
    compliancePortalFiles: compliancePortalFiles(first: 100)
      @connection(key: "CompliancePortalFileList_compliancePortalFiles") {
      __id
      edges {
        node {
          id
          ...CompliancePortalFileListItem_fileFragment
        }
      }
    }
  }
`;

export function CompliancePortalFileList(props: {
  compliancePortalRef: CompliancePortalFileList_compliancePortalFragment$key;
  filesConnectionId: DataID;
}) {
  const { compliancePortalRef, filesConnectionId } = props;

  const { t } = useTranslation("organizations/compliance-portals");
  const deleteDialogRef = useDialogRef();

  const compliancePortal = useFragment(compliancePortalFragment, compliancePortalRef);
  const { compliancePortalFiles: files } = compliancePortal;

  const [editingFile, setEditingFile] = useState<
    CompliancePortalFileListItem_fileFragment$data | null>(null);
  const [deletingFileId, setDeletingFileId] = useState<string | null>(null);

  const handleDelete = useCallback(
    (id: string) => {
      setDeletingFileId(id);
      deleteDialogRef.current?.open();
    },
    [deleteDialogRef],
  );

  const handleDeleteComplete = useCallback(() => {
    setDeletingFileId(null);
  }, []);

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{t("fileList.columns.name")}</Th>
            <Th>{t("fileList.columns.category")}</Th>
            <Th>{t("fileList.columns.uploadDate")}</Th>
            <Th>{t("fileList.columns.alias")}</Th>
            <Th>{t("fileList.columns.visibility")}</Th>
            <Th className="w-0" />
          </Tr>
        </Thead>
        <Tbody>
          {files.edges.length === 0
            ? (
                <Tr>
                  <Td colSpan={6} className="text-center text-txt-tertiary">
                    {t("fileList.empty")}
                  </Td>
                </Tr>
              )
            : (
                files.edges.map(({ node }) => (
                  <CompliancePortalFileListItem
                    key={node.id}
                    compliancePortalFragmentRef={compliancePortal}
                    fileFragmentRef={node}
                    onEdit={setEditingFile}
                    onDelete={handleDelete}
                  />
                ))
              )}
        </Tbody>
      </Table>

      {editingFile && (
        <EditCompliancePortalFileDialog
          file={editingFile}
          onClose={() => setEditingFile(null)}
        />
      )}

      <DeleteCompliancePortalFileDialog
        connectionId={filesConnectionId}
        fileId={deletingFileId}
        ref={deleteDialogRef}
        onDelete={handleDeleteComplete}
      />
    </div>
  );
}
