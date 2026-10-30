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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { DateRangeField } from "@probo/ui/src/v2/form/DateRangeField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";

import {
  type AuditLogActorType,
  auditLogActorTypes,
  type AuditLogOutcome,
  auditLogOutcomes,
  useAuditLogListFilters,
} from "../_lib/useAuditLogListFilters";
import { auditLogList } from "../variants";

import { AuditLogFilterField } from "./AuditLogFilterField";
import { AuditLogMoreFilters } from "./AuditLogMoreFilters";

const AUDIT_LOG_NS = "iam/organizations/audit-log";

export function AuditLogListFilters() {
  const { t, i18n } = useTranslation(AUDIT_LOG_NS);
  const {
    action,
    actorType,
    outcome,
    from,
    to,
    setAction,
    setActorType,
    setOutcome,
    setDates,
  } = useAuditLogListFilters();
  const { toolbar, action: actionSlot, filter, range } = auditLogList();
  const allActorTypesLabel = t("auditLogPage.filters.allActorTypes");
  const allOutcomesLabel = t("auditLogPage.filters.allOutcomes");
  const datesLabel = t("auditLogPage.filters.dates");

  return (
    <div className={toolbar()}>
      <AuditLogFilterField
        key={`action:${action}`}
        className={actionSlot()}
        icon={<MagnifyingGlassIcon />}
        label={t("auditLogPage.filters.action")}
        value={action}
        placeholder={t("auditLogPage.filters.actionPlaceholder")}
        onCommit={setAction}
      />
      <div className={filter()}>
        <Select
          value={outcome}
          onValueChange={(value: string | null) => {
            if (value == null) {
              setOutcome(null);
              return;
            }
            if ((auditLogOutcomes as readonly string[]).includes(value)) {
              setOutcome(value as AuditLogOutcome);
            }
          }}
        >
          <SelectTrigger
            size={2}
            placeholder={allOutcomesLabel}
            aria-label={t("auditLogPage.filters.outcome")}
          >
            {(value: AuditLogOutcome | null) => (
              value != null
                ? t(`auditLogPage.outcomes.${value}`)
                : allOutcomesLabel
            )}
          </SelectTrigger>
          <SelectPopup align="end">
            <SelectItem value={null}>{allOutcomesLabel}</SelectItem>
            {auditLogOutcomes.map(item => (
              <SelectItem key={item} value={item}>
                {t(`auditLogPage.outcomes.${item}`)}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
      <div className={filter()}>
        <Select
          value={actorType}
          onValueChange={(value: string | null) => {
            if (value == null) {
              setActorType(null);
              return;
            }
            if ((auditLogActorTypes as readonly string[]).includes(value)) {
              setActorType(value as AuditLogActorType);
            }
          }}
        >
          <SelectTrigger
            size={2}
            placeholder={allActorTypesLabel}
            aria-label={t("auditLogPage.filters.actorType")}
          >
            {(value: AuditLogActorType | null) => (
              value != null
                ? t(`auditLogPage.actorTypes.${value}`)
                : allActorTypesLabel
            )}
          </SelectTrigger>
          <SelectPopup align="end">
            <SelectItem value={null}>{allActorTypesLabel}</SelectItem>
            {auditLogActorTypes.map(item => (
              <SelectItem key={item} value={item}>
                {t(`auditLogPage.actorTypes.${item}`)}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
      <div className={range()}>
        <DateRangeField
          nullable
          value={{ from, to }}
          locale={i18n.language}
          placeholder={t("auditLogPage.filters.datesPlaceholder")}
          hint={t("auditLogPage.filters.datesHint")}
          aria-label={datesLabel}
          onValueChange={next => setDates(next.from, next.to)}
        />
      </div>
      <AuditLogMoreFilters />
    </div>
  );
}
