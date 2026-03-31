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

import {
  formatError,
  getMeasureStateLabel,
  type GraphQLError,
  sprintf,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  usePageTitle(__("Measures"));

  const { organization } = usePreloadedQuery(measuresPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid node type");
  }

  const [searchParams, setSearchParams] = useSearchParams();
  const urlCategory = searchParams.get("category") ?? null;

  const [isPending, startTransition] = useTransition();
  const [queryFilter, setQueryFilter] = useState<string | null>(null);
  const [stateFilter, setStateFilter] = useState<MeasureState | null>(null);
  const [categoryFilter, setCategoryFilter] = useState<string | null>(urlCategory);

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
          category: categoryFilter,
          ...overrides,
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  const initialUrlCategory = useRef(urlCategory);
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
    if (urlCategory !== categoryFilter) {
      setCategoryFilter(urlCategory);
      refetchFilters({ category: urlCategory });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlCategory]);

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
    setCategoryFilter(newCategory);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (newCategory) {
        next.set("category", newCategory);
      } else {
        next.delete("category");
      }
      return next;
    }, { replace: true });
    refetchFilters({ category: newCategory });
  };

  const currentFilter = {
    query: queryFilter,
    state: stateFilter,
    category: categoryFilter,
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
  const hasActiveFilter = queryFilter || stateFilter || categoryFilter;
  const createConnectionIds = hasActiveFilter
    ? [allFiltersNullConnectionId, connectionId]
    : [connectionId];

  const measures = data?.measures?.edges?.map(edge => edge.node) ?? [];
  const categories = organization.measureCategories ?? [];

  const [importMeasures] = useMutationWithToasts<MeasuresPageImportMutation>(
    importMeasuresMutation,
    {
      successMessage: __("Measures imported successfully."),
      errorMessage: __("Failed to import measures"),
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
        title={__("Measures")}
        description={__(
          "Measures are actions taken to reduce the risk. Add them to track their implementation status.",
        )}
      >
        {organization.canCreateMeasure && (
          <>
            <FileButton
              ref={importFileRef}
              variant="secondary"
              icon={IconFolderUpload}
              onChange={handleImport}
            >
              {__("Import")}
            </FileButton>
            <MeasureFormDialog connection={connectionId}>
              <Button variant="primary" icon={IconPlusLarge}>
                {__("New measure")}
              </Button>
            </MeasureFormDialog>
          </>
        )}
      </PageHeader>

      <div className="flex items-center gap-4">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search measures...")}
          value={queryFilter ?? ""}
          onValueChange={handleQueryFilterChange}
        />
        <Select
          value={stateFilter ?? "ALL"}
          onValueChange={handleStateFilterChange}
        >
          <Option value="ALL">{__("All states")}</Option>
          <Option value="NOT_STARTED">{getMeasureStateLabel(__, "NOT_STARTED")}</Option>
          <Option value="IN_PROGRESS">{getMeasureStateLabel(__, "IN_PROGRESS")}</Option>
          <Option value="IMPLEMENTED">{getMeasureStateLabel(__, "IMPLEMENTED")}</Option>
          <Option value="NOT_APPLICABLE">{getMeasureStateLabel(__, "NOT_APPLICABLE")}</Option>
        </Select>
        <Select
          value={categoryFilter ?? "ALL"}
          onValueChange={handleCategoryFilterChange}
        >
          <Option value="ALL">{__("All categories")}</Option>
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
                      <Th>{__("Measure")}</Th>
                      <Th>{__("Category")}</Th>
                      <Th>{__("State")}</Th>
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
                      {isLoadingNext ? __("Loading...") : __("Load more")}
                    </Button>
                  </div>
                )}
              </Card>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {__("No measures yet")}
                  </h3>
                  <p className="text-txt-tertiary mb-4">
                    {__("Create your first measure to get started.")}
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
  const { __ } = useTranslate();
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
                  title: __("Error"),
                  description: formatError(
                    __("Failed to delete measure"),
                    error as GraphQLError[],
                  ),
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Measure deleted successfully"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete measure"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the measure \"%s\". This action cannot be undone.",
          ),
          measure.name,
        ),
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
                    {__("Edit")}
                  </DropdownItem>
                )}
                {measure.canDelete && (
                  <DropdownItem
                    onClick={handleDelete}
                    variant="danger"
                    icon={IconTrashCan}
                  >
                    {__("Delete")}
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
