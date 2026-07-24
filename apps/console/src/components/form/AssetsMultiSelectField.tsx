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

import { useTranslate } from "@probo/i18n";
import { Badge, Button, Field, IconCrossLargeX, Option, Select } from "@probo/ui";
import { type ComponentProps, Suspense, useEffect, useState } from "react";
import { type Control, Controller, type FieldValues, type Path } from "react-hook-form";
import { type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";

import type { AssetsMultiSelectFieldQuery } from "#/__generated__/core/AssetsMultiSelectFieldQuery.graphql";

const assetsQuery = graphql`
  query AssetsMultiSelectFieldQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        assets(
          first: 100
          orderBy: { direction: ASC, field: NAME }
        ) {
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
  const [queryRef, loadQuery]
    = useQueryLoader<AssetsMultiSelectFieldQuery>(assetsQuery);

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
              <AssetsMultiSelectWithQuery
                queryRef={queryRef}
                control={control}
                name={props.name}
                disabled={props.disabled}
                selectedAssets={selectedAssets}
              />
            </Suspense>
          )
        : (
            loadingState
          )}
    </Field>
  );
}

function AssetsMultiSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<Props<T>, "control" | "name" | "disabled" | "selectedAssets"> & {
    queryRef: PreloadedQuery<AssetsMultiSelectFieldQuery>;
  },
) {
  const { __ } = useTranslate();
  const { name, control, selectedAssets = [] } = props;
  const data = usePreloadedQuery<AssetsMultiSelectFieldQuery>(assetsQuery, props.queryRef);
  const assets = data.organization?.assets?.edges.map(edge => edge.node) ?? [];
  const [isOpen, setIsOpen] = useState(false);

  const allAssets: Asset[] = [...assets];
  selectedAssets.forEach((selectedAsset) => {
    if (!allAssets.find(asset => asset.id === selectedAsset.id)) {
      allAssets.push(selectedAsset);
    }
  });

  return (
    <Controller
      control={control}
      name={name as Path<T>}
      render={({ field }) => {
        const selectedAssetIds = (Array.isArray(field.value) ? field.value : []) as string[];
        const selected = allAssets.filter(asset => selectedAssetIds.includes(asset.id));
        const available = allAssets.filter(asset => !selectedAssetIds.includes(asset.id));

        const handleAddAsset = (assetId: string) => {
          field.onChange([...selectedAssetIds, assetId]);
          setIsOpen(false);
        };

        const handleRemoveAsset = (assetId: string) => {
          field.onChange(selectedAssetIds.filter(id => id !== assetId));
        };

        return (
          <div className="space-y-2">
            {available.length > 0 && !props.disabled && (
              <Select
                disabled={props.disabled}
                id={name}
                variant="editor"
                placeholder={__("Add assets...")}
                onValueChange={handleAddAsset}
                key={`${selectedAssetIds.length}-${assets.length}`}
                className="w-full"
                value=""
                open={isOpen}
                onOpenChange={setIsOpen}
              >
                {available.map(asset => (
                  <Option key={asset.id} value={asset.id}>
                    {asset.name}
                  </Option>
                ))}
              </Select>
            )}

            {selected.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {selected.map(asset => (
                  <Badge key={asset.id} variant="neutral" className="flex items-center gap-2">
                    <span>{asset.name}</span>
                    {!props.disabled && (
                      <Button
                        type="button"
                        variant="tertiary"
                        icon={IconCrossLargeX}
                        onClick={() => handleRemoveAsset(asset.id)}
                        className="h-4 w-4 p-0 hover:bg-transparent"
                      />
                    )}
                  </Badge>
                ))}
              </div>
            )}

            {selected.length === 0 && available.length === 0 && (
              <div className="text-sm text-txt-secondary py-2">
                {__("No assets available")}
              </div>
            )}
          </div>
        );
      }}
    />
  );
}
