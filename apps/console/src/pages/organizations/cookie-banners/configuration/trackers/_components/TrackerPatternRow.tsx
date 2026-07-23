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

import { DownloadSimpleIcon, EyeIcon, EyeSlashIcon } from "@phosphor-icons/react";
import { formatError } from "@probo/helpers";
import { dateTimeFormat, humanizeSeconds } from "@probo/i18n";
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
      kind
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
          name
          kind
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
  const { t, i18n } = useTranslation("organizations/cookie-banners");
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
                toast({ title: t("trackerPatternRow.errors.title"), description: errors[0].message, variant: "error" });
              } else {
                toast({ title: t("trackerPatternRow.messages.successTitle"), description: t("trackerPatternRow.messages.cookieDeleted"), variant: "success" });
              }
              resolve();
            },
            onError(error) {
              toast({ title: t("trackerPatternRow.errors.title"), description: formatError(t("trackerPatternRow.errors.deleteCookie"), error), variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: t("trackerPatternRow.deleteConfirmation", { name: pattern.displayName }),
        variant: "danger",
        label: t("trackerPatternRow.actions.delete"),
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
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: t("trackerPatternRow.errors.title"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: t("trackerPatternRow.messages.successTitle"), description: t("trackerPatternRow.messages.cookieMoved"), variant: "success" });
      },
      onError(error) {
        toast({ title: t("trackerPatternRow.errors.title"), description: formatError(t("trackerPatternRow.errors.moveCookie"), error), variant: "error" });
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
          toast({ title: t("trackerPatternRow.errors.title"), description: errors[0].message, variant: "error" });
          return;
        }
      },
      onError(error) {
        toast({ title: t("trackerPatternRow.errors.title"), description: formatError(t("trackerPatternRow.errors.updateCookie"), error), variant: "error" });
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
          toast({ title: t("trackerPatternRow.errors.title"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: t("trackerPatternRow.messages.successTitle"), description: t("trackerPatternRow.messages.thirdPartyImported"), variant: "success" });
      },
      onError(error) {
        toast({ title: t("trackerPatternRow.errors.title"), description: formatError(t("trackerPatternRow.errors.importThirdParty"), error), variant: "error" });
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
          toast({ title: t("trackerPatternRow.errors.title"), description: errors[0].message, variant: "error" });
          return;
        }
        toast({ title: t("trackerPatternRow.messages.successTitle"), description: t("trackerPatternRow.messages.cookieUpdated"), variant: "success" });
        setIsEditing(false);
      },
      onError(error) {
        toast({ title: t("trackerPatternRow.errors.title"), description: formatError(t("trackerPatternRow.errors.updateCookie"), error), variant: "error" });
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

  const typeBadges = {
    COOKIE: { variant: "warning" as const, label: t("trackerPatternRow.types.cookie") },
    LOCAL_STORAGE: { variant: "info" as const, label: t("trackerPatternRow.types.localStorage") },
    SESSION_STORAGE: { variant: "highlight" as const, label: t("trackerPatternRow.types.sessionStorage") },
    INDEXED_DB: { variant: "success" as const, label: t("trackerPatternRow.types.indexedDb") },
    CACHE_STORAGE: { variant: "outline" as const, label: t("trackerPatternRow.types.cacheStorage") },
  };
  const sourceBadges = {
    SCRIPT: { variant: "info" as const, label: t("trackerPatternRow.sources.script") },
    PRE_EXISTING: { variant: "outline" as const, label: t("trackerPatternRow.sources.preExisting") },
    HTTP: { variant: "neutral" as const, label: t("trackerPatternRow.sources.http") },
    EXTENSION: { variant: "warning" as const, label: t("trackerPatternRow.sources.extension") },
  };
  const typeBadge = typeBadges[pattern.trackerType]
    ?? { variant: "neutral" as const, label: pattern.trackerType };
  const srcBadge = pattern.source
    ? sourceBadges[pattern.source]
    ?? { variant: "neutral" as const, label: pattern.source }
    : null;
  const formatDuration = (seconds: number | null) => {
    if (seconds === null || seconds <= 0) {
      return ["LOCAL_STORAGE", "INDEXED_DB", "CACHE_STORAGE"].includes(pattern.trackerType)
        ? t("trackerPatternRow.duration.persistent")
        : t("trackerPatternRow.duration.session");
    }

    return humanizeSeconds(seconds, t);
  };
  const duration = formatDuration(pattern.maxAgeSeconds ?? null);

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
                  <Badge variant="info">{t("trackerPatternRow.commonCatalog")}</Badge>
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
        <span className="pl-2">{duration}</span>
      </Td>
      <Td>
        {pattern.lastMatchedAt
          ? (
              <time dateTime={pattern.lastMatchedAt}>
                {dateTimeFormat(i18n.language, pattern.lastMatchedAt)}
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
            title={t("trackerPatternRow.actions.edit")}
          >
            <IconPencil size={14} />
          </button>
          <ActionDropdown>
            {!pattern.thirdParty && pattern.commonThirdParty && (
              <DropdownItem
                icon={DownloadSimpleIcon}
                onSelect={handleImport}
              >
                {t("trackerPatternRow.actions.importThirdParties")}
              </DropdownItem>
            )}
            <DropdownItem
              icon={pattern.excluded ? EyeIcon : EyeSlashIcon}
              onSelect={handleToggleExcluded}
            >
              {pattern.excluded ? t("trackerPatternRow.actions.include") : t("trackerPatternRow.actions.exclude")}
            </DropdownItem>
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onSelect={handleDelete}
            >
              {t("trackerPatternRow.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        </div>
      </Td>
    </Tr>
  );
}
