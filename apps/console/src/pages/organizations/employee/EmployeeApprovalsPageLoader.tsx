import { Skeleton } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { EmployeeApprovalsPageQuery } from "#/__generated__/core/EmployeeApprovalsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  EmployeeApprovalsPage,
  employeeApprovalsPageQuery,
} from "./EmployeeApprovalsPage";

function EmployeeApprovalsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<EmployeeApprovalsPageQuery>(
    employeeApprovalsPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <Skeleton className="w-full h-64" />;
  }

  return <EmployeeApprovalsPage queryRef={queryRef} />;
}

export default function EmployeeApprovalsPageLoader() {
  return (
    <Suspense fallback={<Skeleton className="w-full h-64" />}>
      <EmployeeApprovalsPageQueryLoader />
    </Suspense>
  );
}
