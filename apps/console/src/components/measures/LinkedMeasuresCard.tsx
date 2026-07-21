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
import { MeasureBadge } from "@probo/ui/src/Molecules/Badge/MeasureBadge";
import { clsx } from "clsx";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedMeasuresCardFragment$key } from "#/__generated__/core/LinkedMeasuresCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedMeasureDialog } from "./LinkedMeasuresDialog";

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
  variant?: "card" | "table";
  readOnly?: boolean;
};

/**
 * Reusable component that displays a list of linked measures
 */
export function LinkedMeasuresCard<Params>(props: Props<Params>) {
  const { t } = useTranslation();
  const [limit, setLimit] = useState<number | null>(
    props.variant === "card" ? 4 : null,
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
          <div className="text-lg font-semibold">
            {t("linkedMeasuresCard.title")}
          </div>
          {!props.readOnly && (
            <LinkedMeasureDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedMeasures={props.measures}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {t("linkedMeasuresCard.actions.link")}
              </Button>
            </LinkedMeasureDialog>
          )}
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{t("linkedMeasuresCard.columns.name")}</Th>
            <Th>{t("linkedMeasuresCard.columns.state")}</Th>
            {!props.readOnly && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {measures.length === 0 && (
            <Tr>
              <Td
                colSpan={props.readOnly ? 2 : 3}
                className="text-center text-txt-secondary"
              >
                {t("linkedMeasuresCard.empty")}
              </Td>
            </Tr>
          )}
          {measures.map(measure => (
            <MeasureRow
              key={measure.id}
              measure={measure}
              onClick={onDetach}
              readOnly={props.readOnly}
            />
          ))}
          {variant === "table" && !props.readOnly && (
            <LinkedMeasureDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedMeasures={props.measures}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={3} icon={IconPlusLarge}>
                {t("linkedMeasuresCard.actions.link")}
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
          {t("linkedMeasuresCard.actions.showMore", {
            count: props.measures.length - limit,
          })}
        </Button>
      )}
    </Wrapper>
  );
}

function MeasureRow(props: {
  measure: LinkedMeasuresCardFragment$key & { id: string };
  onClick: (measureId: string) => void;
  readOnly?: boolean;
}) {
  const measure = useFragment(linkedMeasureFragment, props.measure);
  const organizationId = useOrganizationId();
  const { t } = useTranslation();

  return (
    <Tr to={`/organizations/${organizationId}/measures/${measure.id}`}>
      <Td>{measure.name}</Td>
      <Td>
        <MeasureBadge state={measure.state} />
      </Td>
      {!props.readOnly && (
        <Td noLink width={50} className="text-end">
          <Button
            variant="secondary"
            onClick={() => props.onClick(measure.id)}
            icon={IconTrashCan}
          >
            {t("linkedMeasuresCard.actions.unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
