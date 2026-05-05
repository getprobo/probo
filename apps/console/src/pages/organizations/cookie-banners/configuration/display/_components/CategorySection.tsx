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

import { EyeIcon, EyeSlashIcon } from "@phosphor-icons/react";
import { formatError, type GraphQLError, humanizeSeconds } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  Dropdown,
  DropdownItem,
  IconArrowBoxLeft,
  IconArrowDown,
  IconArrowUp,
  IconPencil,
  IconPlusSmall,
  IconTrashCan,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { CategorySectionCreatePatternMutation } from "#/__generated__/core/CategorySectionCreatePatternMutation.graphql";
import type { CategorySectionDeleteCategoryMutation } from "#/__generated__/core/CategorySectionDeleteCategoryMutation.graphql";
import type { CategorySectionDeletePatternMutation } from "#/__generated__/core/CategorySectionDeletePatternMutation.graphql";
import type { CategorySectionFragment$key } from "#/__generated__/core/CategorySectionFragment.graphql";
import type { CategorySectionMovePatternMutation } from "#/__generated__/core/CategorySectionMovePatternMutation.graphql";
import type { CategorySectionReorderMutation } from "#/__generated__/core/CategorySectionReorderMutation.graphql";
import type { CategorySectionUpdateMutation } from "#/__generated__/core/CategorySectionUpdateMutation.graphql";
import type { CategorySectionUpdatePatternMutation } from "#/__generated__/core/CategorySectionUpdatePatternMutation.graphql";

import { AddCookieRow } from "./AddCookieRow";
import { EditCategoryForm } from "./EditCategoryForm";
import { EditCookieRow } from "./EditCookieRow";

export interface CookieEntry {
  name: string;
  maxAgeSeconds: number | null;
  description: string;
  excluded: boolean;
}

