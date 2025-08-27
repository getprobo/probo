import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Badge,
  InfiniteScrollTrigger,
  Input,
  Spinner,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { getSnapshotTypeLabel } from "@probo/helpers";
import { Suspense, useMemo, useState, type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery, usePaginationFragment } from "react-relay";
import type { LinkedSnapshotsDialogQuery } from "./__generated__/LinkedSnapshotsDialogQuery.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { NodeOf } from "/types";
import type {
  LinkedSnapshotsDialogFragment$data,
  LinkedSnapshotsDialogFragment$key,
} from "./__generated__/LinkedSnapshotsDialogFragment.graphql";

const snapshotsQuery = graphql`
  query LinkedSnapshotsDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...LinkedSnapshotsDialogFragment
      }
    }
  }
`;

const snapshotsFragment = graphql`
  fragment LinkedSnapshotsDialogFragment on Organization
  @refetchable(queryName: "LinkedSnapshotsDialogQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "SnapshotOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    snapshots(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "LinkedSnapshotsDialogQuery_snapshots") {
      edges {
        node {
          id
          name
          description
          type
          createdAt
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  disabled?: boolean;
  linkedSnapshots?: { id: string }[];
  onLink: (snapshotId: string) => void;
  onUnlink: (snapshotId: string) => void;
};

export function LinkedSnapshotsDialog({ children, ...props }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog trigger={children} title={__("Link snapshots")}>
      <DialogContent>
        <Suspense fallback={<Spinner centered />}>
          <LinkedSnapshotsDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function LinkedSnapshotsDialogContent(props: Omit<Props, "children">) {
  const organizationId = useOrganizationId();
  const query = useLazyLoadQuery<LinkedSnapshotsDialogQuery>(snapshotsQuery, {
    organizationId,
  });
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    LinkedSnapshotsDialogQuery,
    LinkedSnapshotsDialogFragment$key>(
    snapshotsFragment,
    query.organization
  );

  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const snapshots = data.snapshots?.edges?.map((edge) => edge.node) ?? [];
  const linkedIds = useMemo(() => {
    return new Set(props.linkedSnapshots?.map((s) => s.id) ?? []);
  }, [props.linkedSnapshots]);

  const filteredSnapshots = useMemo(() => {
    return snapshots.filter((snapshot) =>
      snapshot.name.toLowerCase().includes(search.toLowerCase())
    );
  }, [snapshots, search]);

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search snapshots...")}
          onValueChange={setSearch}
        />
      </div>
      <div className="divide-y divide-border-low">
        {filteredSnapshots.map((snapshot) => (
          <SnapshotRow
            key={snapshot.id}
            snapshot={snapshot}
            linkedSnapshots={linkedIds}
            onLink={props.onLink}
            onUnlink={props.onUnlink}
            disabled={props.disabled}
          />
        ))}
        {hasNext && (
          <InfiniteScrollTrigger
            loading={isLoadingNext}
            onView={() => loadNext(20)}
          />
        )}
      </div>
    </>
  );
}

type Snapshot = NodeOf<LinkedSnapshotsDialogFragment$data["snapshots"]>;

type RowProps = {
  snapshot: Snapshot;
  linkedSnapshots: Set<string>;
  disabled?: boolean;
  onLink: (snapshotId: string) => void;
  onUnlink: (snapshotId: string) => void;
};

function SnapshotRow(props: RowProps) {
  const { __, dateFormat } = useTranslate();

  const isLinked = props.linkedSnapshots.has(props.snapshot.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;

  return (
    <button
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full"
      onClick={() => onClick(props.snapshot.id)}
    >
      <div className="flex-1 flex items-center gap-4">
        <div className="font-medium min-w-0 flex-shrink-0">
          {props.snapshot.name}
        </div>
        <Badge variant="neutral" className="flex-shrink-0 ml-6">
          {getSnapshotTypeLabel(__, props.snapshot.type)}
        </Badge>
        <div className="text-sm text-txt-secondary min-w-0 flex-1 text-left">
          {props.snapshot.description || __("No description")}
        </div>
        <div className="text-sm text-txt-tertiary flex-shrink-0">
          {dateFormat(props.snapshot.createdAt, { year: "numeric", month: "short", day: "numeric" })}
        </div>
      </div>
      <Button
        disabled={props.disabled}
        variant={isLinked ? "secondary" : "primary"}
        asChild
      >
        <span>
          <IconComponent size={16} /> {isLinked ? __("Unlink") : __("Link")}
        </span>
      </Button>
    </button>
  );
}
