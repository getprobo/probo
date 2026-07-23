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

import type { RiskAssessmentDetailPageDeleteMutation } from "#/__generated__/core/RiskAssessmentDetailPageDeleteMutation.graphql";
import type { RiskAssessmentDetailPageQuery } from "#/__generated__/core/RiskAssessmentDetailPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateScopeDialog } from "./_components/CreateScopeDialog";
import { ScopeCard } from "./_components/ScopeCard";

export const riskAssessmentDetailPageQuery = graphql`
  query RiskAssessmentDetailPageQuery($riskAssessmentId: ID!) {
    node(id: $riskAssessmentId) {
      ... on RiskAssessment {
        id
        name
        description
        createdAt
        updatedAt
        canDelete: permission(action: "core:risk-assessment:delete")
        scopes(first: 50)
          @connection(key: "RiskAssessmentDetailPage_scopes", filters: []) {
          __id
          edges {
            node {
              id
              ...ScopeCardFragment
            }
          }
        }
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation RiskAssessmentDetailPageDeleteMutation(
    $input: DeleteRiskAssessmentInput!
    $connections: [ID!]!
  ) {
    deleteRiskAssessment(input: $input) {
      deletedRiskAssessmentId @deleteEdge(connections: $connections)
    }
  }
`;

const RiskAssessmentsConnectionKey = "RiskAssessmentsPage_riskAssessments";

interface RiskAssessmentDetailPageProps {
  queryRef: PreloadedQuery<RiskAssessmentDetailPageQuery>;
}

export default function RiskAssessmentDetailPage({ queryRef }: RiskAssessmentDetailPageProps) {
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const { toast } = useToast();
  const data = usePreloadedQuery<RiskAssessmentDetailPageQuery>(riskAssessmentDetailPageQuery, queryRef);
  const ra = data.node;
  const [deleteRiskAssessment] = useMutation<RiskAssessmentDetailPageDeleteMutation>(deleteMutation);

  usePageTitle(ra?.name ?? t("riskAssessmentDetailPage.title"));

  if (!ra?.id) {
    return null;
  }

  const raId = ra.id;
  const scopes = ra.scopes?.edges.map(e => e.node) ?? [];
  const scopesConnectionId = ra.scopes?.__id ?? "";
  const listConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    RiskAssessmentsConnectionKey,
  );
  const listUrl = `/organizations/${organizationId}/risk-assessments`;

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve, reject) => {
          deleteRiskAssessment({
            variables: {
              input: { riskAssessmentId: raId },
              connections: [listConnectionId],
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: t("riskAssessmentDetailPage.messages.error"),
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
                title: t("riskAssessmentDetailPage.messages.error"),
                description: formatError(t("riskAssessmentDetailPage.errors.delete"), error),
                variant: "error",
              });
              reject(error);
            },
          });
        }),
      { message: t("riskAssessmentDetailPage.deleteConfirmation") },
    );
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          { label: t("riskAssessmentDetailPage.breadcrumb.assessments"), to: listUrl },
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
              {t("riskAssessmentDetailPage.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </PageHeader>

      <div className="space-y-4">
        <h2 className="text-base font-medium">{t("riskAssessmentDetailPage.details")}</h2>
        <Card className="space-y-4" padded>
          {ra.description && (
            <div className="text-sm text-txt-secondary">{ra.description}</div>
          )}
          <div className="grid grid-cols-3 gap-4">
            <div>
              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                {t("riskAssessmentDetailPage.fields.createdAt")}
              </div>
              <div className="text-sm text-txt-primary">
                {dateFormat(i18n.language, ra.createdAt)}
              </div>
            </div>
            <div>
              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                {t("riskAssessmentDetailPage.fields.updatedAt")}
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
          <h2 className="text-base font-medium">{t("riskAssessmentDetailPage.scopes")}</h2>
          <CreateScopeDialog
            connectionId={scopesConnectionId}
          />
        </div>

        {scopes.length === 0 && (
          <Card padded>
            <div className="text-center text-txt-secondary">
              {t("riskAssessmentDetailPage.emptyScopes")}
            </div>
          </Card>
        )}

        {scopes.map(scope => (
          <ScopeCard
            key={scope.id}
            scopeRef={scope}
            scopesConnectionId={scopesConnectionId}
          />
        ))}
      </div>
    </div>
  );
}
