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
} from "@probo/ui";
import { InternalControlBadge } from "@probo/ui/src/Molecules/Badge/InternalControlBadge";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
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

import type { InternalControlsPageDeleteMutation } from "#/__generated__/core/InternalControlsPageDeleteMutation.graphql";
import type { InternalControlsPageFragment$key } from "#/__generated__/core/InternalControlsPageFragment.graphql";
import type { InternalControlsPageImportMutation } from "#/__generated__/core/InternalControlsPageImportMutation.graphql";
import type { InternalControlsPageListQuery } from "#/__generated__/core/InternalControlsPageListQuery.graphql";
import type {
  InternalControlsPageRefetchQuery,
  InternalControlState,
} from "#/__generated__/core/InternalControlsPageRefetchQuery.graphql";
import type { InternalControlsPageRowFragment$key } from "#/__generated__/core/InternalControlsPageRowFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import InternalControlFormDialog from "./dialog/InternalControlFormDialog";

export const InternalControlsConnectionKey = "InternalControlsPage_internalControls";

export const internalControlsPageQuery = graphql`
  query InternalControlsPageListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateInternalControl: permission(action: "core:internal-control:create")
        internalControlCategories
        ...InternalControlsPageFragment
      }
    }
  }
`;

const internalControlRowFragment = graphql`
  fragment InternalControlsPageRowFragment on InternalControl {
    id
    name
    code
    category
    state
    implementationStatus
    canUpdate: permission(action: "core:internal-control:update")
    canDelete: permission(action: "core:internal-control:delete")
    ...InternalControlFormDialogInternalControlFragment
  }
`;

const deleteInternalControlMutation = graphql`
  mutation InternalControlsPageDeleteMutation(
    $input: DeleteInternalControlInput!
    $connections: [ID!]!
  ) {
    deleteInternalControl(input: $input) {
      deletedInternalControlId @deleteEdge(connections: $connections)
    }
  }
`;

const internalControlsPageFragment = graphql`
  fragment InternalControlsPageFragment on Organization
  @refetchable(queryName: "InternalControlsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    after: { type: "CursorKey" }
    query: { type: "String", defaultValue: null }
    state: { type: "InternalControlState", defaultValue: null }
    category: { type: "String", defaultValue: null }
  ) {
    id
    internalControls(
      first: $first
      after: $after
      filter: { query: $query, state: $state, category: $category }
    )
      @connection(
        key: "InternalControlsPage_internalControls"
        filters: ["filter"]
      ) {
      edges {
        node {
          id
          canUpdate: permission(action: "core:internal-control:update")
          canDelete: permission(action: "core:internal-control:delete")
          ...InternalControlsPageRowFragment
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

const importInternalControlsMutation = graphql`
  mutation InternalControlsPageImportMutation(
    $input: ImportInternalControlInput!
    $connections: [ID!]!
  ) {
    importInternalControl(input: $input) {
      internalControlEdges @appendEdge(connections: $connections) {
        node {
          ...InternalControlsPageRowFragment
        }
      }
    }
  }
