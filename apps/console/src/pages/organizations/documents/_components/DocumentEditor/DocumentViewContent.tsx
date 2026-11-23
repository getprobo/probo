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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { DocumentViewContentQuery } from "#/__generated__/core/DocumentViewContentQuery.graphql";

import { DocumentReadOnlyContent } from "./DocumentReadOnlyContent";

// The server renders the live document without its pending suggestions.
export const documentViewContentQuery = graphql`
  query DocumentViewContentQuery($documentId: ID!) {
    node(id: $documentId) {
      __typename
      ... on Document {
        orientation
        content {
          prosemirror
        }
      }
    }
  }
`;

interface DocumentViewContentProps {
  queryRef: PreloadedQuery<DocumentViewContentQuery>;
  // Must be stable: a new callback rebuilds the editor.
  onPages: (pages: number) => void;
}

// The live document as readers see it, read only.
export function DocumentViewContent(props: DocumentViewContentProps) {
  const { t } = useTranslation("organizations/documents");

  return (
    <ErrorBoundary fallback={<Text size={2} color="faint">{t("details.editor.openError")}</Text>}>
      <Suspense fallback={<Text size={2} color="faint">{t("details.editor.loading")}</Text>}>
        <DocumentViewContentLoaded {...props} />
      </Suspense>
    </ErrorBoundary>
  );
}

function DocumentViewContentLoaded({ queryRef, onPages }: DocumentViewContentProps) {
  const data = usePreloadedQuery<DocumentViewContentQuery>(documentViewContentQuery, queryRef);

  if (data.node.__typename !== "Document") {
    return null;
  }

  return (
    <DocumentReadOnlyContent
      content={data.node.content.prosemirror}
      orientation={data.node.orientation}
      onPages={onPages}
    />
  );
}
