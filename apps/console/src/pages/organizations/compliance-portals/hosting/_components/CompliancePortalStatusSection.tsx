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

import { BroadcastIcon, MagnifyingGlassIcon } from "@phosphor-icons/react";
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalStatusSectionFragment$key } from "#/__generated__/core/CompliancePortalStatusSectionFragment.graphql";
import { useUpdateCompliancePortalMutation } from "#/hooks/graph/CompliancePortalGraph";

import { statusSection } from "../variants";

import { CompliancePortalStatusCard } from "./CompliancePortalStatusCard";

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
  const { root, intro, headingRow, grid } = statusSection();

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

  return (
    <div className={root()}>
      <div className={intro()}>
        <div className={headingRow()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("statusSection.title")}
          </Heading>
          {isUpdating && <Spinner size={2} />}
        </div>
        <Text size={2} color="neutral">
          {t("statusSection.description")}
        </Text>
      </div>
      <div className={grid()}>
        <CompliancePortalStatusCard
          title={t(
            compliancePortal.active
              ? "statusSection.activation.titleActive"
              : "statusSection.activation.titleInactive",
          )}
          description={t("statusSection.activation.description")}
          icon={<BroadcastIcon size={24} weight="duotone" />}
          checked={compliancePortal.active}
          disabled={!compliancePortal.canUpdate}
          onCheckedChange={handleToggleActive}
        />
        <CompliancePortalStatusCard
          title={t("statusSection.indexing.title")}
          description={t("statusSection.indexing.description")}
          icon={<MagnifyingGlassIcon size={24} weight="duotone" />}
          checked={
            compliancePortal.active && compliancePortal.searchEngineIndexing === "INDEXABLE"
          }
          disabled={indexingDisabled}
          disabledHint={
            compliancePortal.active
              ? undefined
              : t("statusSection.indexing.disabledHint")
          }
          onCheckedChange={handleToggleSearchEngineIndexing}
        />
      </div>
    </div>
  );
}
