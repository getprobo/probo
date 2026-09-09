// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { getAuditStateVariant, parseDate } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  DropdownItem,
  IconTrashCan,
  Td,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { AuditListItem_audit$key } from "#/__generated__/core/AuditListItem_audit.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useDeleteAudit } from "#/hooks/graph/AuditGraph";

const auditListItemFragment = graphql`
  fragment AuditListItem_audit on Audit {
    id
    name
    validity {
      start
      end
    }
    auditDates {
      start
      end
    }
    reportFile {
      id
    }
    state
    framework {
      id
      name
    }
    canDelete: permission(action: "core:audit:delete")
  }
`;

interface AuditListItemProps {
  auditKey: AuditListItem_audit$key;
  connectionId: string;
  hasAnyAction: boolean;
}

export function AuditListItem({
  auditKey,
  connectionId,
  hasAnyAction,
}: AuditListItemProps) {
  const audit = useFragment(auditListItemFragment, auditKey);
  const organizationId = useOrganizationId();
  const { i18n, t } = useTranslation();
  const deleteAudit = useDeleteAudit(audit, connectionId);

  function formatPeriod(
    start: string | null | undefined,
    end: string | null | undefined,
  ) {
    const compactDateOptions: Intl.DateTimeFormatOptions = {
      year: "numeric",
      month: "short",
      day: "numeric",
    };

    if (start && end) {
      return new Intl.DateTimeFormat(
        i18n.language,
        compactDateOptions,
      ).formatRange(parseDate(start), parseDate(end));
    }
    if (start) {
      return t("auditsPage.row.period.from", {
        date: dateFormat(i18n.language, start, compactDateOptions),
      });
    }
    if (end) {
      return t("auditsPage.row.period.until", {
        date: dateFormat(i18n.language, end, compactDateOptions),
      });
    }
    return t("auditsPage.row.notSet");
  }

  return (
    <Tr to={`/organizations/${organizationId}/governance/audits/${audit.id}`}>
      <Td>{audit.name || t("auditsPage.row.untitled")}</Td>
      <Td>
        {audit.framework?.name ?? t("auditsPage.row.unknownFramework")}
      </Td>
      <Td>
        <Badge variant={getAuditStateVariant(audit.state)}>
          {t(`auditsPage.states.${audit.state.toLowerCase()}`)}
        </Badge>
      </Td>
      <Td className="whitespace-nowrap">
        {formatPeriod(audit.auditDates?.start, audit.auditDates?.end)}
      </Td>
      <Td className="whitespace-nowrap">
        {formatPeriod(audit.validity?.start, audit.validity?.end)}
      </Td>
      <Td>
        {audit.reportFile
          ? (
              <div className="flex flex-col">
                <Badge variant="success">
                  {t("auditsPage.row.uploaded")}
                </Badge>
              </div>
            )
          : (
              <Badge variant="neutral">
                {t("auditsPage.row.notUploaded")}
              </Badge>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {audit.canDelete && (
              <DropdownItem
                onClick={deleteAudit}
                variant="danger"
                icon={IconTrashCan}
              >
                {t("auditsPage.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
