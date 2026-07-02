// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
