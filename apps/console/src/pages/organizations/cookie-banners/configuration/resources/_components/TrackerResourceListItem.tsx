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

import { DotsThreeVerticalIcon, EyeIcon, EyeSlashIcon, TrashIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { MoveToCategorySelect_cookieBanner$key } from "#/__generated__/core/MoveToCategorySelect_cookieBanner.graphql";
import type { TrackerResourceListItem_trackerResource$key } from "#/__generated__/core/TrackerResourceListItem_trackerResource.graphql";
import type { TrackerResourceListItemMoveMutation } from "#/__generated__/core/TrackerResourceListItemMoveMutation.graphql";
import type { TrackerResourceListItemUpdateMutation } from "#/__generated__/core/TrackerResourceListItemUpdateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { trackerResourceListItem } from "../../../variants";
import { MoveToCategorySelect } from "../../_components/MoveToCategorySelect";
import { resourceTypeBadges } from "../_lib/resourceBadges";

import { DeleteTrackerResourceDialog } from "./DeleteTrackerResourceDialog";

const trackerResourceFragment = graphql`
  fragment TrackerResourceListItem_trackerResource on TrackerResource {
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
    canUpdate: permission(action: "core:tracker-resource:update")
    canDelete: permission(action: "core:tracker-resource:delete")
  }
`;

const moveResourceMutation = graphql`
  mutation TrackerResourceListItemMoveMutation(
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
  mutation TrackerResourceListItemUpdateMutation(
    $input: UpdateTrackerResourceInput!
  ) {
    updateTrackerResource(input: $input) {
      trackerResource {
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

interface TrackerResourceListItemProps {
  resourceKey: TrackerResourceListItem_trackerResource$key;
  cookieBannerKey: MoveToCategorySelect_cookieBanner$key;
  onRemoved: () => void;
}

export function TrackerResourceListItem({
  resourceKey,
  cookieBannerKey,
  onRemoved,
}: TrackerResourceListItemProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const resource = useFragment(trackerResourceFragment, resourceKey);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const { origin, title, description, path, date, actions } = trackerResourceListItem({
    excluded: resource.excluded,
  });

  const [moveResource] = useMutation<TrackerResourceListItemMoveMutation>(
    moveResourceMutation,
    {
      successMessage: t("trackerResourceRow.messages.moved"),
      errorToast: t("trackerResourceRow.errors.move"),
    },
  );
  const [updateResource] = useMutation<TrackerResourceListItemUpdateMutation>(
    updateResourceMutation,
    {
      errorToast: t("trackerResourceRow.errors.update"),
    },
  );

  function handleMove(targetCategoryId: string) {
    void moveResource({
      variables: {
        input: {
          trackerResourceId: resource.id,
          targetCookieCategoryId: targetCategoryId,
        },
      },
    }).then(
      () => {
        onRemoved();
      },
      () => undefined,
    );
  }

  function handleToggleExcluded() {
    void updateResource({
      variables: {
        input: {
          trackerResourceId: resource.id,
          excluded: !resource.excluded,
        },
      },
    }).catch(() => undefined);
  }

  const typeBadge = resourceTypeBadges[resource.type];

  return (
    <>
      <TableRow align="center">
        <TableCell>
          {typeBadge == null
            ? <Text size={2}>{resource.type}</Text>
            : (
                <Badge
                  variant={typeBadge.variant}
                  color={typeBadge.color}
                >
                  {t(`trackerResourceRow.types.${typeBadge.labelKey}`)}
                </Badge>
              )}
        </TableCell>
        <TableCell>
          <div className={origin()}>
            <Text size={2} weight="medium" highContrast className={title()}>
              {resource.origin}
            </Text>
            {resource.description
              ? (
                  <Text size={1} color="faint" className={description()}>
                    {resource.description}
                  </Text>
                )
              : null}
          </div>
        </TableCell>
        <TableCell>
          <Text size={1} className={path()}>
            {resource.path}
          </Text>
        </TableCell>
        {resource.canUpdate
          ? (
              <TableCell interactive>
                <MoveToCategorySelect
                  cookieBannerKey={cookieBannerKey}
                  currentCategoryId={resource.cookieCategory?.id}
                  currentCategoryName={resource.cookieCategory?.name}
                  onSelect={handleMove}
                />
              </TableCell>
            )
          : (
              <TableCell>
                <Text size={2} color={resource.cookieCategory == null ? "faint" : undefined}>
                  {resource.cookieCategory?.name ?? "-"}
                </Text>
              </TableCell>
            )}
        <TableCell>
          {resource.lastDetectedAt == null
            ? <Text size={2} color="faint">-</Text>
            : (
                <time dateTime={resource.lastDetectedAt} className={date()}>
                  <Text size={2}>
                    {dateTimeFormat(i18n.language, resource.lastDetectedAt)}
                  </Text>
                </time>
              )}
        </TableCell>
        <TableCell interactive={resource.canUpdate || resource.canDelete ? true : undefined} justify="end">
          {(resource.canUpdate || resource.canDelete) && (
            <div className={actions()}>
              <Dropdown>
                <DropdownTrigger
                  render={(
                    <IconButton
                      variant="ghost"
                      color="neutral"
                      size={1}
                      aria-label={t("trackerResourceRow.actions.more")}
                    >
                      <DotsThreeVerticalIcon />
                    </IconButton>
                  )}
                />
                <DropdownPopup align="end">
                  {resource.canUpdate && (
                    <DropdownItem
                      iconStart={resource.excluded ? <EyeIcon /> : <EyeSlashIcon />}
                      onClick={handleToggleExcluded}
                    >
                      {resource.excluded
                        ? t("trackerResourceRow.actions.include")
                        : t("trackerResourceRow.actions.exclude")}
                    </DropdownItem>
                  )}
                  {resource.canDelete && (
                    <DropdownItem
                      color="error"
                      iconStart={<TrashIcon />}
                      onClick={() => setDeleteOpen(true)}
                    >
                      {t("trackerResourceRow.actions.delete")}
                    </DropdownItem>
                  )}
                </DropdownPopup>
              </Dropdown>
            </div>
          )}
        </TableCell>
      </TableRow>
      {resource.canDelete && (
        <DeleteTrackerResourceDialog
          trackerResourceId={resource.id}
          displayName={resource.displayName}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
          onRemoved={onRemoved}
        />
      )}
    </>
  );
}
