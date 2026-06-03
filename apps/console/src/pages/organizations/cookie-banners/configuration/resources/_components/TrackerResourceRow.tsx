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

import type { TrackerResourceRowDeleteMutation } from "#/__generated__/core/TrackerResourceRowDeleteMutation.graphql";
import type { TrackerResourceRowFragment$key } from "#/__generated__/core/TrackerResourceRowFragment.graphql";
import type { TrackerResourceRowMoveMutation } from "#/__generated__/core/TrackerResourceRowMoveMutation.graphql";
import type { TrackerResourceRowUpdateMutation } from "#/__generated__/core/TrackerResourceRowUpdateMutation.graphql";

import { MoveToCategorySelect } from "../../trackers/_components/MoveToCategorySelect";

import { TrackerResourceRowEdit } from "./TrackerResourceRowEdit";

const trackerResourceFragment = graphql`
  fragment TrackerResourceRowFragment on TrackerResource {
    id
    type
    origin
    path
    displayName
    description
    excluded
    lastDetectedAt
    cookieCategory {
      id
      name
    }
  }
`;

const deleteResourceMutation = graphql`
  mutation TrackerResourceRowDeleteMutation(
    $input: DeleteTrackerResourceInput!
    $connections: [ID!]!
  ) {
    deleteTrackerResource(input: $input) {
      deletedTrackerResourceId @deleteEdge(connections: $connections)
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

const moveResourceMutation = graphql`
  mutation TrackerResourceRowMoveMutation(
    $input: MoveTrackerResourceToCategoryInput!
  ) {
    moveTrackerResourceToCategory(input: $input) {
      trackerResource {
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

const updateResourceMutation = graphql`
  mutation TrackerResourceRowUpdateMutation(
    $input: UpdateTrackerResourceInput!
  ) {
    updateTrackerResource(input: $input) {
      trackerResource {
        id
        displayName
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

function resourceTypeBadge(type: string, __: (s: string) => string) {
  switch (type) {
    case "SCRIPT": return { label: __("Script"), variant: "info" as const };
    case "IFRAME": return { label: __("Iframe"), variant: "warning" as const };
    case "IMAGE": return { label: __("Image"), variant: "neutral" as const };
    case "STYLESHEET": return { label: __("Stylesheet"), variant: "highlight" as const };
    case "FONT": return { label: __("Font"), variant: "outline" as const };
    case "BEACON": return { label: __("Beacon"), variant: "danger" as const };
    case "FETCH": return { label: __("Fetch"), variant: "success" as const };
    case "MEDIA": return { label: __("Media"), variant: "neutral" as const };
    case "SERVICE_WORKER": return { label: __("Service Worker"), variant: "warning" as const };
    default: return { label: type, variant: "neutral" as const };
  }
}

interface TrackerResourceRowProps {
  resourceKey: TrackerResourceRowFragment$key;
  connectionId: string;
}

export function TrackerResourceRow({ resourceKey, connectionId }: TrackerResourceRowProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const resource = useFragment(trackerResourceFragment, resourceKey);
  const typeBadge = resourceTypeBadge(resource.type, __);

  const [isEditing, setIsEditing] = useState(false);

  const [deleteResource]
    = useMutation<TrackerResourceRowDeleteMutation>(deleteResourceMutation);
  const [moveResource]
    = useMutation<TrackerResourceRowMoveMutation>(moveResourceMutation);
  const [updateResource, isUpdating]
    = useMutation<TrackerResourceRowUpdateMutation>(updateResourceMutation);

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteResource({
            variables: {
              input: { trackerResourceId: resource.id },
              connections: [connectionId],
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({ title: __("Error"), description: errors[0].message, variant: "error" });
              } else {
                toast({ title: __("Success"), description: __("Resource deleted"), variant: "success" });
              }
              resolve();
            },
            onError(error) {
              toast({ title: __("Error"), description: formatError(__("Failed to delete resource"), error as GraphQLError), variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: __("Are you sure you want to delete \"%s\"?").replace("%s", resource.displayName),
        variant: "danger",
        label: __("Delete"),
      },
    );
  };

  const handleMove = (targetCategoryId: string) => {
    moveResource({
      variables: {
        input: {
          trackerResourceId: resource.id,
          targetCookieCategoryId: targetCategoryId,
        },
      },
      updater(store) {
        const payload = store.getRootField("moveTrackerResourceToCategory");
        if (!payload?.getLinkedRecord("trackerResource")) {
          return;
        }

        const conn = store.get(connectionId);
        if (conn) {
          ConnectionHandler.deleteNode(conn, resource.id);
        }
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: __("Success"), description: __("Resource moved"), variant: "success" });
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to move resource"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const handleToggleExcluded = () => {
    updateResource({
      variables: {
        input: {
          trackerResourceId: resource.id,
          excluded: !resource.excluded,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
          return;
        }
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to update resource"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const handleSaveEdit = (data: { displayName: string; description: string }) => {
    updateResource({
      variables: {
        input: {
          trackerResourceId: resource.id,
          displayName: data.displayName,
          description: data.description,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: __("Success"), description: __("Resource updated"), variant: "success" });
        setIsEditing(false);
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to update resource"), error as GraphQLError), variant: "error" });
      },
    });
  };

  if (isEditing) {
    return (
      <TrackerResourceRowEdit
        displayName={resource.displayName}
        description={resource.description}
        isUpdating={isUpdating}
        onSave={handleSaveEdit}
        onCancel={() => setIsEditing(false)}
      />
    );
  }

  return (
    <Tr className={resource.excluded ? "bg-txt-quaternary opacity-80 line-through" : undefined}>
      <Td>
        <Badge variant={typeBadge.variant}>
          {typeBadge.label}
        </Badge>
      </Td>
      <Td>
        <div className="flex flex-col min-w-0">
          <span className={resource.excluded ? undefined : "font-medium"}>{resource.origin}</span>
          {resource.description && (
            <span className="text-xs text-txt-tertiary wrap-break-word line-clamp-1">
              {resource.description}
            </span>
          )}
        </div>
      </Td>
      <Td>
        <span className="font-mono text-xs break-all max-w-xs inline-block">{resource.path}</span>
      </Td>
      <Td>
        <MoveToCategorySelect
          currentCategoryId={resource.cookieCategory?.id}
          currentCategoryName={resource.cookieCategory?.name}
          onSelect={handleMove}
        />
      </Td>
      <Td>
        {resource.lastDetectedAt
          ? (
              <time dateTime={resource.lastDetectedAt}>
                {new Date(resource.lastDetectedAt).toLocaleString()}
              </time>
            )
          : <span className="text-txt-tertiary">-</span>}
      </Td>
      <Td className="w-px whitespace-nowrap">
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
            <DropdownItem
              icon={resource.excluded ? EyeIcon : EyeSlashIcon}
              onSelect={handleToggleExcluded}
            >
              {resource.excluded ? __("Include") : __("Exclude")}
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
