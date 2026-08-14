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

import { useTranslation } from "react-i18next";
import { useParams } from "react-router";

import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NavPanelGroup } from "#/pages/iam/organizations/_components/shell/NavPanelGroup";
import { NavPanelItem } from "#/pages/iam/organizations/_components/shell/NavPanelItem";

/**
 * Settings and Pages clusters for the portal in the URL.
 *
 * Hidden until a portal is selected: these paths need an id, and the create
 * page is not one of the sections.
 */
export function CompliancePortalNavItems() {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();

  if (compliancePortalId == null) {
    return null;
  }

  const prefix = `/organizations/${organizationId}/compliance-portals/${compliancePortalId}`;

  return (
    <>
      <NavPanelGroup label={t("nav.compliancePortalsSettings")}>
        <NavPanelItem label={t("nav.compliancePortalsHosting")} to={`${prefix}/hosting`} />
        <NavPanelItem
          label={t("nav.compliancePortalsPermissions")}
          to={`${prefix}/permissions`}
        />
        <NavPanelItem
          label={t("nav.compliancePortalsIntegrations")}
          to={`${prefix}/integrations`}
        />
      </NavPanelGroup>
      <NavPanelGroup label={t("nav.compliancePortalsPages")}>
        <NavPanelItem label={t("nav.compliancePortalsLanding")} to={`${prefix}/landing`} />
        <NavPanelItem label={t("nav.compliancePortalsDocuments")} to={`${prefix}/documents`} />
        <NavPanelItem
          label={t("nav.compliancePortalsSubprocessors")}
          to={`${prefix}/subprocessors`}
        />
        <NavPanelItem label={t("nav.compliancePortalsUpdates")} to={`${prefix}/updates`} />
        <NavPanelItem
          label={t("nav.compliancePortalsRightRequests")}
          to={`${prefix}/right-requests`}
        />
      </NavPanelGroup>
    </>
  );
}
