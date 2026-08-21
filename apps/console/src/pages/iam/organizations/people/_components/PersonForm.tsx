// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { formatDatetime, getAssignableRoles, getMembershipRoles, roles } from "@probo/helpers";
import { Button, Field, Input, Option } from "@probo/ui";
import { use } from "react";
import { useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { PersonForm_createMutation } from "#/__generated__/iam/PersonForm_createMutation.graphql";
import type { PersonForm_updateMutation } from "#/__generated__/iam/PersonForm_updateMutation.graphql";
import type { PersonFormFragment$key } from "#/__generated__/iam/PersonFormFragment.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { EmailsField } from "#/components/form/EmailsField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { z } from "#/lib/zod";
import { CurrentUser } from "#/providers/CurrentUser";

const fragment = graphql`
  fragment PersonFormFragment on Profile {
    id
    fullName
    emailAddress
    source
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
        ...PersonFormFragment
        ...PeopleListItemFragment
      }
    }
  }
`;

const emptyToNull = (value: unknown) => (value === "" || value == null ? null : value);

const sharedPersonFields = {
  fullName: z.string().min(1),
  // SCIM often syncs title/userType as ""; treat blanks as unset so the form
  // stays valid and contract dates remain savable for SCIM-managed people.
  position: z.preprocess(emptyToNull, z.string().min(1).nullable().optional()),
  additionalEmailAddresses: z.preprocess(
    // Empty additional emails are skipped
    v => (v as string[]).filter(v => !!v),
    z.array(z.string().email()),
  ),
  kind: z.preprocess(emptyToNull, z.string().min(1).nullable().optional()),
  contractStartDate: z.string().optional().nullable(),
  contractEndDate: z.string().optional().nullable(),
};

const createSchema = z.object({
  ...sharedPersonFields,
  emailAddress: z.string().email(),
  role: z.enum(roles),
});

const updateSchema = z.object(sharedPersonFields);

type PersonFormValues = z.infer<typeof createSchema>;

export function PersonForm(props: {
  id?: string;
  connectionId?: DataID;
  disabled?: boolean;
  scimManaged?: boolean;
  defaultValues?: PersonFormValues;
  onSubmit?: () => void;
}) {
  const {
    id,
    connectionId = "",
    disabled = false,
    scimManaged = false,
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
  const { t } = useTranslation();
  const { role } = use(CurrentUser);
  const availableRoles = getAssignableRoles(role);

  // The edit path validates against updateSchema (no emailAddress/role, which
  // the update mutation ignores and whose inputs are not rendered), while the
  // form is typed against createSchema so role/email stay available to the
  // create branch. The runtime resolver is what fixes the reported bug: the
  // edit form no longer fails validation on two fields it never shows.
  const { control, formState, handleSubmit: handleSubmitWrapper, register, reset }
    = useFormWithSchema(
      (id ? updateSchema : createSchema) as typeof createSchema,
      { defaultValues },
    );
  const watchedRole = useWatch({
    control,
    name: "role",
    defaultValue: "EMPLOYEE",
  });
  const [createPerson, isCreating] = useMutationWithToasts<PersonForm_createMutation>(
    createPersonMutation,
    {
      successMessage: t("personForm.messages.created"),
      errorMessage: t("personForm.errors.create"),
    },
  );
  const [updatePerson, isUpdating] = useMutationWithToasts<PersonForm_updateMutation>(
    updatePersonMutation,
    {
      successMessage: t("personForm.messages.updated"),
      errorMessage: t("personForm.errors.update"),
    },
  );
  const handleSubmit = handleSubmitWrapper(async (data) => {
    const sharedInput = {
      fullName: data.fullName,
      additionalEmailAddresses: data.additionalEmailAddresses,
      kind: data.kind,
      position: data.position,
      contractStartDate: formatDatetime(data.contractStartDate) ?? null,
      contractEndDate: formatDatetime(data.contractEndDate) ?? null,
    };

    if (id) {
      await updatePerson({
        variables: { input: { ...sharedInput, id } },
        onCompleted: () => {
          reset(data);
          onSubmit?.();
        },
      });
    } else {
      await createPerson({
        variables: {
          input: {
            ...sharedInput,
            emailAddress: data.emailAddress,
            role: data.role,
            organizationId,
          },
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
      <Field
        label={t("personForm.fields.fullName")}
        {...register("fullName")}
        type="text"
        disabled={disabled}
        readOnly={scimManaged}
      />
      {!id && (
        <>
          <Field label={t("personForm.fields.emailAddress")} {...register("emailAddress")} type="email" disabled={disabled} />
          <ControlledField
            control={control}
            name="role"
            type="select"
            label={t("personForm.fields.role")}
            disabled={disabled}
          >
            {getMembershipRoles(t)
              .filter(({ value }) => availableRoles.includes(value))
              .map(({ value, label }) => (
                <Option key={value} value={value}>
                  {label}
                </Option>
              ))}
          </ControlledField>

          <div className="mt-4 space-y-2 text-sm text-txt-tertiary">
            {watchedRole === "OWNER" && (
              <p>{t("personForm.roleDescriptions.owner")}</p>
            )}
            {watchedRole === "ADMIN" && (
              <p>
                {t("personForm.roleDescriptions.admin")}
              </p>
            )}
            {watchedRole === "VIEWER" && <p>{t("personForm.roleDescriptions.viewer")}</p>}
            {watchedRole === "AUDITOR" && (
              <p>
                {t("personForm.roleDescriptions.auditor")}
              </p>
            )}
            {watchedRole === "EMPLOYEE" && (
              <p>{t("personForm.roleDescriptions.employee")}</p>
            )}
            {watchedRole === "COMPLIANCE_PORTAL_MANAGER" && (
              <p>{t("personForm.roleDescriptions.compliancePortalManager")}</p>
            )}
            {watchedRole === "COMPLIANCE_PORTAL_ACCESS_MANAGER" && (
              <p>{t("personForm.roleDescriptions.compliancePortalAccessManager")}</p>
            )}
          </div>
        </>
      )}
      <Field
        {...register("kind")}
        type="text"
        label={t("personForm.fields.type")}
        disabled={disabled}
        readOnly={scimManaged}
      />
      <Field
        label={t("personForm.fields.position")}
        {...register("position")}
        type="text"
        placeholder={t("personForm.fields.positionPlaceholder")}
        disabled={disabled}
        readOnly={scimManaged}
      />
      <EmailsField
        control={control}
        register={register}
        disabled={disabled}
        readOnly={scimManaged}
      />
      <Field label={t("personForm.fields.contractStartDate")}>
        <Input
          {...register("contractStartDate")}
          type="date"
          disabled={disabled}
        />
      </Field>
      <Field label={t("personForm.fields.contractEndDate")}>
        <Input
          {...register("contractEndDate")}
          type="date"
          disabled={disabled}
        />
      </Field>
      <div className="flex justify-end">
        {(!id || formState.isDirty) && !disabled && (
          <Button type="submit" disabled={isUpdating || isCreating || !formState.isValid}>
            {id ? t("personForm.actions.update") : t("personForm.actions.create")}
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
      scimManaged={person.source === "SCIM"}
      defaultValues={
        {
          kind: person.kind || null,
          fullName: person.fullName,
          emailAddress: person.emailAddress,
          role: person.membership.role,
          position: person.position || null,
          additionalEmailAddresses: [...person.additionalEmailAddresses],
          contractStartDate: person.contractStartDate?.split("T")[0] || "",
          contractEndDate: person.contractEndDate?.split("T")[0] || "",
        }
      }
    />
  );
}
