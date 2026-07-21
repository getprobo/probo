// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { faviconUrl, formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
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
  useConfirm,
  useToast,
} from "@probo/ui";
import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
  useRelayEnvironment,
} from "react-relay";
import { Link, Outlet, useNavigate, useParams } from "react-router";
import { ConnectionHandler, fetchQuery } from "relay-runtime";

import type { ThirdPartyDetailLayoutDeleteMutation } from "#/__generated__/core/ThirdPartyDetailLayoutDeleteMutation.graphql";
import type { ThirdPartyDetailLayoutQuery } from "#/__generated__/core/ThirdPartyDetailLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { VettingDialog } from "./_components/VettingDialog";
import { ThirdPartiesConnectionFilter, ThirdPartiesConnectionKey } from "./ThirdPartiesPage";

export const thirdPartyDetailLayoutQuery = graphql`
  query ThirdPartyDetailLayoutQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        id
        name
        websiteUrl
        level
        ancestors {
          id
          name
        }
        vettingStatus
        canVet: permission(action: "core:thirdParty:vet")
        canDelete: permission(action: "core:thirdParty:delete")
        complianceReportsInfo: complianceReports(first: 100) {
          edges {
            node {
              id
            }
          }
        }
        measuresInfo: measures(first: 0) {
          totalCount
        }
      }
    }
  }
`;

const deleteThirdPartyMutation = graphql`
  mutation ThirdPartyDetailLayoutDeleteMutation(
    $input: DeleteThirdPartyInput!
    $connections: [ID!]!
  ) {
    deleteThirdParty(input: $input) {
      deletedThirdPartyId @deleteEdge(connections: $connections)
    }
  }
`;

interface ThirdPartyDetailLayoutProps {
  queryRef: PreloadedQuery<ThirdPartyDetailLayoutQuery>;
}

