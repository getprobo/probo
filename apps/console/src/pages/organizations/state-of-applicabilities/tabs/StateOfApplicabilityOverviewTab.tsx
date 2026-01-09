import { useOutletContext } from "react-router";
import { Card, Field } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { formatDate } from "@probo/helpers";
import type { StateOfApplicabilityGraphNodeQuery$data } from "/hooks/graph/__generated__/StateOfApplicabilityGraphNodeQuery.graphql";

type StateOfApplicabilityNode = NonNullable<StateOfApplicabilityGraphNodeQuery$data["node"]>;

export default function StateOfApplicabilityOverviewTab() {
  const { stateOfApplicability } = useOutletContext<{
    stateOfApplicability: StateOfApplicabilityNode;
  }>();
  const { __ } = useTranslate();

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{__("State of Applicability details")}</h2>
      <Card className="space-y-4" padded>
        <Field
          label={__("Name")}
          value={stateOfApplicability.name}
          disabled
        />
        {stateOfApplicability.sourceId && (
          <Field
            label={__("Source")}
            value={stateOfApplicability.sourceId}
            disabled
          />
        )}
        {stateOfApplicability.snapshotId && (
          <Field
            label={__("Snapshot")}
            value={stateOfApplicability.snapshotId}
            disabled
          />
        )}
        <Field
          label={__("Created at")}
          value={formatDate(stateOfApplicability.createdAt)}
          disabled
        />
        <Field
          label={__("Updated at")}
          value={formatDate(stateOfApplicability.updatedAt)}
          disabled
        />
      </Card>
    </div>
  );
}
