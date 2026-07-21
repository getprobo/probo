// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import {
  ActionDropdown,
  Avatar,
  Badge,
  Breadcrumb,
  Button,
  Drawer,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  PageHeader,
  PropertyRow,
  TabBadge,
  TabLink,
  Tabs,
  useConfirm,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { Outlet, useNavigate, useParams } from "react-router";
import { ConnectionHandler } from "relay-runtime";

import type { RiskDetailLayoutDeleteMutation } from "#/__generated__/core/RiskDetailLayoutDeleteMutation.graphql";
import type { RiskDetailLayoutQuery } from "#/__generated__/core/RiskDetailLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { RisksConnectionKey } from "#/pages/organizations/risks/RisksPage";

import { FormRiskDialog } from "./_components/FormRiskDialog";

export const riskDetailLayoutQuery = graphql`
  query RiskDetailLayoutQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        name
        description
        treatment
        owner {
          fullName
        }
        note
        inherentRiskScore
        residualRiskScore
        measuresInfo: measures(first: 0) {
          totalCount
        }
        documentsInfo: documents(first: 0) {
          totalCount
        }
        controlsInfo: controls(first: 0) {
          totalCount
        }
        obligationsInfo: obligations(first: 0) {
          totalCount
        }
        scenariosInfo: scenarios(first: 0) {
          totalCount
        }
        canUpdate: permission(action: "core:risk:update")
        canDelete: permission(action: "core:risk:delete")
        ...FormRiskDialog_risk
      }
    }
  }
`;

const deleteRiskMutation = graphql`
  mutation RiskDetailLayoutDeleteMutation(
    $input: DeleteRiskInput!
    $connections: [ID!]!
  ) {
    deleteRisk(input: $input) {
      deletedRiskId @deleteEdge(connections: $connections)
    }
  }
`;

interface RiskDetailLayoutProps {
  queryRef: PreloadedQuery<RiskDetailLayoutQuery>;
}

export default function RiskDetailLayout(props: RiskDetailLayoutProps) {
  const { riskId } = useParams<{
    riskId: string;
  }>();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  if (!riskId) {
    throw new Error("Cannot load risk detail page without riskId parameter");
  }

  const { t } = useTranslation();
  const data = usePreloadedQuery<RiskDetailLayoutQuery>(riskDetailLayoutQuery, props.queryRef);
  if (data.node?.__typename !== "Risk") {
    throw new Error("Risk not found");
  }
  const risk = data.node;

  const [deleteRisk] = useMutation<RiskDetailLayoutDeleteMutation>(deleteRiskMutation);

  usePageTitle(risk.name);
  const confirm = useConfirm();

  const onDelete = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      RisksConnectionKey,
    );
    confirm(
      () =>
        new Promise<void>((resolve, reject) => {
          void deleteRisk({
            variables: {
              input: { riskId },
              connections: [connectionId],
            },
            onCompleted() {
              void navigate(`/organizations/${organizationId}/risks`);
              resolve();
            },
            onError(error) {
              reject(error);
            },
          });
        }),
      {
        message: t("riskDetailLayout.deleteConfirmation", { name: risk.name }),
      },
    );
  };

  const documentsCount = risk.documentsInfo?.totalCount ?? 0;
  const measuresCount = risk.measuresInfo?.totalCount ?? 0;
  const controlsCount = risk.controlsInfo?.totalCount ?? 0;
  const obligationsCount = risk.obligationsInfo?.totalCount ?? 0;
  const scenariosCount = risk.scenariosInfo?.totalCount ?? 0;

  const risksUrl = `/organizations/${organizationId}/risks`;
  const baseTabUrl = `/organizations/${organizationId}/risks/${riskId}`;

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center mb-4">
        <Breadcrumb
          items={[
            {
              label: t("riskDetailLayout.breadcrumb.risks"),
              to: risksUrl,
            },
            {
              label: t("riskDetailLayout.breadcrumb.detail"),
            },
          ]}
        />
        <div className="flex gap-2">
          {risk.canUpdate && (
            <FormRiskDialog
              trigger={(
                <Button icon={IconPencil} variant="secondary">
                  {t("riskDetailLayout.actions.edit")}
                </Button>
              )}
              risk={risk}
            />
          )}
          {risk.canDelete && (
            <ActionDropdown variant="secondary">
              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                onClick={onDelete}
              >
                {t("riskDetailLayout.actions.delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </div>
      </div>

      <PageHeader title={risk.name} description={risk.description} />
      <Tabs>
        <TabLink to={`${baseTabUrl}/overview`}>{t("riskDetailLayout.tabs.overview")}</TabLink>
        <TabLink to={`${baseTabUrl}/measures`}>
          {t("riskDetailLayout.tabs.measures")}
          <TabBadge>{measuresCount}</TabBadge>
        </TabLink>
        <TabLink to={`${baseTabUrl}/documents`}>
          {t("riskDetailLayout.tabs.documents")}
          <TabBadge>{documentsCount}</TabBadge>
        </TabLink>
        <TabLink to={`${baseTabUrl}/controls`}>
          {t("riskDetailLayout.tabs.controls")}
          <TabBadge>{controlsCount}</TabBadge>
        </TabLink>
        <TabLink to={`${baseTabUrl}/obligations`}>
          {t("riskDetailLayout.tabs.obligations")}
          <TabBadge>{obligationsCount}</TabBadge>
        </TabLink>
        <TabLink to={`${baseTabUrl}/scenarios`}>
          {t("riskDetailLayout.tabs.scenarios")}
          <TabBadge>{scenariosCount}</TabBadge>
        </TabLink>
      </Tabs>

      <Outlet />

      <Drawer>
        <PropertyRow label={t("riskDetailLayout.fields.owner")}>
          <Badge variant="highlight" size="md" className="gap-2">
            <Avatar name={risk.owner?.fullName ?? ""} />
            {risk.owner?.fullName}
          </Badge>
        </PropertyRow>
        <PropertyRow label={t("riskDetailLayout.fields.treatment")}>
          <Badge variant="highlight" size="md" className="gap-2">
            {t(`riskDetailLayout.treatments.${(risk.treatment ?? "UNKNOWN").toLowerCase()}`)}
          </Badge>
        </PropertyRow>
        <PropertyRow label={t("riskDetailLayout.fields.initialRiskScore")}>
          <div className="text-sm text-txt-secondary">
            {risk.inherentRiskScore}
          </div>
        </PropertyRow>
        <PropertyRow label={t("riskDetailLayout.fields.residualRiskScore")}>
          <div className="text-sm text-txt-secondary">
            {risk.residualRiskScore}
          </div>
        </PropertyRow>
        <PropertyRow label={t("riskDetailLayout.fields.note")}>
          <div className="text-sm text-txt-secondary">{risk.note}</div>
        </PropertyRow>
      </Drawer>
    </div>
  );
}
