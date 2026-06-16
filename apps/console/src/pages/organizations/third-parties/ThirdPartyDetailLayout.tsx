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

import { faviconUrl, formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
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
  useConfirm,
  useToast,
} from "@probo/ui";
import { useEffect, useRef } from "react";
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
import { ThirdPartiesConnectionKey } from "./ThirdPartiesPage";

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
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const { toast } = useToast();

  if (!thirdPartyId) {
    throw new Error("Cannot load third party detail layout without thirdPartyId parameter");
  }

  const data = usePreloadedQuery(thirdPartyDetailLayoutQuery, props.queryRef);
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

  usePageTitle(thirdParty.name ?? __("Third party"));

  const onDelete = () => {
    if (!thirdParty.name) {
      return;
    }
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      ThirdPartiesConnectionKey,
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
                title: __("Error"),
                description: formatError(
                  __("Failed to delete third party"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete thirdParty \"%s\". This action cannot be undone.",
          ),
          thirdParty.name,
        ),
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
            <Badge variant={thirdParty.level === 1 ? "info" : "neutral"}>
              {`${__("Level")} ${thirdParty.level}`}
            </Badge>
          </div>
          {ancestors.length > 0 && (
            <div className="flex items-center gap-1 text-sm text-txt-secondary">
              <span className="text-txt-tertiary">{__("From:")}</span>
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
                {__("Start Vetting")}
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

      <Outlet />
    </div>
  );
}
