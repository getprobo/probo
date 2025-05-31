import { graphql, useFragment, useMutation } from "react-relay";
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
  IconTrashCan,
} from "@probo/ui";
import {
  detachMeasureMutation,
  MeasureLinkDialog,
} from "/components/risks/MeasureLinkDialog";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useOutletContext } from "react-router";
import type { NodeOf } from "/types";
import { useMemo } from "react";

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
  const linkedIds = useMemo(
    () => new Set(measures.map((measure) => measure.id)),
    [measures]
  );

  if (measures.length === 0) {
    return (
      <div className="text-sm text-txt-secondary text-center flex flex-col gap-4 items-center justify-center py-10">
        {__("No measures associated with this risk.")}
        <MeasureLinkDialog
          linkedIds={linkedIds}
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
    <div className="relative">
      <MeasureLinkDialog
        linkedIds={linkedIds}
        connectionId={connectionId}
        organizationId={organizationId}
        riskId={data.id}
        trigger={
          <Button
            icon={IconPlusLarge}
            variant="primary"
            className="absolute -top-18 right-0"
          >
            {__("Link measure")}
          </Button>
        }
      />
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Measure")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {measures.map((measure) => (
            <MeasureRow
              key={measure.id}
              measure={measure}
              riskId={data.id}
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
}

function MeasureRow({
  measure,
  riskId,
  connectionId,
}: {
  measure: NodeOf<RiskMeasuresTabFragment$data["measures"]>;
  riskId: string;
  connectionId: string;
}) {
  const { __ } = useTranslate();

  const [detachMeasure, isFetching] = useMutation(detachMeasureMutation);

  const onClick = () => {
    detachMeasure({
      variables: {
        input: {
          riskId: riskId,
          measureId: measure.id,
        },
        connections: [connectionId],
      },
    });
  };
  return (
    <Tr>
      <Td>{measure.name}</Td>
      <Td>
        <Button
          disabled={isFetching}
          icon={IconTrashCan}
          className="ml-auto"
          variant="quaternary"
          onClick={onClick}
        >
          {__("Unlink")}
        </Button>
      </Td>
    </Tr>
  );
}
