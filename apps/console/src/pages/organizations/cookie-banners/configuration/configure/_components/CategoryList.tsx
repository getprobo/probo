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

import { PlusIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { List } from "@probo/ui/src/v2/List/List";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CategoryList_cookieBanner$key } from "#/__generated__/core/CategoryList_cookieBanner.graphql";

import { cookieBannerCategoriesSection } from "../../../variants";

import { CategoryCreateDialog } from "./CategoryCreateDialog";
import { CategoryListItem } from "./CategoryListItem";

const categoryListFragment = graphql`
  fragment CategoryList_cookieBanner on CookieBanner {
    id
    canCreate: permission(action: "core:cookie-category:create")
    categories(first: 50, orderBy: { field: RANK, direction: ASC }, filter: { excludeKind: UNCATEGORISED })
      @connection(key: "CategoryList_categories")
      @required(action: THROW) {
      __id
      edges {
        node {
          id
          rank
          ...CategoryListItem_cookieCategory
        }
      }
    }
  }
`;

interface CategoryListProps {
  cookieBannerKey: CategoryList_cookieBanner$key;
}

export function CategoryList({ cookieBannerKey }: CategoryListProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const banner = useFragment(categoryListFragment, cookieBannerKey);
  const { root, intro, heading, empty } = cookieBannerCategoriesSection();
  const [createOpen, setCreateOpen] = useState(false);
  const connectionId = banner.categories.__id;
  const categories = banner.categories.edges.map(edge => edge.node);
  const lastRank = categories.length > 0 ? categories[categories.length - 1].rank : -1;

  return (
    <section className={root()}>
      <div className={intro()}>
        <div className={heading()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("categoryList.title")}
          </Heading>
          <Text size={2} color="faint">
            {t("categoryList.description")}
          </Text>
        </div>
        {banner.canCreate && (
          <Button
            size={2}
            variant="solid"
            color="neutral"
            highContrast
            iconStart={<PlusIcon />}
            onClick={() => setCreateOpen(true)}
          >
            {t("categoryList.actions.add")}
          </Button>
        )}
      </div>
      {categories.length === 0
        ? (
            <Card variant="soft">
              <div className={empty()}>
                <Text size={3} weight="medium" highContrast>
                  {t("categoryList.empty.title")}
                </Text>
                <Text size={2} color="faint">
                  {t("categoryList.empty.description")}
                </Text>
              </div>
            </Card>
          )
        : (
            <List>
              {categories.map((category, index) => (
                <CategoryListItem
                  key={category.id}
                  categoryKey={category}
                  connectionId={connectionId}
                  isFirst={index === 0}
                  isLast={index === categories.length - 1}
                  aboveRank={index > 0 ? categories[index - 1].rank : undefined}
                  belowRank={index < categories.length - 1 ? categories[index + 1].rank : undefined}
                />
              ))}
            </List>
          )}
      <CategoryCreateDialog
        cookieBannerId={banner.id}
        connectionId={connectionId}
        nextRank={lastRank + 1}
        open={createOpen}
        onOpenChange={setCreateOpen}
      />
    </section>
  );
}
