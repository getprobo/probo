import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { EmployeeDocumentsPageQuery } from "#/__generated__/core/EmployeeDocumentsPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import {
  EmployeeDocumentsPage,
  employeeDocumentsPageQuery,
} from "./EmployeeDocumentsPage";

function EmployeeDocumentsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<EmployeeDocumentsPageQuery>(
    employeeDocumentsPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <EmployeeDocumentsPage queryRef={queryRef} />;
}

export default function EmployeeDocumentsPageLoader() {
  return (
    <CoreRelayProvider>
      <Suspense fallback={<PageSkeleton />}>
        <EmployeeDocumentsPageQueryLoader />
      </Suspense>
    </CoreRelayProvider>
  );
}
