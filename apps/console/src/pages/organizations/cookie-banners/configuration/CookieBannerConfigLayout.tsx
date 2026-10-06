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

import { CopyIcon, DotsThreeVerticalIcon, TrashIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannerConfigLayoutActivateMutation } from "#/__generated__/core/CookieBannerConfigLayoutActivateMutation.graphql";
import type { CookieBannerConfigLayoutDeactivateMutation } from "#/__generated__/core/CookieBannerConfigLayoutDeactivateMutation.graphql";
import type { CookieBannerConfigLayoutPublishMutation } from "#/__generated__/core/CookieBannerConfigLayoutPublishMutation.graphql";
import type { CookieBannerConfigLayoutQuery } from "#/__generated__/core/CookieBannerConfigLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { navGroupByKey, navHref } from "#/pages/iam/organizations/_lib/navigation";

import { cookieBannerConfigLayout } from "../variants";

import { DeleteCookieBannerDialog } from "./_components/DeleteCookieBannerDialog";

export const cookieBannerConfigLayoutQuery = graphql`
  query CookieBannerConfigLayoutQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        id
        name
        origin
        state
        capabilities {
          corsless
        }
        canDelete: permission(action: "core:cookie-banner:delete")
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

export function CookieBannerConfigLayout({
  queryRef,
}: CookieBannerConfigLayoutProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const { root, header, titleRow, title, version, meta, id, actions } = cookieBannerConfigLayout();
  const [deleteOpen, setDeleteOpen] = useState(false);

  const data = usePreloadedQuery<CookieBannerConfigLayoutQuery>(
    cookieBannerConfigLayoutQuery,
    queryRef,
  );
  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;
  usePageTitle(banner.name);

  const [activate, isActivating] = useMutation<CookieBannerConfigLayoutActivateMutation>(
    activateMutation,
    {
      successMessage: t("configLayout.messages.activated"),
      errorToast: t("configLayout.errors.activate"),
    },
  );
  const [deactivate, isDeactivating] = useMutation<CookieBannerConfigLayoutDeactivateMutation>(
    deactivateMutation,
    {
      successMessage: t("configLayout.messages.deactivated"),
      errorToast: t("configLayout.errors.deactivate"),
    },
  );
  const [publish, isPublishing] = useMutation<CookieBannerConfigLayoutPublishMutation>(
    publishMutation,
    {
      successMessage: t("configLayout.messages.published"),
      errorToast: t("configLayout.errors.publish"),
    },
  );

  const hasDraft = banner.latestVersion?.state === "DRAFT";
  const policyDocumentId = banner.policyDocument?.id;

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
      <div className={header()}>
        <div className={titleRow()}>
          <div className={title()}>
            <Heading level={1} size={6} weight="medium" highContrast>
              {banner.name}
            </Heading>
            {banner.latestVersion?.version != null && (
              <Text size={2} color="faint" className={version()}>
                {t("configLayout.version", { version: banner.latestVersion.version })}
                {banner.latestVersion.state === "DRAFT" && t("configLayout.draft")}
              </Text>
            )}
            <Badge
              color={banner.state === "ACTIVE" ? "green" : "red"}
              variant="soft"
            >
              {banner.state === "ACTIVE"
                ? t("configLayout.status.active")
                : t("configLayout.status.inactive")}
            </Badge>
          </div>
          <div className={actions()}>
            {hasDraft && (
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
            )}
            <Button
              size={2}
              variant="soft"
              color="neutral"
              disabled={isActivating || isDeactivating}
              onClick={() => {
                const mutate = banner.state === "ACTIVE" ? deactivate : activate;
                void mutate({
                  variables: { input: { cookieBannerId: banner.id } },
                }).catch(() => {
                  // Error toast is already shown by useMutation.
                });
              }}
            >
              {banner.state === "ACTIVE"
                ? t("configLayout.actions.deactivate")
                : t("configLayout.actions.activate")}
            </Button>
            {banner.canDelete && banner.state !== "ACTIVE" && (
              <Dropdown>
                <DropdownTrigger
                  render={(
                    <IconButton
                      size={2}
                      variant="ghost"
                      color="neutral"
                      aria-label={t("configLayout.actions.more")}
                    >
                      <DotsThreeVerticalIcon />
                    </IconButton>
                  )}
                />
                <DropdownPopup>
                  <DropdownItem
                    color="error"
                    iconStart={<TrashIcon />}
                    onClick={() => setDeleteOpen(true)}
                  >
                    {t("configLayout.actions.delete")}
                  </DropdownItem>
                </DropdownPopup>
              </Dropdown>
            )}
          </div>
        </div>
        <div className={meta()}>
          {!banner.capabilities.corsless && (
            <Text size={2} color="faint">
              {t("configLayout.metadata.origin")}
              {" "}
              {banner.origin}
            </Text>
          )}
          <div className={id()}>
            <Text size={2} color="faint">
              {t("configLayout.metadata.id")}
              {" "}
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
          </div>
          {policyDocumentId != null && (
            <Link
              size={2}
              to={navHref(
                organizationId,
                navGroupByKey("governance"),
                `documents/${encodeURIComponent(policyDocumentId)}`,
              )}
            >
              {t("configLayout.metadata.cookiePolicy")}
            </Link>
          )}
        </div>
      </div>
      <Outlet />
      <DeleteCookieBannerDialog
        cookieBannerId={banner.id}
        name={banner.name}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
      />
    </div>
  );
}
