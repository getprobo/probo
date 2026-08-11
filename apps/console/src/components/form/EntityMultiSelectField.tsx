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

import {
  Badge,
  Button,
  IconCrossLargeX,
  InfiniteScrollTrigger,
  Option,
  Select,
} from "@probo/ui";
import { type ReactNode, useState } from "react";
import {
  type Control,
  Controller,
  type FieldValues,
  type Path,
} from "react-hook-form";

export type EntityMultiSelectItem = {
  id: string;
};

type Props<
  TItem extends EntityMultiSelectItem,
  TForm extends FieldValues = FieldValues,
> = {
  control: Control<TForm>;
  name: Path<TForm> | string;
  disabled?: boolean;
  placeholder: string;
  emptyLabel: string;
  items: TItem[];
  selectedItems?: TItem[];
  renderOption: (item: TItem) => ReactNode;
  renderBadgeLabel: (item: TItem) => ReactNode;
  getRemoveAriaLabel?: (item: TItem) => string;
  hasNext?: boolean;
  isLoadingNext?: boolean;
  onLoadMore?: (count?: number) => void;
};

export function EntityMultiSelectField<
  TItem extends EntityMultiSelectItem,
  TForm extends FieldValues = FieldValues,
>({
  control,
  name,
  disabled,
  placeholder,
  emptyLabel,
  items,
  selectedItems = [],
  renderOption,
  renderBadgeLabel,
  getRemoveAriaLabel,
  hasNext = false,
  isLoadingNext = false,
  onLoadMore,
}: Props<TItem, TForm>) {
  const [isOpen, setIsOpen] = useState(false);

  const allItems: TItem[] = [...items];
  selectedItems.forEach((selectedItem) => {
    if (!allItems.find(item => item.id === selectedItem.id)) {
      allItems.push(selectedItem);
    }
  });

  return (
    <Controller
      control={control}
      name={name as Path<TForm>}
      render={({ field }) => {
        const selectedIds = (Array.isArray(field.value) ? field.value : []) as string[];
        const selected = allItems.filter(item => selectedIds.includes(item.id));
        const available = allItems.filter(item => !selectedIds.includes(item.id));

        const handleAdd = (itemId: string) => {
          field.onChange([...selectedIds, itemId]);
          setIsOpen(false);
        };

        const handleRemove = (itemId: string) => {
          field.onChange(selectedIds.filter(id => id !== itemId));
        };

        return (
          <div className="space-y-2">
            {(available.length > 0 || hasNext) && !disabled && (
              <Select
                disabled={disabled}
                id={String(name)}
                variant="editor"
                placeholder={placeholder}
                onValueChange={handleAdd}
                key={selectedIds.length}
                className="w-full"
                value=""
                open={isOpen}
                onOpenChange={(open) => {
                  setIsOpen(open);
                  if (!open) {
                    field.onBlur();
                  }
                }}
              >
                {available.map(item => (
                  <Option key={item.id} value={item.id} className="flex gap-2">
                    {renderOption(item)}
                  </Option>
                ))}
                {hasNext && onLoadMore && (
                  <InfiniteScrollTrigger
                    loading={isLoadingNext}
                    onView={() => onLoadMore(50)}
                  />
                )}
              </Select>
            )}

            {selected.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {selected.map(item => (
                  <Badge key={item.id} variant="neutral" className="flex items-center gap-2">
                    {renderBadgeLabel(item)}
                    {!disabled && (
                      <Button
                        type="button"
                        variant="tertiary"
                        icon={IconCrossLargeX}
                        onClick={() => handleRemove(item.id)}
                        className="h-4 w-4 p-0 hover:bg-transparent"
                        aria-label={getRemoveAriaLabel?.(item) ?? "Remove"}
                      />
                    )}
                  </Badge>
                ))}
              </div>
            )}

            {selected.length === 0 && available.length === 0 && !hasNext && (
              <div className="text-sm text-txt-secondary py-2">
                {emptyLabel}
              </div>
            )}
          </div>
        );
      }}
    />
  );
}
