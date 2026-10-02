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

import { CaretLeftIcon, EyeIcon, EyeSlashIcon } from "@phosphor-icons/react";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { TrackerPatternDetailHeader_trackerPattern$key } from "#/__generated__/core/TrackerPatternDetailHeader_trackerPattern.graphql";
import type { TrackerPatternDetailHeaderUpdateMutation } from "#/__generated__/core/TrackerPatternDetailHeaderUpdateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { cookieBannerPath } from "../../../_lib/cookieBannerPaths";
import { trackerPatternDetailHeader } from "../../../variants";
import { cookieSourceBadges, trackerTypeBadges } from "../_lib/trackerBadges";

const trackerPatternDetailHeaderFragment = graphql`
  fragment TrackerPatternDetailHeader_trackerPattern on TrackerPattern {
    id
    displayName
    trackerType
    source
    excluded
    canUpdate: permission(action: "core:tracker-pattern:update")
  }
`;

const updatePatternMutation = graphql`
  mutation TrackerPatternDetailHeaderUpdateMutation(
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

interface TrackerPatternDetailHeaderProps {
  trackerPatternKey: TrackerPatternDetailHeader_trackerPattern$key;
}

export function TrackerPatternDetailHeader({
  trackerPatternKey,
}: TrackerPatternDetailHeaderProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const organizationId = useOrganizationId();
  const location = useLocation();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const pattern = useFragment(trackerPatternDetailHeaderFragment, trackerPatternKey);
  const { root, back, bar, titleRow, title, badges, actions } = trackerPatternDetailHeader();
  const typeBadge = trackerTypeBadges[pattern.trackerType];
  const sourceBadge = pattern.source == null ? null : cookieSourceBadges[pattern.source];
  const [updatePattern, isUpdating] = useMutation<TrackerPatternDetailHeaderUpdateMutation>(
    updatePatternMutation,
    {
      successMessage: t("trackerProperties.messages.updated"),
      errorToast: t("trackerProperties.errors.update"),
    },
  );

  if (cookieBannerId == null) {
    throw new Error(":cookieBannerId missing in route params");
  }

  const trackersPath = `${cookieBannerPath(organizationId, cookieBannerId)}/trackers`;

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

  return (
    <div className={root()}>
      <Link
        to={{ pathname: trackersPath, search: location.search }}
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("trackerProperties.actions.back")}
      </Link>
      <div className={bar()}>
        <div className={titleRow()}>
          <Heading level={1} size={6} weight="medium" highContrast className={title()}>
            {pattern.displayName}
          </Heading>
          <div className={badges()}>
            {typeBadge != null && (
              <Badge variant={typeBadge.variant} color={typeBadge.color}>
                {t(`trackerPatternRow.types.${typeBadge.labelKey}`)}
              </Badge>
            )}
            {sourceBadge != null && (
              <Badge variant={sourceBadge.variant} color={sourceBadge.color}>
                {t(`trackerPatternRow.sources.${sourceBadge.labelKey}`)}
              </Badge>
            )}
          </div>
        </div>
        {pattern.canUpdate && (
          <div className={actions()}>
            <Button
              type="button"
              size={2}
              variant="soft"
              color="neutral"
              loading={isUpdating}
              iconStart={pattern.excluded ? <EyeIcon /> : <EyeSlashIcon />}
              onClick={handleToggleExcluded}
            >
              {pattern.excluded
                ? t("trackerPatternRow.actions.include")
                : t("trackerPatternRow.actions.exclude")}
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
