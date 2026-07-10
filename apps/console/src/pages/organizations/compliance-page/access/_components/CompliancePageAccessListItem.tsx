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
    canUpdate: permission(action: "compliance-portal:portal-access:update")
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
