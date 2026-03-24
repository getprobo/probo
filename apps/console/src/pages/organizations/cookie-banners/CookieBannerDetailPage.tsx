import { formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  DropdownItem,
  IconTrashCan,
  TabLink,
  Tabs,
  useConfirm,
  useToast,
} from "@probo/ui";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { Outlet } from "react-router";

import type { CookieBannerDetailPageDeleteMutation } from "#/__generated__/core/CookieBannerDetailPageDeleteMutation.graphql";
import type { CookieBannerDetailPageDisableMutation } from "#/__generated__/core/CookieBannerDetailPageDisableMutation.graphql";
import type { CookieBannerDetailPagePublishMutation } from "#/__generated__/core/CookieBannerDetailPagePublishMutation.graphql";
import type { CookieBannerDetailPageQuery } from "#/__generated__/core/CookieBannerDetailPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CookieBannerStateBadge } from "./_components/CookieBannerStateBadge";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const cookieBannerNodeQuery = graphql`
  query CookieBannerDetailPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      ... on CookieBanner {
        id
        name
        domain
        state
        title
        description
        acceptAllLabel
        rejectAllLabel
        savePreferencesLabel
        privacyPolicyUrl
        consentExpiryDays
        version
        embedSnippet
        createdAt
        updatedAt
        canUpdate: permission(action: "core:cookie-banner:update")
        canDelete: permission(action: "core:cookie-banner:delete")
        canPublish: permission(action: "core:cookie-banner:update")
        ...CookieBannerOverviewTabFragment
        ...CookieBannerAppearanceTabFragment
        ...CookieBannerCategoriesTabFragment
        ...CookieBannerConsentRecordsTabFragment
      }
    }
  }
`;

const deleteCookieBannerMutation = graphql`
  mutation CookieBannerDetailPageDeleteMutation(
    $input: DeleteCookieBannerInput!
    $connections: [ID!]!
  ) {
    deleteCookieBanner(input: $input) {
      deletedCookieBannerId @deleteEdge(connections: $connections)
    }
  }
`;

const publishCookieBannerMutation = graphql`
  mutation CookieBannerDetailPagePublishMutation(
    $input: PublishCookieBannerInput!
  ) {
    publishCookieBanner(input: $input) {
      cookieBanner {
        id
        state
        version
        updatedAt
      }
    }
  }
`;

const disableCookieBannerMutation = graphql`
  mutation CookieBannerDetailPageDisableMutation(
    $input: DisableCookieBannerInput!
  ) {
    disableCookieBanner(input: $input) {
      cookieBanner {
        id
        state
        updatedAt
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<CookieBannerDetailPageQuery>;
};

export default function CookieBannerDetailPage(props: Props) {
  const { node: banner } = usePreloadedQuery(
    cookieBannerNodeQuery,
    props.queryRef,
  );
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const organizationId = useOrganizationId();

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "CookieBannersPage_cookieBanners",
  );

  const [deleteCookieBanner] = useMutation<CookieBannerDetailPageDeleteMutation>(deleteCookieBannerMutation);
  const [publishBanner] = useMutation<CookieBannerDetailPagePublishMutation>(publishCookieBannerMutation);
  const [disableBanner] = useMutation<CookieBannerDetailPageDisableMutation>(disableCookieBannerMutation);

  const bannersUrl = `/organizations/${organizationId}/cookie-banners`;
  const baseBannerUrl = `/organizations/${organizationId}/cookie-banners/${banner.id}`;

  const handlePublish = () => {
    publishBanner({
      variables: {
        input: { id: banner.id },
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Cookie banner published successfully."),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to publish cookie banner"), error as GraphQLError),
          variant: "error",
        });
      },
    });
  };

  const handleDisable = () => {
    disableBanner({
      variables: {
        input: { id: banner.id },
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Cookie banner disabled successfully."),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to disable cookie banner"), error as GraphQLError),
          variant: "error",
        });
      },
    });
  };

  const handleDelete = () => {
    if (!banner.id || !banner.name) {
      return alert(__("Failed to delete cookie banner: missing id or name"));
    }
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteCookieBanner({
            variables: {
              input: { id: banner.id },
              connections: [connectionId],
            },
            onCompleted() {
              toast({
                title: __("Success"),
                description: __("Cookie banner deleted successfully."),
                variant: "success",
              });
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(__("Failed to delete cookie banner"), error as GraphQLError),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete cookie banner \"%s\". This action cannot be undone.",
          ),
          banner.name,
        ),
      },
    );
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
                onClick={handleDelete}
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
