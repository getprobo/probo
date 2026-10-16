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

import { CaretDownIcon } from "@phosphor-icons/react";
import { type ReactNode, useState } from "react";

import { Button } from "../Button/Button";
import { Text } from "../typography/Text";
import { TextSkeleton } from "../typography/TextSkeleton";

import { Combobox } from "./Combobox";
import { ComboboxEmpty } from "./ComboboxEmpty";
import { ComboboxInput } from "./ComboboxInput";
import { ComboboxInputGroup } from "./ComboboxInputGroup";
import { ComboboxItem } from "./ComboboxItem";
import { ComboboxJoined } from "./ComboboxJoined";
import { ComboboxList } from "./ComboboxList";
import { ComboboxPopup } from "./ComboboxPopup";

export interface ComboboxPagedOption {
  id: string;
  label: string;
  leading?: ReactNode;
}

export interface ComboboxPagedProps<TItem> {
  items: readonly TItem[];
  loading: boolean;
  failed: boolean;
  query: string;
  onQueryChange: (query: string) => void;
  hasNext: boolean;
  onLoadMore: () => void;
  selectedId: string | null;
  selectedLabel: string;
  ariaLabel: string;
  placeholder: string;
  emptyLabel: string;
  errorLabel: string;
  loadMoreLabel: string;
  disabled?: boolean;
  pinned?: ComboboxPagedOption;
  onPinnedSelect?: () => void;
  onSelect: (item: TItem) => void;
  itemId: (item: TItem) => string;
  itemLabel: (item: TItem) => string;
  itemLeading?: (item: TItem) => ReactNode;
}

const skeletonWidths = ["w-28", "w-36", "w-24"] as const;

// A single-select combobox whose options arrive in pages. The caller owns
// the search string and the page request. This component only draws the field.
export function ComboboxPaged<TItem>(props: ComboboxPagedProps<TItem>) {
  const {
    items,
    loading,
    failed,
    query,
    onQueryChange,
    hasNext,
    onLoadMore,
    selectedId,
    selectedLabel,
    ariaLabel,
    placeholder,
    emptyLabel,
    errorLabel,
    loadMoreLabel,
    disabled,
    pinned,
    onPinnedSelect,
    onSelect,
    itemId,
    itemLabel,
    itemLeading,
  } = props;
  const [open, setOpen] = useState(false);
  const showEmpty = !loading && !failed && items.length === 0 && pinned == null;

  return (
    <Combobox
      value={selectedId}
      onValueChange={(value) => {
        if (typeof value !== "string") {
          return;
        }
        if (pinned != null && value === pinned.id) {
          onPinnedSelect?.();
          return;
        }
        const item = items.find(candidate => itemId(candidate) === value);
        if (item != null) {
          onSelect(item);
        }
      }}
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (next) {
          onQueryChange("");
        }
      }}
      inputValue={open ? query : (selectedId == null ? "" : selectedLabel)}
      onInputValueChange={(value, details) => {
        if (!open) {
          return;
        }
        if (details.reason !== "input-change") {
          return;
        }
        onQueryChange(value);
      }}
      disabled={disabled}
      filter={null}
    >
      <ComboboxJoined
        open={open}
        field={(
          <ComboboxInputGroup>
            <ComboboxInput
              placeholder={open ? placeholder : selectedLabel}
              aria-label={ariaLabel}
              disabled={disabled}
            />
            <CaretDownIcon className="size-4 shrink-0 text-sand-11" />
          </ComboboxInputGroup>
        )}
      >
        <ComboboxPopup joined>
          {(pinned != null || items.length > 0) && (
            <ComboboxList>
              {pinned != null && (
                <ComboboxItem
                  key={pinned.id}
                  value={pinned.id}
                >
                  <OptionLabel icon={pinned.leading} label={pinned.label} />
                </ComboboxItem>
              )}
              {items.map(item => (
                <ComboboxItem
                  key={itemId(item)}
                  value={itemId(item)}
                >
                  <OptionLabel icon={itemLeading?.(item)} label={itemLabel(item)} />
                </ComboboxItem>
              ))}
            </ComboboxList>
          )}
          {loading && items.length === 0 && <OptionSkeletons />}
          {failed && (
            <Text size={2} color="faint" className="px-3 py-2">
              {errorLabel}
            </Text>
          )}
          {showEmpty && <ComboboxEmpty>{emptyLabel}</ComboboxEmpty>}
          {hasNext && (
            <div
              onMouseDown={(event) => {
                event.preventDefault();
              }}
            >
              <Button
                variant="ghost"
                color="neutral"
                size={1}
                className="w-full justify-center"
                loading={loading}
                onClick={onLoadMore}
              >
                {loadMoreLabel}
              </Button>
            </div>
          )}
        </ComboboxPopup>
      </ComboboxJoined>
    </Combobox>
  );
}

function OptionSkeletons() {
  return (
    <div className="flex flex-col gap-1 py-1" aria-hidden>
      {skeletonWidths.map(width => (
        <span key={width} className="flex h-8 items-center gap-2 px-3">
          <span className="size-4 shrink-0 animate-pulse rounded-full bg-sand-3" />
          <TextSkeleton size={2} className={width} />
        </span>
      ))}
    </div>
  );
}

function OptionLabel({ icon, label }: { icon?: ReactNode; label: string }) {
  return (
    <span className="flex min-w-0 items-center gap-2">
      {icon != null && (
        <span className="inline-flex size-4 shrink-0 items-center justify-center [&_svg]:size-4">
          {icon}
        </span>
      )}
      <span className="min-w-0 truncate">{label}</span>
    </span>
  );
}
