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

import { DotsThreeVerticalIcon, EyeIcon, EyeSlashIcon, PencilSimpleIcon, TrashIcon } from "@phosphor-icons/react";
import { dateTimeFormat, humanizeSeconds } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableLink } from "@probo/ui/src/v2/Table/TableLink";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { MoveToCategorySelect_cookieBanner$key } from "#/__generated__/core/MoveToCategorySelect_cookieBanner.graphql";
import type { TrackerPatternListItem_trackerPattern$key } from "#/__generated__/core/TrackerPatternListItem_trackerPattern.graphql";
import type { TrackerPatternListItemMoveMutation } from "#/__generated__/core/TrackerPatternListItemMoveMutation.graphql";
import type { TrackerPatternListItemUpdateMutation } from "#/__generated__/core/TrackerPatternListItemUpdateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { cookieBannerPath } from "../../../_lib/cookieBannerPaths";
import { trackerPatternListItem } from "../../../variants";

import { DeleteTrackerPatternDialog } from "./DeleteTrackerPatternDialog";
import { MoveToCategorySelect } from "./MoveToCategorySelect";
import { TrackerAttributionLabel } from "./TrackerAttributionLabel";
import { TrackerPatternListItemEdit } from "./TrackerPatternListItemEdit";

const trackerPatternFragment = graphql`
  fragment TrackerPatternListItem_trackerPattern on TrackerPattern {
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
    commonThirdParty {
      id
      name
    }
    attribution
  }
`;