`;

interface InternalControlsPageProps {
  queryRef: PreloadedQuery<InternalControlsPageListQuery>;
}

export default function InternalControlsPage({ queryRef }: InternalControlsPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();

  usePageTitle(t("internalControlsPage.title"));

  const { organization } = usePreloadedQuery<InternalControlsPageListQuery>(internalControlsPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid node type");
  }

  const [searchParams, setSearchParams] = useSearchParams();
  const urlCategory = searchParams.get("category") ?? null;

  const [isPending, startTransition] = useTransition();
  const [queryFilter, setQueryFilter] = useState<string | null>(null);
  const [stateFilter, setStateFilter] = useState<InternalControlState | null>(null);
  const { data, loadNext, hasNext, isLoadingNext, refetch }
    = usePaginationFragment<InternalControlsPageRefetchQuery, InternalControlsPageFragment$key>(
      internalControlsPageFragment,
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
    const newState = value === "ALL" ? null : (value as InternalControlState);
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
    InternalControlsConnectionKey,
    { filter: currentFilter },
  );
  const allFiltersNullConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    InternalControlsConnectionKey,
    { filter: { query: null, state: null, category: null } },
  );
  const hasActiveFilter = queryFilter || stateFilter || urlCategory;
  const createConnectionIds = hasActiveFilter
    ? [allFiltersNullConnectionId, connectionId]
    : [connectionId];

  const internalControls = data?.internalControls?.edges?.map(edge => edge.node) ?? [];
  const categories = organization.internalControlCategories ?? [];

  const [importInternalControls] = useMutationWithToasts<InternalControlsPageImportMutation>(
    importInternalControlsMutation,
    {
      successMessage: t("internalControlsPage.messages.imported"),
      errorMessage: t("internalControlsPage.errors.import"),
    },
  );
  const importFileRef = useRef<HTMLInputElement>(null);

  const handleImport: ChangeEventHandler<HTMLInputElement> = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    void importInternalControls({
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

  const hasAnyAction = internalControls.some(
    ({ canUpdate, canDelete }) => canUpdate || canDelete,
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("internalControlsPage.title")}
        description={t("internalControlsPage.description")}
      >
        {organization.canCreateInternalControl && (
          <>
            <FileButton
              ref={importFileRef}
              variant="secondary"
              icon={IconFolderUpload}
              onChange={handleImport}
            >
              {t("internalControlsPage.actions.import")}
            </FileButton>
            <InternalControlFormDialog connection={connectionId}>
              <Button variant="primary" icon={IconPlusLarge}>
                {t("internalControlsPage.actions.newInternalControl")}
              </Button>
            </InternalControlFormDialog>
          </>
        )}
      </PageHeader>

      <div className="flex items-center gap-4">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={t("internalControlsPage.filters.searchPlaceholder")}
          value={queryFilter ?? ""}
          onValueChange={handleQueryFilterChange}
        />
        <Select
          value={stateFilter ?? "ALL"}
          onValueChange={handleStateFilterChange}
        >
          <Option value="ALL">{t("internalControlsPage.filters.allStates")}</Option>
          <Option value="NOT_STARTED">{t("internalControlsPage.states.not_started")}</Option>
          <Option value="IN_PROGRESS">{t("internalControlsPage.states.in_progress")}</Option>
          <Option value="IMPLEMENTED">{t("internalControlsPage.states.implemented")}</Option>
          <Option value="NOT_APPLICABLE">{t("internalControlsPage.states.not_applicable")}</Option>
        </Select>
        <Select
          value={urlCategory ?? "ALL"}
          onValueChange={handleCategoryFilterChange}
        >
          <Option value="ALL">{t("internalControlsPage.filters.allCategories")}</Option>
          {categories.map(category => (
            <Option key={category} value={category}>
              {category}
            </Option>
          ))}
        </Select>
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {internalControls.length > 0
          ? (
              <Card>
                <Table>
                  <Thead>
                    <Tr>
                      <Th>{t("internalControlsPage.columns.code")}</Th>
                      <Th>{t("internalControlsPage.columns.internalControl")}</Th>
                      <Th>{t("internalControlsPage.columns.category")}</Th>
                      <Th>{t("internalControlsPage.columns.status")}</Th>
                      <Th>{t("internalControlsPage.columns.state")}</Th>
                      {hasAnyAction && <Th />}
                    </Tr>
                  </Thead>
                  <Tbody>
                    {internalControls.map(internalControl => (
                      <InternalControlRow
                        key={internalControl.id}
                        internalControlKey={internalControl}
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
                      {isLoadingNext ? t("internalControlsPage.actions.loading") : t("internalControlsPage.actions.loadMore")}
                    </Button>
                  </div>
                )}
              </Card>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {t("internalControlsPage.empty.title")}
                  </h3>
                  <p className="text-txt-tertiary mb-4">
                    {t("internalControlsPage.empty.description")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}

type InternalControlRowProps = {
  internalControlKey: InternalControlsPageRowFragment$key;
  connectionId: string;
  hasAnyAction: boolean;
};

function InternalControlRow(props: InternalControlRowProps) {
  const internalControl = useFragment(internalControlRowFragment, props.internalControlKey);
  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const [deleteInternalControl] = useMutation<InternalControlsPageDeleteMutation>(deleteInternalControlMutation);
  const toast = useToast();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteInternalControl({
            variables: {
              input: { internalControlId: internalControl.id },
              connections: [props.connectionId],
            },
            onCompleted(_, error) {
              if (error) {
                toast.add({
                  title: t("internalControlsPage.messages.error"),
                  description: formatError(
                    t("internalControlsPage.errors.delete"),
                    error,
                  ),
                  type: "error",
                });
              } else {
                toast.add({
                  title: t("internalControlsPage.messages.success"),
                  description: t("internalControlsPage.messages.deleted"),
                  type: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast.add({
                title: t("internalControlsPage.messages.error"),
                description: formatError(
                  t("internalControlsPage.errors.delete"),
                  error,
                ),
                type: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("internalControlsPage.deleteConfirmation", { name: internalControl.name }),
      },
    );
  };

  return (
    <>
      <InternalControlFormDialog internalControl={internalControl} ref={dialogRef} />
      <Tr to={`/organizations/${organizationId}/governance/internal-controls/${internalControl.id}`}>
        <Td>{internalControl.code}</Td>
        <Td>{internalControl.name}</Td>
        <Td>{internalControl.category}</Td>
        <Td>
          {internalControl.implementationStatus
            ? t(`internalControlDetailPage.implementationStatuses.${internalControl.implementationStatus.toLowerCase()}`)
            : null}
        </Td>
        <Td width={120}>
          <InternalControlBadge state={internalControl.state} />
        </Td>
        {props.hasAnyAction && (
          <Td noLink width={50} className="text-end">
            {(internalControl.canUpdate || internalControl.canDelete) && (
              <ActionDropdown>
                {internalControl.canUpdate && (
                  <DropdownItem
                    icon={IconPencil}
                    onClick={() => dialogRef.current?.open()}
                  >
                    {t("internalControlsPage.actions.edit")}
                  </DropdownItem>
                )}
                {internalControl.canDelete && (
                  <DropdownItem
                    onClick={handleDelete}
                    variant="danger"
                    icon={IconTrashCan}
                  >
                    {t("internalControlsPage.actions.delete")}
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
