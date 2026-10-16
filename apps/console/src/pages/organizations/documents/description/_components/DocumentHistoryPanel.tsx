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

import { dateTimeFormat } from "@probo/i18n";
import { Button, IconClock, type RichEditorHistoryHighlight } from "@probo/ui";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, useMutation, usePaginationFragment } from "react-relay";

import type { DocumentHistoryPanel_version$key } from "#/__generated__/core/DocumentHistoryPanel_version.graphql";
import type { DocumentHistoryPanelPaginationQuery } from "#/__generated__/core/DocumentHistoryPanelPaginationQuery.graphql";
import type { DocumentHistoryPanelQuery } from "#/__generated__/core/DocumentHistoryPanelQuery.graphql";
import type { DocumentHistoryPanelRestoreMutation } from "#/__generated__/core/DocumentHistoryPanelRestoreMutation.graphql";
import type { DocumentHistoryPanelSnapshotQuery } from "#/__generated__/core/DocumentHistoryPanelSnapshotQuery.graphql";
import { RichEditor } from "#/pages/organizations/_components/RichEditor";

const documentHistoryPanelQuery = graphql`
  query DocumentHistoryPanelQuery($versionId: ID!) {
    version: node(id: $versionId) {
      __typename
      ... on DocumentVersion {
        ...DocumentHistoryPanel_version
      }
    }
  }
`;

const documentHistoryPanelFragment = graphql`
  fragment DocumentHistoryPanel_version on DocumentVersion
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 25 }
    after: { type: "CursorKey" }
  )
  @refetchable(queryName: "DocumentHistoryPanelPaginationQuery") {
    collaborationHistory(first: $first, after: $after)
      @connection(key: "DocumentHistoryPanel_collaborationHistory") {
      edges {
        node {
          id
          heads
          previousHeads
          authorName
          startedAt
          endedAt
        }
      }
    }
  }
`;

const documentHistorySnapshotQuery = graphql`
  query DocumentHistoryPanelSnapshotQuery(
    $versionId: ID!
    $heads: [String!]!
    $previousHeads: [String!]
  ) {
    version: node(id: $versionId) {
      __typename
      ... on DocumentVersion {
        collaborationHistorySnapshot(heads: $heads, previousHeads: $previousHeads) {
          content {
            prosemirror
          }
          highlights {
            kind
            from
            to
            text
          }
        }
      }
    }
  }
`;

const restoreMutation = graphql`
  mutation DocumentHistoryPanelRestoreMutation($input: RestoreDocumentCollaborationInput!) {
    restoreDocumentCollaboration(input: $input) {
      documentVersion {
        id
        content {
          prosemirror
        }
        status
      }
    }
  }
`;

type HistorySession = {
  id: string;
  heads: readonly string[];
  previousHeads: readonly string[];
  authorName: string | null | undefined;
  startedAt: string;
  endedAt: string;
};

export function DocumentHistoryView(props: {
  versionId: string;
  canRestore: boolean;
  onClose: () => void;
  onRestored: () => void;
}) {
  const { versionId, canRestore, onClose, onRestored } = props;
  const data = useLazyLoadQuery<DocumentHistoryPanelQuery>(
    documentHistoryPanelQuery,
    { versionId },
    { fetchPolicy: "network-only" },
  );

  if (data.version?.__typename !== "DocumentVersion") {
    throw new Error("invalid type for node");
  }

  return (
    <DocumentHistorySessions
      versionKey={data.version}
      versionId={versionId}
      canRestore={canRestore}
      onClose={onClose}
      onRestored={onRestored}
    />
  );
}

