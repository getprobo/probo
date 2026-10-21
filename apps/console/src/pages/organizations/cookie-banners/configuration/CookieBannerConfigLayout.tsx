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

import { CheckCircleIcon, CopyIcon, InfoIcon, WarningIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Fragment, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannerConfigLayoutPublishMutation } from "#/__generated__/core/CookieBannerConfigLayoutPublishMutation.graphql";
import type { CookieBannerConfigLayoutQuery } from "#/__generated__/core/CookieBannerConfigLayoutQuery.graphql";
import type { CookieBannerConfigLayoutStartSampleMutation } from "#/__generated__/core/CookieBannerConfigLayoutStartSampleMutation.graphql";
import type { CookieBannerConfigLayoutStopSampleMutation } from "#/__generated__/core/CookieBannerConfigLayoutStopSampleMutation.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import type { TonedCardTone } from "#/components/TonedCard/variants";
import { useMutation } from "#/lib/relay/useMutation";

import { cookieBannerConfigLayout } from "../variants";

import { FamilyLogo } from "./_components/FamilyLogo";
import { familyLabel } from "./_lib/familyLabel";

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
        latestSample {
          id
          createdAt
          frozenAt
          pageLoads {
            family
            count
          }
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
        latestSample {
          id
          createdAt
          frozenAt
          pageLoads {
            family
            count
          }
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

const startSampleMutation = graphql`
  mutation CookieBannerConfigLayoutStartSampleMutation($input: StartCookieBannerSampleInput!) {
    startCookieBannerSample(input: $input) {
      cookieBannerSample {
        id
        createdAt
        frozenAt
        pageLoads {
          family
          count
        }
      }
      cookieBanner {
        id
        latestSample {
          id
          createdAt
          frozenAt
          pageLoads {
            family
            count
          }
        }
      }
    }
  }
`;

const stopSampleMutation = graphql`
  mutation CookieBannerConfigLayoutStopSampleMutation($input: StopCookieBannerSampleInput!) {
    stopCookieBannerSample(input: $input) {
      cookieBannerSample {
        id
        createdAt
        frozenAt
        pageLoads {
          family
          count
        }
      }
      cookieBanner {
        id
        latestSample {
          id
          createdAt
          frozenAt
          pageLoads {
            family
            count
          }
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
    sample,
    sampleRow,
    sampleItem,
    sampleLogo,
    sampleSeparator,
    sampleActions,
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
  const [startSample, isStartingSample] = useMutation<CookieBannerConfigLayoutStartSampleMutation>(
    startSampleMutation,
    {
      successMessage: t("configLayout.messages.sampleStarted"),
      errorToast: t("configLayout.errors.startSample"),
    },
  );
  const [stopSample, isStoppingSample] = useMutation<CookieBannerConfigLayoutStopSampleMutation>(
    stopSampleMutation,
    {
      successMessage: t("configLayout.messages.sampleStopped"),
      errorToast: t("configLayout.errors.stopSample"),
    },
  );

  const hasDraft = banner.latestVersion?.state === "DRAFT";
  const isDeactivated = banner.state !== "ACTIVE";
  const latestSample = banner.latestSample;
  const inSample = latestSample != null && latestSample.frozenAt == null;
  const version = banner.latestVersion?.version;
  const pageLoads = latestSample?.pageLoads ?? [];

  let message: string;
  if (isDeactivated && hasDraft && version != null) {
    message = t("configLayout.callout.deactivatedDraft", { version });
  } else if (isDeactivated) {
    message = t("configLayout.callout.deactivated");
  } else if (inSample && hasDraft) {
    message = t("configLayout.callout.samplingUnpublished");
  } else if (inSample) {
    message = t("configLayout.callout.sampling");
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
  } else if (inSample) {
    tone = "sky";
    icon = <InfoIcon size={24} weight="duotone" />;
  } else if (hasDraft) {
    tone = "amber";
    icon = <WarningIcon size={24} weight="duotone" />;
  } else {
    tone = "green";
    icon = <CheckCircleIcon size={24} weight="duotone" />;
  }

  function handleCopyId() {
    void navigator.clipboard.writeText(banner.id).then(
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
        control={(
          <div className={sampleActions()}>
            {inSample
              ? (
                  <Button
                    size={2}
                    variant="soft"
                    color="neutral"
                    loading={isStoppingSample}
                    onClick={() => {
                      void stopSample({
                        variables: { input: { cookieBannerId: banner.id } },
                      }).catch(() => {
                        // Error toast is already shown by useMutation.
                      });
                    }}
                  >
                    {t("configLayout.actions.stopSample")}
                  </Button>
                )
              : (
                  <Button
                    size={2}
                    variant="soft"
                    color="neutral"
                    loading={isStartingSample}
                    onClick={() => {
                      void startSample({
                        variables: { input: { cookieBannerId: banner.id } },
                      }).catch(() => {
                        // Error toast is already shown by useMutation.
                      });
                    }}
                  >
                    {t("configLayout.actions.startSample")}
                  </Button>
                )}
            {hasDraft
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
          </div>
        )}
      >
        <div className={sample()}>
          <Text size={2} color="neutral">{message}</Text>
          {latestSample != null
            ? (
                <SampleLoadRow
                  label={t("configLayout.sample.pageLoads")}
                  loads={pageLoads}
                  empty={t("configLayout.sample.empty")}
                  sampleRow={sampleRow()}
                  sampleItem={sampleItem()}
                  sampleLogo={sampleLogo()}
                  sampleSeparator={sampleSeparator()}
                />
              )
            : null}
        </div>
      </TonedCard>
      <Outlet />
    </div>
  );
}

interface SampleLoadRowProps {
  label: string;
  loads: readonly { family: string; count: number }[];
  empty: string;
  sampleRow: string;
  sampleItem: string;
  sampleLogo: string;
  sampleSeparator: string;
}

function SampleLoadRow({
  label,
  loads,
  empty,
  sampleRow,
  sampleItem,
  sampleLogo,
  sampleSeparator,
}: SampleLoadRowProps) {
  const { t } = useTranslation("organizations/cookie-banners");

  return (
    <div className={sampleRow}>
      <Text size={2} color="faint">
        {label}
      </Text>
      {loads.length > 0
        ? loads.map((entry, index) => (
            <Fragment key={entry.family}>
              {index > 0
                ? (
                    <Text
                      size={2}
                      color="faint"
                      className={sampleSeparator}
                    >
                      ·
                    </Text>
                  )
                : null}
              <div className={sampleItem}>
                <FamilyLogo
                  family={entry.family}
                  size={16}
                  className={sampleLogo}
                />
                <Text size={2} color="faint">
                  {familyLabel(entry.family, t)}
                </Text>
                <Text size={2} weight="medium" highContrast>
                  {entry.count}
                </Text>
              </div>
            </Fragment>
          ))
        : (
            <Text size={2} color="faint">
              {empty}
            </Text>
          )}
    </div>
  );
}
