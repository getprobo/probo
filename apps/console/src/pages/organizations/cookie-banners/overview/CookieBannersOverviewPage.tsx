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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  Card,
  DropdownItem,
  IconTrashCan,
  PageHeader,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannersOverviewPageDeleteMutation } from "#/__generated__/core/CookieBannersOverviewPageDeleteMutation.graphql";
import type { CookieBannersOverviewPageQuery } from "#/__generated__/core/CookieBannersOverviewPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CookieBannerEmptyState } from "./_components/CookieBannerEmptyState";

export const cookieBannersOverviewPageQuery = graphql`
  query CookieBannersOverviewPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        cookieBanners(first: 50, orderBy: { field: CREATED_AT, direction: DESC })
          @connection(key: "CookieBannersOverviewPage_cookieBanners", filters: [])
          @required(action: THROW) {
          __id
          edges {
            node {
              id
              name
              origin
              state
              createdAt
              canDelete: permission(action: "core:cookie-banner:delete")
            }
          }
        }
      }
    }
  }
`;

const deleteCookieBannerMutation = graphql`
  mutation CookieBannersOverviewPageDeleteMutation(
    $input: DeleteCookieBannerInput!
    $connections: [ID!]!
  ) {
    deleteCookieBanner(input: $input) {
      deletedCookieBannerId @deleteEdge(connections: $connections)
    }
  }
`;

interface CookieBannersOverviewPageProps {
  queryRef: PreloadedQuery<CookieBannersOverviewPageQuery>;
}

export function CookieBannersOverviewPage({ queryRef }: CookieBannersOverviewPageProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const organizationId = useOrganizationId();
  const { toast } = useToast();
  const confirm = useConfirm();

  usePageTitle(t("overviewPage.pageTitle"));

  const { organization } = usePreloadedQuery<CookieBannersOverviewPageQuery>(cookieBannersOverviewPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const connectionId = organization.cookieBanners.__id;
  const banners = organization.cookieBanners.edges.map(e => e.node);
  const newBannerHref = `/organizations/${organizationId}/cookie-banners/new`;

  const [deleteCookieBanner] = useMutation<CookieBannersOverviewPageDeleteMutation>(deleteCookieBannerMutation);

  const handleDelete = (bannerId: string, bannerName: string) => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteCookieBanner({
            variables: {
              input: { cookieBannerId: bannerId },
              connections: [connectionId],
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: t("overviewPage.errors.title"),
                  description: errors[0].message,
                  variant: "error",
                });
              } else {
                toast({
                  title: t("overviewPage.messages.successTitle"),
                  description: t("overviewPage.messages.deleted"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: t("overviewPage.errors.title"),
                description: formatError(
                  t("overviewPage.errors.delete"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("overviewPage.deleteConfirmation", { name: bannerName }),
        variant: "danger",
        label: t("overviewPage.actions.delete"),
      },
    );
  };

  if (banners.length === 0) {
    return (
      <div className="space-y-6">
        <PageHeader
          title={t("overviewPage.title")}
          description={t("overviewPage.description")}
        />
        <CookieBannerEmptyState>
          <Button to={newBannerHref}>
            {t("overviewPage.actions.createFirst")}
          </Button>
        </CookieBannerEmptyState>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("overviewPage.title")}
        description={t("overviewPage.description")}
      />

      <div className="space-y-4">
        <div className="flex justify-end">
          <Button to={newBannerHref}>{t("overviewPage.actions.create")}</Button>
        </div>

        <Card className="divide-y divide-border-low rounded-lg">
          {banners.map(banner => (
            <Link
              key={banner.id}
              to={`/organizations/${organizationId}/cookie-banners/${banner.id}`}
              className="flex items-center justify-between gap-4 p-4 hover:bg-muted/50 transition-colors"
            >
              <div className="min-w-0 flex-1">
                <div className="font-medium">{banner.name}</div>
                <div className="text-sm text-muted-foreground truncate">{banner.origin}</div>
              </div>
              <div className="flex items-center gap-3">
                <Badge variant={banner.state === "ACTIVE" ? "success" : "danger"}>
                  {banner.state === "ACTIVE"
                    ? t("overviewPage.status.active")
                    : t("overviewPage.status.inactive")}
                </Badge>
                <span className="text-xs text-muted-foreground">
                  {dateFormat(i18n.language, banner.createdAt)}
                </span>
                {banner.canDelete && banner.state !== "ACTIVE" && (
                  <div onClick={e => e.preventDefault()}>
                    <ActionDropdown>
                      <DropdownItem
                        onClick={() => handleDelete(banner.id, banner.name)}
                        variant="danger"
                        icon={IconTrashCan}
                      >
                        {t("overviewPage.actions.delete")}
                      </DropdownItem>
                    </ActionDropdown>
                  </div>
                )}
              </div>
            </Link>
          ))}
        </Card>
      </div>
    </div>
  );
}
