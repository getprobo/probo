// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { getAssignableRoles } from "@probo/helpers";
import { roles } from "@probo/helpers/src/roles";
import { useTranslate } from "@probo/i18n";
import { Button, Field, Option } from "@probo/ui";
import { use } from "react";
import { useWatch } from "react-hook-form";
import { type DataID, graphql } from "relay-runtime";
import { z } from "zod";

import type { AddMemberForm_createMutation } from "#/__generated__/iam/AddMemberForm_createMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

const createMemberMutation = graphql`
  mutation AddMemberForm_createMutation($input: CreateUserInput! $connections: [ID!]!) {
    createUser(input: $input) {
      profileEdge @prependEdge(connections: $connections) {
        node {
          ...MembersListItemFragment
        }
      }
    }
  }
`;

const schema = z.object({
  emailAddress: z.string().email(),
  role: z.enum(roles),
});

export function AddMemberForm(props: {
  connectionId: DataID;
  onSubmit?: () => void;
}) {
  const { connectionId, onSubmit } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const { role } = use(CurrentUser);
  const availableRoles = getAssignableRoles(role);

  const { control, formState, handleSubmit: handleSubmitWrapper, register, reset }
    = useFormWithSchema(schema, {
      defaultValues: {
        emailAddress: "",
        role: "EMPLOYEE" as const,
      },
    });
  const watchedRole = useWatch({
    control,
    name: "role",
    defaultValue: "EMPLOYEE",
  });
  const [createMember, isCreating] = useMutationWithToasts<AddMemberForm_createMutation>(
    createMemberMutation,
    {
      successMessage: __("Member added successfully."),
      errorMessage: __("Failed to add member"),
    },
  );

  const handleSubmit = handleSubmitWrapper(async (data: z.infer<typeof schema>) => {
    await createMember({
      variables: {
        input: {
          fullName: data.emailAddress.split("@")[0] ?? data.emailAddress,
          emailAddress: data.emailAddress,
          role: data.role,
          organizationId,
        },
        connections: [connectionId],
      },
      onCompleted: () => {
        reset();
        onSubmit?.();
      },
    });
  });

  return (
    <form onSubmit={e => void handleSubmit(e)} className="space-y-4">
      <Field label={__("Email Address *")} {...register("emailAddress")} type="email" />
      <ControlledField
        control={control}
        name="role"
        type="select"
        label={__("Role *")}
      >
        {availableRoles.includes("OWNER") && (
          <Option value="OWNER">{__("Owner")}</Option>
        )}
        {availableRoles.includes("ADMIN") && (
          <Option value="ADMIN">{__("Admin")}</Option>
        )}
        {availableRoles.includes("VIEWER") && (
          <Option value="VIEWER">{__("Viewer")}</Option>
        )}
        {availableRoles.includes("AUDITOR") && (
          <Option value="AUDITOR">{__("Auditor")}</Option>
        )}
        {availableRoles.includes("EMPLOYEE") && (
          <Option value="EMPLOYEE">{__("Employee")}</Option>
        )}
      </ControlledField>

      <div className="mt-4 space-y-2 text-sm text-txt-tertiary">
        {watchedRole === "OWNER" && (
          <p>{__("Full access to everything")}</p>
        )}
        {watchedRole === "ADMIN" && (
          <p>
            {__("Full access except organization setup and API keys")}
          </p>
        )}
        {watchedRole === "VIEWER" && <p>{__("Read-only access")}</p>}
        {watchedRole === "AUDITOR" && (
          <p>
            {__("Read-only access without settings and tasks")}
          </p>
        )}
        {watchedRole === "EMPLOYEE" && (
          <p>{__("Access to employee page")}</p>
        )}
      </div>

      <div className="flex justify-end">
        <Button type="submit" disabled={isCreating || !formState.isValid}>
          {__("Add member")}
        </Button>
      </div>
    </form>
  );
}
