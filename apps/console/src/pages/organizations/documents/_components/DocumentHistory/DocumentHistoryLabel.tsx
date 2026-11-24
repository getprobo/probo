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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { Trans, useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";

import type { DocumentHistoryContentQuery } from "#/__generated__/core/DocumentHistoryContentQuery.graphql";
import type { DocumentHistoryLabel_revision$key } from "#/__generated__/core/DocumentHistoryLabel_revision.graphql";

import { useRevisionDate } from "../../_lib/useRevisionDate";

import { documentHistoryContentQuery } from "./DocumentHistoryContent";

const documentHistoryLabelFragment = graphql`
  fragment DocumentHistoryLabel_revision on DocumentRevision {
    authorName
    endedAt
    versions {
      major
      minor
    }
  }
`;

interface DocumentHistoryLabelProps {
  queryRef: PreloadedQuery<DocumentHistoryContentQuery> | null;
  onShowLatest: () => void;
}

// Says which revision the document shows: the live one, or a previous one
// with its version number or author and its date, and a way back to the live
// one.
export function DocumentHistoryLabel({ queryRef, onShowLatest }: DocumentHistoryLabelProps) {
  const { t } = useTranslation("organizations/documents");

  if (queryRef == null) {
    return <Text size={2} color="faint">{t("details.history.viewingCurrent")}</Text>;
  }

  return (
    <>
      <ErrorBoundary fallback={null}>
        <Suspense fallback={null}>
          <DocumentHistoryLabelLoaded queryRef={queryRef} />
        </Suspense>
      </ErrorBoundary>
      <Button size={1} variant="ghost" color="neutral" className="shrink-0" onClick={onShowLatest}>
        {t("details.history.showLatest")}
      </Button>
    </>
  );
}

function DocumentHistoryLabelLoaded({ queryRef }: { queryRef: PreloadedQuery<DocumentHistoryContentQuery> }) {
  const data = usePreloadedQuery<DocumentHistoryContentQuery>(documentHistoryContentQuery, queryRef);

  if (data.node.__typename !== "Document" || data.node.revision == null) {
    return null;
  }

  return <DocumentHistoryLabelText revisionKey={data.node.revision} />;
}

function DocumentHistoryLabelText({ revisionKey }: { revisionKey: DocumentHistoryLabel_revision$key }) {
  const revision = useFragment(documentHistoryLabelFragment, revisionKey);
  const { t } = useTranslation("organizations/documents");
  const formatDate = useRevisionDate();
  const version = revision.versions.at(-1);

  return (
    <Text size={2} weight="medium" className="truncate">
      {version != null
        ? (
            <Trans
              ns="organizations/documents"
              i18nKey="details.history.viewingVersion"
              values={{
                version: `${version.major}.${version.minor}`,
                date: formatDate(revision.endedAt),
              }}
              components={{ number: <span className="font-mono tabular-nums" /> }}
            />
          )
        : t("details.history.viewingRevision", {
            author: revision.authorName ?? t("details.history.unknownAuthor"),
            date: formatDate(revision.endedAt),
          })}
    </Text>
  );
}
