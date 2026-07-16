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

import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import type { DocumentViewerPageQuery } from "./__generated__/DocumentViewerPageQuery.graphql";
import { DocumentLocked } from "./_components/DocumentLocked";
import { DocumentViewer } from "./_components/DocumentViewer";
import { useAccessRequest } from "./_lib/useAccessRequest";
import type { DocumentKind } from "./_lib/useDocumentExport";
import { useDocumentExport } from "./_lib/useDocumentExport";

export const documentViewerPageQuery = graphql`
  query DocumentViewerPageQuery($alias: String!) {
    aliasedNode(alias: $alias) @required(action: THROW) {
      __typename
      ... on Document {
        id
        title
        isUserAuthorized
      }
      ... on TrustCenterFile {
        id
        name
        isUserAuthorized
      }
      ... on AuditReport {
        id
        fileName
        isUserAuthorized
      }
    }
  }
`;

interface DocumentViewerPageProps {
  queryRef: PreloadedQuery<DocumentViewerPageQuery>;
}

interface ResolvedNode {
  kind: DocumentKind;
  id: string;
  title: string;
  isAuthorized: boolean;
}

function resolveNode(node: DocumentViewerPageQuery["response"]["aliasedNode"]): ResolvedNode {
  switch (node.__typename) {
    case "Document":
      return { kind: "Document", id: node.id, title: node.title, isAuthorized: node.isUserAuthorized };
    case "TrustCenterFile":
      return { kind: "TrustCenterFile", id: node.id, title: node.name, isAuthorized: node.isUserAuthorized };
    case "AuditReport":
      return { kind: "AuditReport", id: node.id, title: node.fileName, isAuthorized: node.isUserAuthorized };
    default:
      throw new Error(`Unexpected aliased node type: ${node.__typename}`);
  }
}

// Full-page viewer for a single document/file/report resolved by its alias. It
// exports the (watermarked) bytes and renders them; unauthorized visitors get a
// locked state instead.
export function DocumentViewerPage({ queryRef }: DocumentViewerPageProps) {
  const data = usePreloadedQuery<DocumentViewerPageQuery>(documentViewerPageQuery, queryRef);
  const node = resolveNode(data.aliasedNode);
  const { dataUri } = useDocumentExport(node.kind, node.id, node.isAuthorized);
  const { requestAccess, isRequesting } = useAccessRequest(node.kind, node.id);

  if (!node.isAuthorized) {
    return <DocumentLocked onGetAccess={requestAccess} isRequesting={isRequesting} />;
  }

  return <DocumentViewer title={node.title} dataUri={dataUri} downloadName={node.title} />;
}
