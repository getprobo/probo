import { useQueryLoader } from "react-relay";
import { SCIMSettingsPage, scimSettingsPageQuery } from "./SCIMSettingsPage";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";
import { useEffect } from "react";
import type { SCIMSettingsPageQuery } from "/__generated__/iam/SCIMSettingsPageQuery.graphql";

function SCIMSettingsPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<SCIMSettingsPageQuery>(
    scimSettingsPageQuery
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

export default function () {
  return (
    <IAMRelayProvider>
      <SCIMSettingsPageLoader />
    </IAMRelayProvider>
  );
}
