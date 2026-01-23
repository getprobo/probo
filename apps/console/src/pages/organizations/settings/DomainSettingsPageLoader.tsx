import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { DomainSettingsPageQuery } from "#/__generated__/core/DomainSettingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import {
  DomainSettingsPage,
  domainSettingsPageQuery,
} from "./DomainSettingsPage";

function DomainSettingsPageQueryLoader() {
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

export default function DomainSettingsPageLoader() {
  return (
    <CoreRelayProvider>
      <Suspense>
        <DomainSettingsPageQueryLoader />
      </Suspense>
    </CoreRelayProvider>
  );
}
