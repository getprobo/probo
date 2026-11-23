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

import { ScrollArea } from "@base-ui/react/scroll-area";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { DocumentVersionList_document$key } from "#/__generated__/core/DocumentVersionList_document.graphql";
import type { DocumentVersionListPaginationQuery } from "#/__generated__/core/DocumentVersionListPaginationQuery.graphql";
import type { DocumentVersionListQuery } from "#/__generated__/core/DocumentVersionListQuery.graphql";

import { documentVersionList } from "../../variants";

import { DocumentVersionListItem } from "./DocumentVersionListItem";

const versionPageSize = 20;

export const documentVersionListQuery = graphql`
  query DocumentVersionListQuery($documentId: ID!) {
    node(id: $documentId) {
      __typename
      ... on Document {
        ...DocumentVersionList_document
      }
    }
  }
`;

const documentVersionListFragment = graphql`
  fragment DocumentVersionList_document on Document
  @refetchable(queryName: "DocumentVersionListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    versions(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: DESC }
    ) @connection(key: "DocumentVersionList_versions", filters: ["orderBy"]) {
      edges {
        node {
          id
          ...DocumentVersionListItem_version
        }
      }
    }
  }
`;

interface DocumentVersionListProps {
  queryRef: PreloadedQuery<DocumentVersionListQuery> | null | undefined;
  activeVersionId: string | null;
  onSelect: (versionId: string | null) => void;
}

// Every version of the document, newest first, after the live one. The
// column never blocks the document: while it loads or after it fails, it
// stays empty.
export function DocumentVersionList({ queryRef, ...props }: DocumentVersionListProps) {
  if (queryRef == null) {
    return null;
  }

  return (
    <ErrorBoundary fallback={null}>
      <Suspense fallback={null}>
        <DocumentVersionListLoaded queryRef={queryRef} {...props} />
      </Suspense>
    </ErrorBoundary>
  );
}

function DocumentVersionListLoaded({
  queryRef,
  ...props
}: Omit<DocumentVersionListProps, "queryRef"> & {
  queryRef: PreloadedQuery<DocumentVersionListQuery>;
}) {
  const data = usePreloadedQuery<DocumentVersionListQuery>(documentVersionListQuery, queryRef);

  if (data.node.__typename !== "Document") {
    return null;
  }

  return <DocumentVersionListContent documentKey={data.node} {...props} />;
}

function DocumentVersionListContent({
  documentKey,
  activeVersionId,
  onSelect,
}: Omit<DocumentVersionListProps, "queryRef"> & {
  documentKey: DocumentVersionList_document$key;
}) {
  const { t } = useTranslation("organizations/documents");
  const { data, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    DocumentVersionListPaginationQuery,
    DocumentVersionList_document$key
  >(documentVersionListFragment, documentKey);
  const { root, heading, body, viewport, content, more } = documentVersionList();

  return (
    <div className={root()}>
      <Text size={2} weight="medium" className={heading()}>{t("details.versions.title")}</Text>
      <ScrollArea.Root className={body()}>
        <ScrollArea.Viewport className={viewport()}>
          <ScrollArea.Content className={content()}>
            <DocumentVersionListItem
              versionKey={null}
              active={activeVersionId == null}
              onSelect={() => onSelect(null)}
            />
            {data.versions.edges.map(({ node }) => (
              <DocumentVersionListItem
                key={node.id}
                versionKey={node}
                active={activeVersionId === node.id}
                onSelect={() => onSelect(node.id)}
              />
            ))}
            {hasNext && (
              <Button
                size={1}
                variant="soft"
                color="neutral"
                className={more()}
                loading={isLoadingNext}
                onClick={() => loadNext(versionPageSize)}
              >
                {t("details.versions.loadMore")}
              </Button>
            )}
          </ScrollArea.Content>
        </ScrollArea.Viewport>
      </ScrollArea.Root>
    </div>
  );
}
