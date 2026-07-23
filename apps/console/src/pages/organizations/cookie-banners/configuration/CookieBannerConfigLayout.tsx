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

import { ClipboardTextIcon, CodeIcon, MagnifyingGlassIcon } from "@phosphor-icons/react";
import { formatError } from "@probo/helpers";
import {
  Badge,
  Breadcrumb,
  Button,
  IconGlobe,
  IconPageTextLine,
  IconSettingsGear2,
  IconSquareBehindSquare2,
  PageHeader,
  TabLink,
  Tabs,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { Link, Outlet, useParams } from "react-router";
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
        policyDocument {
          id
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
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();

  const data = usePreloadedQuery<CookieBannerConfigLayoutQuery>(cookieBannerConfigLayoutQuery, queryRef);
  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;

  const [activate, isActivating] = useMutation<CookieBannerConfigLayoutActivateMutation>(activateMutation);
  const [deactivate, isDeactivating] = useMutation<CookieBannerConfigLayoutDeactivateMutation>(
    deactivateMutation,
  );
  const [publish, isPublishing] = useMutation<CookieBannerConfigLayoutPublishMutation>(publishMutation);

  const handleToggleState = () => {
    if (banner.state === "ACTIVE") {
      deactivate({
        variables: { input: { cookieBannerId: banner.id } },
        onCompleted() {
          toast({ title: t("configLayout.messages.successTitle"), description: t("configLayout.messages.deactivated"), variant: "success" });
        },
        onError(error) {
          toast({ title: t("configLayout.errors.title"), description: formatError(t("configLayout.errors.deactivate"), error), variant: "error" });
        },
      });
    } else {
      activate({
        variables: { input: { cookieBannerId: banner.id } },
        onCompleted() {
          toast({ title: t("configLayout.messages.successTitle"), description: t("configLayout.messages.activated"), variant: "success" });
        },
        onError(error) {
          toast({ title: t("configLayout.errors.title"), description: formatError(t("configLayout.errors.activate"), error), variant: "error" });
        },
      });
    }
  };

  const handlePublish = () => {
    publish({
      variables: { input: { cookieBannerId: banner.id } },
      onCompleted() {
        toast({ title: t("configLayout.messages.successTitle"), description: t("configLayout.messages.published"), variant: "success" });
      },
      onError(error) {
        toast({ title: t("configLayout.errors.title"), description: formatError(t("configLayout.errors.publish"), error), variant: "error" });
      },
    });
  };

  const hasDraft = banner.latestVersion?.state === "DRAFT";

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("configLayout.breadcrumbs.index"),
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
                {t("configLayout.version", { version: banner.latestVersion.version })}
                {banner.latestVersion.state === "DRAFT" && (
                  <span className="text-xs font-sans">
                    {t("configLayout.draft")}
                  </span>
                )}
              </span>
            )}
          </div>
        )}
        description={(
          <span className="flex items-center gap-3 text-sm text-txt-secondary">
            <span>
              <span className="font-medium text-txt-primary">{t("configLayout.metadata.origin")}</span>
              {" "}
              {banner.origin}
            </span>
            <span className="text-border-primary">·</span>
            <span className="flex items-center gap-1">
              <span className="font-medium text-txt-primary">{t("configLayout.metadata.id")}</span>
              {" "}
              {banner.id}
              <button
                type="button"
                className="p-1 rounded hover:bg-bg-hover transition-colors cursor-pointer"
                onClick={() => {
                  void navigator.clipboard.writeText(banner.id);
                  toast({ title: t("configLayout.messages.copiedTitle"), description: t("configLayout.messages.idCopied"), variant: "success" });
                }}
              >
                <IconSquareBehindSquare2 size={16} />
              </button>
            </span>
            {banner.policyDocument && (
              <>
                <span className="text-border-primary">·</span>
                <Link
                  to={`/organizations/${organizationId}/documents/${banner.policyDocument.id}`}
                  className="font-medium text-txt-primary underline"
                >
                  {t("configLayout.metadata.cookiePolicy")}
                </Link>
              </>
            )}
          </span>
        )}
      >
        <Badge variant={banner.state === "ACTIVE" ? "success" : "danger"}>
          {banner.state === "ACTIVE" ? t("configLayout.status.active") : t("configLayout.status.inactive")}
        </Badge>
        {hasDraft && (
          <Button onClick={handlePublish} disabled={isPublishing}>
            {isPublishing ? t("configLayout.actions.publishing") : t("configLayout.actions.publish")}
          </Button>
        )}
        <Button
          variant="secondary"
          onClick={handleToggleState}
          disabled={isActivating || isDeactivating}
        >
          {banner.state === "ACTIVE" ? t("configLayout.actions.deactivate") : t("configLayout.actions.activate")}
        </Button>
      </PageHeader>

      <Tabs>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/display`}>
          <IconPageTextLine size={20} />
          {t("configLayout.tabs.display")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/settings`}>
          <IconSettingsGear2 size={20} />
          {t("configLayout.tabs.settings")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/translations`}>
          <IconGlobe size={20} />
          {t("configLayout.tabs.translations")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/trackers`}>
          <MagnifyingGlassIcon size={20} />
          {t("configLayout.tabs.trackers")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/resources`}>
          <CodeIcon size={20} />
          {t("configLayout.tabs.resources")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/cookie-banners/${cookieBannerId}/consent-records`}>
          <ClipboardTextIcon size={20} />
          {t("configLayout.tabs.consentRecords")}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
