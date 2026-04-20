// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card, IconPlusSmall, useConfirm, useToast } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieBannerCookiesPageDeleteMutation } from "#/__generated__/core/CookieBannerCookiesPageDeleteMutation.graphql";
import type { CookieBannerCookiesPageQuery } from "#/__generated__/core/CookieBannerCookiesPageQuery.graphql";

import { CategoryDialog } from "../_components/CategoryDialog";

import { CategorySection } from "./_components/CategorySection";

export const cookieBannerCookiesPageQuery = graphql`
  query CookieBannerCookiesPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        id
        categories(first: 50, orderBy: { field: RANK, direction: ASC })
          @connection(key: "CookieBannerCookiesPage_categories")
          @required(action: THROW) {
          __id
          edges {
            node {
              id
              rank
              name
              kind
              ...CategorySectionFragment
            }
          }
        }
      }
    }
  }
`;

const deleteCategoryMutation = graphql`
  mutation CookieBannerCookiesPageDeleteMutation(
    $input: DeleteCookieCategoryInput!
    $connections: [ID!]!
  ) {
    deleteCookieCategory(input: $input) {
      deletedCookieCategoryId @deleteEdge(connections: $connections)
      cookieBanner {
        id
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface CookieBannerCookiesPageProps {
  queryRef: PreloadedQuery<CookieBannerCookiesPageQuery>;
}

export default function CookieBannerCookiesPage({
  queryRef,
}: CookieBannerCookiesPageProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const data = usePreloadedQuery(cookieBannerCookiesPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;
  const connectionId = banner.categories.__id;
  const categories = banner.categories.edges.map(e => e.node);
  const sorted = [...categories].sort((a, b) => a.rank - b.rank);

  const [deleteCategory]
    = useMutation<CookieBannerCookiesPageDeleteMutation>(deleteCategoryMutation);

  const [showCreateDialog, setShowCreateDialog] = useState(false);

  const handleDeleteCategory = (categoryId: string, categoryName: string) => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteCategory({
            variables: {
              input: { cookieCategoryId: categoryId },
              connections: [connectionId],
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: __("Error"),
                  description: errors[0].message,
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Category deleted"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete category"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: __("Are you sure you want to delete the category \"%s\"? Any cookies in this category will be moved to Uncategorised.").replace("%s", categoryName),
        variant: "danger",
        label: __("Delete"),
      },
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">{__("Cookies")}</h3>
        <Button variant="secondary" onClick={() => setShowCreateDialog(true)}>
          <IconPlusSmall size={16} />
          {__("Add Category")}
        </Button>
      </div>

      {sorted.length === 0 && (
        <Card className="border p-8 text-center text-muted-foreground">
          {__("No categories yet. Add a category to start managing cookies.")}
        </Card>
      )}

      {sorted.map(category => (
        <CategorySection
          key={category.id}
          categoryKey={category}
          onDelete={
            category.kind === "NORMAL"
              ? () => handleDeleteCategory(category.id, category.name)
              : undefined
          }
        />
      ))}

      {showCreateDialog && (
        <CategoryDialog
          cookieBannerId={banner.id}
          connectionId={connectionId}
          nextRank={sorted.length > 0 ? sorted[sorted.length - 1].rank + 1 : 0}
          onOpenChange={setShowCreateDialog}
        />
      )}
    </div>
  );
}
