import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import type { StateOfApplicabilityDetailPageQuery } from "/__generated__/core/StateOfApplicabilityDetailPageQuery.graphql";
import StateOfApplicabilityDetailPage, {
    stateOfApplicabilityDetailPageQuery,
} from "./StateOfApplicabilityDetailPage";
import { useParams } from "react-router";

function StateOfApplicabilityDetailPageLoader() {
    const { stateOfApplicabilityId } = useParams<{
        stateOfApplicabilityId: string;
    }>();
    const [queryRef, loadQuery] =
        useQueryLoader<StateOfApplicabilityDetailPageQuery>(
            stateOfApplicabilityDetailPageQuery,
        );

    useEffect(() => {
        if (stateOfApplicabilityId) {
            loadQuery({ stateOfApplicabilityId });
        }
    }, [loadQuery, stateOfApplicabilityId]);

    if (!queryRef) {
        return <PageSkeleton />;
    }

    return (
        <Suspense fallback={<PageSkeleton />}>
            <StateOfApplicabilityDetailPage queryRef={queryRef} />
        </Suspense>
    );
}

export default StateOfApplicabilityDetailPageLoader;
