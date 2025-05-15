import { Suspense, useEffect, useTransition } from "react";
import {
  graphql,
  type PreloadedQuery,
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
import type { PeopleListViewQuery as PeopleListViewQueryType } from "./__generated__/PeopleListViewQuery.graphql";
import type { PeopleListViewDeletePeopleMutation } from "./__generated__/PeopleListViewDeletePeopleMutation.graphql";
import type { PeopleListViewPaginationQuery } from "./__generated__/PeopleListViewPaginationQuery.graphql";
import type { PeopleListView_peoples$key } from "./__generated__/PeopleListView_peoples.graphql";
import { PageTemplate } from "@/components/PageTemplate";
import { PeopleListViewSkeleton } from "./PeopleListPage";

const ITEMS_PER_PAGE = 25;

const peopleListViewQuery = graphql`
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
  @refetchable(queryName: "PeopleListViewPaginationQuery")
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
    ) @connection(key: "PeopleListView_peoples") {
      __id
      edges {
        node {
          id
          fullName
          primaryEmailAddress
          additionalEmailAddresses
          kind
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
  queryRef: PreloadedQuery<PeopleListViewQueryType>;
}) {
  const data = usePreloadedQuery<PeopleListViewQueryType>(
    peopleListViewQuery,
    queryRef,
  );
  const [, setSearchParams] = useSearchParams();
  const [, startTransition] = useTransition();
  const [deletePeople] =
    useMutation<PeopleListViewDeletePeopleMutation>(deletePeopleMutation);
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
    PeopleListViewPaginationQuery,
    PeopleListView_peoples$key
  >(peopleListFragment, data.organization);

  const peoples =
    peoplesConnection.peoples.edges.map((edge) => edge.node) ?? [];
  const pageInfo = peoplesConnection.peoples.pageInfo;

  return (
    <PageTemplate
      title="People"
      description="Keep track of your company's workforce and their progress
      towards completing tasks assigned to them."
      actions={
        <Button asChild variant="secondary" className="gap-2">
          <Link to={`/organizations/${organizationId}/people/new`}>
            <UserPlus className="h-4 w-4" />
            Add a person
          </Link>
        </Button>
      }
    >
      <div className="space-y-6">
        <div className="space-y-2">
          {peoples.map((person) => (
            <Link
              key={person?.id}
              to={`/organizations/${organizationId}/people/${person?.id}`}
              className="block"
            >
              <div className="flex items-center justify-between p-4 rounded-xl border bg-level-1 hover:bg-accent-bg/5 transition-colors">
                <div className="flex items-center gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>{person?.fullName?.[0]}</AvatarFallback>
                  </Avatar>
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{person?.fullName}</p>
                    {person?.primaryEmailAddress && (
                      <>
                        <span className="text-tertiary">•</span>
                        <p className="text-sm text-tertiary">
                          {person.primaryEmailAddress}
                        </p>
                      </>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge className="px-3 py-0.5 text-xs font-medium">
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
                    className="h-8 w-8 text-tertiary hover:bg-transparent hover:[&>svg]:text-danger"
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
    </PageTemplate>
  );
}

export default function PeopleListView() {
  const [searchParams] = useSearchParams();
  const [queryRef, loadQuery] =
    useQueryLoader<PeopleListViewQueryType>(peopleListViewQuery);

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
    return <PeopleListViewSkeleton />;
  }

  return (
    <Suspense fallback={<PeopleListViewSkeleton />}>
      <PeopleListContent queryRef={queryRef} />
    </Suspense>
  );
}
