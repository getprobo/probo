import { useTranslate } from "@probo/i18n";
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
import { type ReactNode, Suspense, useMemo, useState } from "react";

import { usePaginatedMeasures } from "#/hooks/graph/usePaginatedMeasures";
import { useOrganizationId } from "#/hooks/useOrganizationId";

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
  const { data, loadNext, hasNext, isLoadingNext }
    = usePaginatedMeasures(organizationId);
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<string | null>(null);
  const measures = useMemo(() => data.measures?.edges?.map(edge => edge.node) ?? [], [data.measures]);
  const linkedIds = useMemo(() => {
    return new Set(props.linkedMeasures?.map(m => m.id) ?? []);
  }, [props.linkedMeasures]);

  const filteredMeasures = useMemo(() => {
    return measures.filter(
      measure =>
        (category === null || measure.category === category)
        && (measure.name.toLowerCase().includes(search.toLowerCase())
          || measure.description?.toLowerCase().includes(search.toLowerCase())),
    );
  }, [measures, search, category]);

  const categories = useMemo(
    () => Array.from(new Set(measures.map(m => m.category))),
    [measures],
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
          {categories.map(category => (
            <Option key={category} value={category}>
              {category}
            </Option>
          ))}
        </Select>
      </div>
      <div className="divide-y divide-border-low">
        {filteredMeasures.map(measure => (
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
          <IconComponent size={16} />
          {" "}
          {isLinked ? __("Unlink") : __("Link")}
        </span>
      </Button>
    </button>
  );
}
