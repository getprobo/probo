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
import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Dropdown,
  IconArrowBoxLeft,
  IconPencil,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Suspense, useCallback, useState } from "react";
import { graphql, useFragment, useMutation, useQueryLoader } from "react-relay";
import { useParams } from "react-router";
import { ConnectionHandler } from "relay-runtime";

import type { MoveToCategoryDropdownQuery } from "#/__generated__/core/MoveToCategoryDropdownQuery.graphql";
import type { TrackerPatternRowDeleteMutation } from "#/__generated__/core/TrackerPatternRowDeleteMutation.graphql";
import type { TrackerPatternRowFragment$key } from "#/__generated__/core/TrackerPatternRowFragment.graphql";
import type { TrackerPatternRowMoveMutation } from "#/__generated__/core/TrackerPatternRowMoveMutation.graphql";
import type { TrackerPatternRowUpdateMutation } from "#/__generated__/core/TrackerPatternRowUpdateMutation.graphql";

import {
  MoveToCategoryDropdown,
  moveToCategoryDropdownQuery,
} from "./MoveToCategoryDropdown";
import { TrackerPatternRowEdit } from "./TrackerPatternRowEdit";

const trackerPatternFragment = graphql`
  fragment TrackerPatternRowFragment on TrackerPattern {
    id
    trackerType
    displayName
    source
    description
    maxAgeSeconds
    excluded
    lastMatchedAt
  }
`;

