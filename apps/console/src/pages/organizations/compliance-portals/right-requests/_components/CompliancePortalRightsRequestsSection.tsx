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

import { Card, Toggle } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalRightsRequestsSectionFragment$key } from "#/__generated__/core/CompliancePortalRightsRequestsSectionFragment.graphql";
import { useUpdateCompliancePortalMutation } from "#/hooks/graph/CompliancePortalGraph";

const fragment = graphql`
  fragment CompliancePortalRightsRequestsSectionFragment on CompliancePortal {
    id
    capabilities {
      rightsRequests
    }
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

export function CompliancePortalRightsRequestsSection(props: {
  fragmentRef: CompliancePortalRightsRequestsSectionFragment$key;
}) {
  const { fragmentRef } = props;

  const { t } = useTranslation("organizations/compliance-portals");

  const compliancePortal = useFragment<CompliancePortalRightsRequestsSectionFragment$key>(
    fragment,
    fragmentRef,
  );

  const [updateCompliancePortal] = useUpdateCompliancePortalMutation();

  const handleToggleRightsRequests = async (rightsRequests: boolean) => {
    await updateCompliancePortal({
      variables: {
        input: {
          compliancePortalId: compliancePortal.id,
          capabilities: { rightsRequests },
        },
      },
    });
  };

  return (
    <div className="space-y-4">
      <Card padded>
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h3 className="font-medium">{t("statusSection.rightsRequests.title")}</h3>
            <p className="text-sm text-txt-tertiary">
              {t("statusSection.rightsRequests.description")}
            </p>
          </div>
          <Toggle
            checked={compliancePortal.capabilities.rightsRequests}
            onChange={checked => void handleToggleRightsRequests(checked)}
            disabled={!compliancePortal.canUpdate}
            aria-label={t("statusSection.rightsRequests.title")}
          />
        </div>
      </Card>
    </div>
  );
}
