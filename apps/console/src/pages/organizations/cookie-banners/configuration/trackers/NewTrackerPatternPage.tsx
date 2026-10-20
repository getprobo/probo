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

import { Form } from "@base-ui/react/form";
import { formatError, fromMaxAgeSeconds, toFieldErrors, toMaxAgeSeconds } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation, useNavigate, useParams } from "react-router";

import type { NewTrackerPatternPageMutation } from "#/__generated__/core/NewTrackerPatternPageMutation.graphql";
import type { NewTrackerPatternPageQuery } from "#/__generated__/core/NewTrackerPatternPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { CookieBannerPageHeader } from "../../_components/CookieBannerPageHeader";
import { cookieBannerTrackersPath } from "../../_lib/cookieBannerPaths";
import { cookieBannerPage, cookieBannerSettingsSection } from "../../variants";

import { TrackerMaxAgeField } from "./_components/TrackerMaxAgeField";
import { trackerTypeBadges } from "./_lib/trackerBadges";

const PATTERN_MAX_LENGTH = 255;
const DESCRIPTION_MAX_LENGTH = 1000;

const trackerTypes = Object.keys(trackerTypeBadges) as Array<keyof typeof trackerTypeBadges>;

const serverFields: Record<string, string> = {
  cookie_category_id: "cookieCategoryId",
  tracker_type: "trackerType",
  pattern: "pattern",
  match_type: "pattern",
  display_name: "pattern",
  description: "description",
};

function starCount(pattern: string): number {
  return pattern.split("*").length - 1;
}

function mapFieldErrors(fieldErrors: Record<string, string>): Record<string, string> {
  return Object.fromEntries(
    Object.entries(fieldErrors).map(([field, message]) => [
      serverFields[field] ?? field,
      message,
    ]),
  );
}

export const newTrackerPatternPageQuery = graphql`
  query NewTrackerPatternPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        categories(first: 50, orderBy: { field: RANK, direction: ASC })
          @required(action: THROW) {
          edges {
            node {
              id
              name
            }
          }
        }
      }
    }
  }
`;

