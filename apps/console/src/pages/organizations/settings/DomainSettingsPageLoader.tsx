import { Suspense, useEffect } from "react";
import { useOrganizationId } from "/hooks/useOrganizationId";
import {
  DomainSettingsPage,
  domainSettingsPageQuery,
} from "./DomainSettingsPage";
import { useQueryLoader } from "react-relay";
import { CoreRelayProvider } from "/providers/CoreRelayProvider";
import type { DomainSettingsPageQuery } from "/__generated__/core/DomainSettingsPageQuery.graphql";
import { PermissionsProviderLoader } from "/providers/NewPermissionsProvider";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";
import { domainSettingsPagePermissionsQuery } from "./domainSettingsPage.iam";
import type { domainSettingsPage_permissionsQuery } from "/__generated__/iam/domainSettingsPage_permissionsQuery.graphql";

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
    <IAMRelayProvider>
      <PermissionsProviderLoader<domainSettingsPage_permissionsQuery>
        query={domainSettingsPagePermissionsQuery}
      >
        <CoreRelayProvider>
          <Suspense>
            <DomainSettingsPageLoader />
          </Suspense>
        </CoreRelayProvider>
      </PermissionsProviderLoader>
    </IAMRelayProvider>
  );
}
