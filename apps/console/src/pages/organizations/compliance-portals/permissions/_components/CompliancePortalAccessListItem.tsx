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

import { PencilSimpleIcon } from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { MouseEvent } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePortalAccessListItemFragment$key } from "#/__generated__/core/CompliancePortalAccessListItemFragment.graphql";

import { accessListItem } from "../variants";

import { NdaSignatureBadge } from "./NdaSignatureBadge";

const fragment = graphql`
  fragment CompliancePortalAccessListItemFragment on CompliancePortalAccess {
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
  }
`;

interface CompliancePortalAccessListItemProps {
  accessKey: CompliancePortalAccessListItemFragment$key;
}

export function CompliancePortalAccessListItem({
  accessKey,
}: CompliancePortalAccessListItemProps) {
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const navigate = useNavigate();
  const access = useFragment(fragment, accessKey);
  const isActive = access.profile.state === "ACTIVE";
  const { row, person, personCopy } = accessListItem({
    interactive: true,
    inactive: !isActive,
  });

  function handleRowClick() {
    void navigate(access.id);
  }

  function handleLinkClick(event: MouseEvent) {
    event.stopPropagation();
  }

  return (
    <TableRow
      align="center"
      className={row()}
      onClick={handleRowClick}
    >
      <TableRowHeaderCell minWidth="12rem">
        <div className={person()}>
          <Avatar
            size={2}
            variant="soft"
            color="gold"
            fallback={access.profile.fullName.charAt(0).toUpperCase() || "?"}
          />
          <Link
            to={access.id}
            size={2}
            color="neutral"
            highContrast
            underline={false}
            onClick={handleLinkClick}
            className={personCopy()}
          >
            {access.profile.fullName}
          </Link>
        </div>
      </TableRowHeaderCell>
      <TableCell>
        <Text size={1} color="gold" className="truncate">
          {access.profile.emailAddress}
        </Text>
      </TableCell>
      <TableCell>
        <Text size={1} color="faint">
          {dateFormat(i18n.language, access.createdAt)}
        </Text>
      </TableCell>
      <TableCell>
        <Text size={1} color="faint">
          {access.activeCount}
        </Text>
      </TableCell>
      <TableCell>
        <Text size={1} color="faint">
          {access.pendingRequestCount}
        </Text>
      </TableCell>
      <TableCell>
        {access.ndaSignature
          ? <NdaSignatureBadge status={access.ndaSignature.status} />
          : null}
      </TableCell>
      <TableCell>
        <ButtonLink
          to={access.id}
          size={1}
          variant="ghost"
          color="neutral"
          aria-label={t("accessListItem.actions.open")}
          onClick={handleLinkClick}
        >
          <PencilSimpleIcon />
        </ButtonLink>
      </TableCell>
    </TableRow>
  );
}
