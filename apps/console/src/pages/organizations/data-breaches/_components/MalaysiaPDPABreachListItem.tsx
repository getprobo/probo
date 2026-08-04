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

import { dateTimeFormat } from "@probo/i18n";
import { Badge, Td, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { MalaysiaPDPABreachListItem_incident$key } from "#/__generated__/core/MalaysiaPDPABreachListItem_incident.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  getBreachDecisionBadgeVariant,
  getBreachStatusBadgeVariant,
} from "../_lib/breachDisplay";

const incidentFragment = graphql`
  fragment MalaysiaPDPABreachListItem_incident on MalaysiaPDPABreachIncident {
    id
    title
    status
    awarenessAt
    affectedDataSubjects
    significantHarm
    significantScale
    notificationRecommendation
    commissionerNotificationDueAt
    commissionerNotificationOverdue
  }
`;

interface MalaysiaPDPABreachListItemProps {
  incidentKey: MalaysiaPDPABreachListItem_incident$key;
}

export function MalaysiaPDPABreachListItem({
  incidentKey,
}: MalaysiaPDPABreachListItemProps) {
  const { i18n, t } = useTranslation("organizations/data-breaches");
  const organizationId = useOrganizationId();
  const incident = useFragment(incidentFragment, incidentKey);

  return (
    <Tr to={`/organizations/${organizationId}/data-breaches/${incident.id}`}>
      <Td>
        <div className="space-y-1">
          <span className="font-medium">{incident.title}</span>
          {(incident.significantHarm || incident.significantScale) && (
            <span className="block text-xs text-txt-tertiary">
              {incident.significantHarm
                ? t("list.flags.significantHarm")
                : t("list.flags.significantScale")}
            </span>
          )}
        </div>
      </Td>
      <Td>
        <Badge variant={getBreachStatusBadgeVariant(incident.status)}>
          {t(`statuses.${incident.status}`)}
        </Badge>
      </Td>
      <Td>
        <time dateTime={incident.awarenessAt}>
          {dateTimeFormat(i18n.language, incident.awarenessAt)}
        </time>
      </Td>
      <Td>
        {new Intl.NumberFormat(i18n.language).format(
          incident.affectedDataSubjects,
        )}
      </Td>
      <Td>
        <Badge
          variant={getBreachDecisionBadgeVariant(
            incident.notificationRecommendation,
          )}
        >
          {t(`decisions.${incident.notificationRecommendation}`)}
        </Badge>
      </Td>
      <Td>
        {incident.commissionerNotificationDueAt
          ? (
              <div className="flex flex-col items-start gap-1">
                <time dateTime={incident.commissionerNotificationDueAt}>
                  {dateTimeFormat(
                    i18n.language,
                    incident.commissionerNotificationDueAt,
                  )}
                </time>
                {incident.commissionerNotificationOverdue && (
                  <Badge variant="danger">{t("deadlines.overdue")}</Badge>
                )}
              </div>
            )
          : (
              <span className="text-txt-tertiary">{t("common.notApplicable")}</span>
            )}
      </Td>
    </Tr>
  );
}
