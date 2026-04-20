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
import { Badge, Button, Card, IconArrowDown, IconArrowUp, useToast } from "@probo/ui";
import { useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CategoryList_cookieBanner$key } from "#/__generated__/core/CategoryList_cookieBanner.graphql";
import type { CategoryListDeleteMutation } from "#/__generated__/core/CategoryListDeleteMutation.graphql";
import type { CategoryListUpdateMutation } from "#/__generated__/core/CategoryListUpdateMutation.graphql";

import { CategoryDialog } from "./CategoryDialog";

const categoryListFragment = graphql`
  fragment CategoryList_cookieBanner on CookieBanner {
    id
    categories(first: 50, orderBy: { field: RANK, direction: ASC }) @required(action: THROW) {
      __id
      edges {
        node {
          id
          name
          description
          required
          rank
        }
      }
    }
  }
`;

const deleteCategoryMutation = graphql`
  mutation CategoryListDeleteMutation(
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

const updateCategoryMutation = graphql`
  mutation CategoryListUpdateMutation($input: UpdateCookieCategoryInput!) {
    updateCookieCategory(input: $input) {
      cookieCategory {
        id
        name
        description
        rank
        cookies {
          name
          duration
          description
        }
        updatedAt
      }
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

interface CategoryListProps {
  cookieBannerKey: CategoryList_cookieBanner$key;
}

export function CategoryList({ cookieBannerKey }: CategoryListProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  const banner = useFragment(categoryListFragment, cookieBannerKey);
  const connectionId = banner.categories.__id;
  const categories = banner.categories.edges.map(e => e.node);

  const [deleteCategory] = useMutation<CategoryListDeleteMutation>(deleteCategoryMutation);
  const [updateCategory] = useMutation<CategoryListUpdateMutation>(updateCategoryMutation);

  const sorted = [...categories].sort((a, b) => a.rank - b.rank);

  const handleDelete = (categoryId: string) => {
    deleteCategory({
      variables: {
        input: { cookieCategoryId: categoryId },
        connections: [connectionId],
      },
      onCompleted() {
        toast({ title: __("Success"), description: __("Category deleted"), variant: "success" });
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to delete category"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const handleMoveUp = (index: number) => {
    if (index === 0) return;
    const current = sorted[index];
    const above = sorted[index - 1];
    updateCategory({
      variables: { input: { cookieCategoryId: current.id, rank: above.rank } },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to reorder"), error as GraphQLError), variant: "error" });
      },
    });
    updateCategory({
      variables: { input: { cookieCategoryId: above.id, rank: current.rank } },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to reorder"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const handleMoveDown = (index: number) => {
    if (index >= sorted.length - 1) return;
    const current = sorted[index];
    const below = sorted[index + 1];
    updateCategory({
      variables: { input: { cookieCategoryId: current.id, rank: below.rank } },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to reorder"), error as GraphQLError), variant: "error" });
      },
    });
    updateCategory({
      variables: { input: { cookieCategoryId: below.id, rank: current.rank } },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to reorder"), error as GraphQLError), variant: "error" });
      },
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">{__("Categories Sorting")}</h3>
        <Button variant="secondary" onClick={() => setShowCreateDialog(true)}>
          {__("Add Category")}
        </Button>
      </div>
      <p className="text-sm text-txt-secondary">
        {__("Categories will be displayed in your cookie banner in the same order as below.")}
      </p>
      <Card className="divide-y divide-border-low rounded-lg border">
        {sorted.map((category, index) => (
          <div key={category.id} className="p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className="font-medium">{category.name}</span>
                {category.required && (
                  <Badge variant="neutral">{__("Required")}</Badge>
                )}
              </div>
              <div className="flex items-center gap-2">
                <div className="flex items-center gap-1">
                  <button
                    type="button"
                    onClick={() => handleMoveUp(index)}
                    disabled={index === 0}
                    className="p-0.5 rounded cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
                  >
                    <IconArrowUp size={14} />
                  </button>
                  <button
                    type="button"
                    onClick={() => handleMoveDown(index)}
                    disabled={index === sorted.length - 1}
                    className="p-0.5 rounded cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
                  >
                    <IconArrowDown size={14} />
                  </button>
                </div>
                {!category.required && (
                  <Button
                    variant="danger"
                    className="h-6 px-2 text-xs"
                    onClick={() => handleDelete(category.id)}
                  >
                    {__("Delete")}
                  </Button>
                )}
              </div>
            </div>
            <p className="text-sm text-muted-foreground mb-2">{category.description}</p>
          </div>
        ))}
      </Card>

      {showCreateDialog && (
        <CategoryDialog
          cookieBannerId={banner.id}
          connectionId={connectionId}
          onOpenChange={setShowCreateDialog}
        />
      )}
    </div>
  );
}
