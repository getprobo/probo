// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { faviconUrl } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Avatar, Badge, Button, Field, IconCrossLargeX, Option, Select } from "@probo/ui";
import { type ComponentProps, Suspense, useState } from "react";
import { type Control, Controller, type FieldValues, type Path } from "react-hook-form";

import { useThirdParties } from "#/hooks/graph/ThirdPartyGraph";

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
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" disabled placeholder="Loading..." />}
      >
        <ThirdPartiesMultiSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
          selectedThirdParties={selectedThirdParties}
        />
      </Suspense>
    </Field>
  );
}

function ThirdPartiesMultiSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<Props<T>, "organizationId" | "control" | "name" | "disabled" | "selectedThirdParties">,
) {
  const { __ } = useTranslate();
  const { name, organizationId, control, selectedThirdParties = [] } = props;
  const thirdParties = useThirdParties(organizationId);
  const [isOpen, setIsOpen] = useState(false);

  const allThirdParties = [...thirdParties];
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
