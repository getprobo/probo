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

import {
  DotsThreeVerticalIcon,
  ListIcon,
  PlusIcon,
  UploadSimpleIcon,
  WarningIcon,
} from "@phosphor-icons/react";
import { internalControlImplementationStatuses } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { SegmentedControl } from "@probo/ui/src/v2/SegmentedControl/SegmentedControl";
import { SegmentedControlItem } from "@probo/ui/src/v2/SegmentedControl/SegmentedControlItem";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ChangeEventHandler, useCallback, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { InternalControlsPageImportMutation } from "#/__generated__/core/InternalControlsPageImportMutation.graphql";
import type { InternalControlsPageQuery } from "#/__generated__/core/InternalControlsPageQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { InternalControlList } from "./_components/InternalControlList/InternalControlList";
import { InternalControlsListSearch } from "./_components/InternalControlsListSearch";
import { InternalControlsOwnerFilter } from "./_components/InternalControlsOwnerFilter";
import { internalControlsPage } from "./_components/variants";
import {
  type InternalControlImplementationStatus,
  type InternalControlsListView,
  useInternalControlsListFilters,
} from "./_lib/useInternalControlsListFilters";
import InternalControlFormDialog from "./dialog/InternalControlFormDialog";

const allCategories = "ALL";

export const internalControlsPageQuery = graphql`
  query InternalControlsPageQuery(
    $organizationId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
    $order: InternalControlOrder
    $query: String
    $implementationStatuses: [InternalControlImplementationStatus!]
    $category: String
    $ownerId: ID
    $needsAttention: Boolean
  ) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateInternalControl: permission(action: "core:internal-control:create")
        internalControlCategories
        internalControls(
          first: $first
          after: $after
          last: $last
          before: $before
          orderBy: $order
          filter: {
            query: $query
            implementationStatuses: $implementationStatuses
            category: $category
            ownerId: $ownerId
            needsAttention: $needsAttention
          }
        ) {
          totalCount
        }
        ...InternalControlList_organization
          @arguments(
            first: $first
            after: $after
            last: $last
            before: $before
            order: $order
            query: $query
            implementationStatuses: $implementationStatuses
            category: $category
            ownerId: $ownerId
            needsAttention: $needsAttention
          )
      }
    }
  }
`;

const importInternalControlsMutation = graphql`
  mutation InternalControlsPageImportMutation($input: ImportInternalControlInput!) {
    importInternalControl(input: $input) {
      internalControlEdges {
        node {
          id
        }
      }
    }
  }
`;

interface InternalControlsPageProps {
  queryRef: PreloadedQuery<InternalControlsPageQuery>;
}

