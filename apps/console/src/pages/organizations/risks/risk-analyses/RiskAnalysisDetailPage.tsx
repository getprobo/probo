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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Breadcrumb,
  Card,
  DropdownItem,
  IconTrashCan,
  PageHeader,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";

import type { RiskAnalysisDetailPageDeleteMutation } from "#/__generated__/core/RiskAnalysisDetailPageDeleteMutation.graphql";
import type { RiskAnalysisDetailPageQuery } from "#/__generated__/core/RiskAnalysisDetailPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateDiagramDialog } from "./_components/CreateDiagramDialog";
import { DiagramCard } from "./_components/DiagramCard";

export const riskAnalysisDetailPageQuery = graphql`
  query RiskAnalysisDetailPageQuery($riskAnalysisId: ID!) {
    node(id: $riskAnalysisId) {
      ... on RiskAnalysis {
        id
        name
        description
        createdAt
        updatedAt
        canDelete: permission(action: "core:risk-analysis:delete")
        diagrams(first: 50)
          @connection(key: "RiskAnalysisDetailPage_diagrams", filters: []) {
          __id
          edges {
            node {
              id
              ...DiagramCardFragment
            }
          }
        }
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation RiskAnalysisDetailPageDeleteMutation(
    $input: DeleteRiskAnalysisInput!
    $connections: [ID!]!
  ) {
    deleteRiskAnalysis(input: $input) {
      deletedRiskAnalysisId @deleteEdge(connections: $connections)
    }
  }
`;

const RiskAnalysesConnectionKey = "RiskAnalysesPage_riskAnalyses";

interface RiskAnalysisDetailPageProps {
  queryRef: PreloadedQuery<RiskAnalysisDetailPageQuery>;
}

export default function RiskAnalysisDetailPage({ queryRef }: RiskAnalysisDetailPageProps) {
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const { toast } = useToast();
  const data = usePreloadedQuery<RiskAnalysisDetailPageQuery>(riskAnalysisDetailPageQuery, queryRef);
  const ra = data.node;
  const [deleteRiskAnalysis] = useMutation<RiskAnalysisDetailPageDeleteMutation>(deleteMutation);

  usePageTitle(ra?.name ?? t("riskAnalysisDetailPage.title"));

  if (!ra?.id) {
    return null;
  }

  const raId = ra.id;
  const diagrams = ra.diagrams?.edges.map(e => e.node) ?? [];
  const diagramsConnectionId = ra.diagrams?.__id ?? "";
  const listConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    RiskAnalysesConnectionKey,
  );
  const listUrl = `/organizations/${organizationId}/risk-analyses`;

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve, reject) => {
          deleteRiskAnalysis({
            variables: {
              input: { riskAnalysisId: raId },
              connections: [listConnectionId],
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: t("riskAnalysisDetailPage.messages.error"),
                  description: errors[0].message,
                  variant: "error",
                });
                reject(new Error(errors[0].message));
                return;
              }
              void navigate(listUrl);
              resolve();
            },
            onError(error) {
              toast({
                title: t("riskAnalysisDetailPage.messages.error"),
                description: formatError(t("riskAnalysisDetailPage.errors.delete"), error),
                variant: "error",
              });
              reject(error);
            },
          });
        }),
      { message: t("riskAnalysisDetailPage.deleteConfirmation") },
    );
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          { label: t("riskAnalysisDetailPage.breadcrumb.assessments"), to: listUrl },
          { label: ra.name ?? "" },
        ]}
      />

      <PageHeader
        title={ra.name}
      >
        {ra.canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleDelete}
            >
              {t("riskAnalysisDetailPage.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </PageHeader>

      <div className="space-y-4">
        <h2 className="text-base font-medium">{t("riskAnalysisDetailPage.details")}</h2>
        <Card className="space-y-4" padded>
          {ra.description && (
            <div className="text-sm text-txt-secondary">{ra.description}</div>
          )}
          <div className="grid grid-cols-3 gap-4">
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
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-medium">{t("riskAnalysisDetailPage.diagrams")}</h2>
          <CreateDiagramDialog
            connectionId={diagramsConnectionId}
          />
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
      </div>
    </div>
  );
}
