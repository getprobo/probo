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

import { fromMaxAgeSeconds, toMaxAgeSeconds } from "@probo/helpers";
import { Badge, Button, DurationInput, Input, Td, Toggle, Tr } from "@probo/ui";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EditCookieRowFragment$key } from "#/__generated__/core/EditCookieRowFragment.graphql";

import type { CookieEntry } from "./CategorySection";

export const editCookieRowFragment = graphql`
  fragment EditCookieRowFragment on TrackerPattern {
    displayName
    trackerType
    maxAgeSeconds
    description
    excluded
  }
`;

interface CookieFormValues {
  duration: { value: string; unit: string };
  description: string;
  excluded: boolean;
}

interface EditCookieRowProps {
  cookieKey: EditCookieRowFragment$key;
  isUpdating: boolean;
  onSave: (cookie: CookieEntry) => void;
  onCancel: () => void;
}

export function EditCookieRow({
  cookieKey,
  isUpdating,
  onSave,
  onCancel,
}: EditCookieRowProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const cookie = useFragment(editCookieRowFragment, cookieKey);
  const initial = fromMaxAgeSeconds(cookie.maxAgeSeconds ?? null);

  const { register, handleSubmit, control } = useForm<CookieFormValues>({
    defaultValues: {
      duration: initial,
      description: cookie.description,
      excluded: cookie.excluded,
    },
  });

  const onSubmit = (data: CookieFormValues) => {
    onSave({
      name: cookie.displayName,
      maxAgeSeconds: toMaxAgeSeconds(data.duration.value, data.duration.unit),
      description: data.description,
      excluded: data.excluded,
    });
  };

  const trackerTypeBadges = {
    COOKIE: { variant: "warning" as const, label: t("editCookieRow.trackerTypes.cookie") },
    LOCAL_STORAGE: { variant: "info" as const, label: t("editCookieRow.trackerTypes.localStorage") },
    SESSION_STORAGE: { variant: "highlight" as const, label: t("editCookieRow.trackerTypes.sessionStorage") },
    INDEXED_DB: { variant: "success" as const, label: t("editCookieRow.trackerTypes.indexedDb") },
    CACHE_STORAGE: { variant: "outline" as const, label: t("editCookieRow.trackerTypes.cacheStorage") },
  };
  const typeBadge = trackerTypeBadges[cookie.trackerType]
    ?? { variant: "neutral" as const, label: cookie.trackerType };

  return (
    <Tr>
      <Td className="pr-3">
        <div className="flex flex-col gap-2 min-w-0 max-w-xs">
          <div className="flex items-center gap-2">
            <Controller
              name="excluded"
              control={control}
              render={({ field }) => (
                <Toggle
                  size="sm"
                  checked={!field.value}
                  onChange={checked => field.onChange(!checked)}
                  title={t("editCookieRow.includeTitle")}
                />
              )}
            />
            <code className="text-sm font-mono">{cookie.displayName}</code>
          </div>
          <Input
            {...register("description")}
            placeholder={t("editCookieRow.fields.descriptionPlaceholder")}
          />
        </div>
      </Td>
      <Td>
        <Badge variant={typeBadge.variant}>{typeBadge.label}</Badge>
      </Td>
      <Td className="pr-3">
        <Controller
          name="duration"
          control={control}
          render={({ field }) => (
            <DurationInput
              value={field.value.value}
              unit={field.value.unit}
              onValueChange={v => field.onChange({ ...field.value, value: v })}
              onUnitChange={u => field.onChange({ ...field.value, unit: u })}
            />
          )}
        />
      </Td>
      <Td>
        <div className="flex items-center gap-2">
          <Button
            onClick={() => void handleSubmit(onSubmit)()}
            disabled={isUpdating}
          >
            {t("editCookieRow.actions.save")}
          </Button>
          <Button
            variant="secondary"
            onClick={onCancel}
          >
            {t("editCookieRow.actions.cancel")}
          </Button>
        </div>
      </Td>
    </Tr>
  );
}
