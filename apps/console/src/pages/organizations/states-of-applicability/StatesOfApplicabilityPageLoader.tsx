import { useQueryLoader } from "react-relay";
import { Suspense, useEffect } from "react";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import type { StatesOfApplicabilityPageQuery } from "/__generated__/core/StatesOfApplicabilityPageQuery.graphql";
import StatesOfApplicabilityPage, {
    statesOfApplicabilityPageQuery,
} from "./StatesOfApplicabilityPage";
import { useOrganizationId } from "/hooks/useOrganizationId";

function StatesOfApplicabilityPageLoader() {
    const organizationId = useOrganizationId();
    const [queryRef, loadQuery] =
        useQueryLoader<StatesOfApplicabilityPageQuery>(
            statesOfApplicabilityPageQuery,
        );

    useEffect(() => {
        loadQuery({ organizationId });
    }, [loadQuery, organizationId]);

    if (!queryRef) {
        return <PageSkeleton />;
    }

    return (
        <Suspense fallback={<PageSkeleton />}>
            <StatesOfApplicabilityPage queryRef={queryRef} />
        </Suspense>
    );
}

export default StatesOfApplicabilityPageLoader;
