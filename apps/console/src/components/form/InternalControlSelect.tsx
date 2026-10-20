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

import { Combobox as BaseCombobox } from "@base-ui/react/combobox";
import { CaretDownIcon, CheckIcon, MagnifyingGlassIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Combobox } from "@probo/ui/src/v2/Combobox/Combobox";
import { ComboboxEmpty } from "@probo/ui/src/v2/Combobox/ComboboxEmpty";
import { ComboboxItem } from "@probo/ui/src/v2/Combobox/ComboboxItem";
import { ComboboxList } from "@probo/ui/src/v2/Combobox/ComboboxList";
import { comboboxInput, comboboxItem, comboboxPopup } from "@probo/ui/src/v2/Combobox/variants";
import { selectTrigger } from "@probo/ui/src/v2/Select/variants";
import {
  Suspense,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useState,
  useTransition,
} from "react";
import { useTranslation } from "react-i18next";
import { useDebounceCallback } from "usehooks-ts";

import { usePaginatedInternalControls } from "#/hooks/graph/usePaginatedInternalControls";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const pageSize = 100;

const internalControlOrder = { field: "NAME", direction: "ASC" } as const;

export interface InternalControlOption {
  id: string;
  name: string;
}

interface InternalControlSelectBaseProps {
  placeholder: string;
  searchPlaceholder: string;
  emptyLabel: string;
  ariaLabel: string;
  disabled?: boolean;
  size?: 1 | 2;
  variant?: "classic" | "surface" | "soft" | "ghost";
  className?: string;
}

interface InternalControlSelectSingleProps extends InternalControlSelectBaseProps {
  multiple?: false;
  value: InternalControlOption | null;
  onValueChange: (value: InternalControlOption | null) => void;
  allowClear?: boolean;
}

interface InternalControlSelectMultipleProps extends InternalControlSelectBaseProps {
  multiple: true;
  value: readonly InternalControlOption[];
  onValueChange: (value: InternalControlOption[]) => void;
  allowClear?: false;
}

export type InternalControlSelectProps
  = | InternalControlSelectSingleProps
    | InternalControlSelectMultipleProps;

function sameOptions(
  left: readonly InternalControlOption[],
  right: readonly InternalControlOption[],
) {
  return left.length === right.length
    && left.every((option, index) =>
      option.id === right[index]?.id && option.name === right[index]?.name,
    );
}

function mergeOptions(
  selected: readonly InternalControlOption[],
  results: readonly InternalControlOption[],
) {
  const seen = new Set<string>();
  const merged: InternalControlOption[] = [];
  for (const option of [...selected, ...results]) {
    if (seen.has(option.id)) {
      continue;
    }
    seen.add(option.id);
    merged.push(option);
  }
  return merged;
}

function OptionsFallback() {
  return (
    <div className="flex flex-col gap-1 px-1 py-1">
      <span className="h-8 animate-pulse rounded-2 bg-sand-3" />
      <span className="h-8 animate-pulse rounded-2 bg-sand-3" />
      <span className="h-8 animate-pulse rounded-2 bg-sand-3" />
    </div>
  );
}

function InternalControlOptions({
  query,
  onResults,
}: {
  query: string;
  onResults: (options: InternalControlOption[]) => void;
}) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const [, startTransition] = useTransition();
  const [mountedQuery] = useState(query);
  const { data, hasNext, isLoadingNext, loadNext, refetch } = usePaginatedInternalControls(
    organizationId,
    {
      first: pageSize,
      order: internalControlOrder,
      filter: { query: mountedQuery.trim() || null },
    },
  );
  const results = useMemo(
    () => data.internalControls?.edges?.flatMap(edge => edge?.node ?? []) ?? [],
    [data.internalControls?.edges],
  );
  const refetchSearch = useDebounceCallback(
    useCallback((next: string) => {
      startTransition(() => {
        refetch(
          {
            first: pageSize,
            order: internalControlOrder,
            filter: { query: next.trim() || null },
          },
          { fetchPolicy: "network-only" },
        );
      });
    }, [refetch]),
    300,
  );
  useLayoutEffect(() => {
    onResults(results);
  }, [onResults, results]);

  useEffect(() => {
    if (query === mountedQuery) {
      return;
    }
    refetchSearch(query);
  }, [mountedQuery, query, refetchSearch]);

  useEffect(() => () => {
    refetchSearch.cancel();
  }, [refetchSearch]);

  if (!hasNext) {
    return null;
  }

  return (
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
        loading={isLoadingNext}
        onClick={() => {
          if (isLoadingNext) {
            return;
          }

          loadNext(pageSize);
        }}
      >
        {t("internalControlsPage.actions.loadMore")}
      </Button>
    </div>
  );
}

