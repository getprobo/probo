import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { SAMLSettingsPageQuery } from "#/__generated__/iam/SAMLSettingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { SAMLSettingsPage, samlSettingsPageQuery } from "./SAMLSettingsPage";

function SAMLSettingsPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<SAMLSettingsPageQuery>(
    samlSettingsPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <SAMLSettingsPage queryRef={queryRef} />;
}

export default function () {
  return (
    <IAMRelayProvider>
      <SAMLSettingsPageLoader />
    </IAMRelayProvider>
  );
}
