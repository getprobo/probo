import {
  ActionDropdown,
  Avatar,
  Breadcrumb,
  Button,
  ConfirmDialog,
  Drawer,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  PageHeader,
  PropertyRow,
  RiskOverview,
  Badge,
} from "@probo/ui";
import { useNavigate, useParams } from "react-router";
import { useTranslate } from "@probo/i18n";
import { getTreatment, sprintf } from "@probo/helpers";
import { ConnectionHandler, graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import type {
  RiskDetailPageQuery,
  RiskDetailPageQuery$data,
} from "./__generated__/RiskDetailPageQuery.graphql";
import FormRiskDialog from "./FormRiskDialog";
import { usePageTitle } from "@probo/hooks";
import { useOrganizationId } from "../../../hooks/useOrganizationId";
import { useDeleteRiskMutation } from "../../../mutations/Risks";

const riskQuery = graphql`
  query RiskDetailPageQuery($riskId: ID!) {
    node(id: $riskId) {
      ... on Risk {
        name
        description
        treatment
        owner {
          id
          fullName
        }
        inherentLikelihood
        inherentImpact
        residualLikelihood
        residualImpact
        note
        createdAt
        updatedAt
        ...useRiskFormFragment
      }
    }
  }
`;

export default function RiskDetailPage() {
  const { riskId } = useParams<{ riskId: string }>();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  if (!riskId) {
    throw new Error("Cannot load risk detail page without riskId parameter");
  }

  const { __ } = useTranslate();
  const data = useLazyLoadQuery<RiskDetailPageQuery>(riskQuery, { riskId });
  const risk = data.node as Required<RiskDetailPageQuery$data["node"]>;
  const [deleteRisk] = useDeleteRiskMutation();

  usePageTitle(risk.name ?? "Risk detail");

  const onDelete = (riskId: string) => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      "RisksPage_risks"
    );
    deleteRisk({
      variables: {
        input: { riskId },
        connections: [connectionId],
      },
      onSuccess() {
        navigate(`/organizations/${organizationId}/risks`);
      },
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center mb-4">
        <Breadcrumb
          items={[
            {
              label: __("Risks"),
              to: `/organizations/${organizationId}/risks`,
            },
            {
              label: __("Risk detail"),
            },
          ]}
        />
        <div className="flex gap-2">
          <FormRiskDialog
            trigger={
              <Button icon={IconPencil} variant="secondary">
                {__("Edit")}
              </Button>
            }
            risk={{ id: riskId, ...risk }}
          />
          <ActionDropdown variant="secondary">
            <ConfirmDialog
              message={sprintf(
                __(
                  'This will permanently delete the risk "%s". This action cannot be undone.'
                ),
                risk.name
              )}
              onConfirm={() => onDelete(riskId)}
            >
              <DropdownItem variant="danger" icon={IconTrashCan}>
                {__("Delete")}
              </DropdownItem>
            </ConfirmDialog>
          </ActionDropdown>
        </div>
      </div>

      <PageHeader title={risk.name} />

      <div className="grid grid-cols-2">
        <RiskOverview type="inherent" risk={risk} />
        <RiskOverview type="residual" risk={risk} />
      </div>

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
        <PropertyRow label={__("Note")}>
          <div className="text-sm text-txt-secondary">{risk.note}</div>
        </PropertyRow>
      </Drawer>
    </div>
  );
}
