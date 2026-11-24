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

import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Trans, useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentHistoryListItem_revision$key } from "#/__generated__/core/DocumentHistoryListItem_revision.graphql";

import { documentVersionStatus } from "../../_lib/documentVersionStatus";
import { useRevisionDate } from "../../_lib/useRevisionDate";
import { documentHistoryListItem } from "../../variants";

const documentHistoryListItemFragment = graphql`
  fragment DocumentHistoryListItem_revision on DocumentRevision {
    authorName
    endedAt
    changeCount
    synthetic
    versions {
      id
      status
      major
      minor
      changelog
    }
  }
`;

interface DocumentHistoryListItemProps {
  // Null for the live document, listed first.
  revisionKey: DocumentHistoryListItem_revision$key | null;
  active: boolean;
  onSelect: () => void;
}

// A revision of the history. A version is a tag on it: the version number
// and status replace the author when the revision carries one.
export function DocumentHistoryListItem({ revisionKey, active, onSelect }: DocumentHistoryListItemProps) {
  const revision = useFragment(documentHistoryListItemFragment, revisionKey);
  const { t } = useTranslation("organizations/documents");
  const formatDate = useRevisionDate();
  const { root, header, changelog } = documentHistoryListItem({ active });
  const versions = revision?.versions ?? [];

  return (
    <button type="button" className={root()} aria-current={active ? "true" : undefined} onClick={onSelect}>
      {revision == null
        ? (
            <div className={header()}>
              <Text size={2} weight="medium">{t("details.history.current")}</Text>
            </div>
          )
        : (
            <>
              {versions.length > 0
                ? versions.map((version) => {
                    const status = documentVersionStatus(version.status);
                    return (
                      <div key={version.id} className={header()}>
                        <Text size={2} weight="medium">
                          <Trans
                            ns="organizations/documents"
                            i18nKey="details.history.version"
                            values={{ version: `${version.major}.${version.minor}` }}
                            components={{ number: <span className="font-mono tabular-nums" /> }}
                          />
                        </Text>
                        {status != null && (
                          <Badge size={1} color={status.color}>{t(status.label)}</Badge>
                        )}
                      </div>
                    );
                  })
                : (
                    <div className={header()}>
                      <Text size={2} weight="medium" className="truncate">
                        {revision.authorName ?? t("details.history.unknownAuthor")}
                      </Text>
                    </div>
                  )}
              <Text size={1} color="faint">
                {formatDate(revision.endedAt)}
                {!revision.synthetic && ` · ${t("details.history.changes", { count: revision.changeCount })}`}
              </Text>
              {versions.map(version => version.changelog !== ""
                ? (
                    <Text key={version.id} size={1} color="faint" className={changelog()}>
                      {version.changelog}
                    </Text>
                  )
                : null)}
            </>
          )}
    </button>
  );
}
