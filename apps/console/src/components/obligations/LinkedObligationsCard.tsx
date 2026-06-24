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

import {
  getObligationStatusLabel,
  getObligationStatusVariant,
  sprintf,
} from "@probo/helpers";
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
import { useMemo, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedObligationsCardFragment$key } from "#/__generated__/core/LinkedObligationsCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedObligationDialog } from "./LinkedObligationsDialog";

const linkedObligationFragment = graphql`
  fragment LinkedObligationsCardFragment on Obligation {
    id
    area
    source
    status
    owner {
      fullName
    }
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      obligationId: string;
    } & Params;
    connections: string[];
  };
}) => void;

type Props<Params> = {
  obligations: (LinkedObligationsCardFragment$key & { id: string })[];
  connectionId: string;
  disabled?: boolean;
  variant?: "card" | "table";
  readOnly?: boolean;

  params: Params;

  onAttach: Mutation<Params>;
  onDetach: Mutation<Params>;
};

export function LinkedObligationsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(
    props.variant === "card" ? 4 : null,
  );

  const onAttach = (obligationId: string) => {
    props.onAttach({
      variables: {
        input: {
          obligationId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (obligationId: string) => {
    props.onDetach({
      variables: {
        input: {
          obligationId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const obligations = useMemo(() => {
    return limit ? props.obligations.slice(0, limit) : props.obligations;
  }, [props.obligations, limit]);

  const showMoreButton = limit !== null && props.obligations.length > limit;
  const variant = props.variant ?? "table";

  const Wrapper = variant === "card" ? Card : "div";

  return (
    <Wrapper padded className="space-y-[10px]">
      {variant === "card" && (
        <div className="flex justify-between">
          <div className="text-lg font-semibold">{__("Obligations")}</div>
          {!props.readOnly && (
            <LinkedObligationDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedObligations={props.obligations}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {__("Link obligation")}
              </Button>
            </LinkedObligationDialog>
          )}
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{__("Area")}</Th>
            <Th>{__("Source")}</Th>
            <Th>{__("Status")}</Th>
            <Th>{__("Owner")}</Th>
            {!props.readOnly && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {obligations.length === 0 && (
            <Tr>
              <Td
                colSpan={props.readOnly ? 4 : 5}
                className="text-center text-txt-secondary"
              >
                {__("No obligations linked")}
              </Td>
            </Tr>
          )}
          {obligations.map(obligation => (
            <ObligationRow
              key={obligation.id}
              obligation={obligation}
              onClick={onDetach}
              readOnly={props.readOnly}
            />
          ))}
          {variant === "table" && !props.readOnly && (
            <LinkedObligationDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedObligations={props.obligations}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={5} icon={IconPlusLarge}>
                {__("Link obligation")}
              </TrButton>
            </LinkedObligationDialog>
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
          {sprintf(__("Show %s more"), props.obligations.length - limit)}
        </Button>
      )}
    </Wrapper>
  );
}

function ObligationRow(props: {
  obligation: LinkedObligationsCardFragment$key & { id: string };
  onClick: (obligationId: string) => void;
  readOnly?: boolean;
}) {
  const { __ } = useTranslate();
  const obligation = useFragment(linkedObligationFragment, props.obligation);
  const organizationId = useOrganizationId();

  const onDetach = () => {
    props.onClick(obligation.id);
  };

  const detailsUrl = `/organizations/${organizationId}/obligations/${obligation.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>{obligation.area || __("No area specified")}</Td>
      <Td>{obligation.source || __("No source specified")}</Td>
      <Td>
        <Badge variant={getObligationStatusVariant(obligation.status)}>
          {getObligationStatusLabel(obligation.status)}
        </Badge>
      </Td>
      <Td>{obligation.owner?.fullName || __("Unassigned")}</Td>
      {!props.readOnly && (
        <Td noLink width={50} className="text-end">
          <Button variant="secondary" icon={IconTrashCan} onClick={onDetach}>
            {__("Unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
