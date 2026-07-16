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

import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentListItem_document$key } from "./__generated__/DocumentListItem_document.graphql";
import { DocumentEntry } from "./DocumentEntry";

const documentListItemFragment = graphql`
  fragment DocumentListItem_document on Document @throwOnFieldError {
    id
    alias
    title
    documentType
    isUserAuthorized
    access {
      status
    }
  }
`;

interface DocumentListItemProps {
  documentKey: DocumentListItem_document$key;
}

// A single Probo document entry: title, its document type, and an access action
// linking to the viewer when authorized.
export function DocumentListItem({ documentKey }: DocumentListItemProps) {
  const { t } = useTranslation("documents");
  const document = useFragment(documentListItemFragment, documentKey);

  return (
    <DocumentEntry
      title={document.title}
      meta={t(`types.${document.documentType}`)}
      isAuthorized={document.isUserAuthorized}
      requested={document.access?.status === "REQUESTED"}
      viewHref={`/documents/${encodeURIComponent(document.alias ?? document.id)}`}
    />
  );
}
