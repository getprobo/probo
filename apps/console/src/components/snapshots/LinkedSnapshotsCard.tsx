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
  Badge,
  TrButton,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { LinkedSnapshotsCardFragment$key } from "./__generated__/LinkedSnapshotsCardFragment.graphql";
import { useFragment } from "react-relay";
import { useMemo, useState } from "react";
import { sprintf, getSnapshotTypeLabel, getSnapshotTypeUrlPath } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { LinkedSnapshotsDialog } from "./LinkedSnapshotsDialog";
import clsx from "clsx";

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

  return (
    <Wrapper padded className="space-y-[10px]">
      {variant === "card" && (
        <div className="flex justify-between">
          <div className="text-lg font-semibold">{__("Snapshots")}</div>
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
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Type")}</Th>
            {variant === "table" && <Th>{__("Description")}</Th>}
            <Th>{__("Created")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {snapshots.length === 0 && (
            <Tr>
              <Td colSpan={variant === "table" ? 5 : 3} className="text-center text-txt-secondary">
                {__("No snapshots linked")}
              </Td>
            </Tr>
          )}
          {snapshots.map((snapshot) => (
            <SnapshotRow
              key={snapshot.id}
              snapshot={snapshot}
              onClick={onDetach}
              variant={variant}
            />
          ))}
          {variant === "table" && (
            <LinkedSnapshotsDialog
              disabled={props.disabled}
              linkedSnapshots={props.snapshots}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={variant === "table" ? 5 : 3} icon={IconPlusLarge}>
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
}) {
  const snapshot = useFragment(linkedSnapshotFragment, props.snapshot);
  const organizationId = useOrganizationId();
  const { __, dateFormat } = useTranslate();

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
      <Td className="text-txt-tertiary">
        {dateFormat(snapshot.createdAt, { year: "numeric", month: "short", day: "numeric" })}
      </Td>
      <Td noLink width={50} className="text-end">
        <Button
          variant="secondary"
          onClick={() => props.onClick(snapshot.id)}
          icon={IconTrashCan}
        >
          {__("Unlink")}
        </Button>
      </Td>
    </Tr>
  );
}
