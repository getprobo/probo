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
  Button,
  Card,
  IconPageTextLine,
  IconUpload,
  PageHeader,
  TabLink,
  Tabs,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { Link, Outlet, useNavigate } from "react-router";

import type { RiskAnalysisDetailLayoutQuery } from "#/__generated__/core/RiskAnalysisDetailLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { formatMatrixSize } from "./_components/matrixSize";
import { PublishRiskAnalysisDialog } from "./_components/PublishRiskAnalysisDialog";
import { RiskAnalysisActions } from "./_components/RiskAnalysisActions";
import { RiskAnalysisDescriptionSection } from "./_components/RiskAnalysisDescriptionSection";
import { riskAnalysisDetailSummary } from "./variants";

export const riskAnalysisDetailLayoutQuery = graphql`
  query RiskAnalysisDetailLayoutQuery($riskAnalysisId: ID!) {
    node(id: $riskAnalysisId) {
      __typename
      ... on RiskAnalysis {
        id
        name
        period {
          start
          end
        }
        matrixSize {
          rows
          cols
        }
        createdAt
        document {
          id
          defaultApprovers {
            id
          }
        }
        canPublish: permission(action: "risk-management:risk-analysis:publish")
        ...RiskAnalysisActions_riskAnalysis
        ...RiskAnalysisDescriptionSection_riskAnalysis
      }
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
  const data = usePreloadedQuery<RiskAnalysisDetailLayoutQuery>(
    riskAnalysisDetailLayoutQuery,
    queryRef,
  );
  const ra = data.node;

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
  const { body, meta, label, value } = riskAnalysisDetailSummary();

  return (
    <div className="space-y-6">
      <PageHeader title={ra.name}>
        {ra.document?.id && (
          <Button variant="secondary" asChild>
            <Link
              to={`/organizations/${organizationId}/governance/documents/${ra.document.id}`}
            >
              <IconPageTextLine size={16} />
              {t("riskAnalysisDetailPage.actions.document")}
            </Link>
          </Button>
        )}
        {ra.canPublish && (
          <PublishRiskAnalysisDialog
            riskAnalysisId={ra.id}
            defaultApproverIds={ra.document?.defaultApprovers?.map(approver => approver.id) ?? []}
            onPublished={(documentId) => {
              void navigate(
                `/organizations/${organizationId}/governance/documents/${documentId}`,
              );
            }}
          >
            <Button icon={IconUpload}>
              {t("riskAnalysisDetailPage.actions.publish")}
            </Button>
          </PublishRiskAnalysisDialog>
        )}
        <RiskAnalysisActions
          riskAnalysisKey={ra}
          connectionId={listConnectionId}
          variant="secondary"
          onDeleted={() => {
            void navigate(listUrl);
          }}
        />
      </PageHeader>

      <Card padded>
        <div className={body()}>
          <RiskAnalysisDescriptionSection riskAnalysisKey={ra} />
          <div className={meta()}>
            <div>
              <div className={label()}>
                {t("riskAnalysisDetailPage.fields.period")}
              </div>
              <div className={value()}>
                {periodLabel}
              </div>
            </div>
            <div>
              <div className={label()}>
                {t("riskAnalysisDetailPage.fields.matrixSize")}
              </div>
              <div className={value()}>
                {formatMatrixSize(ra.matrixSize.rows, ra.matrixSize.cols)}
              </div>
            </div>
            <div>
              <div className={label()}>
                {t("riskAnalysisDetailPage.fields.createdAt")}
              </div>
              <div className={value()}>
                {dateFormat(i18n.language, ra.createdAt)}
              </div>
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