const deletePatternMutation = graphql`
  mutation TrackerPatternRowDeleteMutation(
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
  mutation TrackerPatternRowMoveMutation(
    $input: MoveTrackerPatternToCategoryInput!
  ) {
    moveTrackerPatternToCategory(input: $input) {
      trackerPattern {
        id
        cookieCategory {
          id
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
  mutation TrackerPatternRowUpdateMutation(
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

type BadgeVariant = "success" | "warning" | "danger" | "info" | "neutral" | "outline" | "highlight";

function trackerTypeBadge(type: string, __: (s: string) => string): { label: string; variant: BadgeVariant } {
  switch (type) {
    case "COOKIE": return { label: __("Cookie"), variant: "warning" };
    case "LOCAL_STORAGE": return { label: __("localStorage"), variant: "info" };
    case "SESSION_STORAGE": return { label: __("sessionStorage"), variant: "highlight" };
    case "INDEXED_DB": return { label: __("IndexedDB"), variant: "success" };
    case "CACHE_STORAGE": return { label: __("Cache Storage"), variant: "outline" };
    default: return { label: type, variant: "neutral" };
  }
}

function sourceBadge(source: string, __: (s: string) => string): { label: string; variant: BadgeVariant } {
  switch (source) {
    case "SCRIPT": return { label: __("Script"), variant: "info" };
    case "PRE_EXISTING": return { label: __("Pre-existing"), variant: "outline" };
    default: return { label: source, variant: "neutral" };
  }
}

interface TrackerPatternRowProps {
  patternKey: TrackerPatternRowFragment$key;
  connectionId: string;
}

export function TrackerPatternRow({ patternKey, connectionId }: TrackerPatternRowProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const pattern = useFragment(trackerPatternFragment, patternKey);

  const [isEditing, setIsEditing] = useState(false);
  const [categoryQueryRef, loadCategoryQuery]
    = useQueryLoader<MoveToCategoryDropdownQuery>(moveToCategoryDropdownQuery);

  const handleCategoryDropdownOpen = useCallback(
    (open: boolean) => {
      if (open && cookieBannerId) {
        loadCategoryQuery({ cookieBannerId });
      }
    },
    [loadCategoryQuery, cookieBannerId],
  );

  const [deletePattern]
    = useMutation<TrackerPatternRowDeleteMutation>(deletePatternMutation);
  const [movePattern]
    = useMutation<TrackerPatternRowMoveMutation>(movePatternMutation);
  const [updatePattern, isUpdating]
    = useMutation<TrackerPatternRowUpdateMutation>(updatePatternMutation);

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deletePattern({
            variables: {
              input: { trackerPatternId: pattern.id },
              connections: [connectionId],
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({ title: __("Error"), description: errors[0].message, variant: "error" });
              } else {
                toast({ title: __("Success"), description: __("Cookie deleted"), variant: "success" });
              }
              resolve();
            },
            onError(error) {
              toast({ title: __("Error"), description: formatError(__("Failed to delete cookie"), error as GraphQLError), variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: __("Are you sure you want to delete \"%s\"?").replace("%s", pattern.displayName),
        variant: "danger",
        label: __("Delete"),
      },
    );
  };

  const handleMove = (targetCategoryId: string) => {
    movePattern({
      variables: {
        input: {
          trackerPatternId: pattern.id,
          targetCookieCategoryId: targetCategoryId,
        },
      },
      updater(store) {
        const payload = store.getRootField("moveTrackerPatternToCategory");
        if (!payload?.getLinkedRecord("trackerPattern")) {
          return;
        }

        const conn = store.get(connectionId);
        if (conn) {
          ConnectionHandler.deleteNode(conn, pattern.id);
        }
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: __("Success"), description: __("Cookie moved"), variant: "success" });
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to move cookie"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const handleToggleExcluded = () => {
    updatePattern({
      variables: {
        input: {
          trackerPatternId: pattern.id,
          excluded: !pattern.excluded,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
          return;
        }
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to update cookie"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const handleSaveEdit = (data: { description: string; maxAgeSeconds: number | null }) => {
    updatePattern({
      variables: {
        input: {
          trackerPatternId: pattern.id,
          description: data.description,
          maxAgeSeconds: data.maxAgeSeconds,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: __("Success"), description: __("Cookie updated"), variant: "success" });
        setIsEditing(false);
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to update cookie"), error as GraphQLError), variant: "error" });
      },
    });
  };

  if (isEditing) {
    return (
      <TrackerPatternRowEdit
        pattern={pattern.displayName}
        description={pattern.description}
        maxAgeSeconds={pattern.maxAgeSeconds ?? null}
        isUpdating={isUpdating}
        onSave={handleSaveEdit}
        onCancel={() => setIsEditing(false)}
      />
    );
  }

  const typeBadge = trackerTypeBadge(pattern.trackerType, __);
  const srcBadge = pattern.source ? sourceBadge(pattern.source, __) : null;

  return (
    <Tr className={pattern.excluded ? "bg-txt-quaternary opacity-80  line-through" : undefined}>
      <Td>
        <div className="flex flex-col min-w-0">
          <span className={pattern.excluded ? undefined : "font-medium"}>{pattern.displayName}</span>
          {pattern.description && (
            <span className="text-xs text-txt-tertiary wrap-break-word line-clamp-1">
              {pattern.description}
            </span>
          )}
        </div>
      </Td>
      <Td>
        <Badge variant={typeBadge.variant}>{typeBadge.label}</Badge>
      </Td>
      <Td>
        {srcBadge
          ? <Badge variant={srcBadge.variant}>{srcBadge.label}</Badge>
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td>
        {pattern.lastMatchedAt
          ? (
              <time dateTime={pattern.lastMatchedAt}>
                {new Date(pattern.lastMatchedAt).toLocaleString()}
              </time>
            )
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => setIsEditing(true)}
            className="p-1 rounded cursor-pointer"
            title={__("Edit")}
          >
            <IconPencil size={14} />
          </button>
          <Dropdown
            onOpenChange={handleCategoryDropdownOpen}
            toggle={(
              <button
                type="button"
                className="p-1 rounded cursor-pointer"
                title={__("Move to category")}
              >
                <IconArrowBoxLeft size={14} />
              </button>
            )}
          >
            {categoryQueryRef && (
              <Suspense>
                <MoveToCategoryDropdown queryRef={categoryQueryRef} onMove={handleMove} />
              </Suspense>
            )}
          </Dropdown>
          <button
            type="button"
            onClick={handleToggleExcluded}
            className="p-1 rounded cursor-pointer"
            title={pattern.excluded ? __("Include") : __("Exclude")}
          >
            {pattern.excluded ? <EyeIcon size={14} /> : <EyeSlashIcon size={14} />}
          </button>
          <button
            type="button"
            onClick={handleDelete}
            className="p-1 rounded cursor-pointer text-danger-dark"
            title={__("Delete")}
          >
            <IconTrashCan size={14} />
          </button>
        </div>
      </Td>
    </Tr>
  );
}
