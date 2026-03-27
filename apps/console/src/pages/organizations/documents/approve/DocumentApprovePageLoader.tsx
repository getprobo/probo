import { Spinner } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { DocumentApprovePageQuery } from "#/__generated__/core/DocumentApprovePageQuery.graphql";

import {
  DocumentApprovePage,
  documentApprovePageQuery,
} from "./DocumentApprovePage";

function DocumentApprovePageQueryLoader() {
  const { documentId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }

  const [queryRef, loadQuery]
    = useQueryLoader<DocumentApprovePageQuery>(documentApprovePageQuery);

  useEffect(() => {
    loadQuery({ documentId });
  }, [loadQuery, documentId]);

  if (!queryRef) {
    return <Spinner />;
  }

  return (
    <Suspense fallback={<Spinner />}>
      <DocumentApprovePage queryRef={queryRef} />
    </Suspense>
  );
}

export default function DocumentApprovePageLoader() {
  return (
    <Suspense fallback={<Spinner />}>
      <DocumentApprovePageQueryLoader />
    </Suspense>
  );
}
