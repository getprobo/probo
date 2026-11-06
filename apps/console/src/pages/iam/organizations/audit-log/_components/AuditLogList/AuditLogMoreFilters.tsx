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

import { FunnelIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import { useTranslation } from "react-i18next";

import { useAuditLogListFilters } from "../../_lib/useAuditLogListFilters";
import { auditLogList } from "../variants";

import { AuditLogFilterField } from "./AuditLogFilterField";

const AUDIT_LOG_NS = "iam/organizations/audit-log";

export function AuditLogMoreFilters() {
  const { t } = useTranslation(AUDIT_LOG_NS);
  const {
    actorId,
    resourceType,
    resourceId,
    requestId,
    setActorId,
    setResourceType,
    setResourceId,
    setRequestId,
  } = useAuditLogListFilters();
  const { popup, panel, field } = auditLogList();
  const count = [actorId, resourceType, resourceId, requestId]
    .filter(value => value !== "")
    .length;

  return (
    <Popover>
      <PopoverTrigger
        render={(
          <Button
            variant={count > 0 ? "soft" : "surface"}
            color="neutral"
            iconStart={<FunnelIcon />}
          >
            {count > 0
              ? t("auditLogPage.filters.moreCount", { count })
              : t("auditLogPage.filters.more")}
          </Button>
        )}
      />
      <PopoverPopup align="end" className={popup()}>
        <div className={panel()}>
          <Field label={t("auditLogPage.filters.actorId")}>
            <AuditLogFilterField
              className={field()}
              label={t("auditLogPage.filters.actorId")}
              value={actorId}
              onCommit={setActorId}
            />
          </Field>
          <Field label={t("auditLogPage.filters.resourceType")}>
            <AuditLogFilterField
              className={field()}
              label={t("auditLogPage.filters.resourceType")}
              value={resourceType}
              onCommit={setResourceType}
            />
          </Field>
          <Field label={t("auditLogPage.filters.resourceId")}>
            <AuditLogFilterField
              className={field()}
              label={t("auditLogPage.filters.resourceId")}
              value={resourceId}
              onCommit={setResourceId}
            />
          </Field>
          <Field label={t("auditLogPage.filters.requestId")}>
            <AuditLogFilterField
              className={field()}
              label={t("auditLogPage.filters.requestId")}
              value={requestId}
              onCommit={setRequestId}
            />
          </Field>
        </div>
      </PopoverPopup>
    </Popover>
  );
}
