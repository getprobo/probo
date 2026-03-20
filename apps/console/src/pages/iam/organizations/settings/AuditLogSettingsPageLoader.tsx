import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { AuditLogSettingsPageQuery } from "#/__generated__/iam/AuditLogSettingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  AuditLogSettingsPage,
  auditLogSettingsPageQuery,
} from "#/pages/iam/organizations/settings/AuditLogSettingsPage";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

function AuditLogSettingsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<AuditLogSettingsPageQuery>(
    auditLogSettingsPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <AuditLogSettingsPage queryRef={queryRef} />;
}

export default function AuditLogSettingsPageLoader() {
  return (
    <IAMRelayProvider>
      <AuditLogSettingsPageQueryLoader />
    </IAMRelayProvider>
  );
}
