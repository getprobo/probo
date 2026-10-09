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

import { CheckCircleIcon, CopyIcon, WarningIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannerConfigLayoutPublishMutation } from "#/__generated__/core/CookieBannerConfigLayoutPublishMutation.graphql";
import type { CookieBannerConfigLayoutQuery } from "#/__generated__/core/CookieBannerConfigLayoutQuery.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useMutation } from "#/lib/relay/useMutation";

import { cookieBannerConfigLayout } from "../variants";

import { DiscoveryFamilyCounts } from "./_components/DiscoveryFamilyCounts";

export const cookieBannerConfigLayoutQuery = graphql`
  query CookieBannerConfigLayoutQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        id
        name
        state
        latestVersion {
          id
          version
          state
        }
        publishedVersion {
          id
        }
        discoveryPageLoads {
          family
          count
        }
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
        publishedVersion {
          id
          gvlVendorCount
          gvlVendorIds
        }
      }
    }
  }
`;

interface CookieBannerConfigLayoutProps {
  queryRef: PreloadedQuery<CookieBannerConfigLayoutQuery>;
}

export function CookieBannerConfigLayout({
  queryRef,
}: CookieBannerConfigLayoutProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();
  const {
    root,
    lead,
    title,
    meta,
    id,
    version: versionClass,
    discovery,
    discoveryRow,
  } = cookieBannerConfigLayout();

  const data = usePreloadedQuery<CookieBannerConfigLayoutQuery>(
    cookieBannerConfigLayoutQuery,
    queryRef,
  );
  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;
  usePageTitle(banner.name);

  const [publish, isPublishing] = useMutation<CookieBannerConfigLayoutPublishMutation>(
    publishMutation,
    {
      successMessage: t("configLayout.messages.published"),
      errorToast: t("configLayout.errors.publish"),
    },
  );

  const hasDraft = banner.latestVersion?.state === "DRAFT";
  const isDeactivated = banner.state !== "ACTIVE";
  const version = banner.latestVersion?.version;
  const pageLoads = banner.discoveryPageLoads;
  const showDiscoveryLoads = banner.publishedVersion == null;

  let message: string;
  if (isDeactivated && hasDraft && version != null) {
    message = t("configLayout.callout.deactivatedDraft", { version });
  } else if (isDeactivated) {
    message = t("configLayout.callout.deactivated");
  } else if (hasDraft && version != null) {
    message = t("configLayout.callout.draft", { version });
  } else if (version != null) {
    message = t("configLayout.callout.published", { version });
  } else {
    message = t("configLayout.callout.publishedUnknown");
  }

  const tone = isDeactivated || hasDraft ? "amber" : "green";
  const icon = isDeactivated || hasDraft
    ? <WarningIcon size={24} weight="duotone" />
    : <CheckCircleIcon size={24} weight="duotone" />;

  function handleCopyId() {
    void navigator.clipboard.writeText(banner.id).then(
      () => {
        toast({
          title: t("configLayout.messages.copiedTitle"),
          description: t("configLayout.messages.idCopied"),
          variant: "success",
        });
      },
      () => {
        toast({
          title: t("configLayout.errors.title"),
          description: t("configLayout.errors.copy"),
          variant: "error",
        });
      },
    );
  }

  return (
    <div className={root()}>
      <TonedCard
        tone={tone}
        icon={icon}
        lead={(
          <div className={lead()}>
            <Heading
              level={2}
              size={4}
              weight="medium"
              highContrast
              className={title()}
            >
              {banner.name}
            </Heading>
            <div className={meta()}>
              <Text size={2} color="neutral" className={id()}>
                {banner.id}
              </Text>
              <IconButton
                size={1}
                variant="ghost"
                color="neutral"
                aria-label={t("configLayout.actions.copyId")}
                onClick={handleCopyId}
              >
                <CopyIcon />
              </IconButton>
              {version != null && (
                <Text size={2} color="neutral" className={versionClass()}>
                  {t("configLayout.callout.version", { version })}
                </Text>
              )}
            </div>
          </div>
        )}
        control={hasDraft
          ? (
              <Button
                size={2}
                variant="solid"
                color="neutral"
                highContrast
                loading={isPublishing}
                onClick={() => {
                  void publish({
                    variables: { input: { cookieBannerId: banner.id } },
                  }).catch(() => {
                    // Error toast is already shown by useMutation.
                  });
                }}
              >
                {t("configLayout.actions.publish")}
              </Button>
            )
          : undefined}
      >
        <div className={discovery()}>
          <Text size={2} color="neutral">{message}</Text>
          {showDiscoveryLoads
            ? (
                <div className={discoveryRow()}>
                  <Text size={2} color="faint">
                    {t("configLayout.discovery.pageLoads")}
                  </Text>
                  {pageLoads.length > 0
                    ? <DiscoveryFamilyCounts items={pageLoads} compact />
                    : (
                        <Text size={2} color="faint">
                          {t("configLayout.discovery.empty")}
                        </Text>
                      )}
                </div>
              )
            : null}
        </div>
      </TonedCard>
      <Outlet />
    </div>
  );
}
