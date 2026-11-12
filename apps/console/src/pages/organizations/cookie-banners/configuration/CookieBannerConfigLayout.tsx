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
import { Button } from "@probo/ui/src/v2/Button/Button";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Fragment, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannerConfigLayout_cookieBanner$key } from "#/__generated__/core/CookieBannerConfigLayout_cookieBanner.graphql";
import type { CookieBannerConfigLayout_pageLoad$key } from "#/__generated__/core/CookieBannerConfigLayout_pageLoad.graphql";
import type { CookieBannerConfigLayoutPublishMutation } from "#/__generated__/core/CookieBannerConfigLayoutPublishMutation.graphql";
import type { CookieBannerConfigLayoutQuery } from "#/__generated__/core/CookieBannerConfigLayoutQuery.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import type { TonedCardTone } from "#/components/TonedCard/variants";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";

import { cookieBannerConfigLayout } from "../variants";

import { FamilyLogo } from "./_components/FamilyLogo";
import { familyLabel } from "./_lib/familyLabel";

const pageLoadFragment = graphql`
  fragment CookieBannerConfigLayout_pageLoad on BrowserFamilyCount {
    family
    count
  }
`;

const cookieBannerFragment = graphql`
  fragment CookieBannerConfigLayout_cookieBanner on CookieBanner {
    id
    name
    state
    latestVersion {
      id
      version
      state
    }
    pageLoads(days: 7) {
      ...CookieBannerConfigLayout_pageLoad
    }
  }
`;

export const cookieBannerConfigLayoutQuery = graphql`
  query CookieBannerConfigLayoutQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        ...CookieBannerConfigLayout_cookieBanner
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
        ...CookieBannerConfigLayout_cookieBanner
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
  const toast = useToast();
  const {
    root,
    lead,
    title,
    meta,
    id,
    version: versionClass,
    body,
    loadsRow,
    loadsItem,
    loadsLogo,
    loadsSeparator,
  } = cookieBannerConfigLayout();

  const data = usePreloadedQuery<CookieBannerConfigLayoutQuery>(
    cookieBannerConfigLayoutQuery,
    queryRef,
  );
  const banner = useFragment<CookieBannerConfigLayout_cookieBanner$key>(
    cookieBannerFragment,
    data.node.__typename === "CookieBanner" ? data.node : null,
  );
  if (banner == null) {
    throw new NotFoundError("Cookie banner not found");
  }
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

  let message: string;
  if (isDeactivated && hasDraft && version != null) {
    message = t("configLayout.callout.deactivatedDraft", { version });
  } else if (isDeactivated) {
    message = t("configLayout.callout.deactivated");
  } else if (hasDraft && version != null) {
    message = t("configLayout.callout.draft");
  } else if (version != null) {
    message = t("configLayout.callout.published");
  } else {
    message = t("configLayout.callout.unpublished");
  }

  let tone: TonedCardTone;
  let icon: ReactNode;
  if (isDeactivated) {
    tone = "amber";
    icon = <WarningIcon size={24} weight="duotone" />;
  } else if (hasDraft) {
    tone = "amber";
    icon = <WarningIcon size={24} weight="duotone" />;
  } else {
    tone = "green";
    icon = <CheckCircleIcon size={24} weight="duotone" />;
  }

  const bannerId = banner.id;

  function handleCopyId() {
    void navigator.clipboard.writeText(bannerId).then(
      () => {
        toast.add({
          title: t("configLayout.messages.copiedTitle"),
          description: t("configLayout.messages.idCopied"),
          type: "success",
        });
      },
      () => {
        toast.add({
          title: t("configLayout.errors.title"),
          description: t("configLayout.errors.copy"),
          type: "error",
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
          : null}
      >
        <div className={body()}>
          <Text size={2} color="neutral">{message}</Text>
          <PageLoadRow
            label={t("configLayout.pageLoads.label")}
            pageLoads={banner.pageLoads}
            empty={t("configLayout.pageLoads.empty")}
            loadsRow={loadsRow()}
            loadsItem={loadsItem()}
            loadsLogo={loadsLogo()}
            loadsSeparator={loadsSeparator()}
          />
        </div>
      </TonedCard>
      <Outlet />
    </div>
  );
}

interface PageLoadRowProps {
  label: string;
  pageLoads: readonly CookieBannerConfigLayout_pageLoad$key[];
  empty: string;
  loadsRow: string;
  loadsItem: string;
  loadsLogo: string;
  loadsSeparator: string;
}

function PageLoadRow({
  label,
  pageLoads,
  empty,
  loadsRow,
  loadsItem,
  loadsLogo,
  loadsSeparator,
}: PageLoadRowProps) {
  return (
    <div className={loadsRow}>
      <Text size={2} color="faint">
        {label}
      </Text>
      {pageLoads.length > 0
        ? pageLoads.map((pageLoad, index) => (
            <PageLoadItem
              key={index}
              pageLoadKey={pageLoad}
              showSeparator={index > 0}
              loadsItem={loadsItem}
              loadsLogo={loadsLogo}
              loadsSeparator={loadsSeparator}
            />
          ))
        : (
            <Text size={2} color="faint">
              {empty}
            </Text>
          )}
    </div>
  );
}

interface PageLoadItemProps {
  pageLoadKey: CookieBannerConfigLayout_pageLoad$key;
  showSeparator: boolean;
  loadsItem: string;
  loadsLogo: string;
  loadsSeparator: string;
}

function PageLoadItem({
  pageLoadKey,
  showSeparator,
  loadsItem,
  loadsLogo,
  loadsSeparator,
}: PageLoadItemProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const load = useFragment(pageLoadFragment, pageLoadKey);

  return (
    <Fragment>
      {showSeparator
        ? (
            <Text
              size={2}
              color="faint"
              className={loadsSeparator}
            >
              ·
            </Text>
          )
        : null}
      <div className={loadsItem}>
        <FamilyLogo
          family={load.family}
          size={16}
          className={loadsLogo}
        />
        <Text size={2} color="faint">
          {familyLabel(load.family, t)}
        </Text>
        <Text size={2} weight="medium" highContrast>
          {load.count}
        </Text>
      </div>
    </Fragment>
  );
}
