import { graphql } from "relay-runtime";
import {
  Card,
  IconPlusLarge,
  Button,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
  IconChevronDown,
  MeasureBadge,
  IconTrashCan,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { LinkedMeasuresCardFragment$key } from "./__generated__/LinkedMeasuresCardFragment.graphql";
import { useFragment } from "react-relay";
import { useMemo, useState } from "react";
import { sprintf } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { MeasureLinkDialog } from "./MeasureLinkDialog";

const linkedMeasureFragment = graphql`
  fragment LinkedMeasuresCardFragment on Measure {
    id
    name
    state
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      measureId: string;
    } & Params;
    connections: string[];
  };
}) => void;

type Props<Params> = {
  // Measures linked to the element
  measures: (LinkedMeasuresCardFragment$key & { id: string })[];
  // Extra params to send to the mutation
  params: Params;
  // Disable (action when loading for instance)
  disabled?: boolean;
  // ID of the connection to update
  connectionId: string;
  // Mutation to attach a measure (will receive {measureId, ...params})
  onAttach: Mutation<Params>;
  // Mutation to detach a measure (will receive {measureId, ...params})
  onDetach: Mutation<Params>;
};

/**
 * Reusable component that displays a list of linked measures
 */
export function LinkedMeasuresCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(4);
  const measures = useMemo(() => {
    return limit ? props.measures.slice(0, limit) : props.measures;
  }, [props.measures, limit]);
  const showMoreButton = limit !== null && props.measures.length > limit;

  const onAttach = (measureId: string) => {
    props.onAttach({
      variables: {
        input: {
          measureId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (measureId: string) => {
    props.onDetach({
      variables: {
        input: {
          measureId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  return (
    <Card padded className="space-y-[10px]">
      <div className="flex justify-between">
        <div className="text-lg font-semibold">{__("Measures")}</div>
        <MeasureLinkDialog
          connectionId={props.connectionId}
          disabled={props.disabled}
          linkedMeasures={props.measures}
          onLink={onAttach}
          onUnlink={onDetach}
        >
          <Button variant="tertiary" icon={IconPlusLarge}>
            {__("Link measure")}
          </Button>
        </MeasureLinkDialog>
      </div>
      {measures.length > 0 ? (
        <Table className="bg-invert">
          <Thead>
            <Tr>
              <Th>{__("Name")}</Th>
              <Th>{__("State")}</Th>
              <Th></Th>
            </Tr>
          </Thead>
          <Tbody>
            {measures.map((measure) => (
              <MeasureRow
                key={measure.id}
                measure={measure}
                onClick={onDetach}
              />
            ))}
          </Tbody>
        </Table>
      ) : (
        <div className="text-center text-sm text-txt-secondary">
          {__("No measures linked")}
        </div>
      )}
      {showMoreButton && (
        <Button
          variant="tertiary"
          onClick={() => setLimit(null)}
          className="mt-3 mx-auto"
          icon={IconChevronDown}
        >
          {sprintf(__("Show %s more"), props.measures.length - limit)}
        </Button>
      )}
    </Card>
  );
}

function MeasureRow(props: {
  measure: LinkedMeasuresCardFragment$key & { id: string };
  onClick: (measureId: string) => void;
}) {
  const measure = useFragment(linkedMeasureFragment, props.measure);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr to={`/organizations/${organizationId}/measures/${measure.id}`}>
      <Td>{measure.name}</Td>
      <Td>
        <MeasureBadge state={measure.state} />
      </Td>
      <Td noLink width={50} className="text-end">
        <Button
          variant="secondary"
          onClick={() => props.onClick(measure.id)}
          icon={IconTrashCan}
        >
          {__("Unlink")}
        </Button>
      </Td>
    </Tr>
  );
}
