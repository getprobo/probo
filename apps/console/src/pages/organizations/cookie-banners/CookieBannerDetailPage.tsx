import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  DropdownItem,
  IconTrashCan,
  TabLink,
  Tabs,
} from "@probo/ui";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { Outlet } from "react-router";

import type { CookieBannerGraphNodeQuery } from "#/__generated__/core/CookieBannerGraphNodeQuery.graphql";
import {
  cookieBannerNodeQuery,
  useDeleteCookieBanner,
  useDisableCookieBannerMutation,
  usePublishCookieBannerMutation,
} from "#/hooks/graph/CookieBannerGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CookieBannerStateBadge } from "./_components/CookieBannerStateBadge";

type Props = {
  queryRef: PreloadedQuery<CookieBannerGraphNodeQuery>;
};

export default function CookieBannerDetailPage(props: Props) {
  const { node: banner } = usePreloadedQuery(
    cookieBannerNodeQuery,
    props.queryRef,
  );
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "CookieBannersPage_cookieBanners",
  );
  const deleteCookieBanner = useDeleteCookieBanner(banner, connectionId);
  const [publishBanner] = usePublishCookieBannerMutation();
  const [disableBanner] = useDisableCookieBannerMutation();

  const bannersUrl = `/organizations/${organizationId}/cookie-banners`;
  const baseBannerUrl = `/organizations/${organizationId}/cookie-banners/${banner.id}`;

  const handlePublish = () => {
    void publishBanner({
      variables: {
        input: { id: banner.id },
      },
    });
  };

  const handleDisable = () => {
    void disableBanner({
      variables: {
        input: { id: banner.id },
      },
    });
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Cookie Banners"),
            to: bannersUrl,
          },
          {
            label: banner.name ?? "",
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl">{banner.name}</div>
          <CookieBannerStateBadge state={banner.state ?? "DRAFT"} />
        </div>
        <div className="flex gap-2 items-center">
          {banner.canPublish && banner.state !== "PUBLISHED" && (
            <Button onClick={handlePublish} variant="secondary">
              {__("Publish")}
            </Button>
          )}
          {banner.canPublish && banner.state === "PUBLISHED" && (
            <Button onClick={handleDisable} variant="secondary">
              {__("Disable")}
            </Button>
          )}
          {banner.canDelete && (
            <ActionDropdown variant="secondary">
              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                onClick={deleteCookieBanner}
              >
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </div>
      </div>

      <Tabs>
        <TabLink to={`${baseBannerUrl}/overview`}>
          {__("Overview")}
        </TabLink>
        <TabLink to={`${baseBannerUrl}/appearance`}>
          {__("Appearance")}
        </TabLink>
        <TabLink to={`${baseBannerUrl}/categories`}>
          {__("Categories")}
        </TabLink>
        <TabLink to={`${baseBannerUrl}/consent-records`}>
          {__("Consent Records")}
        </TabLink>
        <TabLink to={`${baseBannerUrl}/embed`}>
          {__("Embed")}
        </TabLink>
      </Tabs>

      <Outlet context={{ banner }} />
    </div>
  );
}
