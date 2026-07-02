// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
