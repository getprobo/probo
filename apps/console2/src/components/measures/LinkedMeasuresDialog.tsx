import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  InfiniteScrollTrigger,
  Input,
  Option,
  Select,
  Spinner,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { Suspense, useMemo, useState, type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery, usePaginationFragment } from "react-relay";
import type { LinkedMeasuresDialogQuery } from "./__generated__/LinkedMeasuresDialogQuery.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { LinkedMeasuresDialogFragment$key } from "./__generated__/LinkedMeasuresDialogFragment.graphql";

const measuresQuery = graphql`
  query LinkedMeasuresDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...LinkedMeasuresDialogFragment
      }
    }
  }
`;

const measuresFragment = graphql`
  fragment LinkedMeasuresDialogFragment on Organization
  @refetchable(queryName: "LinkedMeasuresDialogQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "MeasureOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    measures(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "LinkedMeasuresDialogQuery_measures") {
      edges {
        node {
          id
          name
          state
          description
          category
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedMeasures?: { id: string }[];
  onLink: (measureId: string) => void;
  onUnlink: (measureId: string) => void;
};

export function LinkedMeasureDialog({ children, ...props }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog trigger={children} title={__("Link measures")}>
      <DialogContent>
        <Suspense fallback={<Spinner centered />}>
          <LinkedMeasuresDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function LinkedMeasuresDialogContent(props: Omit<Props, "children">) {
  const organizationId = useOrganizationId();
  const query = useLazyLoadQuery<LinkedMeasuresDialogQuery>(measuresQuery, {
    organizationId,
  });
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment(
    measuresFragment,
    query.organization as LinkedMeasuresDialogFragment$key
  );
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<string | null>(null);
  const measures = data.measures?.edges?.map((edge) => edge.node) ?? [];
  const linkedIds = useMemo(() => {
    return new Set(props.linkedMeasures?.map((m) => m.id) ?? []);
  }, [props.linkedMeasures]);

  const filteredMeasures = useMemo(() => {
    return measures.filter(
      (measure) =>
        (category === null || measure.category === category) &&
        (measure.name.toLowerCase().includes(search.toLowerCase()) ||
          measure.description?.toLowerCase().includes(search.toLowerCase()))
    );
  }, [measures, search, category]);

  const categories = useMemo(
    () => Array.from(new Set(measures.map((m) => m.category))),
    [measures]
  );

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search measures...")}
          onValueChange={setSearch}
        />
        <Select
          value={category ?? ""}
          placeholder={__("All categories")}
          onValueChange={setCategory}
          className="max-w-[180px]"
        >
          {categories.map((category) => (
            <Option key={category} value={category}>
              {category}
            </Option>
          ))}
        </Select>
      </div>
      <div className="divide-y divide-border-low">
        {filteredMeasures.map((measure) => (
          <MeasureRow
            key={measure.id}
            measure={measure}
            linkedMeasures={linkedIds}
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

type RowProps = {
  measure: { name: string; category: string; id: string };
  linkedMeasures: Set<string>;
  disabled?: boolean;
  onLink: (measureId: string) => void;
  onUnlink: (measureId: string) => void;
};

function MeasureRow(props: RowProps) {
  const { __ } = useTranslate();

  const isLinked = props.linkedMeasures.has(props.measure.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;

  return (
    <button
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full"
      onClick={() => onClick(props.measure.id)}
    >
      {props.measure.name}
      <Badge variant="neutral">{props.measure.category}</Badge>
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
