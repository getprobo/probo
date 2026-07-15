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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";

import { useMutation } from "#/lib/relay/useMutation";

import { openExportedFile } from "../_lib/openExportedFile";

import type { TrustCenterFileListItem_file$key } from "./__generated__/TrustCenterFileListItem_file.graphql";
import type { TrustCenterFileListItemExportMutation } from "./__generated__/TrustCenterFileListItemExportMutation.graphql";
import { DocumentAccessAction } from "./DocumentAccessAction";
import { documentListItem } from "./variants";

const trustCenterFileListItemFragment = graphql`
  fragment TrustCenterFileListItem_file on TrustCenterFile @throwOnFieldError {
    id
    name
    category
    isUserAuthorized
    access {
      status
    }
  }
`;

const exportTrustCenterFileMutation = graphql`
  mutation TrustCenterFileListItemExportMutation($input: ExportTrustCenterFileInput!) {
    exportTrustCenterFile(input: $input) {
      data
    }
  }
`;

interface TrustCenterFileListItemProps {
  fileKey: TrustCenterFileListItem_file$key;
}

// A single uploaded trust-center file row: name, its category, and an access
// action that opens the exported file when the viewer is authorized.
export function TrustCenterFileListItem({ fileKey }: TrustCenterFileListItemProps) {
  const file = useFragment(trustCenterFileListItemFragment, fileKey);
  const [exportFile, isExporting] = useMutation<TrustCenterFileListItemExportMutation>(exportTrustCenterFileMutation);
  const { root, content } = documentListItem();

  const handleView = () => {
    exportFile({
      variables: { input: { trustCenterFileId: file.id } },
      onCompleted: response => openExportedFile(response.exportTrustCenterFile.data),
    }).catch(() => {
      // The mutation failure is already surfaced through a toast.
    });
  };

  return (
    <div className={root()}>
      <div className={content()}>
        <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
          {file.name}
        </Text>
        <Text size={1} color="gold" className="truncate">
          {file.category}
        </Text>
      </div>
      <DocumentAccessAction
        isAuthorized={file.isUserAuthorized}
        requested={file.access?.status === "REQUESTED"}
        onView={handleView}
        isViewing={isExporting}
      />
    </div>
  );
}
