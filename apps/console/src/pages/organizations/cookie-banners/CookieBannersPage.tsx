import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { CookieBannerGraphListQuery } from "#/__generated__/core/CookieBannerGraphListQuery.graphql";
import type { CookieBannersPageFragment$key } from "#/__generated__/core/CookieBannersPageFragment.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import {
  cookieBannersQuery,
  useDeleteCookieBanner,
} from "#/hooks/graph/CookieBannerGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CookieBannerStateBadge } from "./_components/CookieBannerStateBadge";
import { CreateCookieBannerDialog } from "./dialogs/CreateCookieBannerDialog";

const paginatedCookieBannersFragment = graphql`
  fragment CookieBannersPageFragment on Organization
  @refetchable(queryName: "CookieBannersListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "CookieBannerOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    cookieBanners(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "CookieBannersPage_cookieBanners") {
      __id
      edges {
        node {
          # eslint-disable-next-line relay/unused-fields
          id
          # eslint-disable-next-line relay/unused-fields
          name
          # eslint-disable-next-line relay/unused-fields
          domain
          # eslint-disable-next-line relay/unused-fields
          state
          # eslint-disable-next-line relay/unused-fields
          version
          # eslint-disable-next-line relay/unused-fields
          createdAt
          canUpdate: permission(action: "core:cookie-banner:update")
          canDelete: permission(action: "core:cookie-banner:delete")
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<CookieBannerGraphListQuery>;
};

export default function CookieBannersPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery<CookieBannerGraphListQuery>(
    cookieBannersQuery,
    props.queryRef,
  );
  const pagination = usePaginationFragment(
    paginatedCookieBannersFragment,
    data.node as CookieBannersPageFragment$key,
  );

  const banners = pagination.data?.cookieBanners?.edges.map(edge => edge.node) ?? [];
  const connectionId = pagination.data?.cookieBanners?.__id ?? "";

  usePageTitle(__("Cookie Banners"));

  const hasAnyAction = banners.some(
    ({ canUpdate, canDelete }) => canUpdate || canDelete,
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Cookie Banners")}
        description={__(
          "Manage cookie consent banners for your websites. Configure categories, customize appearance, and track visitor consent.",
        )}
      >
        {data.node.canCreateCookieBanner && (
          <CreateCookieBannerDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>
              {__("Add cookie banner")}
            </Button>
          </CreateCookieBannerDialog>
        )}
      </PageHeader>
      {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
      <SortableTable {...(pagination as any)}>
        <Thead>
          <Tr>
            <SortableTh field="CREATED_AT">{__("Name")}</SortableTh>
            <Th>{__("Domain")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Version")}</Th>
            {hasAnyAction && <Th />}
          </Tr>
        </Thead>
        <Tbody>
          {banners?.map(banner => (
            <CookieBannerRow
              key={banner.id}
              banner={banner}
              organizationId={organizationId}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

type Banner = {
  id: string;
  name: string;
  domain: string;
  state: string;
  version: number;
  createdAt: string;
  canUpdate: boolean;
  canDelete: boolean;
};

function CookieBannerRow({
  banner,
  organizationId,
  connectionId,
  hasAnyAction,
}: {
  banner: Banner;
  organizationId: string;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const { __ } = useTranslate();
  const deleteCookieBanner = useDeleteCookieBanner(banner, connectionId);
  const bannerUrl = `/organizations/${organizationId}/cookie-banners/${banner.id}/overview`;

  return (
    <Tr to={bannerUrl}>
      <Td>{banner.name}</Td>
      <Td>{banner.domain}</Td>
      <Td>
        <CookieBannerStateBadge state={banner.state} />
      </Td>
      <Td>v{banner.version}</Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {banner.canDelete && (
              <DropdownItem
                onClick={deleteCookieBanner}
                variant="danger"
                icon={IconTrashCan}
              >
                {__("Delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
