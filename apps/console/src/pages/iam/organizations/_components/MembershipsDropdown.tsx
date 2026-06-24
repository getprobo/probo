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

import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dropdown,
  DropdownItem,
  DropdownSeparator,
  IconChevronGrabberVertical,
  IconMagnifyingGlass,
  IconPlusLarge,
  Input,
} from "@probo/ui";
import { Suspense, useCallback, useState } from "react";
import { useFragment, useQueryLoader } from "react-relay";
import { Link, useLocation } from "react-router";
import { graphql } from "relay-runtime";

import type { MembershipsDropdown_organizationFragment$key } from "#/__generated__/iam/MembershipsDropdown_organizationFragment.graphql";
import type { MembershipsDropdownMenuQuery } from "#/__generated__/iam/MembershipsDropdownMenuQuery.graphql";

import {
  MembershipsDropdownMenu,
  membershipsDropdownMenuQuery,
} from "./MembershipsDropdownMenu";

const organizationFragment = graphql`
  fragment MembershipsDropdown_organizationFragment on Organization {
    name
  }
`;

export function MembershipsDropdown(props: {
  organizationFKey: MembershipsDropdown_organizationFragment$key;
}) {
  const { organizationFKey } = props;

  const location = useLocation();
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");

  const currentOrganization
    = useFragment<MembershipsDropdown_organizationFragment$key>(
      organizationFragment,
      organizationFKey,
    );
  const [queryRef, loadQuery] = useQueryLoader<MembershipsDropdownMenuQuery>(
    membershipsDropdownMenuQuery,
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
        toggle={(
          <Button
            className="-ml-3"
            variant="tertiary"
            iconAfter={IconChevronGrabberVertical}
          >
            {currentOrganization.name}
          </Button>
        )}
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
              fallback={(
                <div className="px-3 py-2 text-gray-500">
                  {__("Loading organizations...")}
                </div>
              )}
            >
              <MembershipsDropdownMenu search={search} queryRef={queryRef} />
            </Suspense>
          )}
        </div>
        <DropdownSeparator />
        <DropdownItem asChild>
          <Link to="/organizations/new" state={{ from: location.pathname }}>
            <IconPlusLarge size={16} />
            {__("Add organization")}
          </Link>
        </DropdownItem>
      </Dropdown>
    </div>
  );
}
