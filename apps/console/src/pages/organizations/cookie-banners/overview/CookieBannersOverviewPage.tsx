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

import { formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { toast } = useToast();
  const confirm = useConfirm();

  usePageTitle(__("Cookie Banners"));

  const { organization } = usePreloadedQuery(cookieBannersOverviewPageQuery, queryRef);
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
                  title: __("Error"),
                  description: errors[0].message,
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Cookie banner deleted successfully"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete cookie banner"),
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
          __("This will permanently delete the cookie banner \"%s\". This action cannot be undone."),
          bannerName,
        ),
        variant: "danger",
        label: __("Delete"),
      },
    );
  };

  if (banners.length === 0) {
    return (
      <div className="space-y-6">
        <PageHeader
          title={__("Cookie Banners")}
          description={__(
            "Manage cookie consent banners for your websites. Configure categories, cookies, and install the SDK.",
          )}
        />
        <CookieBannerEmptyState>
          <Button to={newBannerHref}>{__("Create your first banner")}</Button>
        </CookieBannerEmptyState>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Cookie Banners")}
        description={__(
          "Manage cookie consent banners for your websites. Configure categories, cookies, and install the SDK.",
        )}
      />

      <div className="space-y-4">
        <div className="flex justify-end">
          <Button to={newBannerHref}>{__("Create Banner")}</Button>
        </div>

        <Card className="divide-y divide-border rounded-lg border">
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
                  {banner.state === "ACTIVE" ? __("Active") : __("Inactive")}
                </Badge>
                <span className="text-xs text-muted-foreground">
                  {new Date(banner.createdAt).toLocaleDateString()}
                </span>
                {banner.canDelete && banner.state !== "ACTIVE" && (
                  <div onClick={e => e.preventDefault()}>
                    <ActionDropdown>
                      <DropdownItem
                        onClick={() => handleDelete(banner.id, banner.name)}
                        variant="danger"
                        icon={IconTrashCan}
                      >
                        {__("Delete")}
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
