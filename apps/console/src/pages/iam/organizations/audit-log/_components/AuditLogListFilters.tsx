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

import { DateField } from "@probo/ui/src/v2/form/DateField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
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

const AUDIT_LOG_NS = "iam/organizations/audit-log";

export function AuditLogListFilters() {
  const { t, i18n } = useTranslation(AUDIT_LOG_NS);
  const {
    action,
    actorType,
    outcome,
    actorId,
    resourceType,
    resourceId,
    from,
    to,
    requestId,
    setAction,
    setActorType,
    setOutcome,
    setActorId,
    setResourceType,
    setResourceId,
    setFrom,
    setTo,
    setRequestId,
  } = useAuditLogListFilters();
  const { toolbar, tools, action: actionSlot, filters, filter, field } = auditLogList();
  const allActorTypesLabel = t("auditLogPage.filters.allActorTypes");
  const allOutcomesLabel = t("auditLogPage.filters.allOutcomes");
  const rangeInverted = from !== "" && to !== "" && from > to;

  return (
    <>
      <div className={toolbar()}>
        <AuditLogFilterField
          key={`action:${action}`}
          className={actionSlot()}
          label={t("auditLogPage.filters.action")}
          value={action}
          placeholder={t("auditLogPage.filters.actionPlaceholder")}
          onCommit={setAction}
        />
        <div className={filters()}>
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
        </div>
      </div>
      <div className={tools()}>
        <AuditLogFilterField
          key={`actor:${actorId}`}
          className={field()}
          label={t("auditLogPage.filters.actorId")}
          value={actorId}
          onCommit={setActorId}
        />
        <AuditLogFilterField
          key={`resource-type:${resourceType}`}
          className={field()}
          label={t("auditLogPage.filters.resourceType")}
          value={resourceType}
          onCommit={setResourceType}
        />
        <AuditLogFilterField
          key={`resource:${resourceId}`}
          className={field()}
          label={t("auditLogPage.filters.resourceId")}
          value={resourceId}
          onCommit={setResourceId}
        />
        <AuditLogFilterField
          key={`request:${requestId}`}
          className={field()}
          label={t("auditLogPage.filters.requestId")}
          value={requestId}
          onCommit={setRequestId}
        />
        <div className={filter()}>
          <DateField
            size={2}
            nullable
            value={from}
            locale={i18n.language}
            placeholder={t("auditLogPage.filters.from")}
            aria-label={t("auditLogPage.filters.from")}
            max={to === "" ? undefined : to}
            onValueChange={setFrom}
          />
        </div>
        <div className={filter()}>
          <DateField
            size={2}
            nullable
            value={to}
            locale={i18n.language}
            placeholder={t("auditLogPage.filters.to")}
            aria-label={t("auditLogPage.filters.to")}
            aria-invalid={rangeInverted || undefined}
            min={from === "" ? undefined : from}
            onValueChange={setTo}
          />
        </div>
      </div>
      {rangeInverted && (
        <Text size={1} color="red">
          {t("auditLogPage.filters.range")}
        </Text>
      )}
    </>
  );
}