const movePatternMutation = graphql`
  mutation TrackerPatternListItemMoveMutation(
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
  mutation TrackerPatternListItemUpdateMutation(
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

const typeBadges = {
  COOKIE: { color: "amber" as const, labelKey: "cookie", variant: "soft" as const },
  LOCAL_STORAGE: { color: "sky" as const, labelKey: "localStorage", variant: "soft" as const },
  SESSION_STORAGE: { color: "indigo" as const, labelKey: "sessionStorage", variant: "soft" as const },
  INDEXED_DB: { color: "green" as const, labelKey: "indexedDb", variant: "soft" as const },
  CACHE_STORAGE: { color: "neutral" as const, labelKey: "cacheStorage", variant: "outline" as const },
};

const sourceBadges = {
  SCRIPT: { color: "sky" as const, labelKey: "script", variant: "soft" as const },
  PRE_EXISTING: { color: "neutral" as const, labelKey: "preExisting", variant: "outline" as const },
  HTTP: { color: "neutral" as const, labelKey: "http", variant: "soft" as const },
  EXTENSION: { color: "amber" as const, labelKey: "extension", variant: "soft" as const },
};

interface TrackerPatternListItemProps {
  patternKey: TrackerPatternListItem_trackerPattern$key;
  cookieBannerKey: MoveToCategorySelect_cookieBanner$key;
  onRemoved: () => void;
}

export function TrackerPatternListItem({
  patternKey,
  cookieBannerKey,
  onRemoved,
}: TrackerPatternListItemProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const pattern = useFragment(trackerPatternFragment, patternKey);
  const [isEditing, setIsEditing] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const { name, heading, title, description, date, actions } = trackerPatternListItem({
    excluded: pattern.excluded,
  });

  const [movePattern] = useMutation<TrackerPatternListItemMoveMutation>(
    movePatternMutation,
    {
      successMessage: t("trackerPatternRow.messages.cookieMoved"),
      errorToast: t("trackerPatternRow.errors.moveCookie"),
    },
  );
  const [updatePattern, isUpdating] = useMutation<TrackerPatternListItemUpdateMutation>(
    updatePatternMutation,
    {
      errorToast: t("trackerPatternRow.errors.updateCookie"),
    },
  );

  if (cookieBannerId == null) {
    throw new Error(":cookieBannerId missing in route params");
  }

  function handleMove(targetCategoryId: string) {
    if (targetCategoryId === pattern.cookieCategory?.id) {
      return;
    }
    void movePattern({
      variables: {
        input: {
          trackerPatternId: pattern.id,
          targetCookieCategoryId: targetCategoryId,
        },
      },
    }).catch(() => undefined);
  }

  function handleToggleExcluded() {
    void updatePattern({
      variables: {
        input: {
          trackerPatternId: pattern.id,
          excluded: !pattern.excluded,
        },
      },
    }).catch(() => undefined);
  }

  function handleSaveEdit(data: { description: string; maxAgeSeconds: number | null }) {
    void updatePattern({
      variables: {
        input: {
          trackerPatternId: pattern.id,
          description: data.description,
          maxAgeSeconds: data.maxAgeSeconds,
        },
      },
    }, {
      successMessage: t("trackerPatternRow.messages.cookieUpdated"),
    }).then(
      () => {
        setIsEditing(false);
      },
      () => undefined,
    );
  }

  if (isEditing) {
    return (
      <TrackerPatternListItemEdit
        pattern={pattern.displayName}
        description={pattern.description}
        maxAgeSeconds={pattern.maxAgeSeconds ?? null}
        isUpdating={isUpdating}
        onSave={handleSaveEdit}
        onCancel={() => setIsEditing(false)}
      />
    );
  }

  const typeBadge = typeBadges[pattern.trackerType];
  const sourceBadge = pattern.source == null ? null : sourceBadges[pattern.source];
  const detailPath = `${cookieBannerPath(organizationId, cookieBannerId)}/trackers/${pattern.id}`;
  const durationSeconds = pattern.maxAgeSeconds ?? null;
  const duration = durationSeconds == null || durationSeconds <= 0
    ? ["LOCAL_STORAGE", "INDEXED_DB", "CACHE_STORAGE"].includes(pattern.trackerType)
      ? t("trackerPatternRow.duration.persistent")
      : t("trackerPatternRow.duration.session")
    : humanizeSeconds(durationSeconds, t);

  return (
    <>
      <TableRow align="center" interactive>
        <TableRowHeaderCell>
          <div className={name()}>
            <div className={heading()}>
              {typeBadge != null && (
                <Badge
                  variant={typeBadge.variant}
                  color={typeBadge.color}
                >
                  {t(`trackerPatternRow.types.${typeBadge.labelKey}`)}
                </Badge>
              )}
              <TableLink to={detailPath}>
                <Text size={2} weight="medium" highContrast className={title()}>
                  {pattern.displayName}
                </Text>
              </TableLink>
            </div>
            {pattern.description
              ? (
                  <Text size={1} color="faint" className={description()}>
                    {pattern.description}
                  </Text>
                )
              : null}
          </div>
        </TableRowHeaderCell>
        <TableCell>
          {pattern.commonThirdParty
            ? (
                <Text size={2} highContrast>
                  {pattern.commonThirdParty.name}
                </Text>
              )
            : <TrackerAttributionLabel attribution={pattern.attribution} />}
        </TableCell>
        <TableCell>
          {sourceBadge == null
            ? <Text size={2} color="faint">-</Text>
            : (
                <Badge
                  variant={sourceBadge.variant}
                  color={sourceBadge.color}
                >
                  {t(`trackerPatternRow.sources.${sourceBadge.labelKey}`)}
                </Badge>
              )}
        </TableCell>
        <TableCell interactive>
          <MoveToCategorySelect
            cookieBannerKey={cookieBannerKey}
            currentCategoryId={pattern.cookieCategory?.id}
            currentCategoryName={pattern.cookieCategory?.name}
            highlight={pattern.cookieCategory != null && pattern.cookieCategory.kind !== "UNCATEGORISED"}
            onSelect={handleMove}
          />
        </TableCell>
        <TableCell>
          <Text size={2} className={date()}>{duration}</Text>
        </TableCell>
        <TableCell>
          {pattern.lastMatchedAt == null
            ? <Text size={2} color="faint">-</Text>
            : (
                <time dateTime={pattern.lastMatchedAt} className={date()}>
                  <Text size={2}>
                    {dateTimeFormat(i18n.language, pattern.lastMatchedAt)}
                  </Text>
                </time>
              )}
        </TableCell>
        <TableCell interactive justify="end">
          <div className={actions()}>
            <IconButton
              variant="ghost"
              color="neutral"
              size={1}
              aria-label={t("trackerPatternRow.actions.edit")}
              onClick={() => setIsEditing(true)}
            >
              <PencilSimpleIcon />
            </IconButton>
            <Dropdown>
              <DropdownTrigger
                render={(
                  <IconButton
                    variant="ghost"
                    color="neutral"
                    size={1}
                    aria-label={t("trackerPatternRow.actions.more")}
                  >
                    <DotsThreeVerticalIcon />
                  </IconButton>
                )}
              />
              <DropdownPopup align="end">
                <DropdownItem
                  iconStart={pattern.excluded ? <EyeIcon /> : <EyeSlashIcon />}
                  onClick={handleToggleExcluded}
                >
                  {pattern.excluded
                    ? t("trackerPatternRow.actions.include")
                    : t("trackerPatternRow.actions.exclude")}
                </DropdownItem>
                <DropdownItem
                  color="error"
                  iconStart={<TrashIcon />}
                  onClick={() => setDeleteOpen(true)}
                >
                  {t("trackerPatternRow.actions.delete")}
                </DropdownItem>
              </DropdownPopup>
            </Dropdown>
          </div>
        </TableCell>
      </TableRow>
      <DeleteTrackerPatternDialog
        trackerPatternId={pattern.id}
        displayName={pattern.displayName}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        onRemoved={onRemoved}
      />
    </>
  );
}
