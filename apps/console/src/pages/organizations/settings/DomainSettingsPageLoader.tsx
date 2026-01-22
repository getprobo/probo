import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import { useOrganizationId } from "/hooks/useOrganizationId";
import { CoreRelayProvider } from "/providers/CoreRelayProvider";
import type { DomainSettingsPageQuery } from "/__generated__/core/DomainSettingsPageQuery.graphql";

import {
  DomainSettingsPage,
  domainSettingsPageQuery,
} from "./DomainSettingsPage";

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
