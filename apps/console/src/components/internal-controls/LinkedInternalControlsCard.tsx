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
import { InternalControlBadge } from "@probo/ui/src/Molecules/Badge/InternalControlBadge";
import { clsx } from "clsx";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedInternalControlsCardFragment$key } from "#/__generated__/core/LinkedInternalControlsCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedInternalControlDialog } from "./LinkedInternalControlsDialog";

const linkedInternalControlFragment = graphql`
  fragment LinkedInternalControlsCardFragment on InternalControl {
    id
    name
    state
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      internalControlId: string;
    } & Params;
    connections: string[];
  };
}) => void;

type Props<Params> = {
  // Measures linked to the element
  internalControls: (LinkedInternalControlsCardFragment$key & { id: string })[];
  // Extra params to send to the mutation
  params: Params;
  // Disable (action when loading for instance)
  disabled?: boolean;
  // ID of the connection to update
  connectionId: string;
  // Mutation to attach a internalControl (will receive {internalControlId, ...params})
  onAttach: Mutation<Params>;
  // Mutation to detach a internalControl (will receive {internalControlId, ...params})
  onDetach: Mutation<Params>;
  variant?: "card" | "table";
  readOnly?: boolean;
};

/**
 * Reusable component that displays a list of linked internalControls
 */
export function LinkedInternalControlsCard<Params>(props: Props<Params>) {
  const { t } = useTranslation();
  const [limit, setLimit] = useState<number | null>(
    props.variant === "card" ? 4 : null,
  );
  const internalControls = useMemo(() => {
    return limit ? props.internalControls.slice(0, limit) : props.internalControls;
  }, [props.internalControls, limit]);
  const showMoreButton = limit !== null && props.internalControls.length > limit;
  const variant = props.variant ?? "table";

  const onAttach = (internalControlId: string) => {
    props.onAttach({
      variables: {
        input: {
          internalControlId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (internalControlId: string) => {
    props.onDetach({
      variables: {
        input: {
          internalControlId,
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
            {t("linkedInternalControlsCard.title")}
          </div>
          {!props.readOnly && (
            <LinkedInternalControlDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedInternalControls={props.internalControls}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {t("linkedInternalControlsCard.actions.link")}
              </Button>
            </LinkedInternalControlDialog>
          )}
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{t("linkedInternalControlsCard.columns.name")}</Th>
            <Th>{t("linkedInternalControlsCard.columns.state")}</Th>
            {!props.readOnly && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {internalControls.length === 0 && (
            <Tr>
              <Td
                colSpan={props.readOnly ? 2 : 3}
                className="text-center text-txt-secondary"
              >
                {t("linkedInternalControlsCard.empty")}
              </Td>
            </Tr>
          )}
          {internalControls.map(internalControl => (
            <InternalControlRow
              key={internalControl.id}
              internalControl={internalControl}
              onClick={onDetach}
              readOnly={props.readOnly}
            />
          ))}
          {variant === "table" && !props.readOnly && (
            <LinkedInternalControlDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedInternalControls={props.internalControls}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={3} icon={IconPlusLarge}>
                {t("linkedInternalControlsCard.actions.link")}
              </TrButton>
            </LinkedInternalControlDialog>
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
          {t("linkedInternalControlsCard.actions.showMore", {
            count: props.internalControls.length - limit,
          })}
        </Button>
      )}
    </Wrapper>
  );
}

function InternalControlRow(props: {
  internalControl: LinkedInternalControlsCardFragment$key & { id: string };
  onClick: (internalControlId: string) => void;
  readOnly?: boolean;
}) {
  const internalControl = useFragment(linkedInternalControlFragment, props.internalControl);
  const organizationId = useOrganizationId();
  const { t } = useTranslation();

  return (
    <Tr to={`/organizations/${organizationId}/governance/internal-controls/${internalControl.id}`}>
      <Td>{internalControl.name}</Td>
      <Td>
        <InternalControlBadge state={internalControl.state} />
      </Td>
      {!props.readOnly && (
        <Td noLink width={50} className="text-end">
          <Button
            variant="secondary"
            onClick={() => props.onClick(internalControl.id)}
            icon={IconTrashCan}
          >
            {t("linkedInternalControlsCard.actions.unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
