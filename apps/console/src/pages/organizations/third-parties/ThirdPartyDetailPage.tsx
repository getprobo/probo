// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { faviconUrl, validateSnapshotConsistency } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  DropdownItem,
  IconPageTextLine,
  IconTrashCan,
  TabBadge,
  TabLink,
  Tabs,
} from "@probo/ui";
import {
  ConnectionHandler,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { Outlet, useParams } from "react-router";

import type { ThirdPartyComplianceTabFragment$key } from "#/__generated__/core/ThirdPartyComplianceTabFragment.graphql";
import type { ThirdPartyGraphNodeQuery } from "#/__generated__/core/ThirdPartyGraphNodeQuery.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import {
  useDeleteThirdParty,
  thirdPartyConnectionKey,
  thirdPartyNodeQuery,
} from "#/hooks/graph/ThirdPartyGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { ImportAssessmentDialog } from "./dialogs/ImportAssessmentDialog";
import { complianceReportsFragment } from "./tabs/ThirdPartyComplianceTab";

type Props = {
  queryRef: PreloadedQuery<ThirdPartyGraphNodeQuery>;
};

export default function ThirdPartyDetailPage(props: Props) {
  const { node: thirdParty } = usePreloadedQuery(thirdPartyNodeQuery, props.queryRef);
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  validateSnapshotConsistency(thirdParty, snapshotId);
  const deleteThirdParty = useDeleteThirdParty(
    thirdParty,
    ConnectionHandler.getConnectionID(organizationId, thirdPartyConnectionKey),
  );
  const logo = faviconUrl(thirdParty.websiteUrl);
  const reportsCount = useFragment(
    complianceReportsFragment,
    thirdParty as ThirdPartyComplianceTabFragment$key,
  ).complianceReports.edges.length;

  const thirdPartiesUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/third-parties`
      : `/organizations/${organizationId}/third-parties`;

  const baseThirdPartyUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/third-parties/${thirdParty.id}`
      : `/organizations/${organizationId}/third-parties/${thirdParty.id}`;

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <Breadcrumb
        items={[
          {
            label: __("Third Parties"),
            to: thirdPartiesUrl,
          },
          {
            label: thirdParty.name ?? "",
          },
        ]}
      />
      <div className="flex justify-between items-start">
        <div className="space-y-4">
          {logo && (
            <img
              src={logo}
              alt={thirdParty.name ?? ""}
              className="shadow-mid rounded-2xl"
            />
          )}
          <div className="text-2xl">{thirdParty.name}</div>
        </div>
        {!isSnapshotMode && (
          <div className="flex gap-2 items-center">
            {thirdParty.canAssess && (
              <ImportAssessmentDialog thirdPartyId={thirdParty.id}>
                <Button icon={IconPageTextLine} variant="secondary">
                  {__("Assessment From Website")}
                </Button>
              </ImportAssessmentDialog>
            )}
            {thirdParty.canDelete && (
              <ActionDropdown variant="secondary">
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={deleteThirdParty}
                >
                  {__("Delete")}
                </DropdownItem>
              </ActionDropdown>
            )}
          </div>
        )}
      </div>

      <Tabs>
        <TabLink to={`${baseThirdPartyUrl}/overview`}>{__("Overview")}</TabLink>
        <TabLink to={`${baseThirdPartyUrl}/certifications`}>
          {__("Certifications")}
        </TabLink>
        <TabLink to={`${baseThirdPartyUrl}/compliance`}>
          {__("Compliance reports")}
          {reportsCount > 0 && <TabBadge>{reportsCount}</TabBadge>}
        </TabLink>
        <TabLink to={`${baseThirdPartyUrl}/risks`}>{__("Risk Assessment")}</TabLink>
        <TabLink to={`${baseThirdPartyUrl}/contacts`}>{__("Contacts")}</TabLink>
        <TabLink to={`${baseThirdPartyUrl}/services`}>{__("Services")}</TabLink>
      </Tabs>

      <Outlet context={{ thirdParty }} />
    </div>
  );
}
