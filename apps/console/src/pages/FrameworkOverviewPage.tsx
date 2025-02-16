import { Suspense, useEffect } from "react";
import { useParams } from "react-router";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { Shield, FileText, Clock } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type { FrameworkOverviewPageQuery as FrameworkOverviewPageQueryType } from "./__generated__/FrameworkOverviewPageQuery.graphql";
import { Helmet } from "react-helmet-async";

const FrameworkOverviewPageQuery = graphql`
  query FrameworkOverviewPageQuery($frameworkId: ID!) {
    node(id: $frameworkId) {
      id
      ... on Framework {
        name
        description
        controls {
          edges {
            node {
              id
              name
              description
              state
              category
            }
          }
        }
      }
    }
  }
`;

function FrameworkOverviewPageContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<FrameworkOverviewPageQueryType>;
}) {
  const data = usePreloadedQuery(FrameworkOverviewPageQuery, queryRef);
  const framework = data.node;
  const controls = framework.controls?.edges.map((edge) => edge?.node) ?? [];

  // Group controls by their category
  const controlsByCategory = controls.reduce(
    (acc, control) => {
      if (!control?.category) return acc;
      if (!acc[control.category]) {
        acc[control.category] = [];
      }
      acc[control.category].push(control);
      return acc;
    },
    {} as Record<string, typeof controls>,
  );

  const controlCards = Object.entries(controlsByCategory).map(
    ([category, controls]) => ({
      title: category,
      controls,
      completed: controls.filter((c) => c?.state === "IMPLEMENTED").length,
      total: controls.length,
    }),
  );

  const totalImplemented = controls.filter(
    (c) => c?.state === "IMPLEMENTED",
  ).length;

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="space-y-4 mb-8">
        <h1 className="text-2xl font-semibold">{framework.name}</h1>
        <p className="text-muted-foreground max-w-3xl">
          {framework.description}
        </p>
      </div>

      <div className="mb-8 space-y-4">
        <div className="flex items-center gap-2 text-sm">
          <div className="bg-primary/10 text-primary px-3 py-1 rounded-md flex items-center gap-2">
            <FileText className="w-4 h-4" />
            Frame 153
          </div>
          <div className="bg-warning/10 text-warning px-3 py-1 rounded-md flex items-center gap-2">
            <Clock className="w-4 h-4" />
            12 hours left
          </div>
        </div>
        <div className="bg-card p-4 rounded-lg space-y-4">
          <h3 className="font-medium">Preparation phase</h3>
          <Progress value={30} className="h-2" />
          <div className="flex justify-between text-sm text-muted-foreground">
            <span>Preparation</span>
            <span>Observation period: 3 month</span>
            <span>Audit: 6-9 days</span>
            <span>Report: 10-14 days</span>
          </div>
        </div>
      </div>

      <div className="mb-6 flex justify-between items-center">
        <h2 className="text-xl font-semibold">Controls</h2>
        <div className="text-primary">
          {totalImplemented} out of {controls.length} validated
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {controlCards.map((card, index) => (
          <Card key={index}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Shield className="w-4 h-4" />
                {card.title}
              </CardTitle>
              <Avatar className="h-6 w-6">
                <AvatarFallback>U</AvatarFallback>
              </Avatar>
            </CardHeader>
            <CardContent>
              <div className="flex gap-1 my-2">
                {Array(card.total)
                  .fill(0)
                  .map((_, i) => (
                    <div
                      key={i}
                      className={`h-2 w-2 rounded-sm ${
                        i < card.completed ? "bg-primary" : "bg-muted"
                      }`}
                    />
                  ))}
              </div>
              <p className="text-xs text-muted-foreground">
                {card.completed}/{card.total} Controls validated
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

function FrameworkOverviewPageFallback() {
  return (
    <div className="min-h-screen bg-background p-6">
      <div className="mb-8">
        <div className="h-8 w-48 bg-muted animate-pulse rounded" />
        <div className="h-4 w-96 bg-muted animate-pulse rounded mt-2" />
      </div>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3].map((i) => (
          <Card key={i}>
            <CardContent className="p-6">
              <div className="relative mb-6">
                <div className="bg-muted w-24 h-24 rounded-full animate-pulse mb-4" />
                <div className="h-6 w-48 bg-muted animate-pulse rounded mb-2" />
                <div className="h-20 w-full bg-muted animate-pulse rounded" />
              </div>
              <div className="h-4 w-32 bg-muted animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

export default function FrameworkOverviewPage() {
  const { frameworkId } = useParams();
  const [queryRef, loadQuery] = useQueryLoader<FrameworkOverviewPageQueryType>(
    FrameworkOverviewPageQuery,
  );

  useEffect(() => {
    loadQuery({ frameworkId: frameworkId! });
  }, [loadQuery, frameworkId]);

  return (
    <>
      <Helmet>
        <title>Framework Overview - Probo Console</title>
      </Helmet>
      <Suspense fallback={<FrameworkOverviewPageFallback />}>
        {queryRef && <FrameworkOverviewPageContent queryRef={queryRef} />}
      </Suspense>
    </>
  );
}
