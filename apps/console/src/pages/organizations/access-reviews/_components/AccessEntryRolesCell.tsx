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
import { graphql, useFragment } from "react-relay";

import type { AccessEntryRolesCell_accessEntry$key } from "#/__generated__/core/AccessEntryRolesCell_accessEntry.graphql";

import { AccessEntryExtraRolesPopover } from "./AccessEntryExtraRolesPopover";
import { NotAvailable } from "./accessReviewHelpers";

const VISIBLE_ROLE_COUNT = 2;

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
      <Td className="max-w-0 py-2.5">
        <NotAvailable />
      </Td>
    );
  }

  const visibleRoles = roles.slice(0, VISIBLE_ROLE_COUNT);
  const hiddenRoles = roles.slice(VISIBLE_ROLE_COUNT);
  const primaryRole = visibleRoles[0];
  const secondaryRoles = visibleRoles.slice(1);

  return (
    <Td noLink className="max-w-0 py-2.5">
      <div className="flex min-w-0 items-center gap-1.5">
        <span className="min-w-0 truncate text-sm" title={primaryRole}>
          {primaryRole}
        </span>
        {secondaryRoles.map((role, index) => (
          <Badge
            key={`${index}-${role}`}
            variant="neutral"
            className="shrink-0 text-xs"
            title={role}
          >
            {role}
          </Badge>
        ))}
        {hiddenRoles.length > 0 && (
          <AccessEntryExtraRolesPopover
            roles={hiddenRoles}
            className="inline-flex shrink-0"
            trigger={(
              <Badge variant="neutral" className="cursor-pointer text-xs">
                +
                {hiddenRoles.length}
              </Badge>
            )}
          />
        )}
      </div>
    </Td>
  );
}
