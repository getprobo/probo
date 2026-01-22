import { getRoles, peopleRoles } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card, IconCheckmark1, Option } from "@probo/ui";
import { type PropsWithChildren } from "react";
import { useOutletContext } from "react-router";
import { z } from "zod";

import type { PeopleGraphNodeQuery$data } from "#/__generated__/core/PeopleGraphNodeQuery.graphql";
import type { PeopleGraphUpdateMutation } from "#/__generated__/core/PeopleGraphUpdateMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { updatePeopleMutation } from "#/hooks/graph/PeopleGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const schema = z.object({
  kind: z.enum(peopleRoles),
});

export default function PeopleRoleTab() {
  const { people } = useOutletContext<{
    people: PeopleGraphNodeQuery$data["node"];
  }>();
  const { __ } = useTranslate();
  const { control, formState, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        kind: people.kind,
      },
    },
  );
  const [mutate, isMutating] = useMutationWithToasts<PeopleGraphUpdateMutation>(
    updatePeopleMutation,
    {
      successMessage: __("Member updated successfully."),
      errorMessage: __("Failed to update member"),
    },
  );

  const onSubmit = handleSubmit(async (data) => {
    await mutate({
      variables: {
        input: {
          id: people.id!,
          kind: data.kind,
          // TODO : make these field optional in the query (server side)
          fullName: people.fullName!,
          primaryEmailAddress: people.primaryEmailAddress!,
          additionalEmailAddresses: people.additionalEmailAddresses ?? [],
          contractStartDate: people.contractStartDate,
          contractEndDate: people.contractEndDate,
        },
      },
      onCompleted: () => {
        reset(data);
      },
    });
  });

  return (
    <form onSubmit={e => void onSubmit(e)} className="space-y-4">
      <Card padded className="space-y-4">
        <ControlledField
          control={control}
          name="kind"
          type="select"
          label={__("Role")}
          disabled={!people.canUpdate}
        >
          {getRoles(__).map(role => (
            <Option key={role.value} value={role.value}>
              {role.label}
            </Option>
          ))}
        </ControlledField>
        <div className="space-y-2 ">
          <div className="text-sm font-medium">{__("Permissions")}</div>
          <ul className="text-sm text-txt-tertiary space-y-2">
            <AccessItem>
              {__("Access dashboard & reports relevant to their team")}
            </AccessItem>
            <AccessItem>
              {__("Create and manage own tasks, tickets, or projects")}
            </AccessItem>
            <AccessItem>
              {__("Comment on shared documents or projects")}
            </AccessItem>
            <AccessItem>
              {__("Receive notifications and system alerts")}
            </AccessItem>
            <AccessItem>
              {__("Join and participate in team chats or threads")}
            </AccessItem>
          </ul>
        </div>
      </Card>
      {people.canUpdate && (
        <div className="flex justify-end">
          {formState.isDirty && (
            <Button type="submit" disabled={isMutating}>
              {__("Update")}
            </Button>
          )}
        </div>
      )}
    </form>
  );
}

function AccessItem({ children }: PropsWithChildren) {
  return (
    <li className="flex gap-2 items-center">
      <IconCheckmark1 size={16} />
      {children}
    </li>
  );
}
