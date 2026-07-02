// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