function SelectTriggerButton({
  labelText,
  lines,
  placeholder,
  ariaLabel,
  disabled,
  size,
  variant,
  className,
}: {
  labelText: string | null;
  lines?: readonly InternalControlOption[];
  placeholder: string;
  ariaLabel: string;
  disabled?: boolean;
  size?: 1 | 2;
  variant?: "classic" | "surface" | "soft" | "ghost";
  className?: string;
}) {
  const stacked = lines != null && lines.length > 0;
  const { trigger, value, icon } = selectTrigger({
    size,
    variant,
    layout: stacked ? "stack" : "single",
  });

  return (
    <BaseCombobox.Trigger
      className={trigger({ className })}
      aria-label={ariaLabel}
      disabled={disabled}
      title={stacked ? lines.map(line => line.name).join("\n") : labelText ?? undefined}
    >
      {stacked
        ? (
            <span className={value()}>
              {lines.map(line => (
                <span key={line.id}>{line.name}</span>
              ))}
            </span>
          )
        : (
            <span className={value({ className: labelText == null ? "text-sand-a10" : undefined })}>
              {labelText ?? placeholder}
            </span>
          )}
      <span className={icon()}>
        <CaretDownIcon />
      </span>
    </BaseCombobox.Trigger>
  );
}

export function InternalControlSelect(props: InternalControlSelectProps) {
  if (props.multiple) {
    return <MultipleInternalControlSelect {...props} />;
  }

  return <SingleInternalControlSelect {...props} />;
}

function useInternalControlSearch(selected: readonly InternalControlOption[]) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<InternalControlOption[] | null>(null);
  const handleResults = useCallback((next: InternalControlOption[]) => {
    setResults(current => current != null && sameOptions(current, next) ? current : next);
  }, []);
  const items = useMemo(
    () => results == null ? [...selected] : mergeOptions(selected, results),
    [results, selected],
  );

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setQuery("");
      setResults(null);
    }
  }

  return {
    open,
    query,
    setQuery,
    items,
    loaded: results != null,
    handleResults,
    handleOpenChange,
  };
}

function SingleInternalControlSelect({
  value,
  onValueChange,
  placeholder,
  searchPlaceholder,
  emptyLabel,
  ariaLabel,
  disabled,
  size = 2,
  variant = "surface",
  allowClear = false,
  className,
}: InternalControlSelectSingleProps) {
  const selected = useMemo(() => value == null ? [] : [value], [value]);
  const search = useInternalControlSearch(selected);

  return (
    <Combobox<InternalControlOption>
      items={search.items}
      value={value}
      disabled={disabled}
      open={search.open}
      onOpenChange={search.handleOpenChange}
      inputValue={search.query}
      filter={null}
      itemToStringLabel={option => option.name}
      isItemEqualToValue={(left, right) => left.id === right.id}
      onInputValueChange={(input, details) => {
        if (details.reason !== "input-change" && details.reason !== "input-clear") {
          return;
        }
        search.setQuery(input);
      }}
      onValueChange={(next) => {
        if (next?.id === value?.id) {
          return;
        }
        onValueChange(next);
      }}
    >
      <SelectTriggerButton
        labelText={value?.name ?? null}
        placeholder={placeholder}
        ariaLabel={ariaLabel}
        disabled={disabled}
        size={size}
        variant={variant}
        className={className}
      />
      <InternalControlSelectPopup
        open={search.open}
        query={search.query}
        selected={selected}
        items={search.items}
        loaded={search.loaded}
        onResults={search.handleResults}
        searchPlaceholder={searchPlaceholder}
        emptyLabel={emptyLabel}
        allowClear={allowClear}
        clearLabel={placeholder}
        onClear={() => {
          onValueChange(null);
          search.handleOpenChange(false);
        }}
      />
    </Combobox>
  );
}

