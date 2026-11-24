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

import type { DocumentVersionContentQuery } from "#/__generated__/core/DocumentVersionContentQuery.graphql";
import type { DocumentVersionLabel_version$key } from "#/__generated__/core/DocumentVersionLabel_version.graphql";

import { useVersionDate } from "../../_lib/useVersionDate";

import { documentVersionContentQuery } from "./DocumentVersionContent";

const documentVersionLabelFragment = graphql`
  fragment DocumentVersionLabel_version on DocumentVersion {
    major
    minor
    publishedAt
    createdAt
  }
`;

interface DocumentVersionLabelProps {
  queryRef: PreloadedQuery<DocumentVersionContentQuery> | null;
  onShowLatest: () => void;
}

// Says which version the document shows: the live one, or a previous one
// with its number and date and a way back to the live one.
export function DocumentVersionLabel({ queryRef, onShowLatest }: DocumentVersionLabelProps) {
  const { t } = useTranslation("organizations/documents");

  if (queryRef == null) {
    return <Text size={2} color="faint">{t("details.versions.viewingCurrent")}</Text>;
  }

  return (
    <>
      <ErrorBoundary fallback={null}>
        <Suspense fallback={null}>
          <DocumentVersionLabelLoaded queryRef={queryRef} />
        </Suspense>
      </ErrorBoundary>
      <Button size={1} variant="ghost" color="neutral" className="shrink-0" onClick={onShowLatest}>
        {t("details.versions.showLatest")}
      </Button>
    </>
  );
}

function DocumentVersionLabelLoaded({ queryRef }: { queryRef: PreloadedQuery<DocumentVersionContentQuery> }) {
  const data = usePreloadedQuery<DocumentVersionContentQuery>(documentVersionContentQuery, queryRef);

  if (data.node.__typename !== "DocumentVersion") {
    return null;
  }

  return <DocumentVersionLabelText versionKey={data.node} />;
}

function DocumentVersionLabelText({ versionKey }: { versionKey: DocumentVersionLabel_version$key }) {
  const version = useFragment(documentVersionLabelFragment, versionKey);
  const formatDate = useVersionDate();

  return (
    <Text size={2} weight="medium" className="truncate">
      <Trans
        ns="organizations/documents"
        i18nKey="details.versions.viewingVersion"
        values={{
          version: `${version.major}.${version.minor}`,
          date: formatDate(version.publishedAt ?? version.createdAt),
        }}
        components={{ number: <span className="font-mono tabular-nums" /> }}
      />
    </Text>
  );
}
