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
  Spinner,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { Suspense, useMemo, useState, type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import type {
  MeasureLinkDialogQuery,
  MeasureLinkDialogQuery$data,
} from "./__generated__/MeasureLinkDialogQuery.graphql";
import type { NodeOf } from "../../types";
import { useMutationWithToasts } from "../../hooks/useMutationWithToasts";
import { useToggle } from "@probo/hooks";

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
  ) {
    createRiskMeasureMapping(input: $input) {
      success
    }
  }
`;

type Props = {
  trigger: ReactNode;
  organizationId: string;
  connectionId: string;
  riskId: string;
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
  const measures =
    data.organization?.measures?.edges?.map((edge) => edge.node) ?? [];

  const filteredMeasures = useMemo(() => {
    return measures.filter(
      (measure) =>
        measure.name.toLowerCase().includes(search.toLowerCase()) ||
        measure.description?.toLowerCase().includes(search.toLowerCase())
    );
  }, [measures, search]);

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search measures...")}
          onValueChange={setSearch}
        />
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
  const [isLinked, toggleLinked] = useToggle(false);
  const { __ } = useTranslate();
  const [attachMeasure, isFetching] = useMutationWithToasts(
    attachMeasureMutation,
    {
      successMessage: __("Measure linked successfully"),
      errorMessage: __("Failed to link measure"),
    }
  );

  const onClick = () => {
    attachMeasure({
      variables: {
        input: {
          riskId: props.riskId,
          measureId: props.measure.id,
        },
      },
      onSuccess() {
        toggleLinked();
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
