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

import { CaretLeftIcon, NewspaperIcon } from "@phosphor-icons/react";
import { Link } from "@probo/ui/src/v2/Button/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";
import { formatDate } from "#/lib/datetime/formatDate";
import { NotFoundError } from "#/lib/relay/errors";

import type { UpdateDetailPageQuery } from "./__generated__/UpdateDetailPageQuery.graphql";
import { UpdatesSubscribeButton } from "./_components/UpdatesSubscribeButton";
import { updateArticle } from "./_components/variants";

export const updateDetailPageQuery = graphql`
  query UpdateDetailPageQuery($updateId: ID!) @throwOnFieldError {
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
    throw new NotFoundError("Update not found");
  }
  const update = data.node;

  const { toolbar, content, article, meta, metaIcon, body } = updateArticle();

  return (
    <>
      <HeaderBand>
        <div className={toolbar()}>
          <Link to="/updates" variant="soft" color="neutral" highContrast iconStart={<CaretLeftIcon />}>
            {t("backToUpdates")}
          </Link>
          <UpdatesSubscribeButton />
        </div>
      </HeaderBand>
      <div className={content()}>
        <article className={article()}>
          <div className={meta()}>
            <NewspaperIcon weight="light" className={metaIcon()} />
            <Text size={1} color="gold">
              {formatDate(update.updatedAt, i18n.language)}
            </Text>
          </div>
          <Heading level={1} size={7} weight="medium" highContrast>
            {update.title}
          </Heading>
          <Text size={3} className={body()}>
            {update.body}
          </Text>
        </article>
      </div>
    </>
  );
}
