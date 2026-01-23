import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { SCIMSettingsPageQuery } from "#/__generated__/iam/SCIMSettingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { SCIMSettingsPage, scimSettingsPageQuery } from "./SCIMSettingsPage";

function SCIMSettingsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<SCIMSettingsPageQuery>(
    scimSettingsPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <SCIMSettingsPage queryRef={queryRef} />;
}

export default function SCIMSettingsPageLoader() {
  return (
    <IAMRelayProvider>
      <SCIMSettingsPageQueryLoader />
    </IAMRelayProvider>
  );
}
