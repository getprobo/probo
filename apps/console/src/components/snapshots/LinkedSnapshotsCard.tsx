import {
  formatDate,
  getSnapshotTypeLabel,
  getSnapshotTypeUrlPath,
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

import type { LinkedSnapshotsCardFragment$key } from "#/__generated__/core/LinkedSnapshotsCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedSnapshotsDialog } from "./LinkedSnapshotsDialog";

const linkedSnapshotFragment = graphql`
  fragment LinkedSnapshotsCardFragment on Snapshot {
    id
    name
    description
    type
    createdAt
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      snapshotId: string;
    } & Params;
    connections: string[];
  };
}) => void;

type Props<Params> = {
  snapshots: (LinkedSnapshotsCardFragment$key & { id: string })[];
  params: Params;
  disabled?: boolean;
  connectionId: string;
  onAttach: Mutation<Params>;
  onDetach: Mutation<Params>;
  variant?: "card" | "table";
  readOnly?: boolean;
};

export function LinkedSnapshotsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(4);
  const snapshots = useMemo(() => {
    return limit ? props.snapshots.slice(0, limit) : props.snapshots;
  }, [props.snapshots, limit]);
  const showMoreButton = limit !== null && props.snapshots.length > limit;
  const variant = props.variant ?? "table";

  const onAttach = (snapshotId: string) => {
    props.onAttach({
      variables: {
        input: {
          snapshotId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (snapshotId: string) => {
    props.onDetach({
      variables: {
        input: {
          snapshotId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const Wrapper = variant === "card" ? Card : "div";

  const colSpanTable = props.readOnly ? 4 : 5;
  const colSpanCard = props.readOnly ? 3 : 4;

  return (
    <Wrapper padded className="space-y-[10px]">
      {variant === "card" && (
        <div className="flex justify-between">
          <div className="text-lg font-semibold">{__("Snapshots")}</div>
          {!props.readOnly && (
            <LinkedSnapshotsDialog
              disabled={props.disabled}
              linkedSnapshots={props.snapshots}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {__("Link snapshot")}
              </Button>
            </LinkedSnapshotsDialog>
          )}
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Type")}</Th>
            {variant === "table" && <Th>{__("Description")}</Th>}
            <Th>{__("Created")}</Th>
            {!props.readOnly && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {snapshots.length === 0 && (
            <Tr>
              <Td
                colSpan={variant === "table" ? colSpanTable : colSpanCard}
                className="text-center text-txt-secondary"
              >
                {__("No snapshots linked")}
              </Td>
            </Tr>
          )}
          {snapshots.map(snapshot => (
            <SnapshotRow
              key={snapshot.id}
              snapshot={snapshot}
              onClick={onDetach}
              variant={variant}
              readOnly={props.readOnly}
            />
          ))}
          {variant === "table" && !props.readOnly && (
            <LinkedSnapshotsDialog
              disabled={props.disabled}
              linkedSnapshots={props.snapshots}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={colSpanTable} icon={IconPlusLarge}>
                {__("Link snapshot")}
              </TrButton>
            </LinkedSnapshotsDialog>
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
          {sprintf(__("Show %s more"), props.snapshots.length - limit)}
        </Button>
      )}
    </Wrapper>
  );
}

function SnapshotRow(props: {
  snapshot: LinkedSnapshotsCardFragment$key & { id: string };
  onClick: (snapshotId: string) => void;
  variant: "card" | "table";
  readOnly?: boolean;
}) {
  const snapshot = useFragment(linkedSnapshotFragment, props.snapshot);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const urlPath = getSnapshotTypeUrlPath(snapshot.type);
  const snapshotUrl = `/organizations/${organizationId}/snapshots/${snapshot.id}${urlPath}`;

  return (
    <Tr to={snapshotUrl}>
      <Td className="font-medium">{snapshot.name}</Td>
      <Td>
        <Badge variant="neutral">
          {getSnapshotTypeLabel(__, snapshot.type)}
        </Badge>
      </Td>
      {props.variant === "table" && (
        <Td className="text-txt-secondary">
          {snapshot.description || __("No description")}
        </Td>
      )}
      <Td className="text-txt-tertiary">{formatDate(snapshot.createdAt)}</Td>
      {!props.readOnly && (
        <Td noLink width={50} className="text-end">
          <Button
            variant="secondary"
            onClick={() => props.onClick(snapshot.id)}
            icon={IconTrashCan}
          >
            {__("Unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
