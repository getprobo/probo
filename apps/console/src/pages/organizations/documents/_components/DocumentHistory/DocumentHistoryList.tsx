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
import { XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Switch } from "@probo/ui/src/v2/Switch/Switch";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePaginationFragment, usePreloadedQuery } from "react-relay";

import type { DocumentHistoryList_document$key } from "#/__generated__/core/DocumentHistoryList_document.graphql";
import type { DocumentHistoryListPaginationQuery } from "#/__generated__/core/DocumentHistoryListPaginationQuery.graphql";
import type { DocumentHistoryListQuery } from "#/__generated__/core/DocumentHistoryListQuery.graphql";

import { documentHistoryList } from "../../variants";

import { DocumentHistoryListItem } from "./DocumentHistoryListItem";

const historyPageSize = 20;

export const documentHistoryListQuery = graphql`
  query DocumentHistoryListQuery($documentId: ID!) {
    node(id: $documentId) {
      __typename
      ... on Document {
        ...DocumentHistoryList_document
      }
    }
  }
`;

const documentHistoryListFragment = graphql`
  fragment DocumentHistoryList_document on Document
  @refetchable(queryName: "DocumentHistoryListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
    filter: { type: "DocumentRevisionFilter", defaultValue: null }
  ) {
    publishedVersion {
      id
    }
    pendingApprovalVersion {
      id
    }
    history(
      first: $first
      after: $after
      orderBy: { field: ENDED_AT, direction: DESC }
      filter: $filter
    ) @connection(key: "DocumentHistoryList_history", filters: ["orderBy", "filter"]) {
      edges {
        node {
          id
          ...DocumentHistoryListItem_revision
        }
      }
    }
  }
`;

interface DocumentHistoryListProps {
  queryRef: PreloadedQuery<DocumentHistoryListQuery> | null | undefined;
  activeRevisionId: string | null;
  onSelect: (revisionId: string | null) => void;
  onClose: () => void;
}

// The one history of the document, newest first, after the live document: a
// revision per editing session, with the versions published from it as tags.
// The column never blocks the document: while it loads or after it fails, it
// stays empty.
export function DocumentHistoryList({ queryRef, ...props }: DocumentHistoryListProps) {
  if (queryRef == null) {
    return null;
  }

  return (
    <ErrorBoundary fallback={null}>
      <Suspense fallback={null}>
        <DocumentHistoryListLoaded queryRef={queryRef} {...props} />
      </Suspense>
    </ErrorBoundary>
  );
}

function DocumentHistoryListLoaded({
  queryRef,
  ...props
}: Omit<DocumentHistoryListProps, "queryRef"> & {
  queryRef: PreloadedQuery<DocumentHistoryListQuery>;
}) {
  const data = usePreloadedQuery<DocumentHistoryListQuery>(documentHistoryListQuery, queryRef);

  if (data.node.__typename !== "Document") {
    return null;
  }

  return <DocumentHistoryListContent documentKey={data.node} {...props} />;
}

function DocumentHistoryListContent({
  documentKey,
  activeRevisionId,
  onSelect,
  onClose,
}: Omit<DocumentHistoryListProps, "queryRef"> & {
  documentKey: DocumentHistoryList_document$key;
}) {
  const { t } = useTranslation("organizations/documents");
  const { data, hasNext, isLoadingNext, loadNext, refetch } = usePaginationFragment<
    DocumentHistoryListPaginationQuery,
    DocumentHistoryList_document$key
  >(documentHistoryListFragment, documentKey);
  const [versionedOnly, setVersionedOnly] = useState(false);

  // The history loads once, so a publish, approval or rejection would leave it
  // stale. The document's published and pending versions change with each of
  // them, so a change in either reloads the list with the current filter.
  const versionState = `${data.publishedVersion?.id ?? ""}|${data.pendingApprovalVersion?.id ?? ""}`;
  const loadedVersionState = useRef(versionState);

  useEffect(() => {
    if (loadedVersionState.current === versionState) {
      return;
    }
    loadedVersionState.current = versionState;
    refetch(
      { filter: versionedOnly ? { versioned: true } : null },
      { fetchPolicy: "store-and-network" },
    );
  }, [refetch, versionState, versionedOnly]);

  const { root, heading, filter, body, viewport, content, more } = documentHistoryList();

  function handleVersionedOnlyChange(checked: boolean) {
    setVersionedOnly(checked);
    refetch(
      { filter: checked ? { versioned: true } : null },
      { fetchPolicy: "store-and-network" },
    );
  }

  return (
    <div className={root()}>
      <div className={heading()}>
        <Text size={2} weight="medium">{t("details.history.title")}</Text>
        <IconButton
          size={1}
          variant="ghost"
          color="neutral"
          aria-label={t("details.history.close")}
          onClick={onClose}
        >
          <XIcon />
        </IconButton>
      </div>
      <label className={filter()}>
        <Text size={1} color="faint">{t("details.history.versionedOnly")}</Text>
        <Switch
          size={1}
          checked={versionedOnly}
          aria-label={t("details.history.versionedOnly")}
          onCheckedChange={handleVersionedOnlyChange}
        />
      </label>
      <ScrollArea.Root className={body()}>
        <ScrollArea.Viewport className={viewport()}>
          <ScrollArea.Content className={content()}>
            <DocumentHistoryListItem
              revisionKey={null}
              active={activeRevisionId == null}
              onSelect={() => onSelect(null)}
            />
            {data.history.edges.map(({ node }) => (
              <DocumentHistoryListItem
                key={node.id}
                revisionKey={node}
                active={activeRevisionId === node.id}
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
                onClick={() => loadNext(historyPageSize)}
              >
                {t("details.history.loadMore")}
              </Button>
            )}
          </ScrollArea.Content>
        </ScrollArea.Viewport>
      </ScrollArea.Root>
    </div>
  );
}
