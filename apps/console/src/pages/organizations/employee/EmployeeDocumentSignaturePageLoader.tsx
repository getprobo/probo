import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";
import { useParams } from "react-router";

import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import type { EmployeeDocumentSignaturePageQuery } from "/__generated__/core/EmployeeDocumentSignaturePageQuery.graphql";
import { CoreRelayProvider } from "/providers/CoreRelayProvider";

import {
  EmployeeDocumentSignaturePage,
  employeeDocumentSignaturePageQuery,
} from "./EmployeeDocumentSignaturePage";

function EmployeeDocumentSignaturePageLoader() {
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

export default function () {
  return (
    <CoreRelayProvider>
      <EmployeeDocumentSignaturePageLoader />
    </CoreRelayProvider>
  );
}
