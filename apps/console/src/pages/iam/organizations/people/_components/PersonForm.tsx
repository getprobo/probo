import { formatDatetime, getAssignableRoles, getRoles, peopleRoles } from "@probo/helpers";
import { roles } from "@probo/helpers/src/roles";
import { useTranslate } from "@probo/i18n";
import { Button, Field, Input, Option } from "@probo/ui";
import { use } from "react";
import { useWatch } from "react-hook-form";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";
import { z } from "zod";

import type { PersonForm_createMutation } from "#/__generated__/iam/PersonForm_createMutation.graphql";
import type { PersonForm_updateMutation } from "#/__generated__/iam/PersonForm_updateMutation.graphql";
import type { PersonFormFragment$key } from "#/__generated__/iam/PersonFormFragment.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { EmailsField } from "#/components/form/EmailsField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

const fragment = graphql`
  fragment PersonFormFragment on Profile {
    id
    fullName
    emailAddress
    membership @required(action: THROW) {
      role
    }
    kind
    position
    additionalEmailAddresses
    contractStartDate
    contractEndDate
    canUpdate: permission(action: "iam:membership-profile:update")
  }
`;

const createPersonMutation = graphql`
  mutation PersonForm_createMutation($input: CreateUserInput! $connections: [ID!]!) {
    createUser(input: $input) {
      profileEdge @prependEdge(connections: $connections) {
        node {
          ...PeopleListItemFragment
        }
      }
    }
  }
`;

const updatePersonMutation = graphql`
  mutation PersonForm_updateMutation($input: UpdateUserInput!) {
    updateUser(input: $input) {
      profile {
        id
      }
    }
  }
`;

const schema = z.object({
  fullName: z.string().min(1),
  emailAddress: z.string().email(),
  role: z.enum(roles),
  position: z.string().min(1).optional().nullable(),
  additionalEmailAddresses: z.preprocess(
    // Empty additional emails are skipped
    v => (v as string[]).filter(v => !!v),
    z.array(z.string().email()),
  ),
  kind: z.enum(peopleRoles),
  contractStartDate: z.string().optional().nullable(),
  contractEndDate: z.string().optional().nullable(),
});

export function PersonForm(props: {
  id?: string;
  connectionId?: DataID;
  disabled?: boolean;
  defaultValues?: z.infer<typeof schema>;
  onSubmit?: () => void;
}) {
  const {
    id,
    connectionId = "",
    disabled = false,
    defaultValues = {
      fullName: "",
      emailAddress: "",
      role: "EMPLOYEE",
      additionalEmailAddresses: [],
      kind: "EMPLOYEE",
      position: null,
      contractStartDate: null,
      contractEndDate: null,
    },
    onSubmit,
  } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const { role } = use(CurrentUser);
  const availableRoles = getAssignableRoles(role);

  const { control, formState, handleSubmit: handleSubmitWrapper, register, reset }
    = useFormWithSchema(schema, { defaultValues });
  const watchedRole = useWatch({
    control,
    name: "role",
    defaultValue: "EMPLOYEE",
  });
  const [createPerson, isCreating] = useMutationWithToasts<PersonForm_createMutation>(
    createPersonMutation,
    {
      successMessage: __("Person created successfully."),
      errorMessage: __("Failed to create person"),
    },
  );
  const [updatePerson, isUpdating] = useMutationWithToasts<PersonForm_updateMutation>(
    updatePersonMutation,
    {
      successMessage: __("Person updated successfully."),
      errorMessage: __("Failed to update person"),
    },
  );
  const handleSubmit = handleSubmitWrapper(async (data: z.infer<typeof schema>) => {
    const input = {
      fullName: data.fullName,
      emailAddress: data.emailAddress,
      role: data.role,
      additionalEmailAddresses: data.additionalEmailAddresses,
      kind: data.kind,
      position: data.position,
      contractStartDate: formatDatetime(data.contractStartDate) ?? null,
      contractEndDate: formatDatetime(data.contractEndDate) ?? null,
    };

    if (id) {
      await updatePerson({
        variables: { input: { ...input, id } },
        onCompleted: () => {
          reset(data);
          onSubmit?.();
        },
      });
    } else {
      await createPerson({
        variables: {
          input: { ...input, organizationId },
          connections: [connectionId],
        },
        onCompleted: () => {
          reset(data);
          onSubmit?.();
        },
      });
    }
  });

  return (
    <form onSubmit={e => void handleSubmit(e)} className="space-y-4">
      <Field label={__("Full name *")} {...register("fullName")} type="text" />
      {id
        ? (
            <>
              <input type="hidden" {...register("emailAddress")} disabled />
              <input type="hidden" {...register("role")} disabled />
            </>
          )
        : (
            <>
              <Field label={__("Email Address *")} {...register("emailAddress")} type="email" disabled={disabled || !!id} />
              <ControlledField
                control={control}
                name="role"
                type="select"
                label={__("Role *")}
                disabled={disabled || !!id}
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
                    {__("Read-only access without settings, tasks and meetings")}
                  </p>
                )}
                {watchedRole === "EMPLOYEE" && (
                  <p>{__("Access to employee page")}</p>
                )}
              </div>
            </>
          )}
      <ControlledField
        control={control}
        name="kind"
        type="select"
        label={__("Type")}
        disabled={disabled}
      >
        {getRoles(__).map(role => (
          <Option key={role.value} value={role.value}>
            {role.label}
          </Option>
        ))}
      </ControlledField>
      <Field
        label={__("Position")}
        {...register("position")}
        type="text"
        placeholder={__("e.g. CEO, CFO, etc.")}
        disabled={disabled}
      />
      <EmailsField control={control} register={register} />
      <Field label={__("Contract start date")}>
        <Input
          {...register("contractStartDate")}
          type="date"
          disabled={disabled}
        />
      </Field>
      <Field label={__("Contract end date")}>
        <Input
          {...register("contractEndDate")}
          type="date"
          disabled={disabled}
        />
      </Field>
      <div className="flex justify-end">
        {(!id || formState.isDirty) && !disabled && (
          <Button type="submit" disabled={isUpdating || isCreating || !formState.isValid}>
            {id ? __("Update") : __("Create")}
          </Button>
        )}
      </div>
    </form>
  );
}

export function PersonFormLoader(props: { fragmentRef: PersonFormFragment$key }) {
  const { fragmentRef } = props;

  const person = useFragment<PersonFormFragment$key>(fragment, fragmentRef);

  return (
    <PersonForm
      id={person.id}
      disabled={!person.canUpdate}
      defaultValues={
        {
          kind: person.kind,
          fullName: person.fullName,
          emailAddress: person.emailAddress,
          role: person.membership.role,
          position: person.position,
          additionalEmailAddresses: [...person.additionalEmailAddresses],
          contractStartDate: person.contractStartDate?.split("T")[0] || "",
          contractEndDate: person.contractEndDate?.split("T")[0] || "",
        }
      }
    />
  );
}
