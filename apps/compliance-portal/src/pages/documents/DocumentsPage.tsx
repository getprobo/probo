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

import { groupBy } from "@probo/helpers";
import { useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery, useRefetchableFragment } from "react-relay";

import { ListErrorBoundary } from "#/components/errors/ListErrorBoundary";
import { PageHeader } from "#/components/PageHeader/PageHeader";

import type { DocumentsPage_query$key } from "./__generated__/DocumentsPage_query.graphql";
import type { DocumentsPageQuery } from "./__generated__/DocumentsPageQuery.graphql";
import type { DocumentsPageRefetchQuery } from "./__generated__/DocumentsPageRefetchQuery.graphql";
import { AuditReportListItem } from "./_components/AuditReportListItem";
import { DocumentListItem } from "./_components/DocumentListItem";
import { DocumentSection } from "./_components/DocumentSection";
import { DocumentsEmpty } from "./_components/DocumentsEmpty";
import { DocumentsToolbar } from "./_components/DocumentsToolbar";
import { TrustCenterFileListItem } from "./_components/TrustCenterFileListItem";
import { toQueryVariables } from "./_lib/toQueryVariables";
import { useDocumentTab } from "./_lib/useDocumentTab";
import { documentsLayout } from "./variants";

export const documentsPageQuery = graphql`
  query DocumentsPageQuery($visibility: TrustCenterVisibility) {
    ...DocumentsPage_query @arguments(visibility: $visibility)
  }
`;

const documentsPageFragment = graphql`
  fragment DocumentsPage_query on Query
  @refetchable(queryName: "DocumentsPageRefetchQuery")
  @argumentDefinitions(visibility: { type: "TrustCenterVisibility" }) {
    currentTrustCenter @required(action: THROW) {
      documents(first: 250, filter: { visibility: $visibility }) {
        edges {
          node {
            id
            documentType
            ...DocumentListItem_document
          }
        }
      }
      audits(first: 250, filter: { visibility: $visibility }) {
        edges {
          node {
            id
            reportFile {
              id
            }
            ...AuditReportListItem_audit
          }
        }
      }
      trustCenterFiles(first: 250, filter: { visibility: $visibility }) {
        edges {
          node {
            id
            category
            ...TrustCenterFileListItem_file
          }
        }
      }
    }
  }
`;

interface DocumentsPageProps {
  queryRef: PreloadedQuery<DocumentsPageQuery>;
}

// Trust Center documents page: a unified list of published documents, uploaded
// files, and audit reports, grouped into category sections. The All/Public/
// Private tabs are backed by a server-side visibility filter.
export function DocumentsPage({ queryRef }: DocumentsPageProps) {
  const { t } = useTranslation("documents");
  const root = usePreloadedQuery<DocumentsPageQuery>(documentsPageQuery, queryRef);
  const [data, refetch] = useRefetchableFragment<DocumentsPageRefetchQuery, DocumentsPage_query$key>(
    documentsPageFragment,
    root,
  );

  const { tab } = useDocumentTab();
  const [isRefetching, startTransition] = useTransition();

  // Keep the displayed slice in sync with the active tab. Seed from the tab the
  // preloaded query actually loaded with (`queryRef.variables`) instead of
  // assuming the first render matches the URL: if the tab changed while the
  // initial preload was in flight, this reconciles by refetching rather than
  // showing the wrong slice. Refetch inside a transition so the toolbar and
  // current results stay mounted (dimmed via `isRefetching`) while it loads.
  //
  // `requestedVisibility` de-dupes in-flight requests; `loadedVisibility` only
  // advances when the *latest* refetch settles, so an out-of-order or failed
  // refetch can't leave the list showing a different tab than the toolbar.
  const initialVisibility = queryRef.variables.visibility ?? null;
  const loadedVisibility = useRef(initialVisibility);
  const requestedVisibility = useRef(initialVisibility);
  useEffect(() => {
    const target = toQueryVariables(tab).visibility ?? null;
    if (target === requestedVisibility.current) {
      return;
    }
    requestedVisibility.current = target;
    startTransition(() => {
      refetch(toQueryVariables(tab), {
        fetchPolicy: "store-or-network",
        onComplete: (error) => {
          if (requestedVisibility.current !== target) {
            // A newer tab was requested; ignore this stale settle.
            return;
          }
          if (error) {
            // Allow re-selecting this tab to retry after a failed refetch.
            requestedVisibility.current = loadedVisibility.current;
            return;
          }
          loadedVisibility.current = target;
        },
      });
    });
  }, [refetch, tab]);

  const { currentTrustCenter } = data;
  const documentNodes = currentTrustCenter.documents.edges.map(edge => edge.node);
  const fileNodes = currentTrustCenter.trustCenterFiles.edges.map(edge => edge.node);
  const auditNodes = currentTrustCenter.audits.edges
    .map(edge => edge.node)
    .filter(node => node.reportFile != null);

  const total = documentNodes.length + fileNodes.length + auditNodes.length;

  const documentGroups = Object.entries(groupBy(documentNodes, node => node.documentType))
    .map(([key, nodes]) => ({ key, nodes }))
    .sort((a, b) => t(`types.${a.key}`).localeCompare(t(`types.${b.key}`)));
  const fileGroups = Object.entries(groupBy(fileNodes, node => node.category))
    .map(([key, nodes]) => ({ key, nodes }))
    .sort((a, b) => a.key.localeCompare(b.key));

  const { page, results } = documentsLayout({ busy: isRefetching });

  return (
    <>
      <PageHeader title={t("title")} count={total} flushBottomSpace>
        <DocumentsToolbar />
      </PageHeader>
      <div className={page()}>
        <div aria-busy={isRefetching} className={results()}>
          <ListErrorBoundary
            onRetry={done => startTransition(() => {
              refetch(toQueryVariables(tab), { fetchPolicy: "network-only", onComplete: done });
            })}
          >
            {total === 0
              ? <DocumentsEmpty />
              : (
                  <>
                    {auditNodes.length > 0 && (
                      <DocumentSection title={t("sections.reports")}>
                        {auditNodes.map(node => (
                          <AuditReportListItem key={node.id} auditKey={node} />
                        ))}
                      </DocumentSection>
                    )}
                    {documentGroups.map(group => (
                      <DocumentSection key={`type:${group.key}`} title={t(`types.${group.key}`)}>
                        {group.nodes.map(node => (
                          <DocumentListItem key={node.id} documentKey={node} />
                        ))}
                      </DocumentSection>
                    ))}
                    {fileGroups.map(group => (
                      <DocumentSection key={`category:${group.key}`} title={group.key}>
                        {group.nodes.map(node => (
                          <TrustCenterFileListItem key={node.id} fileKey={node} />
                        ))}
                      </DocumentSection>
                    ))}
                  </>
                )}
          </ListErrorBoundary>
        </div>
      </div>
    </>
  );
}
