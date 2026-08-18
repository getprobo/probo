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

import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Card,
  DropdownItem,
  IconTrashCan,
  PageHeader,
  TabLink,
  Tabs,
  useConfirm,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { Outlet, useNavigate } from "react-router";

import type { RiskAnalysisDetailLayoutDeleteMutation } from "#/__generated__/core/RiskAnalysisDetailLayoutDeleteMutation.graphql";
import type { RiskAnalysisDetailLayoutQuery } from "#/__generated__/core/RiskAnalysisDetailLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";

import { formatMatrixSize } from "./_components/matrixSize";
import { UpdateRiskAnalysisDialog } from "./_components/UpdateRiskAnalysisDialog";

export const riskAnalysisDetailLayoutQuery = graphql`
  query RiskAnalysisDetailLayoutQuery($riskAnalysisId: ID!) {
    node(id: $riskAnalysisId) {
      __typename
      ... on RiskAnalysis {
        id
        name
        description
        period {
          start
          end
        }
        matrixSize {
          rows
          cols
        }
        createdAt
        updatedAt
        canUpdate: permission(action: "risk-management:risk-analysis:update")
        canDelete: permission(action: "risk-management:risk-analysis:delete")
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation RiskAnalysisDetailLayoutDeleteMutation(
    $input: DeleteRiskAnalysisInput!
    $connections: [ID!]!
  ) {
    deleteRiskAnalysis(input: $input) {
      deletedRiskAnalysisId @deleteEdge(connections: $connections)
    }
  }
`;

const RiskAnalysesConnectionKey = "RiskAnalysesPage_riskAnalyses";

interface RiskAnalysisDetailLayoutProps {
  queryRef: PreloadedQuery<RiskAnalysisDetailLayoutQuery>;
}

export default function RiskAnalysisDetailLayout({ queryRef }: RiskAnalysisDetailLayoutProps) {
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const data = usePreloadedQuery<RiskAnalysisDetailLayoutQuery>(
    riskAnalysisDetailLayoutQuery,
    queryRef,
  );
  const ra = data.node;
  const [deleteRiskAnalysis] = useMutation<RiskAnalysisDetailLayoutDeleteMutation>(
    deleteMutation,
    { errorToast: t("riskAnalysisDetailPage.errors.delete") },
  );

  usePageTitle(
    ra?.__typename === "RiskAnalysis"
      ? ra.name
      : t("riskAnalysisDetailPage.title"),
  );

  if (ra?.__typename !== "RiskAnalysis") {
    throw new NotFoundError("Risk analysis not found");
  }

  const raId = ra.id;
  const listConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    RiskAnalysesConnectionKey,
  );
  const listUrl = `/organizations/${organizationId}/risk-management/risk-analyses`;
  const baseTabUrl = `/organizations/${organizationId}/risk-management/risk-analyses/${raId}`;
  const periodLabel = ra.period
    ? `${ra.period.start ? dateFormat(i18n.language, ra.period.start) : "—"} – ${ra.period.end ? dateFormat(i18n.language, ra.period.end) : "—"}`
    : "—";

  const handleDelete = () => {
    confirm(
      async () => {
        await deleteRiskAnalysis({
          variables: {
            input: { riskAnalysisId: raId },
            connections: [listConnectionId],
          },
        });
        void navigate(listUrl);
      },
      { message: t("riskAnalysisDetailPage.deleteConfirmation") },
    );
  };

  return (
    <div className="space-y-6">
      <PageHeader title={ra.name} description={ra.description}>
        {ra.canUpdate
          ? (
              <UpdateRiskAnalysisDialog
                riskAnalysis={{
                  id: raId,
                  name: ra.name ?? "",
                  description: ra.description,
                  period: ra.period
                    ? { start: ra.period.start, end: ra.period.end }
                    : ra.period,
                }}
                canDelete={ra.canDelete}
                onDelete={handleDelete}
              />
            )
          : ra.canDelete
            ? (
                <ActionDropdown variant="secondary">
                  <DropdownItem
                    variant="danger"
                    icon={IconTrashCan}
                    onClick={handleDelete}
                  >
                    {t("riskAnalysisDetailPage.actions.delete")}
                  </DropdownItem>
                </ActionDropdown>
              )
            : null}
      </PageHeader>

      <Card padded>
        <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
          <div>
            <div className="text-xs text-txt-tertiary font-semibold mb-1">
              {t("riskAnalysisDetailPage.fields.period")}
            </div>
            <div className="text-sm text-txt-primary">
              {periodLabel}
            </div>
          </div>
          <div>
            <div className="text-xs text-txt-tertiary font-semibold mb-1">
              {t("riskAnalysisDetailPage.fields.matrixSize")}
            </div>
            <div className="text-sm text-txt-primary">
              {formatMatrixSize(ra.matrixSize.rows, ra.matrixSize.cols)}
            </div>
          </div>
          <div>
            <div className="text-xs text-txt-tertiary font-semibold mb-1">
              {t("riskAnalysisDetailPage.fields.createdAt")}
            </div>
            <div className="text-sm text-txt-primary">
              {dateFormat(i18n.language, ra.createdAt)}
            </div>
          </div>
          <div>
            <div className="text-xs text-txt-tertiary font-semibold mb-1">
              {t("riskAnalysisDetailPage.fields.updatedAt")}
            </div>
            <div className="text-sm text-txt-primary">
              {dateFormat(i18n.language, ra.updatedAt)}
            </div>
          </div>
        </div>
      </Card>

      <Tabs>
        <TabLink to={`${baseTabUrl}/treatment-plans`}>
          {t("riskAnalysisDetailPage.tabs.treatmentPlans")}
        </TabLink>
        <TabLink to={`${baseTabUrl}/diagrams`}>
          {t("riskAnalysisDetailPage.tabs.diagrams")}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
