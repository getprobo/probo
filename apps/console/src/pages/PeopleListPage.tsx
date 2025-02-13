import { Suspense, useEffect, useTransition } from "react";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
} from "react-relay";
import { useSearchParams } from "react-router";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { CircleUser, Globe, Shield } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link } from "react-router";
import { Helmet } from "react-helmet-async";
import type { PeopleListPageQuery as PeopleListPageQueryType } from "./__generated__/PeopleListPageQuery.graphql";

const ITEMS_PER_PAGE = 20;

const PeopleListPageQuery = graphql`
  query PeopleListPageQuery(
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
  ) {
    node(id: "AZSfP_xAcAC5IAAAAAAltA") {
      id
      ... on Organization {
        peoples(first: $first, after: $after, last: $last, before: $before)
          @connection(key: "PeopleListPageQuery_peoples") {
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
            cursor
          }
          pageInfo {
            hasNextPage
            hasPreviousPage
            startCursor
            endCursor
          }
        }
      }
    }
  }
`;

const deletePeopleMutation = graphql`
  mutation PeopleListPageDeletePeopleMutation($input: DeletePeopleInput!) {
    deletePeople(input: $input)
  }
`;

function PeopleListPageContent({
  queryRef,
  onPageChange,
  loadQuery,
}: {
  queryRef: PreloadedQuery<PeopleListPageQueryType>;
  onPageChange: (params: {
    first?: number;
    after?: string;
    last?: number;
    before?: string;
  }) => void;
  loadQuery: LoadQueryType;
}) {
  const data = usePreloadedQuery(PeopleListPageQuery, queryRef);
  const peoples = data.node.peoples?.edges.map((edge) => edge?.node) ?? [];
  const pageInfo = data.node.peoples?.pageInfo;
  const [isPending, startTransition] = useTransition();
  const [deletePeople] = useMutation(deletePeopleMutation);

  const handlePageChange = (direction: "prev" | "next") => {
    startTransition(() => {
      if (direction === "prev") {
        onPageChange({
          last: ITEMS_PER_PAGE,
          before: pageInfo?.startCursor,
        });
      } else {
        onPageChange({
          first: ITEMS_PER_PAGE,
          after: pageInfo?.endCursor,
        });
      }
    });
  };

  return (
    <div className="p-6 space-y-6">
      <div className="space-y-1">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-semibold tracking-tight">People</h2>
            <p className="text-sm text-muted-foreground">
              Manage your organization's people.
            </p>
          </div>
          <Button asChild>
            <Link to="/peoples/create">Create People</Link>
          </Button>
        </div>
      </div>
      <div className="space-y-2">
        {peoples.map((person) => (
          <div
            key={person?.id}
            className="flex items-center justify-between p-4 rounded-lg border bg-card text-card-foreground shadow-sm"
          >
            <div className="flex items-center gap-4">
              <Avatar className="h-10 w-10">
                <AvatarFallback>{person?.fullName?.[0]}</AvatarFallback>
              </Avatar>
              <div className="space-y-1">
                <p className="text-sm font-medium leading-none">
                  {person?.fullName}
                </p>
                <p className="text-sm text-muted-foreground">
                  {person?.primaryEmailAddress}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <Badge variant="secondary" className="font-medium">
                {person?.kind === "EMPLOYEE"
                  ? "Employee"
                  : person?.kind === "CONTRACTOR"
                    ? "Contractor"
                    : "Vendor"}
              </Badge>
              <div className="flex gap-1">
                <CircleUser className="h-4 w-4 text-muted-foreground" />
                <Globe className="h-4 w-4 text-muted-foreground" />
                <Shield className="h-4 w-4 text-muted-foreground" />
              </div>
              <Badge variant="outline" className="text-muted-foreground">
                Not onboarded
              </Badge>
              <Button
                variant="ghost"
                size="sm"
                className="text-destructive hover:text-destructive hover:bg-destructive/10"
                onClick={(e) => {
                  e.preventDefault();
                  if (
                    window.confirm(
                      "Are you sure you want to delete this person?",
                    )
                  ) {
                    deletePeople({
                      variables: {
                        input: {
                          peopleId: person.id,
                        },
                      },
                      onCompleted() {
                        loadQuery(
                          {
                            first: ITEMS_PER_PAGE,
                            after: undefined,
                            last: undefined,
                            before: undefined,
                          },
                          { fetchPolicy: "network-only" },
                        );
                      },
                    });
                  }
                }}
              >
                Delete
              </Button>
            </div>
          </div>
        ))}

        <div className="flex gap-2 justify-end mt-4">
          <Button
            variant="outline"
            onClick={() => handlePageChange("prev")}
            disabled={isPending || !pageInfo?.hasPreviousPage}
          >
            {isPending ? "Loading..." : "Previous"}
          </Button>
          <Button
            variant="outline"
            onClick={() => handlePageChange("next")}
            disabled={isPending || !pageInfo?.hasNextPage}
          >
            {isPending ? "Loading..." : "Next"}
          </Button>
        </div>
      </div>
    </div>
  );
}

function PeopleListPageFallback() {
  return (
    <div className="p-6 space-y-6">
      <div className="space-y-1">
        <div className="h-8 w-48 bg-muted animate-pulse rounded" />
        <div className="h-4 w-96 bg-muted animate-pulse rounded" />
      </div>
      <div className="space-y-2">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="flex items-center justify-between p-4 rounded-lg border bg-card text-card-foreground shadow-sm"
          >
            <div className="flex items-center gap-4">
              <div className="h-10 w-10 rounded-full bg-muted animate-pulse" />
              <div className="space-y-1">
                <div className="h-4 w-32 bg-muted animate-pulse rounded" />
                <div className="h-3 w-48 bg-muted animate-pulse rounded" />
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="h-6 w-16 bg-muted animate-pulse rounded-full" />
              <div className="flex gap-1">
                {[1, 2, 3].map((j) => (
                  <div
                    key={j}
                    className="h-4 w-4 bg-muted animate-pulse rounded"
                  />
                ))}
              </div>
              <div className="h-6 w-24 bg-muted animate-pulse rounded-full" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

type LoadQueryType = ReturnType<
  typeof useQueryLoader<PeopleListPageQueryType>
>[1];

export default function PeopleListPage() {
  const [searchParams] = useSearchParams();
  const [queryRef, loadQuery] =
    useQueryLoader<PeopleListPageQueryType>(PeopleListPageQuery);

  useEffect(() => {
    const after = searchParams.get("after");
    const before = searchParams.get("before");

    loadQuery({
      first: before ? undefined : ITEMS_PER_PAGE,
      after: after || undefined,
      last: before ? ITEMS_PER_PAGE : undefined,
      before: before || undefined,
    });
  }, [loadQuery, searchParams]);

  const handlePageChange = ({
    first,
    after,
    last,
    before,
  }: {
    first?: number;
    after?: string;
    last?: number;
    before?: string;
  }) => {
    loadQuery(
      {
        first,
        after,
        last,
        before,
      },
      { fetchPolicy: "network-only" },
    );
  };

  if (!queryRef) {
    return <PeopleListPageFallback />;
  }

  return (
    <>
      <Helmet>
        <title>People - Probo Console</title>
      </Helmet>
      <Suspense fallback={<PeopleListPageFallback />}>
        <PeopleListPageContent
          queryRef={queryRef}
          onPageChange={handlePageChange}
          loadQuery={loadQuery}
        />
      </Suspense>
    </>
  );
}
