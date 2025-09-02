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

  IconTrashCan,
  TrButton,
} from "@probo/ui";
import { MeasureBadge } from "@probo/ui/src/Molecules/Badge/MeasureBadge";
import { useTranslate } from "@probo/i18n";
import type { LinkedMeasuresCardFragment$key } from "./__generated__/LinkedMeasuresCardFragment.graphql";
import { useFragment } from "react-relay";
import { useMemo, useState } from "react";
import { sprintf } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { LinkedMeasureDialog } from "./LinkedMeasuresDialog.tsx";
import { useParams } from "react-router";
import clsx from "clsx";

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
  measures: (LinkedMeasuresCardFragment$key & { id: string })[];
  params: Params;
  disabled?: boolean;
  hideActions?: boolean;
  connectionId: string;
  onAttach: Mutation<Params>;
  onDetach: Mutation<Params>;
  variant?: "card" | "table";
};

/**
 * Reusable component that displays a list of linked measures
 */
export function LinkedMeasuresCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const [limit, setLimit] = useState<number | null>(
    props.variant === "card" ? 4 : null
  );
  const measures = useMemo(() => {
    return limit ? props.measures.slice(0, limit) : props.measures;
  }, [props.measures, limit]);
  const showMoreButton = limit !== null && props.measures.length > limit;
  const variant = props.variant ?? "table";

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

  const Wrapper = variant === "card" ? Card : "div";

  return (
    <Wrapper padded className="space-y-[10px]">
      {variant === "card" && (
        <div className="flex justify-between">
          <div className="text-lg font-semibold">{__("Measures")}</div>
          {!props.hideActions && (
            <LinkedMeasureDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedMeasures={props.measures}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {__("Link measure")}
              </Button>
            </LinkedMeasureDialog>
          )}
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("State")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {measures.length === 0 && (
            <Tr>
              <Td colSpan={3} className="text-center text-txt-secondary">
                {__("No measures linked")}
              </Td>
            </Tr>
          )}
          {measures.map((measure) => (
            <MeasureRow
              key={measure.id}
              measure={measure}
              onClick={onDetach}
              hideActions={props.hideActions}
              snapshotId={snapshotId}
            />
          ))}
          {variant === "table" && !props.hideActions && (
            <LinkedMeasureDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedMeasures={props.measures}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={3} icon={IconPlusLarge}>
                {__("Link measure")}
              </TrButton>
            </LinkedMeasureDialog>
          )}
        </Tbody>
      </Table>
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
    </Wrapper>
  );
}

function MeasureRow(props: {
  measure: LinkedMeasuresCardFragment$key & { id: string };
  onClick: (measureId: string) => void;
  hideActions?: boolean;
  snapshotId?: string;
}) {
  const measure = useFragment(linkedMeasureFragment, props.measure);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const measureUrl = props.snapshotId
    ? `/organizations/${organizationId}/snapshots/${props.snapshotId}/risks/measures/${measure.id}/evidences`
    : `/organizations/${organizationId}/measures/${measure.id}`;

  return (
    <Tr to={measureUrl}>
      <Td>{measure.name}</Td>
      <Td>
        <MeasureBadge state={measure.state} />
      </Td>
      <Td noLink width={50} className="text-end">
        {!props.hideActions && (
          <Button
            variant="secondary"
            onClick={() => props.onClick(measure.id)}
            icon={IconTrashCan}
          >
            {__("Unlink")}
          </Button>
        )}
      </Td>
    </Tr>
  );
}
