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

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { FindingPropertiesSection_finding$key } from "#/__generated__/core/FindingPropertiesSection_finding.graphql";
import type { useUpdateFindingMutation$variables } from "#/__generated__/core/useUpdateFindingMutation.graphql";

import {
  findingDueIsOverdue,
  isFindingPriority,
  isFindingStatus,
} from "../_lib/findingOptions";
import { useUpdateFinding } from "../_lib/useUpdateFinding";

import { FindingPropertiesDate } from "./FindingPropertiesDate";
import { FindingPropertiesKind } from "./FindingPropertiesKind";
import { FindingPropertiesLayout } from "./FindingPropertiesLayout";
import { FindingPropertiesOwner } from "./FindingPropertiesOwner";
import { FindingPropertiesPriority } from "./FindingPropertiesPriority";
import { FindingPropertiesRisk } from "./FindingPropertiesRisk";
import { FindingPropertiesStatus } from "./FindingPropertiesStatus";

const findingPropertiesFragment = graphql`
  fragment FindingPropertiesSection_finding on Finding {
    id
    kind
    status
    priority
    identifiedOn
    dueDate
    canUpdate: permission(action: "core:finding:update")
    owner {
      id
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
    }
    risk {
      id
      name
    }
  }
`;

type UpdateInput = useUpdateFindingMutation$variables["input"];

interface FindingPropertiesSectionProps {
  findingKey: FindingPropertiesSection_finding$key;
}

export function FindingPropertiesSection({ findingKey }: FindingPropertiesSectionProps) {
  const finding = useFragment(findingPropertiesFragment, findingKey);
  const { t } = useTranslation("organizations/findings");
  const [updateFinding, isUpdating] = useUpdateFinding();
  const canUpdate = finding.canUpdate;
  const [acceptingRisk, setAcceptingRisk] = useState(false);
  const disabled = !canUpdate || isUpdating;
  const owner = finding.owner;

  function save(input: Omit<UpdateInput, "id">) {
    return updateFinding({
      variables: { input: { id: finding.id, ...input } },
    }).then(() => true, () => false);
  }

  return (
    <FindingPropertiesLayout>
      <FindingPropertiesKind kind={finding.kind} />
      <FindingPropertiesStatus
        status={finding.status}
        canUpdate={canUpdate}
        disabled={disabled}
        onChange={(status) => {
          if (!isFindingStatus(status)) {
            return;
          }
          if (status === "RISK_ACCEPTED" && finding.risk == null) {
            setAcceptingRisk(true);
            return;
          }
          setAcceptingRisk(false);
          void save({
            status,
            riskId: status === "RISK_ACCEPTED" ? finding.risk?.id ?? null : null,
          });
        }}
      />
      {(acceptingRisk || finding.status === "RISK_ACCEPTED") && (
        <FindingPropertiesRisk
          risk={finding.risk == null ? null : { id: finding.risk.id, name: finding.risk.name }}
          disabled={disabled}
          onSelect={(risk) => {
            if (acceptingRisk) {
              void save({ status: "RISK_ACCEPTED", riskId: risk.id }).then((saved) => {
                if (saved) {
                  setAcceptingRisk(false);
                }
              });
              return;
            }
            void save({ riskId: risk.id });
          }}
        />
      )}
      <FindingPropertiesPriority
        priority={finding.priority}
        canUpdate={canUpdate}
        disabled={disabled}
        onChange={(priority) => {
          if (isFindingPriority(priority)) {
            void save({ priority });
          }
        }}
      />
      <FindingPropertiesOwner
        owner={owner == null
          ? null
          : {
              id: owner.id,
              fullName: owner.fullName,
              emailAddress: owner.emailAddress,
              avatarUrl: owner.avatar?.downloadUrl,
            }}
        canUpdate={canUpdate}
        disabled={disabled}
        onChange={(ownerId) => {
          if (ownerId) {
            void save({ ownerId });
          }
        }}
      />
      <FindingPropertiesDate
        identified
        value={finding.identifiedOn}
        placeholder={t("details.placeholders.dateIdentified")}
        canUpdate={canUpdate}
        disabled={disabled}
        onChange={(identifiedOn) => {
          void save({ identifiedOn });
        }}
      />
      <FindingPropertiesDate
        value={finding.dueDate}
        placeholder={t("details.placeholders.dueDate")}
        overdue={findingDueIsOverdue(finding.dueDate, finding.status)}
        canUpdate={canUpdate}
        disabled={disabled}
        onChange={(dueDate) => {
          void save({ dueDate });
        }}
      />
    </FindingPropertiesLayout>
  );
}
