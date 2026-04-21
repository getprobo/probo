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
import {
  Badge,
  Button,
  Card,
  Dropdown,
  DropdownItem,
  IconArrowBoxLeft,
  IconPencil,
  IconPlusSmall,
  IconTrashCan,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CategorySectionFragment$key } from "#/__generated__/core/CategorySectionFragment.graphql";
import type { CategorySectionMoveCookieMutation } from "#/__generated__/core/CategorySectionMoveCookieMutation.graphql";
import type { CategorySectionUpdateMutation } from "#/__generated__/core/CategorySectionUpdateMutation.graphql";

import { AddCookieRow } from "./AddCookieRow";
import { EditCategoryForm } from "./EditCategoryForm";
import { EditCookieRow } from "./EditCookieRow";

export interface CookieEntry {
  name: string;
  duration: string;
  description: string;
}

export const categorySectionFragment = graphql`
  fragment CategorySectionFragment on CookieCategory {
    id
    name
    description
    kind
    cookies {
      name
      duration
      description
    }
    cookieBanner @required(action: THROW) {
      categories(first: 50, orderBy: { field: RANK, direction: ASC }) @required(action: THROW) {
        edges {
          node {
            id
            name
            cookies {
              name
              duration
              description
            }
          }
        }
      }
    }
  }
`;