function MultipleInternalControlSelect({
  value,
  onValueChange,
  placeholder,
  searchPlaceholder,
  emptyLabel,
  ariaLabel,
  disabled,
  size = 1,
  variant = "surface",
  className,
}: InternalControlSelectMultipleProps) {
  const selectedKey = value.map(option => `${option.id}:${option.name}`).join("\n");
  const selected = useMemo(
    () => [...value],
    // The joined key tracks id and name changes without retriggering on a new array.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [selectedKey],
  );
  const search = useInternalControlSearch(selected);

  return (
    <Combobox<InternalControlOption, true>
      multiple
      items={search.items}
      value={selected}
      disabled={disabled}
      open={search.open}
      onOpenChange={search.handleOpenChange}
      inputValue={search.query}
      filter={null}
      itemToStringLabel={option => option.name}
      isItemEqualToValue={(left, right) => left.id === right.id}
      onInputValueChange={(input, details) => {
        if (details.reason !== "input-change" && details.reason !== "input-clear") {
          return;
        }
        search.setQuery(input);
      }}
      onValueChange={(next) => {
        const nextValue = next ?? [];
        if (sameOptions(nextValue, selected)) {
          return;
        }
        onValueChange(nextValue);
      }}
    >
      <SelectTriggerButton
        labelText={null}
        lines={selected}
        placeholder={placeholder}
        ariaLabel={ariaLabel}
        disabled={disabled}
        size={size}
        variant={variant}
        className={className}
      />
      <InternalControlSelectPopup
        open={search.open}
        query={search.query}
        selected={selected}
        items={search.items}
        loaded={search.loaded}
        onResults={search.handleResults}
        searchPlaceholder={searchPlaceholder}
        emptyLabel={emptyLabel}
        allowClear={false}
        clearLabel={placeholder}
        onClear={() => {}}
      />
    </Combobox>
  );
}

function InternalControlSelectPopup({
  open,
  query,
  selected,
  items,
  loaded,
  onResults,
  searchPlaceholder,
  emptyLabel,
  allowClear,
  clearLabel,
  onClear,
}: {
  open: boolean;
  query: string;
  selected: readonly InternalControlOption[];
  items: readonly InternalControlOption[];
  loaded: boolean;
  onResults: (options: InternalControlOption[]) => void;
  searchPlaceholder: string;
  emptyLabel: string;
  allowClear: boolean;
  clearLabel: string;
  onClear: () => void;
}) {
  const { item, label, indicator } = comboboxItem();
  const showEmpty = loaded && items.length === 0;

  return (
    <BaseCombobox.Portal>
      <BaseCombobox.Positioner
        className="z-3"
        side="bottom"
        align="start"
        sideOffset={4}
        collisionAvoidance={{ align: "flip", fallbackAxisSide: "none" }}
      >
        <BaseCombobox.Popup className={comboboxPopup({ className: "w-72!" })}>
          <div className="flex items-center gap-1.5 px-2 py-1">
            <MagnifyingGlassIcon className="size-3.5 shrink-0 text-sand-a10" />
            <BaseCombobox.Input
              className={comboboxInput()}
              placeholder={searchPlaceholder}
              aria-label={searchPlaceholder}
            />
          </div>
          {allowClear && (
            <button type="button" className={item()} onClick={onClear}>
              <span className={label()}>{clearLabel}</span>
              {selected.length === 0 && (
                <CheckIcon className={indicator()} />
              )}
            </button>
          )}
          {showEmpty && <ComboboxEmpty>{emptyLabel}</ComboboxEmpty>}
          {items.length > 0 && (
            <ComboboxList>
              {(option: InternalControlOption) => (
                <ComboboxItem key={option.id} value={option}>
                  {option.name}
                </ComboboxItem>
              )}
            </ComboboxList>
          )}
          {open && (
            <Suspense fallback={!loaded && items.length === 0 ? <OptionsFallback /> : null}>
              <InternalControlOptions query={query} onResults={onResults} />
            </Suspense>
          )}
        </BaseCombobox.Popup>
      </BaseCombobox.Positioner>
    </BaseCombobox.Portal>
  );
}
