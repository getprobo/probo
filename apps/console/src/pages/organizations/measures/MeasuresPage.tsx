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
  formatError,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import {
  ActionDropdown,
  Button,
  Card,
  DropdownItem,
  FileButton,
  IconFolderUpload,
  IconMagnifyingGlass,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  Input,
  Option,
  PageHeader,
  Select,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { MeasureBadge } from "@probo/ui/src/Molecules/Badge/MeasureBadge";
import { type ChangeEventHandler, useEffect, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useFragment,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useSearchParams } from "react-router";

import type { MeasuresPageDeleteMutation } from "#/__generated__/core/MeasuresPageDeleteMutation.graphql";
import type { MeasuresPageFragment$key } from "#/__generated__/core/MeasuresPageFragment.graphql";
import type { MeasuresPageImportMutation } from "#/__generated__/core/MeasuresPageImportMutation.graphql";
import type { MeasuresPageListQuery } from "#/__generated__/core/MeasuresPageListQuery.graphql";
import type {
  MeasuresPageRefetchQuery,
  MeasureState,
} from "#/__generated__/core/MeasuresPageRefetchQuery.graphql";
import type { MeasuresPageRowFragment$key } from "#/__generated__/core/MeasuresPageRowFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import MeasureFormDialog from "./dialog/MeasureFormDialog";

export const MeasuresConnectionKey = "MeasuresPage_measures";

export const measuresPageQuery = graphql`
  query MeasuresPageListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateMeasure: permission(action: "core:measure:create")
        measureCategories
        ...MeasuresPageFragment
      }
    }
  }
`;

const measureRowFragment = graphql`
  fragment MeasuresPageRowFragment on Measure {
    id
    name
    category
    state
    canUpdate: permission(action: "core:measure:update")
    canDelete: permission(action: "core:measure:delete")
    ...MeasureFormDialogMeasureFragment
  }
`;

const deleteMeasureMutation = graphql`
  mutation MeasuresPageDeleteMutation(
    $input: DeleteMeasureInput!
    $connections: [ID!]!
  ) {
    deleteMeasure(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

const measuresPageFragment = graphql`
  fragment MeasuresPageFragment on Organization
  @refetchable(queryName: "MeasuresPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    after: { type: "CursorKey" }
    query: { type: "String", defaultValue: null }
    state: { type: "MeasureState", defaultValue: null }
    category: { type: "String", defaultValue: null }
  ) {
    id
    measures(
      first: $first
      after: $after
      filter: { query: $query, state: $state, category: $category }
    )
      @connection(
        key: "MeasuresPage_measures"
        filters: ["filter"]
      ) {
      edges {
        node {
          id
          canUpdate: permission(action: "core:measure:update")
          canDelete: permission(action: "core:measure:delete")
          ...MeasuresPageRowFragment
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

const importMeasuresMutation = graphql`
  mutation MeasuresPageImportMutation(
    $input: ImportMeasureInput!
    $connections: [ID!]!
  ) {
    importMeasure(input: $input) {
      measureEdges @appendEdge(connections: $connections) {
        node {
          id
          name
          category
          state
        }
      }
    }
  }
`;

interface MeasuresPageProps {
  queryRef: PreloadedQuery<MeasuresPageListQuery>;
}

export default function MeasuresPage({ queryRef }: MeasuresPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();

  usePageTitle(t("measuresPage.title"));

  const { organization } = usePreloadedQuery<MeasuresPageListQuery>(measuresPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid node type");
  }

  const [searchParams, setSearchParams] = useSearchParams();
  const urlCategory = searchParams.get("category") ?? null;

  const [isPending, startTransition] = useTransition();
  const [queryFilter, setQueryFilter] = useState<string | null>(null);
  const [stateFilter, setStateFilter] = useState<MeasureState | null>(null);
  const { data, loadNext, hasNext, isLoadingNext, refetch }
    = usePaginationFragment<MeasuresPageRefetchQuery, MeasuresPageFragment$key>(
      measuresPageFragment,
      organization,
    );

  const refetchFilters = (overrides: Record<string, unknown> = {}) => {
    startTransition(() => {
      refetch(
        {
          query: queryFilter,
          state: stateFilter,
          category: urlCategory,
          ...overrides,
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  const initialUrlCategory = useRef(urlCategory);
  const prevUrlCategory = useRef(urlCategory);
  useEffect(() => {
    if (initialUrlCategory.current) {
      startTransition(() => {
        refetch(
          {
            query: null,
            state: null,
            category: initialUrlCategory.current,
          },
          { fetchPolicy: "network-only" },
        );
      });
    }
  }, [refetch, startTransition]);

  useEffect(() => {
    if (urlCategory !== prevUrlCategory.current) {
      prevUrlCategory.current = urlCategory;
      refetchFilters({ category: urlCategory });
    }
  });

  const handleQueryFilterChange = (value: string) => {
    const newQuery = value === "" ? null : value;
    setQueryFilter(newQuery);
    refetchFilters({ query: newQuery });
  };

  const handleStateFilterChange = (value: string) => {
    const newState = value === "ALL" ? null : (value as MeasureState);
    setStateFilter(newState);
    refetchFilters({ state: newState });
  };

  const handleCategoryFilterChange = (value: string) => {
    const newCategory = value === "ALL" ? null : value;
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (newCategory) {
        next.set("category", newCategory);
      } else {
        next.delete("category");
      }
      return next;
    }, { replace: true });
  };

  const currentFilter = {
    query: queryFilter,
    state: stateFilter,
    category: urlCategory,
  };

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    MeasuresConnectionKey,
    { filter: currentFilter },
  );
  const allFiltersNullConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    MeasuresConnectionKey,
    { filter: { query: null, state: null, category: null } },
  );
  const hasActiveFilter = queryFilter || stateFilter || urlCategory;
  const createConnectionIds = hasActiveFilter
    ? [allFiltersNullConnectionId, connectionId]
    : [connectionId];

  const measures = data?.measures?.edges?.map(edge => edge.node) ?? [];
  const categories = organization.measureCategories ?? [];

  const [importMeasures] = useMutationWithToasts<MeasuresPageImportMutation>(
    importMeasuresMutation,
    {
      successMessage: t("measuresPage.messages.imported"),
      errorMessage: t("measuresPage.errors.import"),
    },
  );
  const importFileRef = useRef<HTMLInputElement>(null);

  const handleImport: ChangeEventHandler<HTMLInputElement> = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    void importMeasures({
      variables: {
        input: {
          organizationId,
          file: null,
        },
        connections: createConnectionIds,
      },
      uploadables: {
        "input.file": file,
      },
      onCompleted() {
        importFileRef.current!.value = "";
      },
    });
  };

  const hasAnyAction = measures.some(
    ({ canUpdate, canDelete }) => canUpdate || canDelete,
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("measuresPage.title")}
        description={t("measuresPage.description")}
      >
        {organization.canCreateMeasure && (
          <>
            <FileButton
              ref={importFileRef}
              variant="secondary"
              icon={IconFolderUpload}
              onChange={handleImport}
            >
              {t("measuresPage.actions.import")}
            </FileButton>
            <MeasureFormDialog connection={connectionId}>
              <Button variant="primary" icon={IconPlusLarge}>
                {t("measuresPage.actions.newMeasure")}
              </Button>
            </MeasureFormDialog>
          </>
        )}
      </PageHeader>

      <div className="flex items-center gap-4">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={t("measuresPage.filters.searchPlaceholder")}
          value={queryFilter ?? ""}
          onValueChange={handleQueryFilterChange}
        />
        <Select
          value={stateFilter ?? "ALL"}
          onValueChange={handleStateFilterChange}
        >
          <Option value="ALL">{t("measuresPage.filters.allStates")}</Option>
          <Option value="NOT_STARTED">{t("measuresPage.states.not_started")}</Option>
          <Option value="IN_PROGRESS">{t("measuresPage.states.in_progress")}</Option>
          <Option value="IMPLEMENTED">{t("measuresPage.states.implemented")}</Option>
          <Option value="NOT_APPLICABLE">{t("measuresPage.states.not_applicable")}</Option>
        </Select>
        <Select
          value={urlCategory ?? "ALL"}
          onValueChange={handleCategoryFilterChange}
        >
          <Option value="ALL">{t("measuresPage.filters.allCategories")}</Option>
          {categories.map(category => (
            <Option key={category} value={category}>
              {category}
            </Option>
          ))}
        </Select>
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {measures.length > 0
          ? (
              <Card>
                <Table>
                  <Thead>
                    <Tr>
                      <Th>{t("measuresPage.columns.measure")}</Th>
                      <Th>{t("measuresPage.columns.category")}</Th>
                      <Th>{t("measuresPage.columns.state")}</Th>
                      {hasAnyAction && <Th />}
                    </Tr>
                  </Thead>
                  <Tbody>
                    {measures.map(measure => (
                      <MeasureRow
                        key={measure.id}
                        measureKey={measure}
                        connectionId={connectionId}
                        hasAnyAction={hasAnyAction}
                      />
                    ))}
                  </Tbody>
                </Table>

                {hasNext && (
                  <div className="p-4 border-t">
                    <Button
                      variant="secondary"
                      onClick={() => loadNext(20)}
                      disabled={isLoadingNext}
                    >
                      {isLoadingNext ? t("measuresPage.actions.loading") : t("measuresPage.actions.loadMore")}
                    </Button>
                  </div>
                )}
              </Card>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {t("measuresPage.empty.title")}
                  </h3>
                  <p className="text-txt-tertiary mb-4">
                    {t("measuresPage.empty.description")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}

type MeasureRowProps = {
  measureKey: MeasuresPageRowFragment$key;
  connectionId: string;
  hasAnyAction: boolean;
};

function MeasureRow(props: MeasureRowProps) {
  const measure = useFragment(measureRowFragment, props.measureKey);
  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const [deleteMeasure] = useMutation<MeasuresPageDeleteMutation>(deleteMeasureMutation);
  const { toast } = useToast();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteMeasure({
            variables: {
              input: { measureId: measure.id },
              connections: [props.connectionId],
            },
            onCompleted(_, error) {
              if (error) {
                toast({
                  title: t("measuresPage.messages.error"),
                  description: formatError(
                    t("measuresPage.errors.delete"),
                    error,
                  ),
                  variant: "error",
                });
              } else {
                toast({
                  title: t("measuresPage.messages.success"),
                  description: t("measuresPage.messages.deleted"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: t("measuresPage.messages.error"),
                description: formatError(
                  t("measuresPage.errors.delete"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("measuresPage.deleteConfirmation", { name: measure.name }),
      },
    );
  };

  return (
    <>
      <MeasureFormDialog measure={measure} ref={dialogRef} />
      <Tr to={`/organizations/${organizationId}/measures/${measure.id}`}>
        <Td>{measure.name}</Td>
        <Td>{measure.category}</Td>
        <Td width={120}>
          <MeasureBadge state={measure.state} />
        </Td>
        {props.hasAnyAction && (
          <Td noLink width={50} className="text-end">
            {(measure.canUpdate || measure.canDelete) && (
              <ActionDropdown>
                {measure.canUpdate && (
                  <DropdownItem
                    icon={IconPencil}
                    onClick={() => dialogRef.current?.open()}
                  >
                    {t("measuresPage.actions.edit")}
                  </DropdownItem>
                )}
                {measure.canDelete && (
                  <DropdownItem
                    onClick={handleDelete}
                    variant="danger"
                    icon={IconTrashCan}
                  >
                    {t("measuresPage.actions.delete")}
                  </DropdownItem>
                )}
              </ActionDropdown>
            )}
          </Td>
        )}
      </Tr>
    </>
  );
}
