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

import { graphql, useFragment } from "react-relay";

import { useRequestFileAccess } from "../_lib/useAccessRequest";

import type { CompliancePortalFileListItem_file$key } from "./__generated__/CompliancePortalFileListItem_file.graphql";
import { DocumentEntry } from "./DocumentEntry";

const compliancePortalFileListItemFragment = graphql`
  fragment CompliancePortalFileListItem_file on CompliancePortalFile @throwOnFieldError {
    id
    alias
    name
    category
    isUserAuthorized
    access {
      status
    }
  }
`;

interface CompliancePortalFileListItemProps {
  fileKey: CompliancePortalFileListItem_file$key;
}

// A single uploaded trust-center file entry: name, its category, and an access
// action linking to the viewer when authorized.
export function CompliancePortalFileListItem({ fileKey }: CompliancePortalFileListItemProps) {
  const file = useFragment(compliancePortalFileListItemFragment, fileKey);
  const { requestAccess, isRequesting } = useRequestFileAccess(file.id);

  return (
    <DocumentEntry
      title={file.name}
      meta={file.category}
      isAuthorized={file.isUserAuthorized}
      requested={file.access?.status === "REQUESTED"}
      viewHref={`/documents/${encodeURIComponent(file.alias ?? file.id)}`}
      onGetAccess={requestAccess}
      isRequesting={isRequesting}
    />
  );
}
