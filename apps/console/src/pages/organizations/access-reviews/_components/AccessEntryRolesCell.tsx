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

import { Badge, Td } from "@probo/ui";
import * as Popover from "@radix-ui/react-popover";
import { graphql, useFragment } from "react-relay";

import type { AccessEntryRolesCell_accessEntry$key } from "#/__generated__/core/AccessEntryRolesCell_accessEntry.graphql";

import { NotAvailable } from "./accessReviewHelpers";

const VISIBLE_ROLE_COUNT = 3;

const accessEntryRolesCellFragment = graphql`
  fragment AccessEntryRolesCell_accessEntry on AccessReviewEntry {
    roles
  }
`;

type Props = {
  accessEntryKey: AccessEntryRolesCell_accessEntry$key;
};

export function AccessEntryRolesCell({ accessEntryKey }: Props) {
  const entry = useFragment(accessEntryRolesCellFragment, accessEntryKey);
  const roles = entry.roles;

  if (roles.length === 0) {
    return (
      <Td className="max-w-xs">
        <NotAvailable />
      </Td>
    );
  }

  const visibleRoles = roles.slice(0, VISIBLE_ROLE_COUNT);
  const hiddenRoles = roles.slice(VISIBLE_ROLE_COUNT);

  return (
    <Td noLink className="max-w-xs">
      <div className="flex flex-wrap gap-1">
        {visibleRoles.map((role, index) => (
          <Badge key={`${index}-${role}`} variant="neutral" className="text-xs">
            {role}
          </Badge>
        ))}
        {hiddenRoles.length > 0 && (
          <Popover.Root>
            <Popover.Trigger asChild>
              <button type="button" className="inline-flex">
                <Badge variant="neutral" className="text-xs cursor-pointer">
                  +
                  {hiddenRoles.length}
                </Badge>
              </button>
            </Popover.Trigger>
            <Popover.Portal>
              <Popover.Content
                className="z-50 rounded-md border bg-level-0 p-3 shadow-md max-w-sm"
                sideOffset={4}
                align="start"
              >
                <div className="flex flex-wrap gap-1">
                  {hiddenRoles.map((role, index) => (
                    <Badge key={`${index}-${role}`} variant="neutral" className="text-xs">
                      {role}
                    </Badge>
                  ))}
                </div>
              </Popover.Content>
            </Popover.Portal>
          </Popover.Root>
        )}
      </div>
    </Td>
  );
}
