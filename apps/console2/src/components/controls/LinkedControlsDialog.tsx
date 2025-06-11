import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  InfiniteScrollTrigger,
  Input,
  Spinner,
} from "@probo/ui";
import {
  Suspense,
  useMemo,
  useRef,
  useState,
  type ReactNode,
  type RefObject,
} from "react";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useLazyLoadQuery, usePaginationFragment } from "react-relay";
import type { LinkedControlsDialogQuery } from "./__generated__/LinkedControlsDialogQuery.graphql";
import type {
  LinkedControlsDialogFragment$data,
  LinkedControlsDialogFragment$key,
} from "./__generated__/LinkedControlsDialogFragment.graphql";
import type { NodeOf } from "/types";
import { useDebounceCallback } from "usehooks-ts";

const query = graphql`
  query LinkedControlsDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...LinkedControlsDialogFragment
    }
  }
`;

const controlsFragment = graphql`
  fragment LinkedControlsDialogFragment on Organization
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 1 }
    after: { type: "CursorKey" }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    order: { type: "ControlOrder", defaultValue: null }
    filter: { type: "ControlFilter", defaultValue: null }
  )
  @refetchable(queryName: "LinkedControlsDialogControlsQuery") {
    controls(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "LinkedControlsDialogControlsQuery_controls") {
      edges {
        node {
          id
          name
          sectionTitle
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedControls?: { id: string }[];
  onLink: (controlId: string) => void;
  onUnlink: (controlId: string) => void;
};

type SearchRef = RefObject<{ search: (v: string) => void } | null>;

export function LinkedControlsDialog(props: Props) {
  const { __ } = useTranslate();
  const searchRef: SearchRef = useRef(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [minHeight, setMinHeight] = useState(0);
  const onSearch = (v: string) => {
    setMinHeight(contentRef.current?.clientHeight ?? 0);
    searchRef.current?.search(v);
  };
  return (
    <Dialog trigger={props.children} title={__("Link controls")}>
      <DialogContent>
        <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
          <Input
            icon={IconMagnifyingGlass}
            placeholder={__("Search measures...")}
            onValueChange={onSearch}
          />
        </div>
        <div ref={contentRef}>
          <Suspense
            fallback={
              <div style={{ minHeight }}>
                <Spinner centered />
              </div>
            }
          >
            <LinkedControlsDialogContent {...props} ref={searchRef} />
          </Suspense>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function LinkedControlsDialogContent(props: Props & { ref: SearchRef }) {
  const organizationId = useOrganizationId();
  const mainData = useLazyLoadQuery<LinkedControlsDialogQuery>(query, {
    organizationId,
  });
  const { data, loadNext, hasNext, isLoadingNext, refetch } =
    usePaginationFragment(
      controlsFragment,
      mainData.organization as LinkedControlsDialogFragment$key
    );

  const controls = data.controls?.edges?.map((edge) => edge.node) ?? [];
  const controlIds = useMemo(() => {
    return new Set(props.linkedControls?.map((c) => c.id) ?? []);
  }, [props.linkedControls]);

  props.ref.current = {
    search: useDebounceCallback((v: string) => {
      refetch({
        first: 20,
        filter: {
          query: v,
        },
      });
    }, 500),
  };

  return (
    <>
      <div className="divide-y divide-border-low">
        {controls.map((control) => (
          <ControlRow
            key={control.id}
            control={control}
            controlIds={controlIds}
            {...props}
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

function ControlRow(
  props: {
    control: NodeOf<LinkedControlsDialogFragment$data["controls"]>;
    controlIds: Set<string>;
  } & Props
) {
  const { __ } = useTranslate();
  const isLinked = props.controlIds.has(props.control.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;
  return (
    <button
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full text-start"
      onClick={() => onClick(props.control.id)}
    >
      <Badge size="md">{props.control.sectionTitle}</Badge>
      {props.control.name}

      <Button
        disabled={props.disabled}
        className="ml-auto"
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
