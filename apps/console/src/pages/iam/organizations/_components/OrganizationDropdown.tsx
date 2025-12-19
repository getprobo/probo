import { useFragment, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import type { OrganizationDropdownFragment$key } from "./__generated__/OrganizationDropdownFragment.graphql";
import {
  Button,
  Dropdown,
  DropdownSeparator,
  IconChevronGrabberVertical,
  IconMagnifyingGlass,
  Input,
} from "@probo/ui";
import { Suspense, useCallback, useState } from "react";
import { useTranslate } from "@probo/i18n";
import {
  OrganizationDropdownMenu,
  organizationDropdownMenuQuery,
} from "./OrganizationDropdownMenu";
import type { OrganizationDropdownMenuQuery } from "./__generated__/OrganizationDropdownMenuQuery.graphql";

const fragment = graphql`
  fragment OrganizationDropdownFragment on Organization {
    name
  }
`;

export function OrganizationDropdown(props: {
  fKey: OrganizationDropdownFragment$key;
}) {
  const { fKey } = props;

  const { __ } = useTranslate();
  const [search, setSearch] = useState("");

  const currentOrganization = useFragment(fragment, fKey);
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
      </Dropdown>
    </div>
  );
}
