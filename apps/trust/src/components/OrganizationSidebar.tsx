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

import { detectSocialName, externalLinkProps, formatError } from "@probo/helpers";
import { useSystemTheme } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { UnAuthenticatedError } from "@probo/relay";
import {
  Button,
  Card,
  IconBell2,
  IconBlock,
  IconLock,
  IconMedal,
  SocialIcon,
  useToast,
} from "@probo/ui";
import { type PropsWithChildren } from "react";
import { useMutation } from "react-relay";
import { useLocation, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { TrustGraphCurrentQuery$data } from "#/queries/__generated__/TrustGraphCurrentQuery.graphql";
import { getPathPrefix } from "#/utils/pathPrefix";

import type { OrganizationSidebar_requestAllAccessesMutation } from "./__generated__/OrganizationSidebar_requestAllAccessesMutation.graphql";
import type { OrganizationSidebar_subscribeToMailingListMutation } from "./__generated__/OrganizationSidebar_subscribeToMailingListMutation.graphql";
import type { OrganizationSidebar_unsubscribeFromMailingListMutation } from "./__generated__/OrganizationSidebar_unsubscribeFromMailingListMutation.graphql";
import { FrameworkBadge } from "./FrameworkBadge";

const requestAllAccessesMutation = graphql`
  mutation OrganizationSidebar_requestAllAccessesMutation {
    requestAllAccesses {
      trustCenterAccess {
        id
      }
    }
  }
`;

const subscribeToMailingListMutation = graphql`
  mutation OrganizationSidebar_subscribeToMailingListMutation {
    subscribeToMailingList {
      subscription {
        id
        email
        createdAt
        updatedAt
      }
    }
  }
`;

const unsubscribeFromMailingListMutation = graphql`
  mutation OrganizationSidebar_unsubscribeFromMailingListMutation {
    unsubscribeFromMailingList {
      deletedMailingListSubscriberId @deleteRecord
    }
  }
`;

export function OrganizationSidebar({
  trustCenter,
  isAuthenticated,
}: {
  trustCenter: TrustGraphCurrentQuery$data["currentTrustCenter"];
  isAuthenticated: boolean;
}) {
  const trustCenterId = trustCenter?.id;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const theme = useSystemTheme();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();

  const logoFileUrl = theme === "dark" ? (trustCenter?.darkLogoFileUrl ?? trustCenter?.logoFileUrl) : trustCenter?.logoFileUrl;

  const [requestAllAccesses, isRequestingAccess]
    = useMutation<OrganizationSidebar_requestAllAccessesMutation>(
      requestAllAccessesMutation,
    );

  const [subscribeToMailingList, isSubscribing]
    = useMutation<OrganizationSidebar_subscribeToMailingListMutation>(
      subscribeToMailingListMutation,
    );

  const [unsubscribeFromMailingList, isUnsubscribing]
    = useMutation<OrganizationSidebar_unsubscribeFromMailingListMutation>(
      unsubscribeFromMailingListMutation,
    );

  const handleRequestAllAccesses = () => {
    requestAllAccesses({
      variables: {},
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Cannot request access"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Access request submitted successfully."),
          variant: "success",
        });

        window.location.href = location.pathname;
      },
      onError: (error) => {
        if (error instanceof UnAuthenticatedError) {
          const pathPrefix = getPathPrefix();
          searchParams.set("request-all", "true");
          const urlSearchParams = new URLSearchParams([[
            "continue",
            window.location.origin + pathPrefix + location.pathname + "?" + searchParams.toString(),
          ]]);
          void navigate(`/connect?${urlSearchParams.toString()}`);

          return;
        }

        toast({
          title: __("Error"),
          description: error.message ?? __("Cannot request access"),
          variant: "error",
        });
      },
    });
  };

  const handleSubscribe = () => {
    subscribeToMailingList({
      variables: {},
      updater: (store, data) => {
        const subscription = data?.subscribeToMailingList?.subscription;
        if (!subscription?.id || !trustCenterId) return;
        const trustCenterRecord = store.get(trustCenterId);
        if (!trustCenterRecord) return;
        const subscriptionRecord = store.get(subscription.id);
        if (!subscriptionRecord) return;
        trustCenterRecord.setLinkedRecord(subscriptionRecord, "viewerSubscription");
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Cannot subscribe"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Subscribed"),
          description: __("You will be notified of security updates."),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: error.message ?? __("Cannot subscribe"),
          variant: "error",
        });
      },
    });
  };

  const handleUnsubscribe = () => {
    unsubscribeFromMailingList({
      variables: {},
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Cannot unsubscribe"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Unsubscribed"),
          description: __("You will no longer receive security updates."),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: error.message ?? __("Cannot unsubscribe"),
          variant: "error",
        });
      },
    });
  };

  if (!trustCenter) {
    return null;
  }

  return (
    <Card className="p-6 relative overflow-hidden border-b border-border-low isolate">
      <div className="h-21 bg-[#044E4114] absolute top-0 left-0 right-0 -z-1"></div>
      {logoFileUrl
        ? (
            <img
              alt=""
              src={logoFileUrl}
              className="size-24 rounded-2xl border border-border-mid shadow-mid bg-level-1"
            />
          )
        : (
            <div className="size-24 rounded-2xl border border-border-mid shadow-mid bg-level-1" />
          )}
      <h1 className="text-2xl mt-6">{trustCenter.organization.name}</h1>
      <p className="text-sm text-txt-secondary mt-1">
        {trustCenter.organization.description}
      </p>

      <hr className="my-6 -mx-6 h-px bg-border-low border-none" />

      {/* Business information */}
      <div className="space-y-4">
        <h2 className="text-xs text-txt-secondary flex gap-1 items-center">
          <IconBlock size={16} />
          {__("Business information")}
        </h2>
        {trustCenter.organization.websiteUrl && (
          <BusinessInfo label={__("Website")}>
            <a {...externalLinkProps(trustCenter.organization.websiteUrl)}>
              <span className="text-txt-info hover:underline ">
                {new URL(trustCenter.organization.websiteUrl).host}
              </span>
            </a>
          </BusinessInfo>
        )}
        {trustCenter.organization.email && (
          <BusinessInfo label={__("Contact")}>
            <a href={`mailto:${trustCenter.organization.email}`}>
              <span className="text-txt-info hover:underline ">
                {trustCenter.organization.email}
              </span>
            </a>
          </BusinessInfo>
        )}
        {trustCenter.organization.headquarterAddress && (
          <BusinessInfo label={__("HQ address")}>
            {trustCenter.organization.headquarterAddress}
          </BusinessInfo>
        )}
        {trustCenter.externalUrls.edges.length > 0 && (
          <div className="flex flex-wrap gap-x-4 gap-y-2">
            {trustCenter.externalUrls.edges.map(({ node }) => (
              <a
                key={node.id}
                {...externalLinkProps(node.url)}
                className="flex items-center gap-1.5 text-sm text-txt-secondary hover:text-txt-primary transition-colors"
              >
                <SocialIcon socialName={detectSocialName(node.url)} size={14} className="shrink-0" />
                <span>{node.name}</span>
              </a>
            ))}
          </div>
        )}

        <hr className="my-6 -mx-6 h-px bg-border-low border-none" />

        {/* Certifications */}
        {trustCenter.complianceFrameworks.edges.length > 0 && (
          <>
            <div className="space-y-4">
              <h2 className="text-xs text-txt-secondary flex gap-1 items-center">
                <IconMedal size={16} />
                {__("Frameworks")}
              </h2>
              <div
                className="grid grid-cols-4 gap-4"
                style={{
                  gridTemplateColumns: "repeat(auto-fit, 75px",
                }}
              >
                {trustCenter.complianceFrameworks.edges.map(edge => (
                  <FrameworkBadge key={edge.node.id} framework={edge.node.framework} />
                ))}
              </div>
            </div>

            <hr className="my-6 -mx-6 h-px bg-border-low border-none" />
          </>
        )}

        {/* Actions */}
        <div className="space-y-2">
          <Button
            disabled={isRequestingAccess}
            variant="primary"
            icon={IconLock}
            className="w-full h-10"
            onClick={handleRequestAllAccesses}
          >
            {__("Request access")}
          </Button>
          {isAuthenticated && (
            trustCenter.viewerSubscription
              ? (
                  <Button
                    disabled={isUnsubscribing}
                    variant="secondary"
                    icon={IconBell2}
                    className="w-full h-10"
                    onClick={handleUnsubscribe}
                  >
                    {__("Unsubscribe from updates")}
                  </Button>
                )
              : (
                  <Button
                    disabled={isSubscribing}
                    variant="secondary"
                    icon={IconBell2}
                    className="w-full h-10"
                    onClick={handleSubscribe}
                  >
                    {__("Subscribe to updates")}
                  </Button>
                )
          )}
        </div>
      </div>
    </Card>
  );
}

function BusinessInfo({
  children,
  label,
}: PropsWithChildren<{ label: string }>) {
  return (
    <div>
      <div className="text-xs text-txt-secondary">{label}</div>
      <div className="text-sm">{children}</div>
    </div>
  );
}
