import { graphql } from "relay-runtime";
import {
  Button,
  Tr,
  Td,
  Thead,
  Tbody,
  Th,
  IconTrashCan,
  Badge,
  TrButton,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { LinkedControlsCardFragment$key } from "./__generated__/LinkedControlsCardFragment.graphql";
import { useFragment } from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { LinkedControlsDialog } from "./LinkedControlsDialog";
import { SortableTable, SortableTh } from "../SortableTable";
import type { ComponentProps } from "react";

const linkedControlFragment = graphql`
  fragment LinkedControlsCardFragment on Control {
    id
    name
    sectionTitle
    framework {
      name
    }
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      controlId: string;
    } & Params;
    connections: string[];
  };
}) => void;

type Props<Params> = {
  // Controls linked to the element
  controls: (LinkedControlsCardFragment$key & { id: string })[];
  // Extra params to send to the mutation
  params: Params;
  // Disable (action when loading for instance)
  disabled?: boolean;
  // ID of the connection to update
  connectionId: string;
  // Mutation to detach a control (will receive {controlId, ...params})
  onDetach: Mutation<Params>;
  // Mutation to attach a control (will receive {controlId, ...params})
  onAttach?: Mutation<Params>;
  // Allow sorting in the table
  refetch: ComponentProps<typeof SortableTable>["refetch"];
};

/**
 * Reusable component that displays a list of linked controls
 */
export function LinkedControlsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const controls = props.controls;

  const onDetach = (controlId: string) => {
    props.onDetach({
      variables: {
        input: {
          controlId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onAttach = (controlId: string) => {
    if (!props.onAttach) {
      return;
    }
    props.onAttach({
      variables: {
        input: {
          controlId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  return (
    <SortableTable refetch={props.refetch as any}>
      <Thead>
        <Tr>
          <SortableTh field="SECTION_TITLE">{__("Reference")}</SortableTh>
          <Th>{__("Name")}</Th>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {controls.length === 0 && (
          <Tr>
            <Td colSpan={4} className="text-center text-txt-secondary">
              {__("No controls linked")}
            </Td>
          </Tr>
        )}
        {controls.map((control) => (
          <ControlRow
            key={control.id}
            control={control}
            onClick={onDetach}
            onAttach={onAttach}
          />
        ))}
        <LinkedControlsDialog
          connectionId={props.connectionId}
          disabled={props.disabled}
          linkedControls={controls}
          onLink={onAttach}
          onUnlink={onDetach}
        >
          <TrButton colspan={3}>{__("Link control")}</TrButton>
        </LinkedControlsDialog>
      </Tbody>
    </SortableTable>
  );
}

function ControlRow(props: {
  control: LinkedControlsCardFragment$key & { id: string };
  onClick: (controlId: string) => void;
  onAttach?: (controlId: string) => void;
}) {
  const control = useFragment(linkedControlFragment, props.control);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr to={`/organizations/${organizationId}/controls/${control.id}`}>
      <Td>
        <span className="inline-flex gap-2 items-center">
          {control.framework.name}{" "}
          <Badge size="md">{control.sectionTitle}</Badge>
        </span>
      </Td>
      <Td>{control.name}</Td>
      <Td noLink width={50} className="text-end">
        <Button
          variant="secondary"
          onClick={() => props.onClick(control.id)}
          icon={IconTrashCan}
        >
          {__("Unlink")}
        </Button>
      </Td>
    </Tr>
  );
}
