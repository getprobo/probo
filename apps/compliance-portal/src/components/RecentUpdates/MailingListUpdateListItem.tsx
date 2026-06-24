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

import { NewspaperIcon } from "@phosphor-icons/react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { ComplianceArticleItem } from "#/components/ComplianceArticleItem/ComplianceArticleItem";
import { formatRelativeTime } from "#/lib/datetime/relativeTime";

import type { MailingListUpdateListItem_update$key } from "./__generated__/MailingListUpdateListItem_update.graphql";

const mailingListUpdateListItemFragment = graphql`
  fragment MailingListUpdateListItem_update on MailingListUpdate {
    title
    updatedAt
  }
`;

interface MailingListUpdateListItemProps {
  updateKey: MailingListUpdateListItem_update$key;
}

// A single "Recent updates" row. The schema has no category, so we show the
// title and relative date with a single generic icon.
export function MailingListUpdateListItem({ updateKey }: MailingListUpdateListItemProps) {
  const { i18n } = useTranslation();
  const update = useFragment(mailingListUpdateListItemFragment, updateKey);

  return (
    <ComplianceArticleItem
      icon={<NewspaperIcon weight="light" />}
      title={update.title}
      meta={formatRelativeTime(update.updatedAt, i18n.language)}
    />
  );
}
