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
import { formatError } from "@probo/helpers";
import { dateTimeFormat } from "@probo/i18n";
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
import { useTranslation } from "react-i18next";
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

interface TrackerResourceRowProps {
  resourceKey: TrackerResourceRowFragment$key;
  connectionId: string;
}

export function TrackerResourceRow({ resourceKey, connectionId }: TrackerResourceRowProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();
  const confirm = useConfirm();
  const resource = useFragment(trackerResourceFragment, resourceKey);
  const typeBadges = {
    SCRIPT: { label: t("trackerResourceRow.types.script"), variant: "info" as const },
    IFRAME: { label: t("trackerResourceRow.types.iframe"), variant: "warning" as const },
    IMAGE: { label: t("trackerResourceRow.types.image"), variant: "neutral" as const },
    STYLESHEET: { label: t("trackerResourceRow.types.stylesheet"), variant: "highlight" as const },
    FONT: { label: t("trackerResourceRow.types.font"), variant: "outline" as const },
    BEACON: { label: t("trackerResourceRow.types.beacon"), variant: "danger" as const },
    FETCH: { label: t("trackerResourceRow.types.fetch"), variant: "success" as const },
    MEDIA: { label: t("trackerResourceRow.types.media"), variant: "neutral" as const },
    SERVICE_WORKER: { label: t("trackerResourceRow.types.serviceWorker"), variant: "warning" as const },
  };
  const typeBadge = typeBadges[resource.type]
    ?? { label: resource.type, variant: "neutral" as const };

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
                toast({ title: t("trackerResourceRow.errors.title"), description: errors[0].message, variant: "error" });
              } else {
                toast({ title: t("trackerResourceRow.messages.successTitle"), description: t("trackerResourceRow.messages.deleted"), variant: "success" });
              }
              resolve();
            },
            onError(error) {
              toast({ title: t("trackerResourceRow.errors.title"), description: formatError(t("trackerResourceRow.errors.delete"), error), variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: t("trackerResourceRow.deleteConfirmation", { name: resource.displayName }),
        variant: "danger",
        label: t("trackerResourceRow.actions.delete"),
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
          toast({ title: t("trackerResourceRow.errors.title"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: t("trackerResourceRow.messages.successTitle"), description: t("trackerResourceRow.messages.moved"), variant: "success" });
      },
      onError(error) {
        toast({ title: t("trackerResourceRow.errors.title"), description: formatError(t("trackerResourceRow.errors.move"), error), variant: "error" });
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
          toast({ title: t("trackerResourceRow.errors.title"), description: errors[0].message, variant: "error" });
          return;
        }
      },
      onError(error) {
        toast({ title: t("trackerResourceRow.errors.title"), description: formatError(t("trackerResourceRow.errors.update"), error), variant: "error" });
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
          toast({ title: t("trackerResourceRow.errors.title"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: t("trackerResourceRow.messages.successTitle"), description: t("trackerResourceRow.messages.updated"), variant: "success" });
        setIsEditing(false);
      },
      onError(error) {
        toast({ title: t("trackerResourceRow.errors.title"), description: formatError(t("trackerResourceRow.errors.update"), error), variant: "error" });
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
                {dateTimeFormat(i18n.language, resource.lastDetectedAt)}
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
            title={t("trackerResourceRow.actions.edit")}
          >
            <IconPencil size={14} />
          </button>
          <ActionDropdown>
            <DropdownItem
              icon={resource.excluded ? EyeIcon : EyeSlashIcon}
              onSelect={handleToggleExcluded}
            >
              {resource.excluded ? t("trackerResourceRow.actions.include") : t("trackerResourceRow.actions.exclude")}
            </DropdownItem>
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onSelect={handleDelete}
            >
              {t("trackerResourceRow.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        </div>
      </Td>
    </Tr>
  );
}
