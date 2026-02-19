import { getReportStateLabel, getReportStateVariant } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Field, Option, Select } from "@probo/ui";
import { type ComponentProps, Suspense } from "react";
import {
  type Control,
  Controller,
  type FieldValues,
  type Path,
} from "react-hook-form";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { ReportSelectFieldQuery } from "#/__generated__/core/ReportSelectFieldQuery.graphql";

const reportsQuery = graphql`
  query ReportSelectFieldQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        reports(first: 100) {
          edges {
            node {
              id
              name
              framework {
                id
                name
              }
              state
              validFrom
              validUntil
            }
          }
        }
      }
    }
  }
`;

type Props<T extends FieldValues = FieldValues> = {
  organizationId: string;
  control: Control<T>;
  name: Path<T>;
  label?: string;
  error?: string;
} & ComponentProps<typeof Field>;

export function ReportSelectField<T extends FieldValues = FieldValues>({
  organizationId,
  control,
  ...props
}: Props<T>) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" loading placeholder="Loading..." />}
      >
        <ReportSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
        />
      </Suspense>
    </Field>
  );
}

function ReportSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<Props<T>, "organizationId" | "control" | "name" | "disabled">,
) {
  const { __ } = useTranslate();
  const { name, organizationId, control } = props;
  const data = useLazyLoadQuery<ReportSelectFieldQuery>(
    reportsQuery,
    { organizationId },
    { fetchPolicy: "network-only" },
  );
  const reports
    = data?.organization?.reports?.edges
      ?.map(edge => edge.node)
      .filter(node => node !== null) ?? [];

  const NONE_VALUE = "__NONE__";

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <Select
          disabled={props.disabled}
          id={name}
          variant="editor"
          placeholder={__("Select a report")}
          onValueChange={value =>
            field.onChange(value === NONE_VALUE ? "" : value)}
          key={reports?.length.toString() ?? "0"}
          {...field}
          className="w-full"
          value={field.value || NONE_VALUE}
        >
          <Option value={NONE_VALUE}>
            <span className="text-txt-tertiary">{__("None")}</span>
          </Option>
          {reports?.map(report => (
            <Option key={report.id} value={report.id}>
              <div className="flex items-center justify-between w-full">
                <span>
                  {report.name
                    ? `${report.framework.name} - ${report.name}`
                    : report.framework.name}
                </span>
                <div className="ml-3">
                  <Badge variant={getReportStateVariant(report.state)}>
                    {getReportStateLabel(__, report.state)}
                  </Badge>
                </div>
              </div>
            </Option>
          ))}
        </Select>
      )}
    />
  );
}
