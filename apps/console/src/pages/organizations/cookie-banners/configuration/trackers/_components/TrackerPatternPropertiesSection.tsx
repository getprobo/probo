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

import { CopyIcon } from "@phosphor-icons/react";
import { fromMaxAgeSeconds, toMaxAgeSeconds } from "@probo/helpers";
import { humanizeSeconds } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type FocusEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { TrackerPatternPropertiesSection_cookieBanner$key } from "#/__generated__/core/TrackerPatternPropertiesSection_cookieBanner.graphql";
import type { TrackerPatternPropertiesSection_trackerPattern$key } from "#/__generated__/core/TrackerPatternPropertiesSection_trackerPattern.graphql";
import type { TrackerPatternPropertiesSectionMoveMutation } from "#/__generated__/core/TrackerPatternPropertiesSectionMoveMutation.graphql";
import type { TrackerPatternPropertiesSectionUpdateMutation } from "#/__generated__/core/TrackerPatternPropertiesSectionUpdateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { trackerPatternPropertiesSection } from "../../../variants";

import { MoveToCategorySelect } from "./MoveToCategorySelect";
import { TrackerMaxAgeField } from "./TrackerMaxAgeField";

const cookieBannerFragment = graphql`
  fragment TrackerPatternPropertiesSection_cookieBanner on CookieBanner {
    ...MoveToCategorySelect_cookieBanner
  }
`;

const trackerPatternPropertiesSectionFragment = graphql`
  fragment TrackerPatternPropertiesSection_trackerPattern on TrackerPattern {
    id
    trackerType
    maxAgeSeconds
    description
    attribution
    commonTrackerPatternId
    canUpdate: permission(action: "core:tracker-pattern:update")
    cookieCategory {
      id
      name
    }
    commonThirdParty {
      name
    }
  }
`;

