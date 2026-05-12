// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
