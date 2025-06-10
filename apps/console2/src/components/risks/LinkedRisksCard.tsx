import { graphql } from "relay-runtime";
import {
  Button,
  IconTrashCan,
  RiskBadge,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  TrButton,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { LinkedRisksCardFragment$key } from "./__generated__/LinkedRisksCardFragment.graphql";
import { useFragment } from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { LinkedRisksDialog } from "./LinkedRisksDialog.tsx";

const linkedRiskFragment = graphql`
  fragment LinkedRisksCardFragment on Risk {
    id
    name
    inherentRiskScore
    residualRiskScore
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      riskId: string;
    } & Params;
    connections: string[];
  };
}) => void;

type Props<Params> = {
  // Risks linked to the element
  risks: (LinkedRisksCardFragment$key & { id: string })[];
  // Extra params to send to the mutation
  params: Params;
  // Disable (action when loading for instance)
  disabled?: boolean;
  // ID of the connection to update
  connectionId: string;
  // Mutation to attach a risk (will receive {riskId, ...params})
  onAttach: Mutation<Params>;
  // Mutation to detach a risk (will receive {riskId, ...params})
  onDetach: Mutation<Params>;
};

/**
 * Reusable component that displays a list of linked risks
 */
export function LinkedRisksCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();

  const onAttach = (riskId: string) => {
    props.onAttach({
      variables: {
        input: {
          riskId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (riskId: string) => {
    props.onDetach({
      variables: {
        input: {
          riskId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  return (
    <div className="space-y-4 relative">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Inherent Risk")}</Th>
            <Th>{__("Residual Risk")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {props.risks.length === 0 && (
            <Tr>
              <Td colSpan={4} className="text-center text-txt-secondary">
                {__("No risks linked")}
              </Td>
            </Tr>
          )}
          {props.risks.map((risk) => (
            <RiskRow key={risk.id} risk={risk} onClick={onDetach} />
          ))}
          <LinkedRisksDialog
            connectionId={props.connectionId}
            disabled={props.disabled}
            linkedRisks={props.risks}
            onLink={onAttach}
            onUnlink={onDetach}
          >
            <TrButton colspan={4}>{__("Link risk")}</TrButton>
          </LinkedRisksDialog>
        </Tbody>
      </Table>
    </div>
  );
}

function RiskRow(props: {
  risk: LinkedRisksCardFragment$key & { id: string };
  onClick: (riskId: string) => void;
}) {
  const risk = useFragment(linkedRiskFragment, props.risk);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr to={`/organizations/${organizationId}/risks/${risk.id}`}>
      <Td>{risk.name}</Td>
      <Td>
        <RiskBadge level={risk.inherentRiskScore} />
      </Td>
      <Td>
        <RiskBadge level={risk.residualRiskScore} />
      </Td>
      <Td noLink width={50} className="text-end">
        <Button
          variant="secondary"
          onClick={() => props.onClick(risk.id)}
          icon={IconTrashCan}
        >
          {__("Unlink")}
        </Button>
      </Td>
    </Tr>
  );
}
