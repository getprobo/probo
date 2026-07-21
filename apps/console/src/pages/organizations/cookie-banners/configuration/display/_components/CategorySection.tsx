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

import { EyeIcon, EyeSlashIcon } from "@phosphor-icons/react";
import { formatError, type GraphQLError } from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
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
    trackerPatterns(first: 100, orderBy: { field: CREATED_AT, direction: ASC })
      @connection(key: "CategorySection_trackerPatterns", filters: [])
      @required(action: THROW) {
      __id
      edges {
        node {
          id
          displayName
          trackerType
          maxAgeSeconds
          description
          excluded
          ...EditCookieRowFragment
        }
      }
    }
    cookieBanner @required(action: THROW) {
      categories(first: 50, orderBy: { field: RANK, direction: ASC }, filter: { excludeKind: UNCATEGORISED }) @required(action: THROW) {
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
    $input: CreateTrackerPatternInput!
    $connections: [ID!]!
  ) {
    createTrackerPattern(input: $input) {
      trackerPatternEdge @appendEdge(connections: $connections) {
        node {
          id
          displayName
          trackerType
          maxAgeSeconds
          description
          excluded
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
    $input: UpdateTrackerPatternInput!
  ) {
    updateTrackerPattern(input: $input) {
      trackerPattern {
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
    $input: DeleteTrackerPatternInput!
    $connections: [ID!]!
  ) {
    deleteTrackerPattern(input: $input) {
      deletedTrackerPatternId @deleteEdge(connections: $connections)
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
    $input: MoveTrackerPatternToCategoryInput!
  ) {
    moveTrackerPatternToCategory(input: $input) {
      trackerPattern {
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
        categories(first: 50, orderBy: { field: RANK, direction: ASC }, filter: { excludeKind: UNCATEGORISED }) {
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
  const { t, i18n } = useTranslation("organizations/cookie-banners");
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

  const patternsConnectionId = category.trackerPatterns.__id;
  const patterns = category.trackerPatterns.edges.map(e => e.node);
  const isMutating = isUpdating || isCreating || isUpdatingPattern;
  const formatDuration = (seconds: number | null, trackerType: string) => {
    if (seconds === null || seconds <= 0) {
      return ["LOCAL_STORAGE", "INDEXED_DB", "CACHE_STORAGE"].includes(trackerType)
        ? t("categorySection.duration.persistent")
        : t("categorySection.duration.session");
    }
    const units = [["year", 31536000, 21 * 86400], ["month", 2592000, 2 * 86400], ["week", 604800, 43200], ["day", 86400, 7200], ["hour", 3600, 300], ["minute", 60, 5], ["second", 1, 0]] as const;
    let remaining = seconds;
    const parts: string[] = [];
    for (const [unit, unitSeconds, snap] of units) {
      if (remaining < unitSeconds - snap) continue;
      let count = Math.floor(remaining / unitSeconds);
      const leftover = remaining - count * unitSeconds;
      if (leftover >= unitSeconds - snap) count++;
      remaining = leftover <= snap || leftover >= unitSeconds - snap ? 0 : leftover;
      parts.push(t(`categorySection.duration.${unit}`, { count }));
    }
    return new Intl.ListFormat(i18n.language, { style: "long", type: "unit" }).format(parts);
  };
  const trackerTypeBadge = (type: string) => {
    const badges = {
      COOKIE: { variant: "warning" as const, label: t("categorySection.trackerTypes.cookie") },
      LOCAL_STORAGE: { variant: "info" as const, label: t("categorySection.trackerTypes.localStorage") },
      SESSION_STORAGE: { variant: "highlight" as const, label: t("categorySection.trackerTypes.sessionStorage") },
      INDEXED_DB: { variant: "success" as const, label: t("categorySection.trackerTypes.indexedDb") },
      CACHE_STORAGE: { variant: "outline" as const, label: t("categorySection.trackerTypes.cacheStorage") },
    };
    return badges[type as keyof typeof badges] ?? { variant: "neutral" as const, label: type };
  };

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
            title: t("categorySection.errors.title"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }
        toast({
          title: t("categorySection.messages.successTitle"),
          description: t("categorySection.messages.updated"),
          variant: "success",
        });
        setIsEditingCategory(false);
      },
      onError(error) {
        toast({
          title: t("categorySection.errors.title"),
          description: formatError(
            t("categorySection.errors.updateCategory"),
            error,
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
            title: t("categorySection.errors.title"),
            description: isConflict
              ? t("categorySection.errors.duplicateCookie")
              : errors[0].message,
            variant: "error",
          });
          return;
        }
        toast({
          title: t("categorySection.messages.successTitle"),
          description: t("categorySection.messages.cookieAdded"),
          variant: "success",
        });
        setIsAddingCookie(false);
      },
      onError(error) {
        toast({
          title: t("categorySection.errors.title"),
          description: formatError(
            t("categorySection.errors.addCookie"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  const handleSaveEditCookie = (patternId: string, cookie: CookieEntry) => {
    updatePattern({
      variables: {
        input: {
          trackerPatternId: patternId,
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
            title: t("categorySection.errors.title"),
            description: isConflict
              ? t("categorySection.errors.duplicateCookie")
              : errors[0].message,
            variant: "error",
          });
          return;
        }
        toast({
          title: t("categorySection.messages.successTitle"),
          description: t("categorySection.messages.cookieUpdated"),
          variant: "success",
        });
        setEditingCookieId(null);
      },
      onError(error) {
        toast({
          title: t("categorySection.errors.title"),
          description: formatError(
            t("categorySection.errors.updateCookie"),
            error,
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
          trackerPatternId: patternId,
          excluded,
        },
      },
      onCompleted(_response, errors) {
        if (errors?.length) {
          toast({
            title: t("categorySection.errors.title"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }
      },
      onError(error) {
        toast({
          title: t("categorySection.errors.title"),
          description: formatError(
            t("categorySection.errors.updateCookie"),
            error,
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
              input: { trackerPatternId: patternId },
              connections: [patternsConnectionId],
            },
            onCompleted(_response, errors) {
              if (errors?.length) {
                toast({
                  title: t("categorySection.errors.title"),
                  description: errors[0].message,
                  variant: "error",
                });
              } else {
                toast({
                  title: t("categorySection.messages.successTitle"),
                  description: t("categorySection.messages.cookieDeleted"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: t("categorySection.errors.title"),
                description: formatError(
                  t("categorySection.errors.deleteCookie"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("categorySection.deleteCookieConfirmation", { name: patternName }),
        variant: "danger",
        label: t("categorySection.actions.delete"),
      },
    );
  };

  const allCategories = category.cookieBanner.categories.edges.map(e => e.node) ?? [];
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
                toast({ title: t("categorySection.errors.title"), description: errors[0].message, variant: "error" });
              } else {
                toast({ title: t("categorySection.messages.successTitle"), description: t("categorySection.messages.categoryDeleted"), variant: "success" });
              }
              resolve();
            },
            onError(error) {
              toast({ title: t("categorySection.errors.title"), description: formatError(t("categorySection.errors.deleteCategory"), error), variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: t("categorySection.deleteCategoryConfirmation", { name: category.name }),
        variant: "danger",
        label: t("categorySection.actions.delete"),
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
          toast({ title: t("categorySection.errors.title"), description: errors[0].message, variant: "error" });
        }
      },
      onError(error) {
        toast({ title: t("categorySection.errors.title"), description: formatError(t("categorySection.errors.reorder"), error), variant: "error" });
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
          toast({ title: t("categorySection.errors.title"), description: errors[0].message, variant: "error" });
        }
      },
      onError(error) {
        toast({ title: t("categorySection.errors.title"), description: formatError(t("categorySection.errors.reorder"), error), variant: "error" });
      },
    });
  };

  const handleMoveCookie = (patternId: string, targetCategoryId: string) => {
    movePattern({
      variables: {
        input: {
          trackerPatternId: patternId,
          targetCookieCategoryId: targetCategoryId,
        },
      },
      updater(store) {
        const sourceCategory = store.get(category.id);
        if (sourceCategory) {
          const sourceConn = ConnectionHandler.getConnection(
            sourceCategory,
            "CategorySection_trackerPatterns",
          );
          if (sourceConn) {
            ConnectionHandler.deleteNode(sourceConn, patternId);
          }
        }

        const targetCategory = store.get(targetCategoryId);
        if (targetCategory) {
          const targetConn = ConnectionHandler.getConnection(
            targetCategory,
            "CategorySection_trackerPatterns",
          );
          if (targetConn) {
            const patternRecord = store.get(patternId);
            if (patternRecord) {
              const newEdge = ConnectionHandler.createEdge(
                store,
                targetConn,
                patternRecord,
                "TrackerPatternEdge",
              );
              ConnectionHandler.insertEdgeAfter(targetConn, newEdge);
            }
          }
        }
      },
      onCompleted(_response, errors) {
        if (errors?.length) {
          toast({
            title: t("categorySection.errors.title"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }
        toast({
          title: t("categorySection.messages.successTitle"),
          description: t("categorySection.messages.cookieMoved"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: t("categorySection.errors.title"),
          description: formatError(
            t("categorySection.errors.moveCookie"),
            error,
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
                    <Badge variant="neutral">{t("categorySection.required")}</Badge>
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
                    {t("categorySection.actions.edit")}
                  </Button>
                  {canDelete && (
                    <Button variant="danger" onClick={handleDeleteCategory}>
                      <IconTrashCan size={14} />
                      {t("categorySection.actions.delete")}
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
              {t("categorySection.blockHint")}
              <code className="rounded bg-muted px-1 py-0.5 font-mono text-[11px]">
                data-cookie-consent=&quot;
                {category.slug}
                &quot;
              </code>
            </p>
            {category.gcmConsentTypes.length > 0 && (
              <div className="mt-2 flex items-center gap-1.5">
                <span className="text-xs text-txt-secondary/70">
                  {t("categorySection.googleConsentMode")}
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
                  {t("categorySection.posthog")}
                </span>
                <Badge variant="neutral">
                  {t("categorySection.trackingConsent")}
                </Badge>
              </div>
            )}
          </>
        )}
      </div>

      <table className="w-full text-left">
        <Thead>
          <Tr>
            <Th>{t("categorySection.columns.name")}</Th>
            <Th>{t("categorySection.columns.type")}</Th>
            <Th>{t("categorySection.columns.duration")}</Th>
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
                      <div className="flex flex-col min-w-0 max-w-xs">
                        <code className="text-sm font-mono">{pattern.displayName}</code>
                        {pattern.description && (
                          <span className="text-xs text-txt-tertiary wrap-break-word line-clamp-1">
                            {pattern.description}
                          </span>
                        )}
                      </div>
                    </Td>
                    <Td>
                      {(() => {
                        const typeBadge = trackerTypeBadge(pattern.trackerType);
                        return <Badge variant={typeBadge.variant}>{typeBadge.label}</Badge>;
                      })()}
                    </Td>
                    <Td className="text-sm text-muted-foreground">
                      {formatDuration(pattern.maxAgeSeconds ?? null, pattern.trackerType)}
                    </Td>
                    <Td>
                      <div className="flex items-center gap-1">
                        <button
                          type="button"
                          onClick={() => handleToggleExcluded(pattern.id, !pattern.excluded)}
                          className="p-1 rounded cursor-pointer"
                          title={pattern.excluded ? t("categorySection.actions.include") : t("categorySection.actions.exclude")}
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
            {t("categorySection.actions.addCookie")}
          </Button>
        </div>
      )}
    </Card>
  );
}
