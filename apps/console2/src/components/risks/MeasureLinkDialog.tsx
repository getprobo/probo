import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  Input,
  Option,
  Select,
  Spinner,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { Suspense, useMemo, useState, type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery, useMutation } from "react-relay";
import type {
  MeasureLinkDialogQuery,
  MeasureLinkDialogQuery$data,
} from "./__generated__/MeasureLinkDialogQuery.graphql";
import type { NodeOf } from "/types";

const measuresQuery = graphql`
  query MeasureLinkDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        measures(first: 100) @connection(key: "Organization__measures") {
          edges {
            node {
              id
              name
              description
              category
              state
            }
          }
        }
      }
    }
  }
`;

const attachMeasureMutation = graphql`
  mutation MeasureLinkDialogCreateMutation(
    $input: CreateRiskMeasureMappingInput!
    $connections: [ID!]!
  ) {
    createRiskMeasureMapping(input: $input) {
      measureEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          category
          state
        }
      }
    }
  }
`;

export const detachMeasureMutation = graphql`
  mutation MeasureLinkDialogDetachMutation(
    $input: DeleteRiskMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskMeasureMapping(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

type Props = {
  trigger: ReactNode;
  organizationId: string;
  connectionId: string;
  riskId: string;
  linkedIds?: Set<string>;
};

export function MeasureLinkDialog({ trigger, ...props }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog trigger={trigger} title={__("Manage Risk Measures")}>
      <DialogContent className="px-6">
        <p className="text-sm text-txt-secondary mt-6">
          {__("Link or unlink measures to manage this risk.")}
        </p>
        <Suspense fallback={<Spinner centered />}>
          <MeasureLinkDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function MeasureLinkDialogContent(props: Omit<Props, "trigger">) {
  const data = useLazyLoadQuery<MeasureLinkDialogQuery>(measuresQuery, {
    organizationId: props.organizationId,
  });
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<string | null>(null);
  const measures =
    data.organization?.measures?.edges?.map((edge) => edge.node) ?? [];

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
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0">
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
      <div className="space-y-2 divide-y divide-border-low">
        {filteredMeasures.map((measure) => (
          <MeasureRow key={measure.id} measure={measure} {...props} />
        ))}
      </div>
    </>
  );
}

type RowProps = {
  measure: NodeOf<
    Required<MeasureLinkDialogQuery$data["organization"]>["measures"]
  >;
} & Omit<Props, "trigger">;

function MeasureRow(props: RowProps) {
  const isLinked = props.linkedIds?.has(props.measure.id) ?? false;
  const { __ } = useTranslate();
  const [attachMeasure, isFetchingAttach] = useMutation(attachMeasureMutation);
  const [detachMeasure, isFetchingDetach] = useMutation(detachMeasureMutation);

  const isFetching = isFetchingAttach || isFetchingDetach;

  const onClick = () => {
    const action = isLinked ? detachMeasure : attachMeasure;
    action({
      variables: {
        input: {
          riskId: props.riskId,
          measureId: props.measure.id,
        },
        connections: [props.connectionId],
      },
    });
  };

  return (
    <div className="py-4 flex items-center gap-4 hover:bg-subtle">
      {props.measure.name}
      <Badge variant="neutral">{props.measure.category}</Badge>
      <Button
        disabled={isFetching}
        icon={isLinked ? IconTrashCan : IconPlusLarge}
        className="ml-auto"
        variant={isLinked ? "secondary" : "primary"}
        onClick={onClick}
      >
        {isLinked ? __("Unlink") : __("Link")}
      </Button>
    </div>
  );
}
