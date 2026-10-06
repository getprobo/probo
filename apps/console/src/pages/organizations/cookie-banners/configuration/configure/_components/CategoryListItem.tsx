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

import { ArrowDownIcon, ArrowUpIcon, PencilSimpleIcon, TrashIcon } from "@phosphor-icons/react";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CategoryListItem_cookieCategory$key } from "#/__generated__/core/CategoryListItem_cookieCategory.graphql";
import type { CategoryListItemReorderMutation } from "#/__generated__/core/CategoryListItemReorderMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { cookieBannerCategoriesSection } from "../../../variants";

import { CategoryDrawer } from "./CategoryDrawer";
import { DeleteCategoryDialog } from "./DeleteCategoryDialog";

const fragment = graphql`
  fragment CategoryListItem_cookieCategory on CookieCategory {
    id
    name
    description
    kind
    canUpdate: permission(action: "core:cookie-category:update")
    canDelete: permission(action: "core:cookie-category:delete")
    ...CategoryDrawer_cookieCategory
    ...DeleteCategoryDialog_cookieCategory
  }
`;

const reorderMutation = graphql`
  mutation CategoryListItemReorderMutation($input: ReorderCookieCategoryInput!) {
    reorderCookieCategory(input: $input) {
      cookieBanner {
        id
        categories(first: 50, orderBy: { field: RANK, direction: ASC }, filter: { excludeKind: UNCATEGORISED })
          @connection(key: "CategoryList_categories") {
          edges {
            node {
              id
              rank
              ...CategoryListItem_cookieCategory
            }
          }
        }
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface CategoryListItemProps {
  categoryKey: CategoryListItem_cookieCategory$key;
  connectionId: string;
  isFirst: boolean;
  isLast: boolean;
  aboveRank?: number;
  belowRank?: number;
}

export function CategoryListItem({
  categoryKey,
  connectionId,
  isFirst,
  isLast,
  aboveRank,
  belowRank,
}: CategoryListItemProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const category = useFragment(fragment, categoryKey);
  const { title, description, actions } = cookieBannerCategoriesSection();
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [reorderCategory, isReordering] = useMutation<CategoryListItemReorderMutation>(
    reorderMutation,
    {
      errorToast: t("categoryListItem.errors.reorder"),
    },
  );
  const canDelete = category.canDelete && category.kind === "NORMAL";

  function moveTo(rank: number | undefined) {
    if (rank == null) {
      return;
    }

    void reorderCategory({
      variables: {
        input: {
          cookieCategoryId: category.id,
          rank,
        },
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <ListItem>
      <ListItemContent>
        <div className={title()}>
          <Text size={2} weight="medium" color="neutral" highContrast>
            {category.name}
          </Text>
          {category.kind === "NECESSARY" && (
            <Badge color="neutral" variant="soft">
              {t("categoryListItem.required")}
            </Badge>
          )}
        </div>
        <Text size={1} color="faint" className={description()}>
          {category.description}
        </Text>
      </ListItemContent>
      <div className={actions()}>
        {category.canUpdate && (
          <>
            <IconButton
              size={1}
              variant="ghost"
              color="neutral"
              disabled={isFirst || isReordering}
              aria-label={t("categoryListItem.actions.moveUp")}
              onClick={() => moveTo(aboveRank)}
            >
              <ArrowUpIcon />
            </IconButton>
            <IconButton
              size={1}
              variant="ghost"
              color="neutral"
              disabled={isLast || isReordering}
              aria-label={t("categoryListItem.actions.moveDown")}
              onClick={() => moveTo(belowRank)}
            >
              <ArrowDownIcon />
            </IconButton>
            <IconButton
              size={1}
              variant="ghost"
              color="neutral"
              aria-label={t("categoryListItem.actions.edit")}
              onClick={() => setEditOpen(true)}
            >
              <PencilSimpleIcon />
            </IconButton>
          </>
        )}
        {canDelete && (
          <IconButton
            size={1}
            variant="ghost"
            color="red"
            aria-label={t("categoryListItem.actions.delete")}
            onClick={() => setDeleteOpen(true)}
          >
            <TrashIcon />
          </IconButton>
        )}
      </div>
      <CategoryDrawer
        categoryKey={category}
        open={editOpen}
        onOpenChange={setEditOpen}
      />
      <DeleteCategoryDialog
        categoryKey={category}
        connectionId={connectionId}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
      />
    </ListItem>
  );
}
