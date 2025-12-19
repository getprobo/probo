import { useFragment, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import {
  Badge,
  Button,
  Dropdown,
  DropdownItem,
  DropdownSeparator,
  IconChevronGrabberVertical,
  IconMagnifyingGlass,
  IconPeopleAdd,
  IconPlusLarge,
  Input,
} from "@probo/ui";
import { Suspense, useCallback, useState } from "react";
import { useTranslate } from "@probo/i18n";
import {
  OrganizationDropdownMenu,
  organizationDropdownMenuQuery,
} from "./OrganizationDropdownMenu";
import type { OrganizationDropdownMenuQuery } from "./__generated__/OrganizationDropdownMenuQuery.graphql";
import { Link } from "react-router";
import type { OrganizationDropdown_organizationFragment$key } from "./__generated__/OrganizationDropdown_organizationFragment.graphql";
import type { OrganizationDropdown_viewerFragment$key } from "./__generated__/OrganizationDropdown_viewerFragment.graphql";

const organizationFragment = graphql`
  fragment OrganizationDropdown_organizationFragment on Organization {
    name
  }
`;
const viewerFragment = graphql`
  fragment OrganizationDropdown_viewerFragment on Identity {
    pendingInvitations @required(action: THROW) {
      totalCount @required(action: THROW)
    }
  }
`;

export function OrganizationDropdown(props: {
  organizationFKey: OrganizationDropdown_organizationFragment$key;
  viewerFKey: OrganizationDropdown_viewerFragment$key;
}) {
  const { organizationFKey, viewerFKey } = props;

  const { __ } = useTranslate();
  const [search, setSearch] = useState("");

  const currentOrganization =
    useFragment<OrganizationDropdown_organizationFragment$key>(
      organizationFragment,
      organizationFKey,
    );
  const {
    pendingInvitations: { totalCount: pendingInvitationCount },
  } = useFragment<OrganizationDropdown_viewerFragment$key>(
    viewerFragment,
    viewerFKey,
  );
  const [queryRef, loadQuery] = useQueryLoader<OrganizationDropdownMenuQuery>(
    organizationDropdownMenuQuery,
  );

  const handleOpenMenu = useCallback(
    (open: boolean) => {
      if (open) loadQuery({});
    },
    [loadQuery],
  );

  return (
    <div className="flex items-center gap-1">
      <Dropdown
        onOpenChange={handleOpenMenu}
        toggle={
          <Button
            className="-ml-3"
            variant="tertiary"
            iconAfter={IconChevronGrabberVertical}
          >
            {currentOrganization.name}
          </Button>
        }
      >
        <div className="px-3 py-2">
          <Input
            icon={IconMagnifyingGlass}
            placeholder={__("Search organizations...")}
            value={search}
            onValueChange={setSearch}
            onKeyDown={(e) => {
              e.stopPropagation();
            }}
            autoFocus
          />
        </div>
        <div className="max-h-150 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-transparent hover:scrollbar-thumb-gray-400">
          {queryRef && (
            <Suspense
              fallback={
                <div className="px-3 py-2 text-gray-500">
                  {__("Loading organizations...")}
                </div>
              }
            >
              <OrganizationDropdownMenu search={search} queryRef={queryRef} />
            </Suspense>
          )}
        </div>
        <DropdownSeparator />
        {pendingInvitationCount > 0 && (
          <DropdownItem asChild>
            <Link to="/">
              <IconPeopleAdd size={16} />
              <span className="flex-1">{__("Invitations")}</span>
              <Badge variant="info" size="sm">
                {pendingInvitationCount}
              </Badge>
            </Link>
          </DropdownItem>
        )}
        <DropdownItem asChild>
          <Link to="/organizations/new">
            <IconPlusLarge size={16} />
            {__("Add organization")}
          </Link>
        </DropdownItem>
      </Dropdown>
    </div>
  );
}
