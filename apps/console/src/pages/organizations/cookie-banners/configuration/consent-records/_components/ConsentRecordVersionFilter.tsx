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
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectList } from "@probo/ui/src/v2/Select/SelectList";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";
import { usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentRecordVersionFilter_cookieBanner$key } from "#/__generated__/core/ConsentRecordVersionFilter_cookieBanner.graphql";
import type { ConsentRecordVersionFilterPaginationQuery } from "#/__generated__/core/ConsentRecordVersionFilterPaginationQuery.graphql";

const versionFilterPageSize = 10;

export const consentRecordVersionFilterFragment = graphql`
  fragment ConsentRecordVersionFilter_cookieBanner on CookieBanner
  @refetchable(queryName: "ConsentRecordVersionFilterPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    versions(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: CREATED_AT, direction: DESC }
    ) @connection(key: "ConsentRecordVersionFilter_versions") {
      edges {
        node {
          id
          version
          state
        }
      }
    }
  }
`;

interface ConsentRecordVersionFilterProps {
  cookieBannerKey: ConsentRecordVersionFilter_cookieBanner$key;
  version: number | null;
  onVersionChange: (version: number | null) => void;
}

export function ConsentRecordVersionFilter({
  cookieBannerKey,
  version,
  onVersionChange,
}: ConsentRecordVersionFilterProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    ConsentRecordVersionFilterPaginationQuery,
    ConsentRecordVersionFilter_cookieBanner$key
  >(consentRecordVersionFilterFragment, cookieBannerKey);
  const allVersionsLabel = t("consentRecordsPage.filters.all");
  const versions = (data.versions?.edges ?? []).flatMap((edge) => {
    const node = edge?.node;
    if (node == null || node.state !== "PUBLISHED") {
      return [];
    }
    return [node];
  });

  return (
    <Select
      value={version}
      onValueChange={(value: number | null) => {
        onVersionChange(value);
      }}
    >
      <SelectTrigger
        size={2}
        placeholder={t("consentRecordsPage.filters.bannerVersion")}
        aria-label={t("consentRecordsPage.filters.bannerVersion")}
      >
        {(value: number | null) => (
          value != null
            ? t("configLayout.callout.version", { version: value })
            : allVersionsLabel
        )}
      </SelectTrigger>
      <SelectPopup align="end">
        <SelectList>
          <SelectItem value={null}>{allVersionsLabel}</SelectItem>
          {versions.map(node => (
            <SelectItem key={node.id} value={node.version}>
              {t("configLayout.callout.version", { version: node.version })}
            </SelectItem>
          ))}
        </SelectList>
        {hasNext && (
          <Button
            variant="ghost"
            color="neutral"
            size={1}
            className="w-full"
            loading={isLoadingNext}
            onClick={() => {
              loadNext(versionFilterPageSize);
            }}
          >
            {t("consentRecordsPage.filters.seeMore")}
          </Button>
        )}
      </SelectPopup>
    </Select>
  );
}
