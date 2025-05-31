import { useTranslate } from "@probo/i18n";
import { Button, Card, IconCheckmark1 } from "@probo/ui";
import type { PropsWithChildren } from "react";
import z from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { ControlledField } from "/components/form/ControlledField";
import { Option } from "@probo/ui";
import { getRoles, peopleRoles } from "@probo/helpers";
import type { PeopleGraphNodeQuery$data } from "/hooks/graph/__generated__/PeopleGraphNodeQuery.graphql";
import { useOutletContext } from "react-router";
import { updatePeopleMutation } from "/hooks/graph/PeopleGraph";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { PeopleGraphUpdateMutation } from "/hooks/graph/__generated__/PeopleGraphUpdateMutation.graphql";

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
    }
  );
  const [mutate, isMutating] = useMutationWithToasts<PeopleGraphUpdateMutation>(
    updatePeopleMutation,
    {
      successMessage: __("Member updated successfully."),
      errorMessage: __("Failed to update member. Please try again."),
    }
  );

  const onSubmit = handleSubmit((data) => {
    mutate({
      variables: {
        input: {
          id: people.id!,
          kind: data.kind,
          // TODO : make these field optional in the query (server side)
          fullName: people.fullName!,
          primaryEmailAddress: people.primaryEmailAddress!,
          additionalEmailAddresses: people.additionalEmailAddresses ?? [],
        },
      },
      onCompleted: () => {
        reset(data);
      },
    });
  });

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <Card padded className="space-y-4">
        <ControlledField
          control={control}
          name="kind"
          type="select"
          label={__("Role")}
        >
          {getRoles(__).map((role) => (
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
      <div className="flex justify-end">
        {formState.isDirty && (
          <Button type="submit" disabled={isMutating}>
            {__("Update")}
          </Button>
        )}
      </div>
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
