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

import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconTrashCan,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  TrButton,
} from "@probo/ui";
import type { ComponentProps } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedControlsCardFragment$key } from "#/__generated__/core/LinkedControlsCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { SortableTable, SortableTh } from "../SortableTable";

import { LinkedControlsDialog } from "./LinkedControlsDialog";

const linkedControlFragment = graphql`
  fragment LinkedControlsCardFragment on Control {
    id
    name
    sectionTitle
    framework {
      id
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
  readOnly?: boolean;
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
    <SortableTable refetch={props.refetch}>
      <Thead>
        <Tr>
          <SortableTh field="SECTION_TITLE">{__("Reference")}</SortableTh>
          <Th>{__("Name")}</Th>
          {!props.readOnly && <Th></Th>}
        </Tr>
      </Thead>
      <Tbody>
        {controls.length === 0 && (
          <Tr>
            <Td
              colSpan={props.readOnly ? 2 : 3}
              className="text-center text-txt-secondary"
            >
              {__("No controls linked")}
            </Td>
          </Tr>
        )}
        {controls.map(control => (
          <ControlRow
            key={control.id}
            control={control}
            onClick={onDetach}
            onAttach={onAttach}
            readOnly={props.readOnly}
          />
        ))}
        {!props.readOnly && (
          <LinkedControlsDialog
            connectionId={props.connectionId}
            disabled={props.disabled}
            linkedControls={controls}
            onLink={onAttach}
            onUnlink={onDetach}
          >
            <TrButton colspan={3}>{__("Link control")}</TrButton>
          </LinkedControlsDialog>
        )}
      </Tbody>
    </SortableTable>
  );
}

function ControlRow(props: {
  control: LinkedControlsCardFragment$key & { id: string };
  onClick: (controlId: string) => void;
  onAttach?: (controlId: string) => void;
  readOnly?: boolean;
}) {
  const control = useFragment(linkedControlFragment, props.control);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr
      to={`/organizations/${organizationId}/frameworks/${control.framework.id}/controls/${control.id}`}
    >
      <Td>
        <span className="inline-flex gap-2 items-center">
          {control.framework.name}
          {" "}
          <Badge size="md">{control.sectionTitle}</Badge>
        </span>
      </Td>
      <Td>{control.name}</Td>
      {!props.readOnly && (
        <Td noLink width={50} className="text-end">
          <Button
            variant="secondary"
            onClick={() => props.onClick(control.id)}
            icon={IconTrashCan}
          >
            {__("Unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
