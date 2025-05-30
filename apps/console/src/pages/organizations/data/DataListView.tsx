import { Suspense, useEffect, useTransition, useRef } from "react";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
  usePaginationFragment,
} from "react-relay";
import { useSearchParams, useParams } from "react-router";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Database, Plus, Trash2, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link } from "react-router";
import type { DataListViewQuery as DataListViewQueryType } from "./__generated__/DataListViewQuery.graphql";
import type { DataListViewDeleteDataMutation } from "./__generated__/DataListViewDeleteDataMutation.graphql";
import { DataListViewPaginationQuery } from "./__generated__/DataListViewPaginationQuery.graphql";
import { DataListView_data$key } from "./__generated__/DataListView_data.graphql";
import { PageTemplate } from "@/components/PageTemplate";
import { DataListViewSkeleton } from "./DataListPage";

const ITEMS_PER_PAGE = 25;

const dataListViewQuery = graphql`
  query DataListViewQuery(
    $organizationId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
  ) {
    organization: node(id: $organizationId) {
      ...DataListView_data
        @arguments(first: $first, after: $after, last: $last, before: $before)
    }
  }
`;

const dataListFragment = graphql`
  fragment DataListView_data on Organization
  @refetchable(queryName: "DataListViewPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int" }
    after: { type: "CursorKey" }
    last: { type: "Int" }
    before: { type: "CursorKey" }
  ) {
    id
    data(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { direction: ASC, field: NAME }
    ) @connection(key: "DataListView_data") {
      __id
      edges {
        node {
          id
          name
          dataSensitivity
          owner {
            id
            fullName
          }
          vendors {
            edges {
              node {
                id
                name
              }
            }
          }
          createdAt
          updatedAt
        }
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
    }
  }
`;

const deleteDataMutation = graphql`
  mutation DataListViewDeleteDataMutation(
    $input: DeleteDatumInput!
    $connections: [ID!]!
  ) {
    deleteDatum(input: $input) {
      deletedDatumId @deleteEdge(connections: $connections)
    }
  }
`;

function LoadAboveButton({
  isLoading,
  hasMore,
  onLoadMore,
}: {
  isLoading: boolean;
  hasMore: boolean;
  onLoadMore: () => void;
}) {
  if (!hasMore) {
    return null;
  }

  return (
    <div className="flex justify-center">
      <Button
        variant="outline"
        onClick={onLoadMore}
        disabled={isLoading}
        className="w-full"
      >
        {isLoading ? "Loading..." : "Load above"}
      </Button>
    </div>
  );
}

function LoadBelowButton({
  isLoading,
  hasMore,
  onLoadMore,
}: {
  isLoading: boolean;
  hasMore: boolean;
  onLoadMore: () => void;
}) {
  if (!hasMore) {
    return null;
  }

  return (
    <div className="flex justify-center">
      <Button
        variant="outline"
        onClick={onLoadMore}
        disabled={isLoading || !hasMore}
        className="w-full"
      >
        {isLoading ? "Loading..." : "Load below"}
      </Button>
    </div>
  );
}

function DataListContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<DataListViewQueryType>;
}) {
  const data = usePreloadedQuery<DataListViewQueryType>(
    dataListViewQuery,
    queryRef,
  );
  const [, setSearchParams] = useSearchParams();
  const [, startTransition] = useTransition();
  const [deleteData] =
    useMutation<DataListViewDeleteDataMutation>(deleteDataMutation);
  const { organizationId } = useParams();
  const isPaginationUpdate = useRef(false);

  const {
    data: dataConnection,
    loadNext,
    loadPrevious,
    hasNext,
    hasPrevious,
    isLoadingNext,
    isLoadingPrevious,
  } = usePaginationFragment<
    DataListViewPaginationQuery,
    DataListView_data$key
  >(dataListFragment, data.organization);

  const dataItems = dataConnection.data.edges.map((edge) => edge.node) ?? [];
  const pageInfo = dataConnection.data.pageInfo;

  return (
    <>
      <PageTemplate
        title="Data"
        description="Keep track of your organization's data and their data sensitivity."
        actions={
          <Button asChild variant="secondary" className="gap-2">
            <Link to={`/organizations/${organizationId}/data/new`}>
              <Plus className="h-4 w-4" />
              Add data
            </Link>
          </Button>
        }
      >
        <div className="space-y-6">
          <div className="space-y-2">
            {dataItems.map((item) => (
              <Link
                key={item?.id}
                to={`/organizations/${organizationId}/data/${item?.id}`}
                className="block"
              >
                <div className="flex items-center justify-between p-4 rounded-xl border bg-level-1 hover:bg-accent-bg/5 transition-colors">
                  <div className="flex items-center gap-3">
                    <Avatar className="h-8 w-8">
                      <AvatarFallback>
                        <Database className="h-4 w-4" />
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex flex-col gap-1">
                      <div className="flex items-center gap-2">
                        <p className="font-medium">{item?.name}</p>
                        {item?.owner?.fullName && (
                          <>
                            <span className="text-tertiary">•</span>
                            <p className="text-sm text-tertiary">
                              Owned by {item.owner.fullName}
                            </p>
                          </>
                        )}
                      </div>
                      {item?.vendors?.edges?.length > 0 && (
                        <div className="text-sm text-tertiary">
                          Vendors: {item.vendors.edges.map(edge => edge?.node?.name).join(", ")}
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge
                      variant={
                        item?.dataSensitivity === "NONE"
                          ? "default"
                          : item?.dataSensitivity === "LOW"
                          ? "success"
                          : item?.dataSensitivity === "MEDIUM"
                          ? "info"
                          : item?.dataSensitivity === "HIGH"
                          ? "warning"
                          : "destructive"
                      }
                      className="px-3 py-0.5 text-xs font-medium"
                    >
                      {item?.dataSensitivity === "NONE"
                        ? "No sensitive data"
                        : item?.dataSensitivity === "LOW"
                        ? "Public or non-sensitive data"
                        : item?.dataSensitivity === "MEDIUM"
                        ? "Internal/restricted data"
                        : item?.dataSensitivity === "HIGH"
                        ? "Confidential data"
                        : item?.dataSensitivity === "CRITICAL"
                        ? "Regulated/PII/financial data"
                        : "No sensitive data"}
                    </Badge>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-tertiary hover:bg-transparent hover:[&>svg]:text-danger"
                      onClick={(e) => {
                        e.preventDefault();
                        if (
                          window.confirm(
                            "Are you sure you want to delete this data?",
                          )
                        ) {
                          deleteData({
                            variables: {
                              connections: [dataConnection.data.__id],
                              input: {
                                datumId: item.id,
                              },
                            },
                          });
                        }
                      }}
                    >
                      <Trash2 className="h-4 w-4 transition-colors" />
                    </Button>
                    <ChevronRight className="h-4 w-4 text-tertiary" />
                  </div>
                </div>
              </Link>
            ))}
          </div>

          <LoadAboveButton
            isLoading={isLoadingPrevious}
            hasMore={hasPrevious}
            onLoadMore={() => {
              startTransition(() => {
                isPaginationUpdate.current = true;
                setSearchParams((prev) => {
                  prev.set("before", pageInfo?.startCursor || "");
                  prev.delete("after");
                  return prev;
                });
                loadPrevious(ITEMS_PER_PAGE);
              });
            }}
          />
          <LoadBelowButton
            isLoading={isLoadingNext}
            hasMore={hasNext}
            onLoadMore={() => {
              startTransition(() => {
                isPaginationUpdate.current = true;
                setSearchParams((prev) => {
                  prev.set("after", pageInfo?.endCursor || "");
                  prev.delete("before");
                  return prev;
                });
                loadNext(ITEMS_PER_PAGE);
              });
            }}
          />
        </div>
      </PageTemplate>
    </>
  );
}

export default function DataListView() {
  const [searchParams] = useSearchParams();
  const [queryRef, loadQuery] =
    useQueryLoader<DataListViewQueryType>(dataListViewQuery);
  const { organizationId } = useParams();
  const isPaginationUpdate = useRef(false);

  useEffect(() => {
    const after = searchParams.get("after");
    const before = searchParams.get("before");

    // Skip the query if this was triggered by pagination
    if (isPaginationUpdate.current) {
      isPaginationUpdate.current = false;
      return;
    }

    loadQuery({
      organizationId: organizationId!,
      first: before ? undefined : ITEMS_PER_PAGE,
      after: after || undefined,
      last: before ? ITEMS_PER_PAGE : undefined,
      before: before || undefined,
    });
  }, [loadQuery, organizationId, searchParams]);

  if (!queryRef) {
    return <DataListViewSkeleton />;
  }

  return (
    <Suspense fallback={<DataListViewSkeleton />}>
      <DataListContent queryRef={queryRef} />
    </Suspense>
  );
}
