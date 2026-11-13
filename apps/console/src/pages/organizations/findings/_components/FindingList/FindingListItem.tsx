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

import { DotsThreeIcon, TrashIcon, UserCircleDashedIcon } from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableLink } from "@probo/ui/src/v2/Table/TableLink";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { useLocation } from "react-router";
import { graphql } from "relay-runtime";

import type { FindingListItem_finding$key } from "#/__generated__/core/FindingListItem_finding.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { findingKindKey } from "../../_lib/findingOptions";
import { findingColumnWidth, findingsPage } from "../variants";

import { FindingDeleteDialog } from "./FindingDeleteDialog";

const findingListItemFragment = graphql`
  fragment FindingListItem_finding on Finding {
    id
    referenceId
    title
    kind
    status
    priority
    dueDate
    canDelete: permission(action: "core:finding:delete")
    owner {
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
    }
  }
`;

interface FindingListItemProps {
  findingKey: FindingListItem_finding$key;
  onDeleted: () => void;
}

export function FindingListItem(props: FindingListItemProps) {
  const { findingKey, onDeleted } = props;
  const { i18n, t } = useTranslation("organizations/findings");
  const organizationId = useOrganizationId();
  const location = useLocation();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const finding = useFragment(findingListItemFragment, findingKey);
  const ownerLabel = finding.owner?.fullName;
  const { name, nameCell, owner, ownerIcon, ownerName } = findingsPage();

  return (
    <>
      <FindingDeleteDialog
        open={deleteOpen}
        findingId={finding.id}
        referenceId={finding.referenceId}
        onOpenChange={setDeleteOpen}
        onDeleted={onDeleted}
      />
      <TableRow align="start" interactive>
        <TableCell width={findingColumnWidth.reference}>
          <Code>{finding.referenceId}</Code>
        </TableCell>
        <TableRowHeaderCell className={nameCell()} minWidth={findingColumnWidth.title}>
          <TableLink
            to={{
              pathname: `/organizations/${organizationId}/governance/findings/${finding.id}`,
              search: location.search,
            }}
          >
            <Text size={2} highContrast className={name()}>
              {finding.title}
            </Text>
          </TableLink>
        </TableRowHeaderCell>
        <TableCell width={findingColumnWidth.kind}>
          <Badge variant="soft" color="neutral">
            {t(`list.kinds.${findingKindKey(finding.kind)}`)}
          </Badge>
        </TableCell>
        <TableCell width={findingColumnWidth.status}>
          <Badge variant="soft" color={statusColor(finding.status)}>
            {t(`list.statuses.${finding.status.toLowerCase()}`)}
          </Badge>
        </TableCell>
        <TableCell width={findingColumnWidth.priority}>
          <Badge variant="soft" color={priorityColor(finding.priority)}>
            {t(`list.priorities.${finding.priority.toLowerCase()}`)}
          </Badge>
        </TableCell>
        <TableCell width={findingColumnWidth.owner}>
          <span className={owner()}>
            {finding.owner == null
              ? <UserCircleDashedIcon className={ownerIcon()} aria-hidden />
              : (
                  <Avatar
                    size={1}
                    radius="full"
                    name={ownerLabel}
                    email={finding.owner.emailAddress}
                    src={finding.owner.avatar?.downloadUrl}
                  />
                )}
            <Text size={2} color={finding.owner == null ? "faint" : undefined} className={ownerName()}>
              {ownerLabel ?? t("list.values.unassigned")}
            </Text>
          </span>
        </TableCell>
        <TableCell width={findingColumnWidth.due}>
          {finding.dueDate
            ? (
                <Text size={2}>
                  {dateFormat(i18n.language, finding.dueDate, { dateStyle: "medium" })}
                </Text>
              )
            : <Text size={2} color="faint">{t("list.values.noDueDate")}</Text>}
        </TableCell>
        <TableCell interactive justify="end" width={findingColumnWidth.actions}>
          {finding.canDelete && (
            <Dropdown>
              <DropdownTrigger
                render={(
                  <IconButton variant="ghost" color="neutral" aria-label={t("list.actions.menu")}>
                    <DotsThreeIcon />
                  </IconButton>
                )}
              />
              <DropdownPopup>
                <DropdownItem
                  color="error"
                  iconStart={<TrashIcon />}
                  onClick={() => setDeleteOpen(true)}
                >
                  {t("list.actions.delete")}
                </DropdownItem>
              </DropdownPopup>
            </Dropdown>
          )}
        </TableCell>
      </TableRow>
    </>
  );
}

function statusColor(status: string) {
  switch (status) {
    case "OPEN":
      return "red" as const;
    case "IN_PROGRESS":
    case "RISK_ACCEPTED":
      return "gold" as const;
    case "CLOSED":
    case "MITIGATED":
      return "green" as const;
    case "FALSE_POSITIVE":
      return "neutral" as const;
    default:
      return "neutral" as const;
  }
}

function priorityColor(priority: string) {
  switch (priority) {
    case "HIGH":
      return "red" as const;
    case "MEDIUM":
      return "gold" as const;
    default:
      return "neutral" as const;
  }
}
