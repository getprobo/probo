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
import { Suspense, useEffect, useLayoutEffect } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { DocumentVersionContentQuery } from "#/__generated__/core/DocumentVersionContentQuery.graphql";
import { isNotFoundError } from "#/lib/relay/errors";

import { DocumentReadOnlyContent } from "../DocumentEditor/DocumentReadOnlyContent";

export const documentVersionContentQuery = graphql`
  query DocumentVersionContentQuery($versionId: ID!) {
    node(id: $versionId) {
      __typename
      ... on DocumentVersion {
        id
        orientation
        document {
          id
        }
        content {
          prosemirror
        }
        # The toolbar label reads this query too, to share its request.
        # eslint-disable-next-line relay/must-colocate-fragment-spreads
        ...DocumentVersionLabel_version
      }
    }
  }
`;

interface DocumentVersionContentProps {
  queryRef: PreloadedQuery<DocumentVersionContentQuery>;
  documentId: string;
  // Must be stable: a new callback rebuilds the editor.
  onPages: (pages: number) => void;
  // The version keeps the orientation it was published with, which sizes the
  // pages around it.
  onOrientation: (orientation: string) => void;
  // A version of another document, or an id the server reports as not found,
  // falls back to the live document.
  onMissing: () => void;
}

// A previous version, read only, laid out on pages like the live document.
export function DocumentVersionContent(props: DocumentVersionContentProps) {
  const { t } = useTranslation("organizations/documents");

  return (
    <ErrorBoundary
      fallback={error => isNotFoundError(error)
        ? <DocumentVersionMissing onMissing={props.onMissing} />
        : <Text size={2} color="faint">{t("details.versions.openError")}</Text>}
    >
      <Suspense fallback={<Text size={2} color="faint">{t("details.editor.loading")}</Text>}>
        <DocumentVersionContentLoaded {...props} />
      </Suspense>
    </ErrorBoundary>
  );
}

function DocumentVersionContentLoaded({
  queryRef,
  documentId,
  onPages,
  onOrientation,
  onMissing,
}: DocumentVersionContentProps) {
  const data = usePreloadedQuery<DocumentVersionContentQuery>(documentVersionContentQuery, queryRef);
  const version = data.node.__typename === "DocumentVersion" && data.node.document.id === documentId
    ? data.node
    : null;
  const orientation = version?.orientation ?? null;

  useLayoutEffect(() => {
    if (orientation != null) {
      onOrientation(orientation);
    }
  }, [onOrientation, orientation]);

  if (version == null) {
    return <DocumentVersionMissing onMissing={onMissing} />;
  }

  return (
    <DocumentReadOnlyContent
      key={version.id}
      content={version.content.prosemirror}
      orientation={version.orientation}
      onPages={onPages}
    />
  );
}

function DocumentVersionMissing({ onMissing }: { onMissing: () => void }) {
  useEffect(() => {
    onMissing();
  }, [onMissing]);

  return null;
}
