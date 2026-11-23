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

import type { DocumentVersionListItem_version$key } from "#/__generated__/core/DocumentVersionListItem_version.graphql";

import { documentVersionStatus } from "../../_lib/documentVersionStatus";
import { useVersionDate } from "../../_lib/useVersionDate";
import { documentVersionListItem } from "../../variants";

const documentVersionListItemFragment = graphql`
  fragment DocumentVersionListItem_version on DocumentVersion {
    status
    major
    minor
    changelog
    publishedAt
    createdAt
  }
`;

interface DocumentVersionListItemProps {
  // Null for the live document, listed first.
  versionKey: DocumentVersionListItem_version$key | null;
  active: boolean;
  onSelect: () => void;
}

export function DocumentVersionListItem({ versionKey, active, onSelect }: DocumentVersionListItemProps) {
  const version = useFragment(documentVersionListItemFragment, versionKey);
  const { t } = useTranslation("organizations/documents");
  const formatDate = useVersionDate();
  const { root, header, changelog } = documentVersionListItem({ active });
  const status = version == null ? null : documentVersionStatus(version.status);

  return (
    <button type="button" className={root()} aria-current={active ? "true" : undefined} onClick={onSelect}>
      <div className={header()}>
        <Text size={2} weight="medium">
          {version == null
            ? t("details.versions.current")
            : (
                <Trans
                  ns="organizations/documents"
                  i18nKey="details.versions.version"
                  values={{ version: `${version.major}.${version.minor}` }}
                  components={{ number: <span className="font-mono tabular-nums" /> }}
                />
              )}
        </Text>
        {status != null && (
          <Badge size={1} color={status.color}>{t(status.label)}</Badge>
        )}
      </div>
      {version != null && (
        <>
          <Text size={1} color="faint">{formatDate(version.publishedAt ?? version.createdAt)}</Text>
          {version.changelog !== "" && (
            <Text size={1} color="faint" className={changelog()}>{version.changelog}</Text>
          )}
        </>
      )}
    </button>
  );
}
