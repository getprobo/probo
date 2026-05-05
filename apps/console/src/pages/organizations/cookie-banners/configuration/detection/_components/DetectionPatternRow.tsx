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

import { Eye as IconEye, EyeSlash as IconEyeSlash } from "@phosphor-icons/react";
import { formatDate, formatError, type GraphQLError } from "@probo/helpers";
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

import type { DetectionPatternRowDeleteMutation } from "#/__generated__/core/DetectionPatternRowDeleteMutation.graphql";
import type { DetectionPatternRowFragment$key } from "#/__generated__/core/DetectionPatternRowFragment.graphql";
import type { DetectionPatternRowMoveMutation } from "#/__generated__/core/DetectionPatternRowMoveMutation.graphql";
import type { DetectionPatternRowUpdateMutation } from "#/__generated__/core/DetectionPatternRowUpdateMutation.graphql";
import type { MoveToCategoryDropdownQuery } from "#/__generated__/core/MoveToCategoryDropdownQuery.graphql";

import { DetectionPatternRowEdit } from "./DetectionPatternRowEdit";
import {
  MoveToCategoryDropdown,
  moveToCategoryDropdownQuery,
} from "./MoveToCategoryDropdown";

const detectionPatternFragment = graphql`
  fragment DetectionPatternRowFragment on CookiePattern {
    id
    displayName
    source
    description
    maxAgeSeconds
    excluded
    lastMatchedAt
    updatedAt
  }
`;

const deletePatternMutation = graphql`
  mutation DetectionPatternRowDeleteMutation(
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
  mutation DetectionPatternRowMoveMutation(
    $input: MoveCookiePatternToCategoryInput!
  ) {
    moveCookiePatternToCategory(input: $input) {
      cookiePattern {
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
  mutation DetectionPatternRowUpdateMutation(
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

interface DetectionPatternRowProps {
  patternKey: DetectionPatternRowFragment$key;
  connectionId: string;
}

export function DetectionPatternRow({ patternKey, connectionId }: DetectionPatternRowProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const pattern = useFragment(detectionPatternFragment, patternKey);

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
    = useMutation<DetectionPatternRowDeleteMutation>(deletePatternMutation);
  const [movePattern]
    = useMutation<DetectionPatternRowMoveMutation>(movePatternMutation);
  const [updatePattern, isUpdating]
    = useMutation<DetectionPatternRowUpdateMutation>(updatePatternMutation);

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deletePattern({
            variables: {
              input: { cookiePatternId: pattern.id },
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
          cookiePatternId: pattern.id,
          targetCookieCategoryId: targetCategoryId,
        },
      },
      updater(store) {
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
          cookiePatternId: pattern.id,
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

  const handleSaveEdit = (data: { displayName: string; description: string; maxAgeSeconds: number | null }) => {
    updatePattern({
      variables: {
        input: {
          cookiePatternId: pattern.id,
          displayName: data.displayName,
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
      <DetectionPatternRowEdit
        displayName={pattern.displayName}
        description={pattern.description}
        maxAgeSeconds={pattern.maxAgeSeconds ?? null}
        isUpdating={isUpdating}
        onSave={handleSaveEdit}
        onCancel={() => setIsEditing(false)}
      />
    );
  }

  return (
    <Tr className={pattern.excluded ? "opacity-80" : undefined}>
      <Td>
        <div className="flex flex-col min-w-0">
          <span className="font-medium">{pattern.displayName}</span>
          {pattern.description && (
            <span className="text-xs text-txt-tertiary wrap-break-word line-clamp-1">
              {pattern.description}
            </span>
          )}
        </div>
      </Td>
      <Td>
        <Badge variant={pattern.source === "SCRIPT" ? "info" : "neutral"}>
          {pattern.source === "SCRIPT" ? __("Script") : __("Pre-existing")}
        </Badge>
      </Td>
      <Td>
        {pattern.lastMatchedAt
          ? (
              <time dateTime={pattern.lastMatchedAt}>
                {formatDate(pattern.lastMatchedAt)}
              </time>
            )
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td>
        <time dateTime={pattern.updatedAt}>
          {formatDate(pattern.updatedAt)}
        </time>
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
            {pattern.excluded ? <IconEye size={14} /> : <IconEyeSlash size={14} />}
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
