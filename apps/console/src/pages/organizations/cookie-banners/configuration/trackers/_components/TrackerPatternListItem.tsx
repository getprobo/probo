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

import { DotsThreeVerticalIcon, EyeIcon, EyeSlashIcon, InfoIcon, TrashIcon } from "@phosphor-icons/react";
import { dateTimeFormat, humanizeSeconds } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
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
import { MoveToCategorySelect } from "../../_components/MoveToCategorySelect";
import { persistentTrackerTypes } from "../../_lib/persistentTrackerTypes";
import { cookieSourceBadges, trackerTypeBadges } from "../_lib/trackerBadges";

import { DeleteTrackerPatternDialog } from "./DeleteTrackerPatternDialog";
import { TrackerAttributionLabel } from "./TrackerAttributionLabel";

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
    }
    commonThirdParty {
      id
      name
    }
    attribution
    canUpdate: permission(action: "core:tracker-pattern:update")
    canDelete: permission(action: "core:tracker-pattern:delete")
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
  const [deleteOpen, setDeleteOpen] = useState(false);
  const { name, heading, title, info, detail, date, actions } = trackerPatternListItem({
    excluded: pattern.excluded,
  });
  const description = pattern.description.trim();

  const [movePattern] = useMutation<TrackerPatternListItemMoveMutation>(
    movePatternMutation,
    {
      successMessage: t("trackerPatternRow.messages.cookieMoved"),
      errorToast: t("trackerPatternRow.errors.moveCookie"),
    },
  );
  const [updatePattern] = useMutation<TrackerPatternListItemUpdateMutation>(
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

  const typeBadge = trackerTypeBadges[pattern.trackerType];
  const sourceBadge = pattern.source == null ? null : cookieSourceBadges[pattern.source];
  const detailPath = `${cookieBannerPath(organizationId, cookieBannerId)}/trackers/${pattern.id}`;
  const durationSeconds = pattern.maxAgeSeconds == null || pattern.maxAgeSeconds <= 0
    ? null
    : pattern.maxAgeSeconds;
  const sessionOrPersistent = persistentTrackerTypes.has(pattern.trackerType)
    ? t("trackerPatternRow.duration.persistent")
    : t("trackerPatternRow.duration.session");
  const duration = durationSeconds == null
    ? sessionOrPersistent
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
              {description !== ""
                ? (
                    <span className={info()}>
                      <Popover>
                        <PopoverTrigger
                          render={(
                            <IconButton
                              size={1}
                              variant="ghost"
                              color="neutral"
                              aria-label={t("trackerPatternRow.actions.description")}
                            />
                          )}
                        >
                          <InfoIcon />
                        </PopoverTrigger>
                        <PopoverPopup>
                          <div className={detail()}>
                            <Text size={2} weight="medium" highContrast>
                              {pattern.displayName}
                            </Text>
                            <Text size={2} color="faint">
                              {description}
                            </Text>
                          </div>
                        </PopoverPopup>
                      </Popover>
                    </span>
                  )
                : null}
            </div>
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
        {pattern.canUpdate
          ? (
              <TableCell interactive>
                <MoveToCategorySelect
                  cookieBannerKey={cookieBannerKey}
                  currentCategoryId={pattern.cookieCategory?.id}
                  currentCategoryName={pattern.cookieCategory?.name}
                  onSelect={handleMove}
                />
              </TableCell>
            )
          : (
              <TableCell>
                <Text size={2} color={pattern.cookieCategory == null ? "faint" : undefined}>
                  {pattern.cookieCategory?.name ?? "-"}
                </Text>
              </TableCell>
            )}
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
        <TableCell interactive={pattern.canUpdate || pattern.canDelete ? true : undefined} justify="end">
          {(pattern.canUpdate || pattern.canDelete) && (
            <div className={actions()}>
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
                  {pattern.canUpdate && (
                    <DropdownItem
                      iconStart={pattern.excluded ? <EyeIcon /> : <EyeSlashIcon />}
                      onClick={handleToggleExcluded}
                    >
                      {pattern.excluded
                        ? t("trackerPatternRow.actions.include")
                        : t("trackerPatternRow.actions.exclude")}
                    </DropdownItem>
                  )}
                  {pattern.canDelete && (
                    <DropdownItem
                      color="error"
                      iconStart={<TrashIcon />}
                      onClick={() => setDeleteOpen(true)}
                    >
                      {t("trackerPatternRow.actions.delete")}
                    </DropdownItem>
                  )}
                </DropdownPopup>
              </Dropdown>
            </div>
          )}
        </TableCell>
      </TableRow>
      {pattern.canDelete && (
        <DeleteTrackerPatternDialog
          trackerPatternId={pattern.id}
          displayName={pattern.displayName}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
          onRemoved={onRemoved}
        />
      )}
    </>
  );
}
