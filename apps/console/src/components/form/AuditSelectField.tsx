import { Field, Option, Select, Badge } from "@probo/ui";
import { Suspense, type ComponentProps } from "react";
import { useTranslate } from "@probo/i18n";
import { type Control, Controller, type FieldValues, type Path } from "react-hook-form";
import { useLazyLoadQuery, graphql } from "react-relay";
import { getAuditStateVariant, getAuditStateLabel } from "@probo/helpers";
import type { AuditSelectFieldQuery } from "./__generated__/AuditSelectFieldQuery.graphql";

const auditsQuery = graphql`
  query AuditSelectFieldQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        audits(first: 100) {
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

export function AuditSelectField<T extends FieldValues = FieldValues>({
  organizationId,
  control,
  ...props
}: Props<T>) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" loading placeholder="Loading..." />}
      >
        <AuditSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
        />
      </Suspense>
    </Field>
  );
}

function AuditSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<Props<T>, "organizationId" | "control" | "name" | "disabled">
) {
  const { __ } = useTranslate();
  const { name, organizationId, control } = props;
  const data = useLazyLoadQuery<AuditSelectFieldQuery>(auditsQuery, { organizationId }, { fetchPolicy: "network-only" });
  const audits = data?.organization?.audits?.edges?.map((edge) => edge.node).filter((node) => node !== null) ?? [];

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <Select
          disabled={props.disabled}
          id={name}
          variant="editor"
          placeholder={__("Select an audit")}
          onValueChange={field.onChange}
          key={audits?.length.toString() ?? "0"}
          {...field}
          className="w-full"
          value={field.value ?? ""}
        >
          {audits?.map((audit) => (
            <Option key={audit.id} value={audit.id}>
              <div className="flex items-center justify-between w-full">
                <span>
                  {audit.name
                    ? `${audit.framework.name} - ${audit.name}`
                    : audit.framework.name
                  }
                </span>
                <div className="ml-3">
                  <Badge variant={getAuditStateVariant(audit.state)}>
                    {getAuditStateLabel(__, audit.state)}
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
