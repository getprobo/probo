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

import { CaretLeftIcon, NewspaperIcon } from "@phosphor-icons/react";
import { Link } from "@probo/ui/src/v2/Button/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";
import { formatDate } from "#/lib/datetime/formatDate";

import type { UpdateDetailPageQuery } from "./__generated__/UpdateDetailPageQuery.graphql";
import { UpdatesSubscribeButton } from "./_components/UpdatesSubscribeButton";

export const updateDetailPageQuery = graphql`
  query UpdateDetailPageQuery($updateId: ID!) {
    node(id: $updateId) {
      __typename
      ... on MailingListUpdate {
        title
        body
        updatedAt
      }
    }
  }
`;

interface UpdateDetailPageProps {
  queryRef: PreloadedQuery<UpdateDetailPageQuery>;
}

export function UpdateDetailPage({ queryRef }: UpdateDetailPageProps) {
  const { t, i18n } = useTranslation("updates");
  const data = usePreloadedQuery<UpdateDetailPageQuery>(updateDetailPageQuery, queryRef);

  if (data.node?.__typename !== "MailingListUpdate") {
    throw new Error("Update not found");
  }
  const update = data.node;

  return (
    <>
      <HeaderBand>
        <div className="flex w-full items-center justify-between gap-4">
          <Link to="/updates" variant="soft" color="neutral" highContrast iconStart={<CaretLeftIcon />}>
            {t("backToUpdates")}
          </Link>
          <UpdatesSubscribeButton />
        </div>
      </HeaderBand>
      <div className="flex w-full flex-col items-center px-8 py-8">
        <article className="flex w-full max-w-2xl flex-col gap-4">
          <div className="flex items-center gap-1.5">
            <NewspaperIcon weight="light" className="size-4 text-gold-9" />
            <Text size={1} color="gold">
              {formatDate(update.updatedAt, i18n.language)}
            </Text>
          </div>
          <Heading level={1} size={7} weight="medium" highContrast>
            {update.title}
          </Heading>
          <Text size={3} className="block whitespace-pre-wrap">
            {update.body}
          </Text>
        </article>
      </div>
    </>
  );
}
