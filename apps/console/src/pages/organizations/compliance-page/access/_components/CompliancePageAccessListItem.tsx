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

import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { ActionDropdown, DropdownItem, IconPencil, Td, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageAccessListItemFragment$key } from "#/__generated__/core/CompliancePageAccessListItemFragment.graphql";

import { CompliancePageAccessEditDialog } from "./CompliancePageAccessEditDialog";
import { NdaSignatureBadge } from "./NdaSignatureBadge";

const fragment = graphql`
  fragment CompliancePageAccessListItemFragment on TrustCenterAccess {
    id
    createdAt
    profile {
      fullName
      emailAddress
      state
    }
    activeCount
    pendingRequestCount
    ndaSignature {
      status
    }
    canUpdate: permission(action: "core:trust-center-access:update")
  }
`;

export function CompliancePageAccessListItem(props: {
  fragmentRef: CompliancePageAccessListItemFragment$key;
}) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();
  const [dialogOpen, setDialogOpen] = useState<boolean>(false);

  const access = useFragment<CompliancePageAccessListItemFragment$key>(fragment, fragmentRef);

  const isActive = access.profile.state === "ACTIVE";

  return (
    <>
      <Tr
        key={access.id}
        onClick={() => access.canUpdate && isActive && setDialogOpen(true)}
        className={`cursor-pointer hover:bg-bg-secondary transition-colors${!isActive ? " opacity-50" : ""}`}
      >
        <Td className="font-medium">{access.profile.fullName}</Td>
        <Td>{access.profile.emailAddress}</Td>
        <Td>{formatDate(access.createdAt)}</Td>
        <Td className="text-center">{access.activeCount}</Td>
        <Td className="text-center">
          {access.pendingRequestCount > 0 ? access.pendingRequestCount : ""}
        </Td>
        <Td>
          <div className="flex justify-center">
            {access.ndaSignature
              ? (
                  <NdaSignatureBadge status={access.ndaSignature.status} />
                )
              : (
                  <span className="text-txt-tertiary">-</span>
                )}
          </div>
        </Td>
        <Td noLink width={160} className="text-end">
          <div
            className="flex gap-2 justify-end"
            onClick={e => e.stopPropagation()}
          >
            {access.canUpdate && (
              <ActionDropdown>
                {isActive && (
                  <DropdownItem
                    icon={IconPencil}
                    onClick={() => setDialogOpen(true)}
                  >
                    {__("Edit")}
                  </DropdownItem>
                )}
              </ActionDropdown>
            )}
          </div>
        </Td>
      </Tr>

      {access.canUpdate && isActive && dialogOpen && (
        <CompliancePageAccessEditDialog
          access={access}
          onClose={() => setDialogOpen(false)}
        />
      )}
    </>
  );
}
