import { Avatar, Option, Select } from "@probo/ui";
import { Suspense } from "react";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";
import type { UserSelectQuery as UserSelectQueryType } from "./__generated__/UserSelectQuery.graphql";
import { useTranslate } from "@probo/i18n";
import { Controller, type Control } from "react-hook-form";

const usersQuery = graphql`
  query UserSelectQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        users(first: 100, orderBy: { direction: ASC, field: CREATED_AT }) {
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

export function UserSelect({
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
      <UserSelectWithQuery
        organization={organization}
        name={name}
        control={control}
      />
    </Suspense>
  );
}

function UserSelectWithQuery({
  organization,
  name,
  control,
}: {
  organization: string;
  name: string;
  control: Control;
}) {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<UserSelectQueryType>(usersQuery, {
    organizationId: organization,
  });

  const users = data.organization?.users?.edges.map((edge) => edge.node);

  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <Select
          id={name}
          variant="editor"
          placeholder={__("Select an user")}
          {...field}
        >
          {users?.map((user) => (
            <Option key={user.id} value={user.id}>
              <Avatar name={user.fullName} />
              {user.fullName}
            </Option>
          ))}
        </Select>
      )}
    />
  );
}
