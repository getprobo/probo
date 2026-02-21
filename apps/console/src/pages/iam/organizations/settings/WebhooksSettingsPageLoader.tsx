import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { WebhooksSettingsPageQuery } from "#/__generated__/core/WebhooksSettingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  WebhooksSettingsPage,
  webhooksSettingsPageQuery,
} from "#/pages/organizations/settings/WebhooksSettingsPage";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

function WebhooksSettingsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<WebhooksSettingsPageQuery>(
    webhooksSettingsPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <WebhooksSettingsPage queryRef={queryRef} />;
}

export default function WebhooksSettingsPageLoader() {
  return (
    <CoreRelayProvider>
      <WebhooksSettingsPageQueryLoader />
    </CoreRelayProvider>
  );
}