function DocumentHistorySessions(props: {
  versionKey: DocumentHistoryPanel_version$key;
  versionId: string;
  canRestore: boolean;
  onClose: () => void;
  onRestored: () => void;
}) {
  const { versionKey, versionId, canRestore, onClose, onRestored } = props;
  const { t, i18n } = useTranslation();
  const toast = useToast();
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    DocumentHistoryPanelPaginationQuery,
    DocumentHistoryPanel_version$key
  >(documentHistoryPanelFragment, versionKey);

  const sessions = data.collaborationHistory.edges.map(edge => edge.node);
  const [selectedId, setSelectedId] = useState<string | null>(sessions[0]?.id ?? null);
  const selected = sessions.find(session => session.id === selectedId) ?? sessions[0];

  const [restore, restoring] = useMutation<DocumentHistoryPanelRestoreMutation>(restoreMutation);

  return (
    <div className="flex flex-1 min-h-0 min-w-0">
      <div className="flex flex-col flex-1 min-w-0 min-h-0">
        <div className="flex justify-end gap-2 px-4 py-2">
          <Button variant="secondary" onClick={onClose}>
            {t("documentDescriptionPage.history.close")}
          </Button>
        </div>
        {selected
          ? (
              <Suspense
                fallback={(
                  <div className="flex-1 px-8 py-6 text-sm text-txt-secondary">
                    {t("documentDescriptionPage.history.loading")}
                  </div>
                )}
              >
                <DocumentHistoryPreview
                  key={selected.id}
                  versionId={versionId}
                  heads={selected.heads}
                  previousHeads={selected.previousHeads}
                />
              </Suspense>
            )
          : (
              <div className="flex-1 px-8 py-6 text-sm text-txt-secondary">
                {t("documentDescriptionPage.history.empty")}
              </div>
            )}
      </div>
      <aside className="flex w-80 shrink-0 flex-col border-l border-border-solid bg-level-1">
        <div className="flex items-center gap-2 border-b border-border-solid px-4 py-3 text-sm font-medium">
          <IconClock size={16} />
          {t("documentDescriptionPage.history.title")}
        </div>
        <div className="flex-1 overflow-auto">
          {groupSessionsByDay(sessions).map(group => (
            <div key={group.key}>
              <div className="px-4 pt-3 pb-1 text-xs font-medium text-txt-secondary">
                {dayLabel(group.key, t, i18n.language)}
              </div>
              {group.sessions.map(session => {
                const active = session.id === selected?.id;
                return (
                  <div key={session.id} className="px-2">
                    <button
                      type="button"
                      className={`w-full rounded px-2 py-2 text-left ${active ? "bg-active" : "hover:bg-tertiary-hover"}`}
                      onClick={() => setSelectedId(session.id)}
                    >
                      <div className="text-sm text-txt-primary">
                        {session.authorName || t("documentDescriptionPage.history.unknownAuthor")}
                      </div>
                      <div className="text-xs text-txt-secondary">
                        {dateTimeFormat(i18n.language, session.endedAt)}
                      </div>
                    </button>
                    {active && canRestore && (
                      <div className="px-2 pb-2">
                        <Button
                          variant="secondary"
                          disabled={restoring}
                          onClick={() => {
                            restore({
                              variables: {
                                input: {
                                  documentVersionId: versionId,
                                  heads: [...session.heads],
                                },
                              },
                              onCompleted: (_data, errors) => {
                                if (errors?.length) {
                                  toast.add({
                                    title: t("documentDescriptionPage.errors.title"),
                                    description: t("documentDescriptionPage.history.restoreFailed"),
                                    type: "error",
                                  });
                                  return;
                                }

                                onRestored();
                              },
                              onError: () => {
                                toast.add({
                                  title: t("documentDescriptionPage.errors.title"),
                                  description: t("documentDescriptionPage.history.restoreFailed"),
                                  type: "error",
                                });
                              },
                            });
                          }}
                        >
                          {t("documentDescriptionPage.history.restore")}
                        </Button>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          ))}
          {hasNext && (
            <div className="px-4 py-3">
              <Button
                variant="secondary"
                disabled={isLoadingNext}
                onClick={() => loadNext(25)}
              >
                {t("documentDescriptionPage.history.loadMore")}
              </Button>
            </div>
          )}
        </div>
      </aside>
    </div>
  );
}

function DocumentHistoryPreview(props: {
  versionId: string;
  heads: readonly string[];
  previousHeads: readonly string[];
}) {
  const { versionId, heads, previousHeads } = props;
  const data = useLazyLoadQuery<DocumentHistoryPanelSnapshotQuery>(
    documentHistorySnapshotQuery,
    {
      versionId,
      heads: [...heads],
      previousHeads: [...previousHeads],
    },
  );

  if (data.version?.__typename !== "DocumentVersion" || !data.version.collaborationHistorySnapshot) {
    throw new Error("invalid type for node");
  }

  const snapshot = data.version.collaborationHistorySnapshot;
  const highlights: RichEditorHistoryHighlight[] = snapshot.highlights.map(highlight => ({
    kind: highlight.kind,
    from: highlight.from,
    to: highlight.to,
    text: highlight.text,
  }));

  return (
    <RichEditor
      className="flex-1"
      content={snapshot.content.prosemirror}
      disabled
      historyHighlights={highlights}
      data-theme="document"
    />
  );
}

function groupSessionsByDay(sessions: readonly HistorySession[]) {
  const groups: { key: string; sessions: HistorySession[] }[] = [];
  for (const session of sessions) {
    const key = new Date(session.endedAt).toDateString();
    const current = groups[groups.length - 1];
    if (current?.key === key) {
      current.sessions.push(session);
      continue;
    }

    groups.push({ key, sessions: [session] });
  }

  return groups;
}

function dayLabel(
  day: string,
  t: (key: string) => string,
  language: string,
) {
  const date = new Date(day);
  const today = startOfDay(new Date());
  const target = startOfDay(date);
  const diffDays = Math.round((today.getTime() - target.getTime()) / 86_400_000);
  if (diffDays === 0) return t("documentDescriptionPage.history.today");
  if (diffDays === 1) return t("documentDescriptionPage.history.yesterday");
  return dateTimeFormat(language, date.toISOString());
}

function startOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}