export const categorySectionFragment = graphql`
  fragment CategorySectionFragment on CookieCategory {
    id
    name
    slug
    description
    kind
    gcmConsentTypes
    posthogConsent
    cookiePatterns(first: 100, orderBy: { field: CREATED_AT, direction: ASC })
      @connection(key: "CategorySection_cookiePatterns", filters: [])
      @required(action: THROW) {
      __id
      edges {
        node {
          id
          displayName
          maxAgeSeconds
          description
          excluded
          source
          ...EditCookieRowFragment
        }
      }
    }
    cookieBanner @required(action: THROW) {
      consentCategories(first: 50, orderBy: { field: RANK, direction: ASC }) @required(action: THROW) {
        edges {
          node {
            id
            name
            rank
            kind
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
        slug
        description
        rank
        gcmConsentTypes
        posthogConsent
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

const createPatternMutation = graphql`
  mutation CategorySectionCreatePatternMutation(
    $input: CreateCookiePatternInput!
    $connections: [ID!]!
  ) {
    createCookiePattern(input: $input) {
      cookiePatternEdge @appendEdge(connections: $connections) {
        node {
          id
          displayName
          maxAgeSeconds
          description
          excluded
          source
          ...EditCookieRowFragment
        }
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

const updatePatternMutation = graphql`
  mutation CategorySectionUpdatePatternMutation(
    $input: UpdateCookiePatternInput!
  ) {
    updateCookiePattern(input: $input) {
      cookiePattern {
        id
        displayName
        maxAgeSeconds
        description
        excluded
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

const deletePatternMutation = graphql`
  mutation CategorySectionDeletePatternMutation(
    $input: DeleteCookiePatternInput!
    $connections: [ID!]!
  ) {
    deleteCookiePattern(input: $input) {
      deletedCookiePatternId @deleteEdge(connections: $connections)
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

const movePatternMutation = graphql`
  mutation CategorySectionMovePatternMutation(
    $input: MoveCookiePatternToCategoryInput!
  ) {
    moveCookiePatternToCategory(input: $input) {
      cookiePattern {
        id
        displayName
        maxAgeSeconds
        description
        cookieCategory {
          id
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

const deleteCategoryMutation = graphql`
  mutation CategorySectionDeleteCategoryMutation(
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

const reorderCategoryMutation = graphql`
  mutation CategorySectionReorderMutation(
    $input: ReorderCookieCategoryInput!
  ) {
    reorderCookieCategory(input: $input) {
      cookieBanner {
        id
        consentCategories(first: 50, orderBy: { field: RANK, direction: ASC }) {
          edges {
            node {
              id
              rank
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

interface CategorySectionProps {
  categoryKey: CategorySectionFragment$key;
  connectionId: string;
}

export function CategorySection({ categoryKey, connectionId }: CategorySectionProps) {
  const category = useFragment(categorySectionFragment, categoryKey);
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();

  const [updateCategory, isUpdating]
    = useMutation<CategorySectionUpdateMutation>(updateCategoryMutation);
  const [createPattern, isCreating]
    = useMutation<CategorySectionCreatePatternMutation>(createPatternMutation);
  const [updatePattern, isUpdatingPattern]
    = useMutation<CategorySectionUpdatePatternMutation>(updatePatternMutation);
  const [deletePattern]
    = useMutation<CategorySectionDeletePatternMutation>(deletePatternMutation);
  const [movePattern]
    = useMutation<CategorySectionMovePatternMutation>(movePatternMutation);
  const [deleteCategory]
    = useMutation<CategorySectionDeleteCategoryMutation>(deleteCategoryMutation);
  const [reorderCategory]
    = useMutation<CategorySectionReorderMutation>(reorderCategoryMutation);

  const [isEditingCategory, setIsEditingCategory] = useState(false);
  const [editingCookieId, setEditingCookieId] = useState<string | null>(null);
  const [isAddingCookie, setIsAddingCookie] = useState(false);

  const patternsConnectionId = category.cookiePatterns.__id;
  const patterns = category.cookiePatterns.edges.map(e => e.node);
  const isMutating = isUpdating || isCreating || isUpdatingPattern;

  const handleSaveCategory = (
    name: string, slug: string, description: string,
    gcmConsentTypes: string[], posthogConsent: boolean,
  ) => {
    updateCategory({
      variables: {
        input: {
          cookieCategoryId: category.id,
          name,
          slug,
          description,
          gcmConsentTypes,
          posthogConsent,
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
        setIsEditingCategory(false);
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

  const handleSaveNewCookie = (cookie: CookieEntry) => {
    if (!cookie.name.trim()) return;
    createPattern({
      variables: {
        input: {
          cookieCategoryId: category.id,
          pattern: cookie.name,
          matchType: "EXACT",
          displayName: cookie.name,
          maxAgeSeconds: cookie.maxAgeSeconds,
          description: cookie.description,
        },
        connections: [patternsConnectionId],
      },
      onCompleted(_response, errors) {
        if (errors?.length) {
          const isConflict = errors.some(
            e => (e as unknown as GraphQLError).extensions?.code === "CONFLICT",
          );
          toast({
            title: __("Error"),
            description: isConflict
              ? __("A cookie with this name already exists in this banner")
              : errors[0].message,
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Cookie added"),
          variant: "success",
        });
        setIsAddingCookie(false);
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to add cookie"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleSaveEditCookie = (patternId: string, cookie: CookieEntry) => {
    if (!cookie.name.trim()) return;
    updatePattern({
      variables: {
        input: {
          cookiePatternId: patternId,
          displayName: cookie.name,
          maxAgeSeconds: cookie.maxAgeSeconds,
          description: cookie.description,
          excluded: cookie.excluded,
        },
      },
      onCompleted(_response, errors) {
        if (errors?.length) {
          const isConflict = errors.some(
            e => (e as unknown as GraphQLError).extensions?.code === "CONFLICT",
          );
          toast({
            title: __("Error"),
            description: isConflict
              ? __("A cookie with this name already exists in this banner")
              : errors[0].message,
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Cookie updated"),
          variant: "success",
        });
        setEditingCookieId(null);
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to update cookie"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleToggleExcluded = (patternId: string, excluded: boolean) => {
    updatePattern({
      variables: {
        input: {
          cookiePatternId: patternId,
          excluded,
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
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to update cookie"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleDeleteCookie = (patternId: string, patternName: string) => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deletePattern({
            variables: {
              input: { cookiePatternId: patternId },
              connections: [patternsConnectionId],
            },
            onCompleted(_response, errors) {
              if (errors?.length) {
                toast({
                  title: __("Error"),
                  description: errors[0].message,
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Cookie deleted"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete cookie"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: __("Are you sure you want to delete \"%s\"?").replace("%s", patternName),
        variant: "danger",
        label: __("Delete"),
      },
    );
  };

  const allCategories = category.cookieBanner.consentCategories.edges.map(e => e.node) ?? [];
  const siblingCategories = allCategories.filter(c => c.id !== category.id);
  const selfIndex = allCategories.findIndex(c => c.id === category.id);
  const isFirst = selfIndex === 0;
  const isLast = selfIndex === allCategories.length - 1;
  const canDelete = category.kind === "NORMAL";

  const handleDeleteCategory = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteCategory({
            variables: {
              input: { cookieCategoryId: category.id },
              connections: [connectionId],
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({ title: __("Error"), description: errors[0].message, variant: "error" });
              } else {
                toast({ title: __("Success"), description: __("Category deleted"), variant: "success" });
              }
              resolve();
            },
            onError(error) {
              toast({ title: __("Error"), description: formatError(__("Failed to delete category"), error as GraphQLError), variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: __("Are you sure you want to delete the category \"%s\"? Any cookies in this category will be moved to Uncategorised.").replace("%s", category.name),
        variant: "danger",
        label: __("Delete"),
      },
    );
  };

  const handleMoveUp = () => {
    if (isFirst) return;
    const above = allCategories[selfIndex - 1];
    reorderCategory({
      variables: { input: { cookieCategoryId: category.id, rank: above.rank } },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
        }
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to reorder"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const handleMoveDown = () => {
    if (isLast) return;
    const below = allCategories[selfIndex + 1];
    reorderCategory({
      variables: { input: { cookieCategoryId: category.id, rank: below.rank } },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
        }
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to reorder"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const handleMoveCookie = (patternId: string, targetCategoryId: string) => {
    movePattern({
      variables: {
        input: {
          cookiePatternId: patternId,
          targetCookieCategoryId: targetCategoryId,
        },
      },
      updater(store) {
        const sourceCategory = store.get(category.id);
        if (sourceCategory) {
          const sourceConn = ConnectionHandler.getConnection(
            sourceCategory,
            "CategorySection_cookiePatterns",
          );
          if (sourceConn) {
            ConnectionHandler.deleteNode(sourceConn, patternId);
          }
        }

        const targetCategory = store.get(targetCategoryId);
        if (targetCategory) {
          const targetConn = ConnectionHandler.getConnection(
            targetCategory,
            "CategorySection_cookiePatterns",
          );
          if (targetConn) {
            const patternRecord = store.get(patternId);
            if (patternRecord) {
              const newEdge = ConnectionHandler.createEdge(
                store,
                targetConn,
                patternRecord,
                "CookiePatternEdge",
              );
              ConnectionHandler.insertEdgeAfter(targetConn, newEdge);
            }
          }
        }
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
                slug={category.slug}
                description={category.description}
                kind={category.kind}
                gcmConsentTypes={[...category.gcmConsentTypes]}
                posthogConsent={category.posthogConsent}
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
                  <div className="flex items-center gap-1">
                    <button
                      type="button"
                      onClick={handleMoveUp}
                      disabled={isFirst}
                      className="p-0.5 rounded cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
                    >
                      <IconArrowUp size={14} />
                    </button>
                    <button
                      type="button"
                      onClick={handleMoveDown}
                      disabled={isLast}
                      className="p-0.5 rounded cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
                    >
                      <IconArrowDown size={14} />
                    </button>
                  </div>
                  <Button
                    variant="secondary"
                    onClick={() => setIsEditingCategory(true)}
                  >
                    <IconPencil size={14} />
                    {__("Edit")}
                  </Button>
                  {canDelete && (
                    <Button variant="danger" onClick={handleDeleteCategory}>
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
                {category.slug}
                &quot;
              </code>
            </p>
            {category.gcmConsentTypes.length > 0 && (
              <div className="mt-2 flex items-center gap-1.5">
                <span className="text-xs text-txt-secondary/70">
                  {__("Google Consent Mode:")}
                </span>
                {category.gcmConsentTypes.map(type => (
                  <Badge key={type} variant="neutral">
                    {type}
                  </Badge>
                ))}
              </div>
            )}
            {category.posthogConsent && (
              <div className="mt-2 flex items-center gap-1.5">
                <span className="text-xs text-txt-secondary/70">
                  {__("PostHog:")}
                </span>
                <Badge variant="neutral">
                  {__("Tracking consent")}
                </Badge>
              </div>
            )}
          </>
        )}
      </div>

      <table className="w-full text-left">
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Source")}</Th>
            <Th>{__("Duration")}</Th>
            <Th>{__("Description")}</Th>
            <Th className="w-20" />
          </Tr>
        </Thead>
        <Tbody>
          {patterns.map(pattern =>
            editingCookieId === pattern.id
              ? (
                  <EditCookieRow
                    key={pattern.id}
                    cookieKey={pattern}
                    isUpdating={isMutating}
                    onSave={updated => handleSaveEditCookie(pattern.id, updated)}
                    onCancel={() => setEditingCookieId(null)}
                  />
                )
              : (
                  <Tr key={pattern.id} className={pattern.excluded ? "opacity-80" : undefined}>
                    <Td>
                      <code className="text-sm font-mono">{pattern.displayName}</code>
                    </Td>
                    <Td>
                      <Badge
                        variant={pattern.source === "SCRIPT" ? "info" : "neutral"}
                        title={
                          pattern.source === "SCRIPT"
                            ? __("Set by a script at runtime")
                            : __("Already present when the page was loaded")
                        }
                      >
                        {pattern.source === "SCRIPT" ? __("Script") : __("Pre-existing")}
                      </Badge>
                    </Td>
                    <Td className="text-sm text-muted-foreground">
                      {humanizeSeconds(pattern.maxAgeSeconds ?? null)}
                    </Td>
                    <Td className="text-sm text-muted-foreground">
                      {pattern.description}
                    </Td>
                    <Td>
                      <div className="flex items-center gap-1">
                        <button
                          type="button"
                          onClick={() => handleToggleExcluded(pattern.id, !pattern.excluded)}
                          className="p-1 rounded cursor-pointer"
                          title={pattern.excluded ? __("Include") : __("Exclude")}
                        >
                          {pattern.excluded ? <EyeIcon size={14} /> : <EyeSlashIcon size={14} />}
                        </button>
                        <button
                          type="button"
                          onClick={() => {
                            setEditingCookieId(pattern.id);
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
                                onSelect={() => handleMoveCookie(pattern.id, cat.id)}
                              >
                                {cat.name}
                              </DropdownItem>
                            ))}
                          </Dropdown>
                        )}
                        <button
                          type="button"
                          onClick={() => handleDeleteCookie(pattern.id, pattern.displayName)}
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
              isUpdating={isMutating}
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
              setEditingCookieId(null);
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
