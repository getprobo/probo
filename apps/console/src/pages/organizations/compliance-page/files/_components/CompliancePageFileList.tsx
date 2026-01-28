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
