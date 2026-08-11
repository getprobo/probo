// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { faviconUrl } from "@probo/helpers";
import { Avatar, Field, Select } from "@probo/ui";
import { type ComponentProps, Suspense, useEffect } from "react";
import { type Control, type FieldValues } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { ThirdPartiesMultiSelectFieldQuery } from "#/__generated__/core/ThirdPartiesMultiSelectFieldQuery.graphql";
import { EntityMultiSelectField } from "#/components/form/EntityMultiSelectField";

const thirdPartiesQuery = graphql`
  query ThirdPartiesMultiSelectFieldQuery(
    $organizationId: ID!
    $filter: ThirdPartyFilter
  ) {
    organization: node(id: $organizationId) {
      ... on Organization {
        thirdParties(
          first: 100
          orderBy: { direction: ASC, field: NAME }
          filter: $filter
        ) {
          edges {
            node {
              id
              name
              websiteUrl
            }
          }
        }
      }
    }
  }
`;

type ThirdParty = {
  id: string;
  name: string;
  websiteUrl?: string | null;
};

type Props<T extends FieldValues = FieldValues> = {
  organizationId: string;
  control: Control<T>;
  name: string;
  label?: string;
  error?: string;
  selectedThirdParties?: ThirdParty[];
  /**
   * Hierarchy level filter. Defaults to 1 (top-level third parties).
   * Pass null to include nested sub-third-parties as well.
   */
  level?: number | null;
} & ComponentProps<typeof Field>;

export function ThirdPartiesMultiSelectField<
  T extends FieldValues = FieldValues,
>({
  organizationId,
  control,
  selectedThirdParties = [],
  level = 1,
  ...props
}: Props<T>) {
  const [queryRef, loadQuery]
    = useQueryLoader<ThirdPartiesMultiSelectFieldQuery>(thirdPartiesQuery);

  useEffect(() => {
    loadQuery(
      {
        organizationId,
        filter: level === null ? null : { level },
      },
      { fetchPolicy: "network-only" },
    );
  }, [loadQuery, organizationId, level]);

  const loadingState = (
    <Select variant="editor" disabled placeholder="Loading..." />
  );

  return (
    <Field {...props}>
      {queryRef
        ? (
            <Suspense fallback={loadingState}>
              <ThirdPartiesMultiSelectWithQuery
                queryRef={queryRef}
                control={control}
                name={props.name}
                disabled={props.disabled}
                selectedThirdParties={selectedThirdParties}
              />
            </Suspense>
          )
        : (
            loadingState
          )}
    </Field>
  );
}

function ThirdPartiesMultiSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<
    Props<T>,
    "control" | "name" | "disabled" | "selectedThirdParties"
  > & {
    queryRef: PreloadedQuery<ThirdPartiesMultiSelectFieldQuery>;
  },
) {
  const { t } = useTranslation();
  const { name, control, selectedThirdParties = [], disabled } = props;
  const data = usePreloadedQuery<ThirdPartiesMultiSelectFieldQuery>(
    thirdPartiesQuery,
    props.queryRef,
  );
  const thirdParties
    = data.organization?.thirdParties?.edges.map(edge => edge.node) ?? [];

  return (
    <EntityMultiSelectField
      control={control}
      name={name}
      disabled={disabled}
      items={thirdParties}
      selectedItems={selectedThirdParties}
      placeholder={t("thirdPartiesMultiSelectField.addPlaceholder")}
      emptyLabel={t("thirdPartiesMultiSelectField.empty")}
      getRemoveAriaLabel={thirdParty =>
        t("thirdPartiesMultiSelectField.remove", { name: thirdParty.name })}
      renderOption={thirdParty => (
        <>
          <Avatar
            name={thirdParty.name}
            src={faviconUrl(thirdParty.websiteUrl)}
            size="s"
          />
          <div className="flex flex-col">
            <span>{thirdParty.name}</span>
            {thirdParty.websiteUrl && (
              <span className="text-xs text-txt-secondary">
                {thirdParty.websiteUrl}
              </span>
            )}
          </div>
        </>
      )}
      renderBadgeLabel={thirdParty => (
        <>
          <Avatar
            name={thirdParty.name}
            src={faviconUrl(thirdParty.websiteUrl)}
            size="s"
          />
          <span>{thirdParty.name}</span>
        </>
      )}
    />
  );
}
