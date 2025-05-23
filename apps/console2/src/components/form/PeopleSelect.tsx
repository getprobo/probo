import { Avatar, Option, Select } from "@probo/ui";
import { Suspense } from "react";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";
import type { PeopleSelectQuery as PeopleSelectQueryType } from "./__generated__/PeopleSelectQuery.graphql";
import { useTranslate } from "@probo/i18n";
import { Controller, type Control } from "react-hook-form";

const peopleQuery = graphql`
  query PeopleSelectQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        peoples(first: 100, orderBy: { direction: ASC, field: CREATED_AT }) {
          edges {
            node {
              id
              fullName
            }
          }
        }
      }
    }
  }
`;

export function PeopleSelect({
  organization,
  name,
  control,
}: {
  organization: string;
  name: string;
  control: Control<any>;
}) {
  return (
    <Suspense
      fallback={<Select variant="editor" disabled placeholder="Loading..." />}
    >
      <PeopleSelectWithQuery
        organization={organization}
        name={name}
        control={control}
      />
    </Suspense>
  );
}

function PeopleSelectWithQuery({
  organization,
  name,
  control,
}: {
  organization: string;
  name: string;
  control: Control;
}) {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<PeopleSelectQueryType>(peopleQuery, {
    organizationId: organization,
  });

  const people = data.organization?.peoples?.edges.map((edge) => edge.node);

  return (
    <>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <Select
            id={name}
            variant="editor"
            placeholder={__("Select an owner")}
            onValueChange={field.onChange}
            key={people?.length ?? 0}
            {...field}
          >
            {people?.map((p) => (
              <Option key={p.id} value={p.id} className="flex gap-2">
                <Avatar name={p.fullName} />
                {p.fullName}
              </Option>
            ))}
          </Select>
        )}
      />
    </>
  );
}