export default function ThirdPartyDetailLayout(props: ThirdPartyDetailLayoutProps) {
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();
  const environment = useRelayEnvironment();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const confirm = useConfirm();
  const { toast } = useToast();

  if (!thirdPartyId) {
    throw new Error("Cannot load third party detail layout without thirdPartyId parameter");
  }

  const data = usePreloadedQuery<ThirdPartyDetailLayoutQuery>(thirdPartyDetailLayoutQuery, props.queryRef);
  if (data.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }
  const thirdParty = data.node;

  const thirdPartyIdRef = useRef(thirdParty.id);
  useEffect(() => {
    thirdPartyIdRef.current = thirdParty.id;
  }, [thirdParty.id]);

  const isVetting = thirdParty.vettingStatus === "PENDING"
    || thirdParty.vettingStatus === "PROCESSING";

  useEffect(() => {
    if (!isVetting) {
      return;
    }

    const interval = setInterval(() => {
      if (document.hidden) {
        return;
      }

      fetchQuery(
        environment,
        thirdPartyDetailLayoutQuery,
        { thirdPartyId: thirdPartyIdRef.current },
        { fetchPolicy: "network-only" },
      ).subscribe({});
    }, 5000);

    return () => clearInterval(interval);
  }, [isVetting, environment]);

  const [deleteThirdParty] = useMutation<ThirdPartyDetailLayoutDeleteMutation>(
    deleteThirdPartyMutation,
  );

  usePageTitle(thirdParty.name ?? t("thirdPartyDetailLayout.title"));

  const onDelete = () => {
    if (!thirdParty.name) {
      return;
    }
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      ThirdPartiesConnectionKey,
      { filter: ThirdPartiesConnectionFilter },
    );
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteThirdParty({
            variables: {
              input: { thirdPartyId: thirdParty.id },
              connections: [connectionId],
            },
            onCompleted() {
              void navigate(`/organizations/${organizationId}/third-parties`);
              resolve();
            },
            onError(error) {
              toast({
                title: t("thirdPartyDetailLayout.messages.error"),
                description: formatError(
                  t("thirdPartyDetailLayout.errors.delete"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("thirdPartyDetailLayout.deleteConfirmation", { name: thirdParty.name }),
      },
    );
  };

  const logo = faviconUrl(thirdParty.websiteUrl);
  const reportsCount = thirdParty.complianceReportsInfo?.edges.length ?? 0;
  const measuresCount = thirdParty.measuresInfo?.totalCount ?? 0;
  const isVettingFailed = thirdParty.vettingStatus === "FAILED";
  const ancestors = thirdParty.ancestors ?? [];

  const thirdPartiesUrl = `/organizations/${organizationId}/third-parties`;
  const baseThirdPartyUrl
    = `/organizations/${organizationId}/third-parties/${thirdParty.id}`;

  return (
    <div className="space-y-6">
      {isVetting && (
        <div className="flex items-center gap-3 rounded-lg bg-warning px-4 py-3 text-sm text-txt-warning">
          <div
            aria-hidden
            className="size-4 shrink-0 animate-spin rounded-full border-2 border-border-warning/30 border-t-border-warning"
          />
          {t("thirdPartyDetailLayout.vettingInProgress")}
        </div>
      )}
      {isVettingFailed && (
        <div className="rounded-lg bg-danger px-4 py-3 text-sm text-txt-danger">
          {t("thirdPartyDetailLayout.vettingFailed")}
        </div>
      )}
      <Breadcrumb
        items={[
          {
            label: t("thirdPartyDetailLayout.breadcrumb.thirdParties"),
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
            <Badge variant={thirdParty.level === 1 ? "info" : "neutral"}>
              {t("thirdPartyDetailLayout.level", { level: thirdParty.level })}
            </Badge>
          </div>
          {ancestors.length > 0 && (
            <div className="flex items-center gap-1 text-sm text-txt-secondary">
              <span className="text-txt-tertiary">{t("thirdPartyDetailLayout.from")}</span>
              {ancestors.map((ancestor, i) => (
                <span key={ancestor.id}>
                  {i > 0 && " / "}
                  <Link
                    to={`/organizations/${organizationId}/third-parties/${ancestor.id}/overview`}
                    className="text-txt-primary underline hover:no-underline"
                  >
                    {ancestor.name}
                  </Link>
                </span>
              ))}
            </div>
          )}
        </div>
        <div className="flex gap-2 items-center">
          {thirdParty.canVet && !isVetting && (
            <VettingDialog
              thirdPartyId={thirdParty.id}
              websiteUrl={thirdParty.websiteUrl}
            >
              <Button icon={IconPageTextLine} variant="secondary">
                {t("thirdPartyDetailLayout.actions.startVetting")}
              </Button>
            </VettingDialog>
          )}
          {thirdParty.canDelete && (
            <ActionDropdown variant="secondary">
              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                onClick={onDelete}
              >
                {t("thirdPartyDetailLayout.actions.delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </div>
      </div>

      <Tabs>
        <TabLink to={`${baseThirdPartyUrl}/overview`}>{t("thirdPartyDetailLayout.tabs.overview")}</TabLink>
        <TabLink to={`${baseThirdPartyUrl}/certifications`}>
          {t("thirdPartyDetailLayout.tabs.certifications")}
        </TabLink>
        <TabLink to={`${baseThirdPartyUrl}/compliance`}>
          {t("thirdPartyDetailLayout.tabs.complianceReports")}
          {reportsCount > 0 && <TabBadge>{reportsCount}</TabBadge>}
        </TabLink>
        <TabLink to={`${baseThirdPartyUrl}/risks`}>{t("thirdPartyDetailLayout.tabs.riskAssessment")}</TabLink>
        <TabLink to={`${baseThirdPartyUrl}/contacts`}>{t("thirdPartyDetailLayout.tabs.contacts")}</TabLink>
        <TabLink to={`${baseThirdPartyUrl}/services`}>{t("thirdPartyDetailLayout.tabs.services")}</TabLink>
        <TabLink to={`${baseThirdPartyUrl}/third-parties`}>
          {t("thirdPartyDetailLayout.tabs.thirdParties")}
        </TabLink>
        <TabLink to={`${baseThirdPartyUrl}/measures`}>
          {t("thirdPartyDetailLayout.tabs.measures")}
          {measuresCount > 0 && <TabBadge>{measuresCount}</TabBadge>}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