export function InternalControlsPage({ queryRef }: InternalControlsPageProps) {
  const { t } = useTranslation("organizations/internal-controls");
  const organizationId = useOrganizationId();
  const filters = useInternalControlsListFilters();
  const importFileRef = useRef<HTMLInputElement>(null);
  const refetchRef = useRef<() => void>(() => {});
  const { organization } = usePreloadedQuery<InternalControlsPageQuery>(
    internalControlsPageQuery,
    queryRef,
  );
  const queriedCount = organization.__typename === "Organization"
    ? organization.internalControls?.totalCount ?? 0
    : 0;
  const [totalCount, setTotalCount] = useState(queriedCount);
  const [seenQueryCount, setSeenQueryCount] = useState(queriedCount);
  if (queriedCount !== seenQueryCount) {
    setSeenQueryCount(queriedCount);
    setTotalCount(queriedCount);
  }

  usePageTitle(t("list.title"));

  const [importInternalControls, isImporting]
    = useMutationWithToasts<InternalControlsPageImportMutation>(
      importInternalControlsMutation,
      {
        successMessage: t("list.messages.imported"),
        errorMessage: t("list.errors.import"),
      },
    );

  const handleRefetch = useCallback((refetch: () => void) => {
    refetchRef.current = refetch;
  }, []);

  const handleTotalCount = useCallback((count: number) => {
    setTotalCount(count);
  }, []);

  if (organization.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  const canCreate = organization.canCreateInternalControl ?? false;
  const categories = categoryOptions(
    organization.internalControlCategories,
    filters.category,
  );
  const {
    root,
    header,
    intro,
    title,
    actions,
    results,
    views,
    tools,
    tableFilters,
    filter,
  } = internalControlsPage();

  const handleImport: ChangeEventHandler<HTMLInputElement> = (event) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) {
      return;
    }

    void importInternalControls({
      variables: {
        input: {
          organizationId,
          file: null,
        },
      },
      uploadables: {
        "input.file": file,
      },
      onSuccess() {
        refetchRef.current();
      },
    });
  };

  function openImport() {
    importFileRef.current?.click();
  }

  const importAction = canCreate
    ? (
        <Button
          variant="soft"
          color="neutral"
          iconStart={<UploadSimpleIcon />}
          disabled={isImporting}
          onClick={openImport}
        >
          {t("list.actions.import")}
        </Button>
      )
    : null;

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={intro()}>
          <div className={title()}>
            <Heading level={1} size={6} weight="medium" highContrast>
              {t("list.title")}
            </Heading>
            <Text size={2} color="faint">{totalCount}</Text>
          </div>
          <Text size={2} color="faint">{t("list.description")}</Text>
        </div>
        {canCreate && (
          <div className={actions()}>
            <InternalControlFormDialog onSaved={() => refetchRef.current()}>
              <Button variant="solid" iconStart={<PlusIcon />}>
                {t("list.actions.new")}
              </Button>
            </InternalControlFormDialog>
            <Dropdown>
              <DropdownTrigger
                render={(
                  <IconButton
                    variant="soft"
                    color="neutral"
                    aria-label={t("list.actions.menu")}
                  >
                    <DotsThreeVerticalIcon />
                  </IconButton>
                )}
              />
              <DropdownPopup align="end">
                <DropdownItem
                  iconStart={<UploadSimpleIcon />}
                  disabled={isImporting}
                  onClick={openImport}
                >
                  {t("list.actions.import")}
                </DropdownItem>
              </DropdownPopup>
            </Dropdown>
          </div>
        )}
      </div>

      <input
        ref={importFileRef}
        type="file"
        accept="application/json,.json"
        className="sr-only"
        tabIndex={-1}
        onChange={handleImport}
      />

      <div className={results()}>
        <div className={views()}>
          <SegmentedControl
            size={1}
            value={filters.view}
            aria-label={t("list.views.label")}
            onValueChange={(value) => {
              if (isListView(value)) {
                filters.setView(value);
              }
            }}
          >
            <SegmentedControlItem value="all">
              <ListIcon />
              {t("list.views.all")}
            </SegmentedControlItem>
            <SegmentedControlItem value="attention">
              <WarningIcon />
              {t("list.views.attention")}
            </SegmentedControlItem>
          </SegmentedControl>
        </div>
        <div className={tools()}>
          <InternalControlsListSearch />
          <div className={tableFilters()}>
            <div className={filter()}>
              <Select
                multiple
                value={filters.statuses}
                onValueChange={(value) => {
                  filters.setStatuses(value.filter(isImplementationStatus));
                }}
              >
                <SelectTrigger
                  size={2}
                  placeholder={t("list.filters.allStatuses")}
                  aria-label={t("list.columns.status")}
                >
                  {(value: readonly string[]) => selectedFilterLabel(
                    value,
                    t("list.filters.allStatuses"),
                    status => t(`list.statuses.${status.toLowerCase()}`),
                    t("list.filters.selectedCount", { count: value.length }),
                  )}
                </SelectTrigger>
                <SelectPopup align="end">
                  {internalControlImplementationStatuses.map(status => (
                    <SelectItem key={status} value={status}>
                      {t(`list.statuses.${status.toLowerCase()}`)}
                    </SelectItem>
                  ))}
                </SelectPopup>
              </Select>
            </div>
            <div className={filter()}>
              <Select
                value={filters.category ?? allCategories}
                onValueChange={(value) => {
                  filters.setCategory(value === allCategories || value == null ? null : value);
                }}
              >
                <SelectTrigger
                  size={2}
                  placeholder={t("list.filters.allCategories")}
                  aria-label={t("list.filters.allCategories")}
                />
                <SelectPopup align="end">
                  <SelectItem value={allCategories}>
                    {t("list.filters.allCategories")}
                  </SelectItem>
                  {categories.map(category => (
                    <SelectItem key={category} value={category}>
                      {category}
                    </SelectItem>
                  ))}
                </SelectPopup>
              </Select>
            </div>
            <div className={filter()}>
              <InternalControlsOwnerFilter />
            </div>
          </div>
        </div>
        <InternalControlList
          organizationKey={organization}
          importAction={importAction}
          onRefetch={handleRefetch}
          onTotalCount={handleTotalCount}
        />
      </div>
    </div>
  );
}

function isListView(value: string): value is InternalControlsListView {
  return value === "all" || value === "attention";
}

function isImplementationStatus(value: string): value is InternalControlImplementationStatus {
  return (internalControlImplementationStatuses as readonly string[]).includes(value);
}

function categoryOptions(categories: readonly string[], selected: string | null): string[] {
  if (selected == null || categories.includes(selected)) {
    return [...categories];
  }
  return [selected, ...categories];
}

function selectedFilterLabel(
  values: readonly string[],
  emptyLabel: string,
  label: (value: string) => string,
  countLabel: string,
): string {
  if (values.length === 0) {
    return emptyLabel;
  }
  if (values.length === 1) {
    return label(values[0] ?? "");
  }
  return countLabel;
}
