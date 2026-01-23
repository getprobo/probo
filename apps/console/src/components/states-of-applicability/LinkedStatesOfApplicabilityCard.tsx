import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  IconChevronDown,
  IconPlusLarge,
  IconTrashCan,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  TrButton,
} from "@probo/ui";
import { clsx } from "clsx";
import { useEffect, useMemo, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedStatesOfApplicabilityCardFragment$key } from "#/__generated__/core/LinkedStatesOfApplicabilityCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedStatesOfApplicabilityDialog } from "./LinkedStatesOfApplicabilityDialog";

const linkedStateOfApplicabilityFragment = graphql`
    fragment LinkedStatesOfApplicabilityCardFragment on ApplicabilityStatement {
        id
        stateOfApplicability {
            id
            name
        }
        control {
            id
        }
        applicability
        justification
    }
`;

type AttachMutation<Params> = (p: {
  variables: {
    input: {
      stateOfApplicabilityId: string;
      applicability: boolean;
      justification: string | null;
    } & Params;
    connections: string[];
  };
}) => void;

type DetachMutation = (p: {
  variables: {
    input: {
      stateOfApplicabilityId: string;
      controlId: string;
    };
    connections: string[];
  };
}) => void;

type Props<Params> = {
  statesOfApplicability: readonly (LinkedStatesOfApplicabilityCardFragment$key & {
    id: string;
  })[];
  params: Params;
  disabled?: boolean;
  connectionId: string;
  onAttach: AttachMutation<Params>;
  onDetach: DetachMutation;
  variant?: "card" | "table";
  readOnly?: boolean;
};

export function LinkedStatesOfApplicabilityCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();

  const [limit, setLimit] = useState<number | null>(
    props.variant === "card" ? 4 : null,
  );

  const [linkedInfo, setLinkedInfo] = useState<
    { stateOfApplicabilityId: string; controlId: string }[]
  >([]);

  const statesOfApplicability = useMemo(() => {
    return limit
      ? props.statesOfApplicability.slice(0, limit)
      : props.statesOfApplicability;
  }, [props.statesOfApplicability, limit]);

  const showMoreButton
    = limit !== null && props.statesOfApplicability.length > limit;
  const variant = props.variant ?? "table";

  const linkedData = linkedInfo;

  const onAttach = (
    stateOfApplicabilityId: string,
    applicability: boolean,
    justification: string | null,
  ) => {
    props.onAttach({
      variables: {
        input: {
          stateOfApplicabilityId,
          applicability,
          justification,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (stateOfApplicabilityId: string, controlId: string) => {
    props.onDetach({
      variables: {
        input: {
          stateOfApplicabilityId,
          controlId,
        },
        connections: [props.connectionId],
      },
    });
  };

  const Wrapper = variant === "card" ? Card : "div";

  return (
    <Wrapper padded className="space-y-[10px]">
      {props.statesOfApplicability.map((soa, idx) => (
        <LinkedInfoExtractor
          key={idx}
          fragment={soa}
          onExtracted={(info) => {
            setLinkedInfo((prev) => {
              const exists = prev.some(
                p =>
                  p.stateOfApplicabilityId
                  === info.stateOfApplicabilityId
                  && p.controlId === info.controlId,
              );
              return exists ? prev : [...prev, info];
            });
          }}
        />
      ))}
      {variant === "card" && (
        <div className="flex justify-between">
          <div className="text-lg font-semibold">
            {__("States of Applicability")}
          </div>
          {!props.readOnly && (
            <LinkedStatesOfApplicabilityDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedStatesOfApplicability={linkedData}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {__("Link state of applicability")}
              </Button>
            </LinkedStatesOfApplicabilityDialog>
          )}
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Applicability")}</Th>
            <Th>{__("Justification")}</Th>
            {!props.readOnly && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {statesOfApplicability.length === 0 && (
            <Tr>
              <Td
                colSpan={props.readOnly ? 3 : 4}
                className="text-center text-txt-secondary"
              >
                {__("No states of applicability linked")}
              </Td>
            </Tr>
          )}
          {statesOfApplicability.map(soa => (
            <StateOfApplicabilityRow
              key={soa.id}
              stateOfApplicability={soa}
              onClick={onDetach}
              readOnly={props.readOnly}
            />
          ))}
          {variant === "table" && !props.readOnly && (
            <LinkedStatesOfApplicabilityDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedStatesOfApplicability={linkedData}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={4} icon={IconPlusLarge}>
                {__("Link state of applicability")}
              </TrButton>
            </LinkedStatesOfApplicabilityDialog>
          )}
        </Tbody>
      </Table>
      {showMoreButton && (
        <Button
          variant="tertiary"
          icon={IconChevronDown}
          onClick={() => setLimit(null)}
        >
          {sprintf(
            __("Show %d more"),
            props.statesOfApplicability.length - limit,
          )}
        </Button>
      )}
    </Wrapper>
  );
}

function LinkedInfoExtractor(props: {
  fragment: LinkedStatesOfApplicabilityCardFragment$key;
  onExtracted: (info: {
    stateOfApplicabilityId: string;
    controlId: string;
  }) => void;
}) {
  const { onExtracted, fragment } = props;

  const data = useFragment(
    linkedStateOfApplicabilityFragment,
    fragment,
  );

  useEffect(() => {
    onExtracted({
      stateOfApplicabilityId: data.stateOfApplicability.id,
      controlId: data.control.id,
    });
  }, [data.stateOfApplicability.id, data.control.id, onExtracted]);

  return null;
}

function StateOfApplicabilityRow(props: {
  stateOfApplicability: LinkedStatesOfApplicabilityCardFragment$key & {
    id: string;
  };
  onClick: (stateOfApplicabilityId: string, controlId: string) => void;
  readOnly?: boolean;
}) {
  const soa = useFragment(
    linkedStateOfApplicabilityFragment,
    props.stateOfApplicability,
  );
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr
      to={`/organizations/${organizationId}/states-of-applicability/${soa.stateOfApplicability.id}`}
    >
      <Td>{soa.stateOfApplicability.name}</Td>
      <Td>
        <Badge variant={soa.applicability ? "success" : "danger"}>
          {soa.applicability
            ? __("Applicable")
            : __("Not Applicable")}
        </Badge>
      </Td>
      <Td>{soa.justification || "-"}</Td>
      {!props.readOnly && (
        <Td noLink width={50} className="text-end">
          <Button
            variant="secondary"
            onClick={() =>
              props.onClick(
                soa.stateOfApplicability.id,
                soa.control.id,
              )}
            icon={IconTrashCan}
          >
            {__("Unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
