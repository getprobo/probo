import { Suspense, useEffect } from "react";
import { useOrganizationId } from "/hooks/useOrganizationId";
import {
  DomainSettingsPage,
  domainSettingsPageQuery,
} from "./DomainSettingsPage";
import { useQueryLoader } from "react-relay";
import { CoreRelayProvider } from "/providers/CoreRelayProvider";
import type { DomainSettingsPageQuery } from "/__generated__/core/DomainSettingsPageQuery.graphql";

function DomainSettingsPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<DomainSettingsPageQuery>(
    domainSettingsPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <DomainSettingsPage queryRef={queryRef} />;
}

export default function () {
  return (
    <CoreRelayProvider>
      <Suspense>
        <DomainSettingsPageLoader />
      </Suspense>
    </CoreRelayProvider>
  );
}