const createTrackerPatternMutation = graphql`
  mutation NewTrackerPatternPageMutation($input: CreateTrackerPatternInput!) {
    createTrackerPattern(input: $input) {
      trackerPatternEdge {
        node {
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

interface NewTrackerPatternPageProps {
  queryRef: PreloadedQuery<NewTrackerPatternPageQuery>;
}

export function NewTrackerPatternPage({ queryRef }: NewTrackerPatternPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const toast = useToast();
  const navigate = useNavigate();
  const location = useLocation();
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const { card, fields, field, pair, actions } = cookieBannerSettingsSection();
  const [pattern, setPattern] = useState("");
  const [trackerType, setTrackerType] = useState<keyof typeof trackerTypeBadges>("COOKIE");
  const [cookieCategoryId, setCookieCategoryId] = useState<string | null>(null);
  const [description, setDescription] = useState("");
  const [duration, setDuration] = useState(() => fromMaxAgeSeconds(null));
  const [errors, setErrors] = useState<Record<string, string>>({});
  const createError = t("newTrackerPatternPage.errors.create");
  const [createTrackerPattern, isCreating] = useMutation<NewTrackerPatternPageMutation>(
    createTrackerPatternMutation,
    {
      successMessage: t("newTrackerPatternPage.messages.created"),
      errorToast: false,
    },
  );

  usePageTitle(t("newTrackerPatternPage.pageTitle"));

  const data = usePreloadedQuery<NewTrackerPatternPageQuery>(
    newTrackerPatternPageQuery,
    queryRef,
  );

  if (cookieBannerId == null) {
    throw new Error("Missing cookieBannerId parameter");
  }
  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const bannerId = cookieBannerId;

  const categories = data.node.categories.edges.map(edge => edge.node);

  function clearFieldError(name: string) {
    setErrors((current) => {
      if (current[name] == null) {
        return current;
      }
      const next = { ...current };
      delete next[name];
      return next;
    });
  }

  function handleSubmit() {
    const trimmedPattern = pattern.trim();
    if (starCount(trimmedPattern) > 1) {
      setErrors(current => ({
        ...current,
        pattern: t("newTrackerPatternPage.errors.patternStars"),
      }));
      return;
    }
    if (cookieCategoryId == null) {
      setErrors(current => ({
        ...current,
        cookieCategoryId: t("newTrackerPatternPage.errors.category"),
      }));
      return;
    }

    const maxAgeSeconds = toMaxAgeSeconds(duration.value, duration.unit);
    if (duration.value.trim() !== "" && maxAgeSeconds == null) {
      setErrors(current => ({
        ...current,
        maxAge: t("newTrackerPatternPage.errors.maxAge"),
      }));
      return;
    }
    const trimmedDescription = description.trim();

    void createTrackerPattern({
      variables: {
        input: {
          cookieCategoryId,
          trackerType,
          pattern: trimmedPattern,
          matchType: starCount(trimmedPattern) === 1 ? "GLOB" : "EXACT",
          displayName: trimmedPattern,
          description: trimmedDescription === "" ? undefined : trimmedDescription,
          ...(maxAgeSeconds == null ? {} : { maxAgeSeconds }),
        },
      },
      onCompleted(_response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(mapFieldErrors(fieldErrors));
          return;
        }
        if (payloadErrors != null && payloadErrors.length > 0) {
          toast.add({
            title: t("newTrackerPatternPage.errors.title"),
            description: formatError(createError, payloadErrors),
            type: "error",
          });
        }
      },
      onError(error) {
        toast.add({
          title: t("newTrackerPatternPage.errors.title"),
          description: formatError(createError, error),
          type: "error",
        });
      },
    }).then((response) => {
      const trackerPatternId = response.createTrackerPattern.trackerPatternEdge.node.id;
      void navigate({
        pathname: `${cookieBannerTrackersPath(organizationId, bannerId)}/${trackerPatternId}`,
        search: location.search,
      });
    }).catch(() => undefined);
  }

  return (
    <div className={cookieBannerPage()}>
      <CookieBannerPageHeader
        title={t("newTrackerPatternPage.title")}
        description={t("newTrackerPatternPage.description")}
      />
      <Card size={2} variant="soft" className={card()}>
        <Form className={fields()} errors={errors} onFormSubmit={handleSubmit}>
          <div className={field()}>
            <Field
              required
              label={t("newTrackerPatternPage.fields.pattern")}
              error={errors.pattern}
            >
              <TextField
                name="pattern"
                required
                maxLength={PATTERN_MAX_LENGTH}
                value={pattern}
                disabled={isCreating}
                placeholder={t("newTrackerPatternPage.fields.patternPlaceholder")}
                onValueChange={(value) => {
                  setPattern(value);
                  clearFieldError("pattern");
                }}
              />
            </Field>
            <Text size={1} color="faint">
              {t("newTrackerPatternPage.fields.patternHelp")}
            </Text>
          </div>
          <div className={pair()}>
            <Field
              required
              label={t("newTrackerPatternPage.fields.trackerType")}
              error={errors.trackerType}
            >
              <Select
                name="trackerType"
                value={trackerType}
                disabled={isCreating}
                onValueChange={(value: string | null) => {
                  if (value != null && value in trackerTypeBadges) {
                    setTrackerType(value as keyof typeof trackerTypeBadges);
                    clearFieldError("trackerType");
                  }
                }}
              >
                <SelectTrigger size={2}>
                  {(value: string | null) => (
                    value != null && value in trackerTypeBadges
                      ? t(`trackerPatternRow.types.${trackerTypeBadges[value as keyof typeof trackerTypeBadges].labelKey}`)
                      : t(`trackerPatternRow.types.${trackerTypeBadges.COOKIE.labelKey}`)
                  )}
                </SelectTrigger>
                <SelectPopup>
                  {trackerTypes.map(type => (
                    <SelectItem key={type} value={type}>
                      {t(`trackerPatternRow.types.${trackerTypeBadges[type].labelKey}`)}
                    </SelectItem>
                  ))}
                </SelectPopup>
              </Select>
            </Field>
            <Field
              required
              label={t("newTrackerPatternPage.fields.category")}
              error={errors.cookieCategoryId}
            >
              <Select
                name="cookieCategoryId"
                value={cookieCategoryId}
                disabled={isCreating}
                onValueChange={(value: string | null) => {
                  setCookieCategoryId(value);
                  clearFieldError("cookieCategoryId");
                }}
              >
                <SelectTrigger
                  size={2}
                  placeholder={t("newTrackerPatternPage.fields.categoryPlaceholder")}
                >
                  {(value: string | null) => (
                    value != null
                      ? categories.find(category => category.id === value)?.name
                      ?? t("newTrackerPatternPage.fields.categoryPlaceholder")
                      : t("newTrackerPatternPage.fields.categoryPlaceholder")
                  )}
                </SelectTrigger>
                <SelectPopup>
                  {categories.length === 0
                    ? (
                        <SelectItem value={null} disabled>
                          {t("moveToCategorySelect.empty")}
                        </SelectItem>
                      )
                    : categories.map(category => (
                        <SelectItem key={category.id} value={category.id}>
                          {category.name}
                        </SelectItem>
                      ))}
                </SelectPopup>
              </Select>
            </Field>
          </div>
          <Field
            label={t("newTrackerPatternPage.fields.description")}
            error={errors.description}
          >
            <Textarea
              name="description"
              rows={3}
              maxLength={DESCRIPTION_MAX_LENGTH}
              value={description}
              disabled={isCreating}
              placeholder={t("newTrackerPatternPage.fields.descriptionPlaceholder")}
              onChange={(event) => {
                setDescription(event.target.value);
                clearFieldError("description");
              }}
            />
          </Field>
          <Field
            label={t("trackerProperties.properties.maxAge")}
            error={errors.maxAge}
          >
            <TrackerMaxAgeField
              value={duration.value}
              unit={duration.unit}
              disabled={isCreating}
              onValueChange={(value) => {
                setDuration(current => ({ ...current, value }));
                clearFieldError("maxAge");
              }}
              onUnitChange={(unit) => {
                setDuration(current => ({ ...current, unit }));
                clearFieldError("maxAge");
              }}
            />
          </Field>
          <div className={actions({ className: "items-center gap-4" })}>
            <Link
              to={{ pathname: "..", search: location.search }}
              size={2}
              color="neutral"
              underline={false}
            >
              {t("newTrackerPatternPage.actions.cancel")}
            </Link>
            <Button
              type="submit"
              variant="solid"
              color="neutral"
              highContrast
              loading={isCreating}
            >
              {t("newTrackerPatternPage.actions.create")}
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  );
}
