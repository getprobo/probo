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

import { Button, Card, IconChevronDown, Spinner } from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { RiskAnalysisDiagramsPage_analysis$key } from "#/__generated__/core/RiskAnalysisDiagramsPage_analysis.graphql";
import type { RiskAnalysisDiagramsPagePaginationQuery } from "#/__generated__/core/RiskAnalysisDiagramsPagePaginationQuery.graphql";
import type { RiskAnalysisDiagramsPageQuery } from "#/__generated__/core/RiskAnalysisDiagramsPageQuery.graphql";

import { CreateDiagramDialog } from "../_components/CreateDiagramDialog";
import { DiagramCard } from "../_components/DiagramCard";

const PAGE_SIZE = 50;

export const riskAnalysisDiagramsPageQuery = graphql`
  query RiskAnalysisDiagramsPageQuery($riskAnalysisId: ID!) {
    node(id: $riskAnalysisId) {
      ... on RiskAnalysis {
        id
        ...RiskAnalysisDiagramsPage_analysis
      }
    }
  }
`;

const diagramsFragment = graphql`
  fragment RiskAnalysisDiagramsPage_analysis on RiskAnalysis
  @refetchable(queryName: "RiskAnalysisDiagramsPagePaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    id
    canCreateDiagram: permission(action: "risk-management:diagram:create")
    diagrams(first: $first, after: $after)
      @connection(key: "RiskAnalysisDiagramsPage_diagrams", filters: []) {
      __id
      edges {
        node {
          id
          ...DiagramCardFragment
        }
      }
    }
  }
`;

interface RiskAnalysisDiagramsPageProps {
  queryRef: PreloadedQuery<RiskAnalysisDiagramsPageQuery>;
}

export default function RiskAnalysisDiagramsPage({ queryRef }: RiskAnalysisDiagramsPageProps) {
  const data = usePreloadedQuery<RiskAnalysisDiagramsPageQuery>(
    riskAnalysisDiagramsPageQuery,
    queryRef,
  );
  const ra = data.node;

  if (!ra?.id) {
    return null;
  }

  return <RiskAnalysisDiagramsList analysisKey={ra} />;
}

function RiskAnalysisDiagramsList({
  analysisKey,
}: {
  analysisKey: RiskAnalysisDiagramsPage_analysis$key;
}) {
  const { t } = useTranslation();
  const { data: ra, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    RiskAnalysisDiagramsPagePaginationQuery,
    RiskAnalysisDiagramsPage_analysis$key
  >(diagramsFragment, analysisKey);

  const diagrams = ra.diagrams?.edges.map(e => e.node) ?? [];
  const diagramsConnectionId = ra.diagrams?.__id ?? "";

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">{t("riskAnalysisDetailPage.diagrams")}</h2>
        {ra.canCreateDiagram
          ? (
              <CreateDiagramDialog
                connectionId={diagramsConnectionId}
              />
            )
          : null}
      </div>

      {diagrams.length === 0 && (
        <Card padded>
          <div className="text-center text-txt-secondary">
            {t("riskAnalysisDetailPage.emptyDiagrams")}
          </div>
        </Card>
      )}

      {diagrams.map(diagram => (
        <DiagramCard
          key={diagram.id}
          diagramRef={diagram}
          diagramsConnectionId={diagramsConnectionId}
        />
      ))}

      {hasNext && (
        <Button
          variant="tertiary"
          className="mx-auto"
          disabled={isLoadingNext}
          icon={isLoadingNext ? Spinner : IconChevronDown}
          onClick={() => loadNext(PAGE_SIZE)}
        >
          {t("riskAnalysisDetailPage.actions.showMore")}
        </Button>
      )}
    </div>
  );
}
