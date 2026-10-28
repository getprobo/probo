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

import { DotsThreeIcon, PencilIcon, TrashIcon, WarningIcon } from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { useDialogRef } from "@probo/ui";
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

import type { InternalControlListItem_internalControl$key } from "#/__generated__/core/InternalControlListItem_internalControl.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import InternalControlFormDialog from "#/pages/organizations/internal-controls/dialog/InternalControlFormDialog";

import {
  implementationStatusColor,
  nextDue,
} from "../../_lib/internalControlDisplay";
import { internalControlsPage } from "../variants";

import { InternalControlDeleteDialog } from "./InternalControlDeleteDialog";

const internalControlListItemFragment = graphql`
  fragment InternalControlListItem_internalControl on InternalControl {
    id
    name
    code
    implementationStatus
    nextEvidenceDue
    nextTestDue
    canUpdate: permission(action: "core:internal-control:update")
    canDelete: permission(action: "core:internal-control:delete")
    owner {
      fullName
    }
    ...InternalControlFormDialogInternalControlFragment
  }
`;

interface InternalControlListItemProps {
  internalControlKey: InternalControlListItem_internalControl$key;
  onChanged: () => void;
}

export function InternalControlListItem(props: InternalControlListItemProps) {
  const { internalControlKey, onChanged } = props;
  const { i18n, t } = useTranslation("organizations/internal-controls");
  const organizationId = useOrganizationId();
  const location = useLocation();
  const dialogRef = useDialogRef();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const internalControl = useFragment(internalControlListItemFragment, internalControlKey);
  const due = nextDue(internalControl.nextEvidenceDue, internalControl.nextTestDue);
  const ownerName = internalControl.owner?.fullName;
  const { identity, owner, due: dueSlot } = internalControlsPage();
  const statusKey = internalControl.implementationStatus.toLowerCase();

  return (
    <>
      <InternalControlFormDialog
        internalControl={internalControl}
        ref={dialogRef}
        onSaved={onChanged}
      />
      <InternalControlDeleteDialog
        open={deleteOpen}
        internalControlId={internalControl.id}
        name={internalControl.name}
        onOpenChange={setDeleteOpen}
        onDeleted={onChanged}
      />
      <TableRow align="center" interactive>
        <TableCell>
          {internalControl.code
            ? <Code className="whitespace-nowrap">{internalControl.code}</Code>
            : <Text size={2} color="faint">{t("list.values.none")}</Text>}
        </TableCell>
        <TableRowHeaderCell>
          <div className={identity()}>
            <TableLink
              to={{
                pathname: `/organizations/${organizationId}/governance/internal-controls/${internalControl.id}`,
                search: location.search,
              }}
            >
              {internalControl.name}
            </TableLink>
          </div>
        </TableRowHeaderCell>
        <TableCell>
          <Badge variant="soft" color={implementationStatusColor(internalControl.implementationStatus)}>
            {t(`list.statuses.${statusKey}`)}
          </Badge>
        </TableCell>
        <TableCell>
          {ownerName
            ? (
                <span className={owner()}>
                  <Avatar size={1} name={ownerName} />
                  <Text size={2}>{ownerName}</Text>
                </span>
              )
            : <Text size={2} color="faint">{t("list.values.unassigned")}</Text>}
        </TableCell>
        <TableCell>
          {due
            ? (
                <Text size={2} color={due.overdue ? "red" : "neutral"} className={dueSlot()}>
                  {due.overdue && <WarningIcon aria-hidden />}
                  {t(dueLabelKey(due.kind, due.overdue), {
                    date: dateFormat(i18n.language, due.at, { dateStyle: "medium" }),
                  })}
                </Text>
              )
            : <Text size={2} color="faint">{t("list.values.none")}</Text>}
        </TableCell>
        <TableCell interactive justify="end" width="3rem">
          {(internalControl.canUpdate || internalControl.canDelete) && (
            <Dropdown>
              <DropdownTrigger
                render={(
                  <IconButton variant="ghost" color="neutral" aria-label={t("list.actions.menu")}>
                    <DotsThreeIcon />
                  </IconButton>
                )}
              />
              <DropdownPopup>
                {internalControl.canUpdate && (
                  <DropdownItem iconStart={<PencilIcon />} onClick={() => dialogRef.current?.open()}>
                    {t("list.actions.edit")}
                  </DropdownItem>
                )}
                {internalControl.canDelete && (
                  <DropdownItem
                    color="error"
                    iconStart={<TrashIcon />}
                    onClick={() => setDeleteOpen(true)}
                  >
                    {t("list.actions.delete")}
                  </DropdownItem>
                )}
              </DropdownPopup>
            </Dropdown>
          )}
        </TableCell>
      </TableRow>
    </>
  );
}

function dueLabelKey(kind: "evidence" | "test", overdue: boolean) {
  if (kind === "evidence") {
    return overdue ? "list.due.evidenceOverdue" : "list.due.evidence";
  }
  return overdue ? "list.due.testOverdue" : "list.due.test";
}
