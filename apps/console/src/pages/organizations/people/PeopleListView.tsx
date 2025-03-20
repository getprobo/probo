import { Suspense, useEffect, useTransition } from "react";
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
import { UserPlus, Trash2, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link } from "react-router";
import { Helmet } from "react-helmet-async";
import type { PeopleListPageQuery as PeopleListPageQueryType } from "./__generated__/PeopleListPageQuery.graphql";
import type { PeopleListPageDeletePeopleMutation } from "./__generated__/PeopleListPageDeletePeopleMutation.graphql";
import { PeopleListPagePaginationQuery } from "./__generated__/PeopleListPagePaginationQuery.graphql";
import { PeopleListPage_peoples$key } from "./__generated__/PeopleListPage_peoples.graphql";
import { PageHeader } from "@/components/PageHeader";

const ITEMS_PER_PAGE = 25;

const peopleListPageQuery = graphql`
  query PeopleListViewQuery(
    $organizationId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
  ) {
    organization: node(id: $organizationId) {
      ...PeopleListView_peoples
        @arguments(first: $first, after: $after, last: $last, before: $before)
    }
  }
`;

const peopleListFragment = graphql`
  fragment PeopleListView_peoples on Organization
  @refetchable(queryName: "PeopleListPagePaginationQuery")
  @argumentDefinitions(
    first: { type: "Int" }
    after: { type: "CursorKey" }
    last: { type: "Int" }
    before: { type: "CursorKey" }
  ) {
    id
    peoples(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { direction: ASC, field: FULL_NAME }
    ) @connection(key: "PeopleListPage_peoples") {
      __id
      edges {
        node {
          id
          fullName
          primaryEmailAddress
          additionalEmailAddresses
          kind
          position
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

const deletePeopleMutation = graphql`
  mutation PeopleListViewDeletePeopleMutation(
    $input: DeletePeopleInput!
    $connections: [ID!]!
  ) {
    deletePeople(input: $input) {
      deletedPeopleId @deleteEdge(connections: $connections)
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

function PeopleListContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<PeopleListPageQueryType>;
}) {
  const data = usePreloadedQuery<PeopleListPageQueryType>(
    peopleListPageQuery,
    queryRef,
  );
  const [, setSearchParams] = useSearchParams();
  const [, startTransition] = useTransition();
  const [deletePeople] =
    useMutation<PeopleListPageDeletePeopleMutation>(deletePeopleMutation);
  const { organizationId } = useParams();

  const {
    data: peoplesConnection,
    loadNext,
    loadPrevious,
    hasNext,
    hasPrevious,
    isLoadingNext,
    isLoadingPrevious,
  } = usePaginationFragment<
    PeopleListPagePaginationQuery,
    PeopleListPage_peoples$key
  >(peopleListFragment, data.organization);

  const peoples =
    peoplesConnection.peoples.edges.map((edge) => edge.node) ?? [];
  const pageInfo = peoplesConnection.peoples.pageInfo;

  return (
    <>
      <Helmet>
        <title>People - Probo</title>
      </Helmet>
      <div className="container space-y-6">
        <PageHeader
          className="mb-17"
          title="Employees"
          description="Keep track of your company's workforce and their progress
            towards completing tasks assigned to them."
          actions={
            <Button
              asChild
              variant="outline"
              style={{ borderRadius: "0.5rem" }}
              className="gap-2"
            >
              <Link to={`/organizations/${organizationId}/people/create`}>
                <UserPlus className="h-4 w-4" />
                Add a person
              </Link>
            </Button>
          }
        />

        <div className="space-y-2">
          {peoples.map((person) => (
            <Link
              key={person?.id}
              to={`/organizations/${organizationId}/people/${person?.id}`}
              className="block"
            >
              <div className="flex items-center justify-between p-4 rounded-xl border bg-card hover:bg-accent/5 transition-colors">
                <div className="flex items-center gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>{person?.fullName?.[0]}</AvatarFallback>
                  </Avatar>
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{person?.fullName}</p>
                    {person?.primaryEmailAddress && (
                      <>
                        <span className="text-muted-foreground">•</span>
                        <p className="text-sm text-muted-foreground">
                          {person.primaryEmailAddress}
                        </p>
                      </>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge
                    variant="secondary"
                    className="bg-lime-9 text-white rounded-full px-3 py-0.5 text-xs font-medium"
                  >
                    {person?.kind === "EMPLOYEE"
                      ? "Employee"
                      : person?.kind === "CONTRACTOR"
                        ? "Contractor"
                        : person?.kind === "SERVICE_ACCOUNT"
                          ? "Service Account"
                          : "Vendor"}
                  </Badge>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-muted-foreground hover:bg-transparent hover:[&>svg]:text-destructive"
                    onClick={(e) => {
                      e.preventDefault();
                      if (
                        window.confirm(
                          "Are you sure you want to delete this person?",
                        )
                      ) {
                        deletePeople({
                          variables: {
                            connections: [peoplesConnection.peoples.__id],
                            input: {
                              peopleId: person.id,
                            },
                          },
                        });
                      }
                    }}
                  >
                    <Trash2 className="h-4 w-4 transition-colors" />
                  </Button>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
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
    </>
  );
}

function PeopleListPageFallback() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-48 bg-muted animate-pulse rounded" />
        <div className="h-4 w-96 bg-muted animate-pulse rounded mt-1" />
      </div>
      <div className="rounded-xl border bg-card p-4 space-y-4">
        <div className="h-5 w-32 bg-muted animate-pulse rounded" />
        <div className="h-10 w-full bg-muted animate-pulse rounded" />
      </div>
      <div className="space-y-2">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-[72px] bg-muted animate-pulse rounded-xl" />
        ))}
      </div>
    </div>
  );
}

export default function PeopleListPage() {
  const [searchParams] = useSearchParams();
  const [queryRef, loadQuery] =
    useQueryLoader<PeopleListPageQueryType>(peopleListPageQuery);

  const { organizationId } = useParams();

  useEffect(() => {
    const after = searchParams.get("after");
    const before = searchParams.get("before");

    loadQuery({
      organizationId: organizationId!,
      first: before ? undefined : ITEMS_PER_PAGE,
      after: after || undefined,
      last: before ? ITEMS_PER_PAGE : undefined,
      before: before || undefined,
    });
  }, [loadQuery, organizationId, searchParams]);

  if (!queryRef) {
    return <PeopleListPageFallback />;
  }

  return (
    <>
      <Helmet>
        <title>People - Probo Console</title>
      </Helmet>
      <Suspense fallback={<PeopleListPageFallback />}>
        <PeopleListContent queryRef={queryRef} />
      </Suspense>
    </>
  );
}
