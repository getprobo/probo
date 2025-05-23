import { graphql, useFragment } from "react-relay";
import type {
  RiskMeasuresTabFragment$data,
  RiskMeasuresTabFragment$key,
} from "./__generated__/RiskMeasuresTabFragment.graphql";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconPlusLarge,
  Table,
  Thead,
  Tbody,
  Td,
  Th,
  Tr,
} from "@probo/ui";
import { MeasureLinkDialog } from "../../../../components/risks/MeasureLinkDialog";
import { useOrganizationId } from "../../../../hooks/useOrganizationId";
import { useOutletContext } from "react-router";
import type { NodeOf } from "../../../../types";

const measuresFragment = graphql`
  fragment RiskMeasuresTabFragment on Risk {
    id
    measures(first: 100) @connection(key: "Risk__measures") {
      __id
      edges {
        node {
          id
          name
          description
          category
          createdAt
          state
        }
      }
    }
  }
`;

export default function RiskMeasuresTab() {
  const { risk } = useOutletContext<{
    risk: RiskMeasuresTabFragment$key & { id: string };
  }>();
  const data = useFragment(measuresFragment, risk);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const connectionId = data.measures.__id;
  const measures = data.measures?.edges?.map((edge) => edge.node) ?? [];

  if (measures.length === 0) {
    return (
      <div className="text-sm text-txt-secondary text-center flex flex-col gap-4 items-center justify-center py-10">
        {__("No measures associated with this risk.")}
        <MeasureLinkDialog
          connectionId={connectionId}
          organizationId={organizationId}
          riskId={data.id}
          trigger={
            <Button icon={IconPlusLarge} variant="quaternary">
              {__("Link measure")}
            </Button>
          }
        />
      </div>
    );
  }

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>{__("Measure")}</Th>
          <Th>{__("Frameworks")}</Th>
          <Th>{__("Risks")}</Th>
          <Th>{__("Lead")}</Th>
          <Th>{__("Status")}</Th>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {measures.map((measure) => (
          <MeasureRow key={measure.id} measure={measure} />
        ))}
      </Tbody>
    </Table>
  );
}

function MeasureRow({
  measure,
}: {
  measure: NodeOf<RiskMeasuresTabFragment$data["measures"]>;
}) {
  return (
    <Tr>
      <Td>{measure.name}</Td>
      <Td>{measure.category}</Td>
      <Td>{measure.createdAt}</Td>
      <Td>{measure.state}</Td>
      <Td></Td>
    </Tr>
  );
}