const movePatternMutation = graphql`
  mutation TrackerPatternPropertiesSectionMoveMutation(
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
  mutation TrackerPatternPropertiesSectionUpdateMutation(
    $input: UpdateTrackerPatternInput!
  ) {
    updateTrackerPattern(input: $input) {
      trackerPattern {
        id
        displayName
        maxAgeSeconds
        description
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

const persistentTrackerTypes = new Set(["LOCAL_STORAGE", "INDEXED_DB", "CACHE_STORAGE"]);

function attributionCopy(
  translate: (key: string, options?: { name?: string }) => string,
  attribution: string | null | undefined,
  thirdPartyName: string | null,
): string | null {
  if (thirdPartyName != null) {
    return translate("trackerProperties.attribution.thirdParty", { name: thirdPartyName });
  }
  switch (attribution) {
    case "FIRST_PARTY":
      return translate("trackerProperties.attribution.firstParty");
    case "NOT_ATTRIBUTABLE":
      return translate("trackerProperties.attribution.visitorSoftware");
    case "THIRD_PARTY":
      return translate("trackerProperties.attribution.thirdPartyUnnamed");
    default:
      return null;
  }
}

function maxAgeLabel(
  trackerType: string,
  maxAgeSeconds: number | null,
  translate: (key: string) => string,
): string {
  if (maxAgeSeconds == null) {
    if (persistentTrackerTypes.has(trackerType)) {
      return translate("trackerPatternRow.duration.persistent");
    }
    return translate("trackerPatternRow.duration.session");
  }
  return humanizeSeconds(maxAgeSeconds, translate);
}

interface TrackerPatternPropertiesSectionProps {
  trackerPatternKey: TrackerPatternPropertiesSection_trackerPattern$key;
  cookieBannerKey: TrackerPatternPropertiesSection_cookieBanner$key;
}

export function TrackerPatternPropertiesSection({
  trackerPatternKey,
  cookieBannerKey,
}: TrackerPatternPropertiesSectionProps) {
  const { toast } = useToast();
  const { t } = useTranslation("organizations/cookie-banners");
  const cookieBanner = useFragment(cookieBannerFragment, cookieBannerKey);
  const pattern = useFragment<TrackerPatternPropertiesSection_trackerPattern$key>(
    trackerPatternPropertiesSectionFragment,
    trackerPatternKey,
  );
  const { root, block, intro, fields, pair, sourceId, sourceIdText } = trackerPatternPropertiesSection();
  const [description, setDescription] = useState(pattern.description);
  const [duration, setDuration] = useState(() => fromMaxAgeSeconds(pattern.maxAgeSeconds ?? null));

  const [movePattern] = useMutation<TrackerPatternPropertiesSectionMoveMutation>(
    movePatternMutation,
    {
      successMessage: t("trackerProperties.messages.moved"),
      errorToast: t("trackerProperties.errors.move"),
    },
  );
  const [updatePattern, isUpdating] = useMutation<TrackerPatternPropertiesSectionUpdateMutation>(
    updatePatternMutation,
    {
      successMessage: t("trackerProperties.messages.updated"),
      errorToast: t("trackerProperties.errors.update"),
    },
  );

  const canUpdate = pattern.canUpdate;
  const attribution = attributionCopy(
    t,
    pattern.attribution,
    pattern.commonThirdParty?.name ?? null,
  );
  const currentMaxAge = pattern.maxAgeSeconds == null || pattern.maxAgeSeconds <= 0
    ? null
    : pattern.maxAgeSeconds;
  const readOnlyDuration = maxAgeLabel(pattern.trackerType, currentMaxAge, t);

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

  function saveDescription() {
    if (description === pattern.description) {
      return;
    }
    void updatePattern({
      variables: {
        input: {
          trackerPatternId: pattern.id,
          description,
        },
      },
    }).catch(() => undefined);
  }

  function saveMaxAge(next = duration) {
    const nextSeconds = toMaxAgeSeconds(next.value, next.unit);
    if (nextSeconds === currentMaxAge) {
      return;
    }
    void updatePattern({
      variables: {
        input: {
          trackerPatternId: pattern.id,
          maxAgeSeconds: nextSeconds,
        },
      },
    }).catch(() => undefined);
  }

  function handleDurationBlur(event: FocusEvent<HTMLDivElement>) {
    if (event.relatedTarget instanceof Node && event.currentTarget.contains(event.relatedTarget)) {
      return;
    }
    saveMaxAge();
  }

  function copySourceId() {
    const commonTrackerPatternId = pattern.commonTrackerPatternId;
    if (commonTrackerPatternId == null) {
      return;
    }
    void (async () => {
      try {
        await navigator.clipboard.writeText(commonTrackerPatternId);
        toast({
          title: t("trackerProperties.messages.copiedTitle"),
          description: t("trackerProperties.messages.idCopied"),
          variant: "success",
        });
      } catch {
        toast({
          title: t("trackerProperties.errors.title"),
          description: t("trackerProperties.errors.copy"),
          variant: "error",
        });
      }
    })();
  }

  return (
    <div className={root()}>
      <section className={block()}>
        <div className={intro()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("trackerProperties.sections.details")}
          </Heading>
        </div>
        <Card variant="soft" size={2}>
          <div className={fields()}>
            <div className={pair()}>
              <Field label={t("trackerProperties.properties.attribution")}>
                <Text size={2} color={attribution == null ? "faint" : undefined}>
                  {attribution ?? "-"}
                </Text>
              </Field>
              <Field label={t("trackerProperties.properties.sourceId")}>
                {pattern.commonTrackerPatternId == null
                  ? (
                      <Text size={2} color="faint">
                        {t("trackerProperties.manual")}
                      </Text>
                    )
                  : (
                      <div className={sourceId()}>
                        <Text size={2} className={sourceIdText()}>
                          {pattern.commonTrackerPatternId}
                        </Text>
                        <IconButton
                          size={1}
                          variant="soft"
                          color="neutral"
                          aria-label={t("trackerProperties.actions.copyId")}
                          onClick={copySourceId}
                        >
                          <CopyIcon />
                        </IconButton>
                      </div>
                    )}
              </Field>
            </div>
            <Field label={t("trackerProperties.properties.description")}>
              {canUpdate
                ? (
                    <Textarea
                      rows={3}
                      value={description}
                      disabled={isUpdating}
                      placeholder={t("trackerProperties.properties.descriptionPlaceholder")}
                      onChange={event => setDescription(event.target.value)}
                      onBlur={saveDescription}
                    />
                  )
                : (
                    <Text size={2} color={pattern.description === "" ? "faint" : undefined}>
                      {pattern.description === "" ? "-" : pattern.description}
                    </Text>
                  )}
            </Field>
            <div className={pair()}>
              <Field label={t("trackerProperties.properties.category")}>
                {canUpdate
                  ? (
                      <MoveToCategorySelect
                        cookieBannerKey={cookieBanner}
                        currentCategoryId={pattern.cookieCategory?.id}
                        currentCategoryName={pattern.cookieCategory?.name}
                        size={2}
                        onSelect={handleMove}
                      />
                    )
                  : (
                      <Text size={2} color={pattern.cookieCategory == null ? "faint" : undefined}>
                        {pattern.cookieCategory?.name ?? "-"}
                      </Text>
                    )}
              </Field>
              <Field label={t("trackerProperties.properties.maxAge")}>
                {canUpdate
                  ? (
                      <TrackerMaxAgeField
                        value={duration.value}
                        unit={duration.unit}
                        disabled={isUpdating}
                        onValueChange={(value) => {
                          setDuration(current => ({ ...current, value }));
                        }}
                        onUnitChange={(unit) => {
                          const next = { ...duration, unit };
                          setDuration(next);
                          saveMaxAge(next);
                        }}
                        onBlur={handleDurationBlur}
                      />
                    )
                  : <Text size={2}>{readOnlyDuration}</Text>}
              </Field>
            </div>
          </div>
        </Card>
      </section>
    </div>
  );
}
