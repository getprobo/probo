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

import { Card, Spinner, Toggle, useToast } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageStatusSectionFragment$key } from "#/__generated__/core/CompliancePageStatusSectionFragment.graphql";
import { useUpdateCompliancePageMutation } from "#/hooks/graph/CompliancePageGraph";

const fragment = graphql`
  fragment CompliancePageStatusSectionFragment on Organization {
    compliancePage: compliancePortal {
      id
      active
      searchEngineIndexing
      canUpdate: permission(action: "compliance-portal:portal:update")
    }
  }
`;

export function CompliancePageStatusSection(props: {
  fragmentRef: CompliancePageStatusSectionFragment$key;
}) {
  const { fragmentRef } = props;

  const { t } = useTranslation("organizations/compliance-page");
  const { toast } = useToast();

  const organization = useFragment<CompliancePageStatusSectionFragment$key>(
    fragment,
    fragmentRef,
  );

  const [updateCompliancePage, isUpdating] = useUpdateCompliancePageMutation();

  const handleToggleActive = async (active: boolean) => {
    if (!organization.compliancePage?.id) {
      toast({
        title: t("statusSection.errors.title"),
        description: t("statusSection.errors.notFound"),
        variant: "error",
      });
      return;
    }

    await updateCompliancePage({
      variables: {
        input: {
          compliancePortalId: organization.compliancePage.id,
          active,
        },
      },
    });
  };

  const handleToggleSearchEngineIndexing = async (indexable: boolean) => {
    if (!organization.compliancePage?.id) {
      toast({
        title: t("statusSection.errors.title"),
        description: t("statusSection.errors.notFound"),
        variant: "error",
      });
      return;
    }

    await updateCompliancePage({
      variables: {
        input: {
          compliancePortalId: organization.compliancePage.id,
          searchEngineIndexing: indexable ? "INDEXABLE" : "NOT_INDEXABLE",
        },
      },
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">
          {t("statusSection.title")}
        </h2>
        {isUpdating && <Spinner />}
      </div>
      <Card padded className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h3 className="font-medium">{t("statusSection.activation.title")}</h3>
            <p className="text-sm text-txt-tertiary">
              {t("statusSection.activation.description")}
            </p>
          </div>
          <Toggle
            checked={!!organization.compliancePage?.active}
            onChange={checked => void handleToggleActive(checked)}
            disabled={!organization.compliancePage?.canUpdate}
          />
        </div>

        <div className="flex items-center justify-between border-t border-border-solid pt-4">
          <div className="space-y-1">
            <h3 className="font-medium">{t("statusSection.indexing.title")}</h3>
            <p className="text-sm text-txt-tertiary">
              {t("statusSection.indexing.description")}
            </p>
          </div>
          <span
            title={
              !organization.compliancePage?.active
                ? t("statusSection.indexing.disabledHint")
                : undefined
            }
          >
            <Toggle
              checked={
                organization.compliancePage?.searchEngineIndexing === "INDEXABLE"
              }
              onChange={checked =>
                void handleToggleSearchEngineIndexing(checked)}
              disabled={
                !organization.compliancePage?.canUpdate
                || !organization.compliancePage?.active
              }
            />
          </span>
        </div>
      </Card>
    </div>
  );
}