const updateCategoryMutation = graphql`
  mutation CategorySectionUpdateMutation(
    $input: UpdateCookieCategoryInput!
  ) {
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

const moveCookieMutation = graphql`
  mutation CategorySectionMoveCookieMutation(
    $input: MoveCookieToCategoryInput!
  ) {
    moveCookieToCategory(input: $input) {
      sourceCookieCategory {
        id
        cookies {
          name
          duration
          description
        }
        updatedAt
      }
      targetCookieCategory {
        id
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

interface CategorySectionProps {
  categoryKey: CategorySectionFragment$key;
  onDelete?: () => void;
}

export function CategorySection({ categoryKey, onDelete }: CategorySectionProps) {
  const category = useFragment(categorySectionFragment, categoryKey);
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [updateCategory, isUpdating]
    = useMutation<CategorySectionUpdateMutation>(updateCategoryMutation);
  const [moveCookie]
    = useMutation<CategorySectionMoveCookieMutation>(moveCookieMutation);

  const [isEditingCategory, setIsEditingCategory] = useState(false);
  const [editingCookieIndex, setEditingCookieIndex] = useState<number | null>(null);
  const [isAddingCookie, setIsAddingCookie] = useState(false);

  const doUpdate = (
    input: Record<string, unknown>,
    onSuccess?: () => void,
  ) => {
    updateCategory({
      variables: {
        input: {
          cookieCategoryId: category.id,
          ...input,
        },
      },
      onCompleted(_response, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Category updated"),
          variant: "success",
        });
        onSuccess?.();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to update category"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleSaveCategory = (name: string, description: string) => {
    doUpdate({ name, description }, () => {
      setIsEditingCategory(false);
    });
  };

  const handleSaveEditCookie = (index: number, cookie: CookieEntry) => {
    if (!cookie.name.trim()) return;
    const newCookies = category.cookies.map((c, i) =>
      i === index
        ? { ...cookie }
        : { name: c.name, duration: c.duration, description: c.description },
    );
    doUpdate({ cookies: newCookies }, () => {
      setEditingCookieIndex(null);
    });
  };

  const handleDeleteCookie = (index: number) => {
    const newCookies = category.cookies
      .filter((_, i) => i !== index)
      .map(c => ({
        name: c.name,
        duration: c.duration,
        description: c.description,
      }));
    doUpdate({ cookies: newCookies });
  };

  const handleSaveNewCookie = (cookie: CookieEntry) => {
    if (!cookie.name.trim()) return;
    const newCookies = [
      ...category.cookies.map(c => ({
        name: c.name,
        duration: c.duration,
        description: c.description,
      })),
      { ...cookie },
    ];
    doUpdate({ cookies: newCookies }, () => {
      setIsAddingCookie(false);
    });
  };

  const allCategories = category.cookieBanner.categories.edges.map(e => e.node) ?? [];
  const siblingCategories = allCategories.filter(c => c.id !== category.id);

  const handleMoveCookie = (cookieIndex: number, targetCategoryId: string) => {
    const cookie = category.cookies[cookieIndex];

    moveCookie({
      variables: {
        input: {
          sourceCookieCategoryId: category.id,
          targetCookieCategoryId: targetCategoryId,
          cookieName: cookie.name,
        },
      },
      onCompleted(_response, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Cookie moved"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to move cookie"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <Card className="border overflow-hidden">
      <div className="p-4">
        {isEditingCategory
          ? (
              <EditCategoryForm
                name={category.name}
                description={category.description}
                isUpdating={isUpdating}
                onSave={handleSaveCategory}
                onCancel={() => setIsEditingCategory(false)}
              />
            )
          : (
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="font-medium">{category.name}</span>
                  {category.kind === "NECESSARY" && (
                    <Badge variant="neutral">{__("Required")}</Badge>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="secondary"
                    onClick={() => setIsEditingCategory(true)}
                  >
                    <IconPencil size={14} />
                    {__("Edit")}
                  </Button>
                  {onDelete && (
                    <Button variant="danger" onClick={onDelete}>
                      <IconTrashCan size={14} />
                      {__("Delete")}
                    </Button>
                  )}
                </div>
              </div>
            )}
        {!isEditingCategory && (
          <>
            <p className="mt-1 text-sm text-muted-foreground">
              {category.description}
            </p>
            <p className="mt-2 text-xs text-txt-secondary/70">
              {__("Block elements until consent is given:")}
              {" "}
              <code className="rounded bg-muted px-1 py-0.5 font-mono text-[11px]">
                data-cookie-consent=&quot;
                {category.name.toLowerCase()}
                &quot;
              </code>
            </p>
          </>
        )}
      </div>

      <table className="w-full text-left">
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Duration")}</Th>
            <Th>{__("Description")}</Th>
            <Th className="w-20" />
          </Tr>
        </Thead>
        <Tbody>
          {category.cookies.map((cookie, index) =>
            editingCookieIndex === index
              ? (
                  <EditCookieRow
                    key={index}
                    cookie={{
                      name: cookie.name,
                      duration: cookie.duration,
                      description: cookie.description,
                    }}
                    isUpdating={isUpdating}
                    onSave={updated => handleSaveEditCookie(index, updated)}
                    onCancel={() => setEditingCookieIndex(null)}
                  />
                )
              : (
                  <Tr key={index}>
                    <Td>
                      <code className="text-sm font-mono">{cookie.name}</code>
                    </Td>
                    <Td className="text-sm text-muted-foreground">
                      {cookie.duration}
                    </Td>
                    <Td className="text-sm text-muted-foreground">
                      {cookie.description}
                    </Td>
                    <Td>
                      <div className="flex items-center gap-1">
                        <button
                          type="button"
                          onClick={() => {
                            setEditingCookieIndex(index);
                            setIsAddingCookie(false);
                          }}
                          className="p-1 rounded cursor-pointer"
                        >
                          <IconPencil size={14} />
                        </button>
                        {siblingCategories.length > 0 && (
                          <Dropdown
                            toggle={(
                              <button
                                type="button"
                                className="p-1 rounded cursor-pointer"
                              >
                                <IconArrowBoxLeft size={14} />
                              </button>
                            )}
                          >
                            {siblingCategories.map(cat => (
                              <DropdownItem
                                className="text-sm"
                                key={cat.id}
                                onSelect={() => handleMoveCookie(index, cat.id)}
                              >
                                {cat.name}
                              </DropdownItem>
                            ))}
                          </Dropdown>
                        )}
                        <button
                          type="button"
                          onClick={() => handleDeleteCookie(index)}
                          className="p-1 rounded cursor-pointer text-danger-dark"
                        >
                          <IconTrashCan size={14} />
                        </button>
                      </div>
                    </Td>
                  </Tr>
                ),
          )}
          {isAddingCookie && (
            <AddCookieRow
              isUpdating={isUpdating}
              onSave={handleSaveNewCookie}
              onCancel={() => setIsAddingCookie(false)}
            />
          )}
        </Tbody>
      </table>

      {!isAddingCookie && (
        <div className="p-3 border-t border-border-low">
          <Button
            variant="secondary"
            onClick={() => {
              setIsAddingCookie(true);
              setEditingCookieIndex(null);
            }}
          >
            <IconPlusSmall size={14} />
            {__("Add Cookie")}
          </Button>
        </div>
      )}
    </Card>
  );
}
