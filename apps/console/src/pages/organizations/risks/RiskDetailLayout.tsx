// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { getTreatment, sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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

  const { __ } = useTranslate();
  const data = usePreloadedQuery(riskDetailLayoutQuery, props.queryRef);
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
        message: sprintf(
          __(
            "This will permanently delete the risk \"%s\". This action cannot be undone.",
          ),
          risk.name,
        ),
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
              label: __("Risks"),
              to: risksUrl,
            },
            {
              label: __("Risk detail"),
            },
          ]}
        />
        <div className="flex gap-2">
          {risk.canUpdate && (
            <FormRiskDialog
              trigger={(
                <Button icon={IconPencil} variant="secondary">
                  {__("Edit")}
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
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </div>
      </div>

      <PageHeader title={risk.name} description={risk.description} />
      <Tabs>
        <TabLink to={`${baseTabUrl}/overview`}>{__("Overview")}</TabLink>
        <TabLink to={`${baseTabUrl}/measures`}>
          {__("Measures")}
          <TabBadge>{measuresCount}</TabBadge>
        </TabLink>
        <TabLink to={`${baseTabUrl}/documents`}>
          {__("Documents")}
          <TabBadge>{documentsCount}</TabBadge>
        </TabLink>
        <TabLink to={`${baseTabUrl}/controls`}>
          {__("Controls")}
          <TabBadge>{controlsCount}</TabBadge>
        </TabLink>
        <TabLink to={`${baseTabUrl}/obligations`}>
          {__("Obligations")}
          <TabBadge>{obligationsCount}</TabBadge>
        </TabLink>
        <TabLink to={`${baseTabUrl}/scenarios`}>
          {__("Scenarios")}
          <TabBadge>{scenariosCount}</TabBadge>
        </TabLink>
      </Tabs>

      <Outlet />

      <Drawer>
        <PropertyRow label={__("Owner")}>
          <Badge variant="highlight" size="md" className="gap-2">
            <Avatar name={risk.owner?.fullName ?? ""} />
            {risk.owner?.fullName}
          </Badge>
        </PropertyRow>
        <PropertyRow label={__("Treatment")}>
          <Badge variant="highlight" size="md" className="gap-2">
            {getTreatment(__, risk.treatment)}
          </Badge>
        </PropertyRow>
        <PropertyRow label={__("Initial Risk Score")}>
          <div className="text-sm text-txt-secondary">
            {risk.inherentRiskScore}
          </div>
        </PropertyRow>
        <PropertyRow label={__("Residual Risk Score")}>
          <div className="text-sm text-txt-secondary">
            {risk.residualRiskScore}
          </div>
        </PropertyRow>
        <PropertyRow label={__("Note")}>
          <div className="text-sm text-txt-secondary">{risk.note}</div>
        </PropertyRow>
      </Drawer>
    </div>
  );
}
