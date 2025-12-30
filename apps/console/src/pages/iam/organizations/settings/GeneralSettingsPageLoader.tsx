import { useOrganizationId } from "/hooks/useOrganizationId";
import { useQueryLoader } from "react-relay";
import { useEffect } from "react";
import {
  GeneralSettingsPage,
  generalSettingsPageQuery,
} from "./GeneralSettingsPage";
import type { GeneralSettingsPageQuery } from "/__generated__/iam/GeneralSettingsPageQuery.graphql";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";

function GeneralSettingsPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<GeneralSettingsPageQuery>(
    generalSettingsPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <GeneralSettingsPage queryRef={queryRef} />;
}

export default function () {
  return (
    <IAMRelayProvider>
      <GeneralSettingsPageLoader />
    </IAMRelayProvider>
  );
}
