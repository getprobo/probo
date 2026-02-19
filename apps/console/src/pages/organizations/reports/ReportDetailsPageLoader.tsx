import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { ReportDetailsPageQuery } from "#/__generated__/core/ReportDetailsPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import ReportDetailsPage, { reportDetailsPageQuery } from "./ReportDetailsPage";

export default function ReportDetailsPageLoader() {
  const { reportId } = useParams<{ reportId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<ReportDetailsPageQuery>(reportDetailsPageQuery);

  useEffect(() => {
    if (reportId) {
      loadQuery({ reportId });
    }
  }, [loadQuery, reportId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <ReportDetailsPage queryRef={queryRef} />
    </Suspense>
  );
}
