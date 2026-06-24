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
import { useTranslate } from "@probo/i18n";
import { Avatar, Badge, Button, Field, IconCrossLargeX, Option, Select } from "@probo/ui";
import { type ComponentProps, Suspense, useEffect, useState } from "react";
import { type Control, Controller, type FieldValues, type Path } from "react-hook-form";
import { type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";

import type { ThirdPartiesMultiSelectFieldQuery } from "#/__generated__/core/ThirdPartiesMultiSelectFieldQuery.graphql";

const thirdPartiesQuery = graphql`
  query ThirdPartiesMultiSelectFieldQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        thirdParties(
          first: 100
          orderBy: { direction: ASC, field: NAME }
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
  websiteUrl: string | null | undefined;
};

type Props<T extends FieldValues = FieldValues> = {
  organizationId: string;
  control: Control<T>;
  name: string;
  label?: string;
  error?: string;
  selectedThirdParties?: ThirdParty[];
} & ComponentProps<typeof Field>;

export function ThirdPartiesMultiSelectField<T extends FieldValues = FieldValues>({
  organizationId,
  control,
  selectedThirdParties = [],
  ...props
}: Props<T>) {
  const [queryRef, loadQuery]
    = useQueryLoader<ThirdPartiesMultiSelectFieldQuery>(thirdPartiesQuery);

  useEffect(() => {
    loadQuery({ organizationId }, { fetchPolicy: "network-only" });
  }, [loadQuery, organizationId]);

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
  props: Pick<Props<T>, "control" | "name" | "disabled" | "selectedThirdParties"> & {
    queryRef: PreloadedQuery<ThirdPartiesMultiSelectFieldQuery>;
  },
) {
  const { __ } = useTranslate();
  const { name, control, selectedThirdParties = [] } = props;
  const data = usePreloadedQuery<ThirdPartiesMultiSelectFieldQuery>(thirdPartiesQuery, props.queryRef);
  const thirdParties = data.organization?.thirdParties?.edges.map(edge => edge.node) ?? [];
  const [isOpen, setIsOpen] = useState(false);

  const allThirdParties: ThirdParty[] = [...thirdParties];
  if (props.disabled) {
    selectedThirdParties.forEach((selectedThirdParty) => {
      if (!allThirdParties.find(v => v.id === selectedThirdParty.id)) {
        allThirdParties.push(selectedThirdParty);
      }
    });
  }

  return (
    <>
      <Controller
        control={control}
        name={name as Path<T>}
        render={({ field }) => {
          const selectedThirdPartyIds = (Array.isArray(field.value) ? field.value : []) as string[];

          const selectedThirdParties = allThirdParties.filter(v => selectedThirdPartyIds.includes(v.id));
          const availableThirdParties = allThirdParties.filter(v => !selectedThirdPartyIds.includes(v.id));

          const handleAddThirdParty = (thirdPartyId: string) => {
            const newValue = [...selectedThirdPartyIds, thirdPartyId];
            field.onChange(newValue);
            setIsOpen(false);
          };

          const handleRemoveThirdParty = (thirdPartyId: string) => {
            const newValue = selectedThirdPartyIds.filter((id: string) => id !== thirdPartyId);
            field.onChange(newValue);
          };

          return (
            <div className="space-y-2">
              {availableThirdParties.length > 0 && !props.disabled && (
                <Select
                  disabled={props.disabled}
                  id={name}
                  variant="editor"
                  placeholder={__("Add third parties...")}
                  onValueChange={handleAddThirdParty}
                  key={`${selectedThirdPartyIds.length}-${thirdParties.length}`}
                  className="w-full"
                  value=""
                  open={isOpen}
                  onOpenChange={setIsOpen}
                >
                  {availableThirdParties.map(thirdParty => (
                    <Option key={thirdParty.id} value={thirdParty.id} className="flex gap-2">
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
                    </Option>
                  ))}
                </Select>
              )}

              {selectedThirdParties.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {selectedThirdParties.map(thirdParty => (
                    <Badge key={thirdParty.id} variant="neutral" className="flex items-center gap-2">
                      <Avatar
                        name={thirdParty.name}
                        src={faviconUrl(thirdParty.websiteUrl)}
                        size="s"
                      />
                      <span>{thirdParty.name}</span>
                      {!props.disabled && (
                        <Button
                          variant="tertiary"
                          icon={IconCrossLargeX}
                          onClick={() => handleRemoveThirdParty(thirdParty.id)}
                          className="h-4 w-4 p-0 hover:bg-transparent"
                        />
                      )}
                    </Badge>
                  ))}
                </div>
              )}

              {selectedThirdParties.length === 0 && availableThirdParties.length === 0 && (
                <div className="text-sm text-txt-secondary py-2">
                  {__("No third parties available")}
                </div>
              )}
            </div>
          );
        }}
      />
    </>
  );
}
