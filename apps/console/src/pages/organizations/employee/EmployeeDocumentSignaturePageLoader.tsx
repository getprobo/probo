import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { EmployeeDocumentSignaturePageQuery } from "#/__generated__/core/EmployeeDocumentSignaturePageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import {
  EmployeeDocumentSignaturePage,
  employeeDocumentSignaturePageQuery,
} from "./EmployeeDocumentSignaturePage";

function EmployeeDocumentSignaturePageQueryLoader() {
  const { documentId } = useParams();
  const [queryRef, loadQuery]
    = useQueryLoader<EmployeeDocumentSignaturePageQuery>(
      employeeDocumentSignaturePageQuery,
    );

  useEffect(() => {
    if (documentId) {
      loadQuery({
        documentId,
      });
    }
  }, [loadQuery, documentId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <EmployeeDocumentSignaturePage queryRef={queryRef} />
    </Suspense>
  );
}

export default function EmployeeDocumentSignaturePageLoader() {
  return (
    <CoreRelayProvider>
      <EmployeeDocumentSignaturePageQueryLoader />
    </CoreRelayProvider>
  );
}
