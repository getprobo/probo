import { useQueryLoader } from "react-relay";
import { SAMLSettingsPage, samlSettingsPageQuery } from "./SAMLSettingsPage";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";
import { useEffect } from "react";
import type { SAMLSettingsPageQuery } from "./__generated__/SAMLSettingsPageQuery.graphql";

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
