import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { GeneralSettingsPageQuery } from "#/__generated__/iam/GeneralSettingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import {
  GeneralSettingsPage,
  generalSettingsPageQuery,
} from "./GeneralSettingsPage";

function GeneralSettingsPageQueryLoader() {
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

export default function GeneralSettingsPageLoader() {
  return (
    <IAMRelayProvider>
      <GeneralSettingsPageQueryLoader />
    </IAMRelayProvider>
  );
}
