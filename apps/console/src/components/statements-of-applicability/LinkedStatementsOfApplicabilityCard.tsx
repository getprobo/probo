// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import type { LinkedStatementsOfApplicabilityCardFragment$key } from "#/__generated__/core/LinkedStatementsOfApplicabilityCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedStatementsOfApplicabilityDialog } from "./LinkedStatementsOfApplicabilityDialog";

const linkedStatementOfApplicabilityFragment = graphql`
    fragment LinkedStatementsOfApplicabilityCardFragment on ApplicabilityStatement {
        id
        statementOfApplicability {
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
      statementOfApplicabilityId: string;
      applicability: boolean;
      justification: string | null;
    } & Params;
    connections: string[];
  };
}) => void;

type DetachMutation = (p: {
  variables: {
    input: {
      statementOfApplicabilityId: string;
      controlId: string;
    };
    connections: string[];
  };
}) => void;

type Props<Params> = {
  statementsOfApplicability: readonly (LinkedStatementsOfApplicabilityCardFragment$key & {
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

export function LinkedStatementsOfApplicabilityCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();

  const [limit, setLimit] = useState<number | null>(
    props.variant === "card" ? 4 : null,
  );

  const [linkedInfo, setLinkedInfo] = useState<
    { statementOfApplicabilityId: string; controlId: string }[]
  >([]);

  const statementsOfApplicability = useMemo(() => {
    return limit
      ? props.statementsOfApplicability.slice(0, limit)
      : props.statementsOfApplicability;
  }, [props.statementsOfApplicability, limit]);

  const showMoreButton
    = limit !== null && props.statementsOfApplicability.length > limit;
  const variant = props.variant ?? "table";

  const linkedData = linkedInfo;

  const onAttach = (
    statementOfApplicabilityId: string,
    applicability: boolean,
    justification: string | null,
  ) => {
    props.onAttach({
      variables: {
        input: {
          statementOfApplicabilityId,
          applicability,
          justification,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (statementOfApplicabilityId: string, controlId: string) => {
    props.onDetach({
      variables: {
        input: {
          statementOfApplicabilityId,
          controlId,
        },
        connections: [props.connectionId],
      },
    });
  };

  const Wrapper = variant === "card" ? Card : "div";

  return (
    <Wrapper padded className="space-y-[10px]">
      {props.statementsOfApplicability.map((soa, idx) => (
        <LinkedInfoExtractor
          key={idx}
          fragment={soa}
          onExtracted={(info) => {
            setLinkedInfo((prev) => {
              const exists = prev.some(
                p =>
                  p.statementOfApplicabilityId
                  === info.statementOfApplicabilityId
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
            {__("Statements of Applicability")}
          </div>
          {!props.readOnly && (
            <LinkedStatementsOfApplicabilityDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedStatementsOfApplicability={linkedData}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {__("Link statement of applicability")}
              </Button>
            </LinkedStatementsOfApplicabilityDialog>
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
          {statementsOfApplicability.length === 0 && (
            <Tr>
              <Td
                colSpan={props.readOnly ? 3 : 4}
                className="text-center text-txt-secondary"
              >
                {__("No statements of applicability linked")}
              </Td>
            </Tr>
          )}
          {statementsOfApplicability.map(soa => (
            <StatementOfApplicabilityRow
              key={soa.id}
              statementOfApplicability={soa}
              onClick={onDetach}
              readOnly={props.readOnly}
            />
          ))}
          {variant === "table" && !props.readOnly && (
            <LinkedStatementsOfApplicabilityDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedStatementsOfApplicability={linkedData}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={4} icon={IconPlusLarge}>
                {__("Link statement of applicability")}
              </TrButton>
            </LinkedStatementsOfApplicabilityDialog>
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
            props.statementsOfApplicability.length - limit,
          )}
        </Button>
      )}
    </Wrapper>
  );
}

function LinkedInfoExtractor(props: {
  fragment: LinkedStatementsOfApplicabilityCardFragment$key;
  onExtracted: (info: {
    statementOfApplicabilityId: string;
    controlId: string;
  }) => void;
}) {
  const { onExtracted, fragment } = props;

  const data = useFragment(
    linkedStatementOfApplicabilityFragment,
    fragment,
  );

  useEffect(() => {
    onExtracted({
      statementOfApplicabilityId: data.statementOfApplicability.id,
      controlId: data.control.id,
    });
  }, [data.statementOfApplicability.id, data.control.id, onExtracted]);

  return null;
}

function StatementOfApplicabilityRow(props: {
  statementOfApplicability: LinkedStatementsOfApplicabilityCardFragment$key & {
    id: string;
  };
  onClick: (statementOfApplicabilityId: string, controlId: string) => void;
  readOnly?: boolean;
}) {
  const soa = useFragment(
    linkedStatementOfApplicabilityFragment,
    props.statementOfApplicability,
  );
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr
      to={`/organizations/${organizationId}/statements-of-applicability/${soa.statementOfApplicability.id}`}
    >
      <Td>{soa.statementOfApplicability.name}</Td>
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
                soa.statementOfApplicability.id,
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
