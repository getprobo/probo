import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { EmployeeDocumentsPageQuery } from "./__generated__/EmployeeDocumentsPageQuery.graphql";
import {
  EmployeeDocumentsPage,
  employeeDocumentsPageQuery,
} from "./EmployeeDocumentsPage";
import { CoreRelayProvider } from "/providers/CoreRelayProvider";

function EmployeeDocumentsPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<EmployeeDocumentsPageQuery>(
    employeeDocumentsPageQuery
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <EmployeeDocumentsPage queryRef={queryRef} />
    </Suspense>
  );
}

export default function () {
  return (
    <CoreRelayProvider>
      <EmployeeDocumentsPageLoader />
    </CoreRelayProvider>
  );
}
