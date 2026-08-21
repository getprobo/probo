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

import { Card } from "@probo/ui/src/v2/Card/Card";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Switch } from "@probo/ui/src/v2/Switch/Switch";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalStatusSectionFragment$key } from "#/__generated__/core/CompliancePortalStatusSectionFragment.graphql";
import { useUpdateCompliancePortalMutation } from "#/hooks/graph/CompliancePortalGraph";

import { statusSection } from "../variants";

const fragment = graphql`
  fragment CompliancePortalStatusSectionFragment on CompliancePortal {
    id
    active
    searchEngineIndexing
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

interface CompliancePortalStatusSectionProps {
  compliancePortalKey: CompliancePortalStatusSectionFragment$key;
}

export function CompliancePortalStatusSection({
  compliancePortalKey,
}: CompliancePortalStatusSectionProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { root, headingRow, cardBody, row, rowText } = statusSection();

  const compliancePortal = useFragment<CompliancePortalStatusSectionFragment$key>(
    fragment,
    compliancePortalKey,
  );

  const [updateCompliancePortal, isUpdating] = useUpdateCompliancePortalMutation();

  const handleToggleActive = (active: boolean) => {
    void updateCompliancePortal({
      variables: {
        input: {
          compliancePortalId: compliancePortal.id,
          active,
        },
      },
    });
  };

  const handleToggleSearchEngineIndexing = (indexable: boolean) => {
    void updateCompliancePortal({
      variables: {
        input: {
          compliancePortalId: compliancePortal.id,
          searchEngineIndexing: indexable ? "INDEXABLE" : "NOT_INDEXABLE",
        },
      },
    });
  };

  const indexingDisabled = !compliancePortal.canUpdate || !compliancePortal.active;
  const indexingSwitch = (
    <Switch
      checked={compliancePortal.searchEngineIndexing === "INDEXABLE"}
      disabled={indexingDisabled}
      aria-label={t("statusSection.indexing.title")}
      onCheckedChange={handleToggleSearchEngineIndexing}
    />
  );

  return (
    <div className={root()}>
      <div className={headingRow()}>
        <Heading level={2} size={3} weight="medium" highContrast>
          {t("statusSection.title")}
        </Heading>
        {isUpdating && <Spinner size={2} />}
      </div>
      <Card variant="soft" size={2}>
        <div className={cardBody()}>
          <div className={row()}>
            <div className={rowText()}>
              <Heading level={3} size={3} weight="medium" highContrast>
                {t("statusSection.activation.title")}
              </Heading>
              <Text size={2} color="neutral">
                {t("statusSection.activation.description")}
              </Text>
            </div>
            <Switch
              checked={compliancePortal.active}
              disabled={!compliancePortal.canUpdate}
              aria-label={t("statusSection.activation.title")}
              onCheckedChange={handleToggleActive}
            />
          </div>

          <Separator />

          <div className={row()}>
            <div className={rowText()}>
              <Heading level={3} size={3} weight="medium" highContrast>
                {t("statusSection.indexing.title")}
              </Heading>
              <Text size={2} color="neutral">
                {t("statusSection.indexing.description")}
              </Text>
            </div>
            {compliancePortal.active
              ? indexingSwitch
              : (
                  <Tooltip>
                    <TooltipTrigger render={<span>{indexingSwitch}</span>} />
                    <TooltipPopup>
                      {t("statusSection.indexing.disabledHint")}
                    </TooltipPopup>
                  </Tooltip>
                )}
          </div>
        </div>
      </Card>
    </div>
  );
}
