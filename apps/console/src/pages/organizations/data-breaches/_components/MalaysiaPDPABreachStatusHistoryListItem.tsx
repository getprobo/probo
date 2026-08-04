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

import type { MalaysiaPDPABreachStatusHistoryListItem_history$key } from "#/__generated__/core/MalaysiaPDPABreachStatusHistoryListItem_history.graphql";

import { getBreachStatusBadgeVariant } from "../_lib/breachDisplay";

const historyFragment = graphql`
  fragment MalaysiaPDPABreachStatusHistoryListItem_history on MalaysiaPDPABreachStatusHistory {
    fromStatus
    toStatus
    changedByProfileId
    reason
    createdAt
  }
`;

interface MalaysiaPDPABreachStatusHistoryListItemProps {
  historyKey: MalaysiaPDPABreachStatusHistoryListItem_history$key;
}

export function MalaysiaPDPABreachStatusHistoryListItem({
  historyKey,
}: MalaysiaPDPABreachStatusHistoryListItemProps) {
  const { i18n, t } = useTranslation("organizations/data-breaches");
  const history = useFragment(historyFragment, historyKey);

  return (
    <Tr>
      <Td>
        <div className="flex flex-wrap items-center gap-2">
          {history.fromStatus
            ? (
                <Badge variant={getBreachStatusBadgeVariant(history.fromStatus)}>
                  {t(`statuses.${history.fromStatus}`)}
                </Badge>
              )
            : (
                <span className="text-txt-tertiary">{t("history.created")}</span>
              )}
          <span aria-hidden="true" className="text-txt-tertiary">→</span>
          <Badge variant={getBreachStatusBadgeVariant(history.toStatus)}>
            {t(`statuses.${history.toStatus}`)}
          </Badge>
        </div>
      </Td>
      <Td>{history.reason || t("common.notProvided")}</Td>
      <Td>
        <time dateTime={history.createdAt}>
          {dateTimeFormat(i18n.language, history.createdAt)}
        </time>
      </Td>
      <Td>
        <span title={history.changedByProfileId}>
          {t("history.profile", {
            id: history.changedByProfileId.slice(-8),
          })}
        </span>
      </Td>
    </Tr>
  );
}
