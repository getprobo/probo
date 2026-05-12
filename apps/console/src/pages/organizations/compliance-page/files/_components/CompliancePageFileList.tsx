// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";
import { Table, Tbody, Td, Th, Thead, Tr, useDialogRef } from "@probo/ui";
import { useCallback, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageFileListFragment$key } from "#/__generated__/core/CompliancePageFileListFragment.graphql";
import type { CompliancePageFileListItem_fileFragment$data } from "#/__generated__/core/CompliancePageFileListItem_fileFragment.graphql";

import { CompliancePageFileListItem } from "./CompliancePageFileListItem";
import { DeleteCompliancePageFileDialog } from "./DeleteCompliancePageFileDialog";
import { EditCompliancePageFileDialog } from "./EditCompliancePageFileDialog";

const fragment = graphql`
  fragment CompliancePageFileListFragment on Organization {
    compliancePage: trustCenter @required(action: THROW) {
      ...CompliancePageFileListItem_compliancePageFragment
    }
    trustCenterFiles(first: 100)
      @connection(key: "CompliancePageFileList_trustCenterFiles") {
      __id
      edges {
        node {
          id
          ...CompliancePageFileListItem_fileFragment
        }
      }
    }
  }
`;

export function CompliancePageFileList(props: { fragmentRef: CompliancePageFileListFragment$key }) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();
  const deleteDialogRef = useDialogRef();

  const {
    compliancePage,
    trustCenterFiles: files,
  } = useFragment<CompliancePageFileListFragment$key>(fragment, fragmentRef);

  const [editingFile, setEditingFile] = useState<
    CompliancePageFileListItem_fileFragment$data | null>(null);
  const [deletingFileId, setDeletingFileId] = useState<string | null>(null);

  const handleDelete = useCallback(
    (id: string) => {
      setDeletingFileId(id);
      deleteDialogRef.current?.open();
    },
    [deleteDialogRef],
  );

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Category")}</Th>
            <Th>{__("Upload Date")}</Th>
            <Th>{__("Visibility")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {files.edges.length === 0 && (
            <Tr>
              <Td colSpan={5} className="text-center text-txt-secondary">
                {__("No files available")}
              </Td>
            </Tr>
          )}
          {files.edges.map(({ node: file }) => (
            <CompliancePageFileListItem
              key={file.id}
              compliancePageFragmentRef={compliancePage}
              fileFragmentRef={file}
              onEdit={setEditingFile}
              onDelete={handleDelete}
            />
          ))}
        </Tbody>
      </Table>

      {editingFile
        && (
          <EditCompliancePageFileDialog
            file={editingFile}
            onClose={() => setEditingFile(null)}
          />
        )}
      <DeleteCompliancePageFileDialog
        connectionId={files.__id}
        fileId={deletingFileId}
        ref={deleteDialogRef}
        onDelete={() => setDeletingFileId(null)}
      />
    </div>
  );
}
