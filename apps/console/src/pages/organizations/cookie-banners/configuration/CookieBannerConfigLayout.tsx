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

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Breadcrumb,
  Button,
  IconImage,
  IconListStack,
  IconPageTextLine,
  IconSettingsGear2,
  PageHeader,
  TabLink,
  Tabs,
  useToast,
} from "@probo/ui";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { Outlet, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannerConfigLayoutActivateMutation } from "#/__generated__/core/CookieBannerConfigLayoutActivateMutation.graphql";
import type { CookieBannerConfigLayoutDeactivateMutation } from "#/__generated__/core/CookieBannerConfigLayoutDeactivateMutation.graphql";
import type { CookieBannerConfigLayoutPublishMutation } from "#/__generated__/core/CookieBannerConfigLayoutPublishMutation.graphql";
import type { CookieBannerConfigLayoutQuery } from "#/__generated__/core/CookieBannerConfigLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const cookieBannerConfigLayoutQuery = graphql`
  query CookieBannerConfigLayoutQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        id
        name
        origin
        state
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

const activateMutation = graphql`
  mutation CookieBannerConfigLayoutActivateMutation($input: ActivateCookieBannerInput!) {
    activateCookieBanner(input: $input) {
      cookieBanner {
        id
        state
      }
    }
  }
`;

const deactivateMutation = graphql`
  mutation CookieBannerConfigLayoutDeactivateMutation($input: DeactivateCookieBannerInput!) {
    deactivateCookieBanner(input: $input) {
      cookieBanner {
        id
        state
      }
    }
  }
`;

const publishMutation = graphql`
  mutation CookieBannerConfigLayoutPublishMutation($input: PublishCookieBannerVersionInput!) {
    publishCookieBannerVersion(input: $input) {
      cookieBannerVersion {
        id
        version
        state
      }
      cookieBanner {
        id
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface CookieBannerConfigLayoutProps {
  queryRef: PreloadedQuery<CookieBannerConfigLayoutQuery>;
}

export default function CookieBannerConfigLayout({ queryRef }: CookieBannerConfigLayoutProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();

  const data = usePreloadedQuery(cookieBannerConfigLayoutQuery, queryRef);
  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;

  const [commitActivate, isActivating] = useMutation<CookieBannerConfigLayoutActivateMutation>(activateMutation);
  const [commitDeactivate, isDeactivating] = useMutation<CookieBannerConfigLayoutDeactivateMutation>(
    deactivateMutation,
  );
  const [commitPublish, isPublishing] = useMutation<CookieBannerConfigLayoutPublishMutation>(publishMutation);

  const handleToggleState = () => {
    if (banner.state === "ACTIVE") {
      commitDeactivate({
        variables: { input: { cookieBannerId: banner.id } },
        onCompleted() {
          toast({ title: __("Success"), description: __("Banner deactivated"), variant: "success" });
        },
        onError(error) {
          toast({ title: __("Error"), description: formatError(__("Failed to deactivate"), error as GraphQLError), variant: "error" });
        },
      });
    } else {
      commitActivate({
        variables: { input: { cookieBannerId: banner.id } },
        onCompleted() {
          toast({ title: __("Success"), description: __("Banner activated"), variant: "success" });
        },
        onError(error) {
          toast({ title: __("Error"), description: formatError(__("Failed to activate"), error as GraphQLError), variant: "error" });
        },
      });
    }
  };

  const handlePublish = () => {
    commitPublish({
      variables: { input: { cookieBannerId: banner.id } },
      onCompleted() {
        toast({ title: __("Success"), description: __("Version published"), variant: "success" });
      },
      onError(error) {
        toast({ title: __("Error"), description: formatError(__("Failed to publish"), error as GraphQLError), variant: "error" });
      },
    });
  };

  const hasDraft = banner.latestVersion?.state === "DRAFT";

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Cookie Banners"),
            to: `/organizations/${organizationId}/cookie-banners`,
          },
          {
            label: banner.name,
          },
        ]}
      />

      <PageHeader
        title={(
          <div className="align-baseline">
            {banner.name}
            {banner.latestVersion?.version != null && (
              <span className="font-mono text-base text-txt-secondary ml-2">
                v
                {banner.latestVersion.version}
                {banner.latestVersion.state === "DRAFT" && (
                  <span className="text-xs font-sans">
                    &nbsp;(draft)
                  </span>
                )}
              </span>
            )}
          </div>
        )}
        description={banner.origin}
      >
        <Badge variant={banner.state === "ACTIVE" ? "success" : "danger"}>
          {banner.state === "ACTIVE" ? __("Active") : __("Inactive")}
        </Badge>
        {hasDraft && (
          <Button onClick={handlePublish} disabled={isPublishing}>
            {isPublishing ? __("Publishing...") : __("Publish Changes")}
          </Button>
        )}
        <Button
          variant="secondary"
          onClick={handleToggleState}
          disabled={isActivating || isDeactivating}
        >
          {banner.state === "ACTIVE" ? __("Deactivate") : __("Activate")}
        </Button>
      </PageHeader>

      <Tabs>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/settings`}>
          <IconSettingsGear2 size={20} />
          {__("Settings")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/cookies`}>
          <IconListStack size={20} />
          {__("Cookies")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/snippet`}>
          <IconPageTextLine size={20} />
          {__("Snippet")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/theme`}>
          <IconImage size={20} />
          {__("Theme")}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
