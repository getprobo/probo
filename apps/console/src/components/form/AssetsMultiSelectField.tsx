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

import { Field, Select } from "@probo/ui";
import { type ComponentProps, Suspense } from "react";
import { type Control, type FieldValues } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { EntityMultiSelectField } from "#/components/form/EntityMultiSelectField";
import { usePaginatedAssets } from "#/hooks/graph/usePaginatedAssets";

type Asset = {
  id: string;
  name: string;
};

type Props<T extends FieldValues = FieldValues> = {
  organizationId: string;
  control: Control<T>;
  name: string;
  label?: string;
  error?: string;
  selectedAssets?: Asset[];
} & ComponentProps<typeof Field>;

export function AssetsMultiSelectField<T extends FieldValues = FieldValues>({
  organizationId,
  control,
  selectedAssets = [],
  ...props
}: Props<T>) {
  const { t } = useTranslation();

  return (
    <Field {...props}>
      <Suspense fallback={<Select variant="editor" disabled placeholder={t("ui.loading")} />}>
        <AssetsMultiSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
          selectedAssets={selectedAssets}
        />
      </Suspense>
    </Field>
  );
}

function AssetsMultiSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<Props<T>, "organizationId" | "control" | "name" | "disabled" | "selectedAssets">,
) {
  const { t } = useTranslation();
  const { name, control, organizationId, selectedAssets = [], disabled } = props;
  const { data, loadNext, hasNext, isLoadingNext } = usePaginatedAssets(organizationId);
  const assets = data?.assets?.edges?.map(edge => edge.node) ?? [];

  return (
    <EntityMultiSelectField
      control={control}
      name={name}
      disabled={disabled}
      items={assets}
      selectedItems={selectedAssets}
      placeholder={t("assetsMultiSelectField.addPlaceholder")}
      emptyLabel={t("assetsMultiSelectField.empty")}
      getRemoveAriaLabel={asset =>
        t("assetsMultiSelectField.remove", { name: asset.name })}
      hasNext={hasNext}
      isLoadingNext={isLoadingNext}
      onLoadMore={count => loadNext(count ?? 50)}
      renderOption={asset => asset.name}
      renderBadgeLabel={asset => <span>{asset.name}</span>}
    />
  );
}
