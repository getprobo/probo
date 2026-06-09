// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { faviconUrl } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  DropdownItem,
  IconPageTextLine,
  IconTrashCan,
  TabBadge,
  TabLink,
  Tabs,
} from "@probo/ui";
import { useEffect, useRef } from "react";
import {
  ConnectionHandler,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
  useRelayEnvironment,
} from "react-relay";
import { Outlet } from "react-router";
import { fetchQuery } from "relay-runtime";

import type { ThirdPartyComplianceTabFragment$key } from "#/__generated__/core/ThirdPartyComplianceTabFragment.graphql";
import type { ThirdPartyGraphNodeQuery } from "#/__generated__/core/ThirdPartyGraphNodeQuery.graphql";
import {
  thirdPartyConnectionKey,
  thirdPartyNodeQuery,
  useDeleteThirdParty,
} from "#/hooks/graph/ThirdPartyGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { VettingDialog } from "./dialogs/VettingDialog";
import { measuresFragment } from "./measures/ThirdPartyMeasuresPage";
import { complianceReportsFragment } from "./tabs/ThirdPartyComplianceTab";

void measuresFragment;

type Props = {
  queryRef: PreloadedQuery<ThirdPartyGraphNodeQuery>;
};

export default function ThirdPartyDetailPage(props: Props) {
  const environment = useRelayEnvironment();
  const { node: thirdParty } = usePreloadedQuery(thirdPartyNodeQuery, props.queryRef);
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const thirdPartyIdRef = useRef(thirdParty.id);

  useEffect(() => {
    thirdPartyIdRef.current = thirdParty.id;
  }, [thirdParty.id]);

  const isVetting = thirdParty.vettingStatus === "PENDING" || thirdParty.vettingStatus === "PROCESSING";

  useEffect(() => {
    if (!isVetting) return;

    const interval = setInterval(() => {
      if (document.hidden) return;

      fetchQuery<ThirdPartyGraphNodeQuery>(
        environment,
        thirdPartyNodeQuery,
        { thirdPartyId: thirdPartyIdRef.current },
        { fetchPolicy: "network-only" },
      ).subscribe({});
    }, 5000);

    return () => clearInterval(interval);
  }, [isVetting, environment]);

  const deleteThirdParty = useDeleteThirdParty(
    thirdParty,
    ConnectionHandler.getConnectionID(organizationId, thirdPartyConnectionKey),
  );
  const logo = faviconUrl(thirdParty.websiteUrl);
  const reportsCount = useFragment(
    complianceReportsFragment,
    thirdParty as ThirdPartyComplianceTabFragment$key,
  ).complianceReports.edges.length;
  const measuresCount = thirdParty.measuresInfos?.totalCount ?? 0;

  const thirdPartiesUrl = `/organizations/${organizationId}/third-parties`;

  const baseThirdPartyUrl
    = `/organizations/${organizationId}/third-parties/${thirdParty.id}`;

  const isVettingFailed = thirdParty.vettingStatus === "FAILED";

  return (
    <div className="space-y-6">
      {isVetting && (
        <div className="flex items-center gap-3 rounded-lg bg-warning px-4 py-3 text-sm text-txt-warning">
          <div
            aria-hidden
            className="size-4 shrink-0 animate-spin rounded-full border-2 border-border-warning/30 border-t-border-warning"
          />
          {__("Vetting is in progress. Results will appear once the analysis is complete.")}
        </div>
      )}
      {isVettingFailed && (
        <div className="rounded-lg bg-danger px-4 py-3 text-sm text-txt-danger">
          {__("Vetting failed. You can start vetting again.")}
        </div>
      )}
      <Breadcrumb
        items={[
          {
            label: __("Third parties"),
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
          <div className="flex items-center gap-3">
            <div className="text-2xl">{thirdParty.name}</div>
            <Badge variant={thirdParty.firstLevel ? "info" : "neutral"}>
              {thirdParty.firstLevel ? __("First Level") : __("Indirect")}
            </Badge>
          </div>
        </div>
        <div className="flex gap-2 items-center">
          {thirdParty.canVet && !isVetting && (
            <VettingDialog
              thirdPartyId={thirdParty.id}
              websiteUrl={thirdParty.websiteUrl}
            >
              <Button icon={IconPageTextLine} variant="secondary">
                {__("Start Vetting")}
              </Button>
            </VettingDialog>
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
        <TabLink to={`${baseThirdPartyUrl}/third-parties`}>
          {__("Third Parties")}
        </TabLink>
        <TabLink to={`${baseThirdPartyUrl}/measures`}>
          {__("Measures")}
          {measuresCount > 0 && <TabBadge>{measuresCount}</TabBadge>}
        </TabLink>
      </Tabs>

      <Outlet context={{ thirdParty }} />
    </div>
  );
}
