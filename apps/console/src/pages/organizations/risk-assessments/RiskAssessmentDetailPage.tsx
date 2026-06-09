// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatDate, formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const { toast } = useToast();
  const data = usePreloadedQuery(riskAssessmentDetailPageQuery, queryRef);
  const ra = data.node;
  const [deleteRiskAssessment] = useMutation<RiskAssessmentDetailPageDeleteMutation>(deleteMutation);

  usePageTitle(ra?.name ?? __("Risk Assessment"));

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
                  title: __("Error"),
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
                title: __("Error"),
                description: formatError(__("Failed to delete risk assessment"), error as GraphQLError),
                variant: "error",
              });
              reject(error);
            },
          });
        }),
      { message: __("This will permanently delete this risk assessment and all its scopes, nodes, processes, and threats. This action cannot be undone.") },
    );
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          { label: __("Risk Assessments"), to: listUrl },
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
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </PageHeader>

      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Details")}</h2>
        <Card className="space-y-4" padded>
          {ra.description && (
            <div className="text-sm text-txt-secondary">{ra.description}</div>
          )}
          <div className="grid grid-cols-3 gap-4">
            <div>
              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                {__("Created at")}
              </div>
              <div className="text-sm text-txt-primary">
                {formatDate(ra.createdAt)}
              </div>
            </div>
            <div>
              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                {__("Updated at")}
              </div>
              <div className="text-sm text-txt-primary">
                {formatDate(ra.updatedAt)}
              </div>
            </div>
          </div>
        </Card>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-medium">{__("Scopes")}</h2>
          <CreateScopeDialog
            connectionId={scopesConnectionId}
          />
        </div>

        {scopes.length === 0 && (
          <Card padded>
            <div className="text-center text-txt-secondary">
              {__("No scopes yet. Create a scope to start defining nodes, processes, and threats.")}
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
