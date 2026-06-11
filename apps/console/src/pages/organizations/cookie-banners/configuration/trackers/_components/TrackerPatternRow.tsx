// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { DownloadSimpleIcon, EyeIcon, EyeSlashIcon } from "@phosphor-icons/react";
import { formatError, getTrackerSourceBadge, getTrackerTypeBadge, type GraphQLError, humanizeSeconds } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { graphql, useFragment, useMutation } from "react-relay";
import { ConnectionHandler } from "relay-runtime";

import type { TrackerPatternRowDeleteMutation } from "#/__generated__/core/TrackerPatternRowDeleteMutation.graphql";
import type { TrackerPatternRowFragment$key } from "#/__generated__/core/TrackerPatternRowFragment.graphql";
import type { TrackerPatternRowImportMutation } from "#/__generated__/core/TrackerPatternRowImportMutation.graphql";
import type { TrackerPatternRowMoveMutation } from "#/__generated__/core/TrackerPatternRowMoveMutation.graphql";
import type { TrackerPatternRowUpdateMutation } from "#/__generated__/core/TrackerPatternRowUpdateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { MoveToCategorySelect } from "./MoveToCategorySelect";
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
    cookieCategory {
      id
      name
    }
    thirdParty {
      id
      name
    }
    commonThirdParty {
      id
      name
    }
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

const importThirdPartyMutation = graphql`
  mutation TrackerPatternRowImportMutation(
    $input: ImportThirdPartyFromCommonInput!
  ) {
    importThirdPartyFromCommon(input: $input) {
      created
      thirdPartyEdge {
        node {
          id
          name
        }
      }
    }
  }
`;

interface TrackerPatternRowProps {
  patternKey: TrackerPatternRowFragment$key;
  connectionId: string;
}

export function TrackerPatternRow({ patternKey, connectionId }: TrackerPatternRowProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const organizationId = useOrganizationId();
  const pattern = useFragment(trackerPatternFragment, patternKey);

  const [isEditing, setIsEditing] = useState(false);

  const [deletePattern]
    = useMutation<TrackerPatternRowDeleteMutation>(deletePatternMutation);
  const [movePattern]
    = useMutation<TrackerPatternRowMoveMutation>(movePatternMutation);
  const [updatePattern, isUpdating]
    = useMutation<TrackerPatternRowUpdateMutation>(updatePatternMutation);
  const [importThirdParty, isImporting]
    = useMutation<TrackerPatternRowImportMutation>(importThirdPartyMutation);

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
    if (targetCategoryId === pattern.cookieCategory?.id) {
      return;
    }
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

  const handleImport = () => {
    const commonThirdParty = pattern.commonThirdParty;
    if (!commonThirdParty || isImporting) {
      return;
    }

    importThirdParty({
      variables: {
        input: {
          organizationId,
          commonThirdPartyId: commonThirdParty.id,
        },
      },
      updater(store) {
        const node = store
          .getRootField("importThirdPartyFromCommon")
          ?.getLinkedRecord("thirdPartyEdge")
          ?.getLinkedRecord("node");
        if (!node) {
          return;
        }

        // Reflect the import on the clicked pattern immediately. Sibling
        // patterns of the same vendor are backfilled server-side and pick
        // up the link on the next fetch of the list.
        const patternRecord = store.get(pattern.id);
        if (patternRecord) {
          patternRecord.setLinkedRecord(node, "thirdParty");
        }
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: __("Success"), description: __("Third party imported"), variant: "success" });
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to import third party"), error as GraphQLError), variant: "error" });
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

  const typeBadge = getTrackerTypeBadge(pattern.trackerType, __);
  const srcBadge = pattern.source ? getTrackerSourceBadge(pattern.source, __) : null;

  return (
    <Tr
      to={pattern.id}
      className={
        pattern.excluded
          ? "bg-txt-quaternary/70 line-through"
          : pattern.source === "SCRIPT"
            ? undefined
            : "bg-txt-quaternary/25"
      }
    >
      <Td>
        <div className="flex flex-col items-start min-w-0 max-w-xs gap-1">
          <Badge variant={typeBadge.variant}>{typeBadge.label}</Badge>
          <span className={pattern.excluded ? undefined : "font-medium"}>{pattern.displayName}</span>
          {pattern.description && (
            <span className="text-xs text-txt-tertiary wrap-break-word line-clamp-1">
              {pattern.description}
            </span>
          )}
        </div>
      </Td>
      <Td>
        {pattern.thirdParty
          ? (
              <span className="truncate">{pattern.thirdParty.name}</span>
            )
          : pattern.commonThirdParty
            ? (
                <div>
                  <Badge variant="info">{__("Common catalog")}</Badge>
                  <span className="truncate">{pattern.commonThirdParty.name}</span>
                </div>
              )
            : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td>
        {srcBadge
          ? <Badge variant={srcBadge.variant}>{srcBadge.label}</Badge>
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td noLink>
        <div className="pr-2 flex justify-start">
          <MoveToCategorySelect
            currentCategoryId={pattern.cookieCategory?.id}
            currentCategoryName={pattern.cookieCategory?.name}
            highlight={!!pattern.cookieCategory && pattern.cookieCategory.kind !== "UNCATEGORISED"}
            onSelect={handleMove}
          />
        </div>
      </Td>
      <Td>
        <span className="pl-2">{humanizeSeconds(pattern.maxAgeSeconds ?? null)}</span>
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
      <Td noLink className="w-px whitespace-nowrap">
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => setIsEditing(true)}
            className="p-1 rounded cursor-pointer"
            title={__("Edit")}
          >
            <IconPencil size={14} />
          </button>
          <ActionDropdown>
            {!pattern.thirdParty && pattern.commonThirdParty && (
              <DropdownItem
                icon={DownloadSimpleIcon}
                onSelect={handleImport}
              >
                {__("Import to third parties")}
              </DropdownItem>
            )}
            <DropdownItem
              icon={pattern.excluded ? EyeIcon : EyeSlashIcon}
              onSelect={handleToggleExcluded}
            >
              {pattern.excluded ? __("Include") : __("Exclude")}
            </DropdownItem>
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onSelect={handleDelete}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </div>
      </Td>
    </Tr>
  );
}
