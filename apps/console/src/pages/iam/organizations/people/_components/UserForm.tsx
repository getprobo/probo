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

import { Form } from "@base-ui/react/form";
import { formatDatetime, getAssignableRoles, getMembershipRole, getMembershipRoles, peopleRoles, type Role } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { DateField } from "@probo/ui/src/v2/form/DateField";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { use, useState } from "react";
import { useTranslation } from "react-i18next";

import { CurrentUser } from "#/providers/CurrentUser";

import { emptyToNull } from "../_lib/useUpdateUser";
import { isUsersListKind, isUsersListRole } from "../_lib/useUsersListFilters";
import { newUserPage } from "../variants";

import { UserEmailsField } from "./UserEmailsField";

const roleDescriptionKeys = {
  OWNER: "owner",
  ADMIN: "admin",
  VIEWER: "viewer",
  AUDITOR: "auditor",
  EMPLOYEE: "employee",
  COMPLIANCE_PORTAL_MANAGER: "compliancePortalManager",
  COMPLIANCE_PORTAL_ACCESS_MANAGER: "compliancePortalAccessManager",
} as const;

export interface UserFormValues {
  fullName: string;
  emailAddress: string;
  role: Role;
  kind: string | null;
  position: string | null;
  additionalEmailAddresses: string[];
  contract: {
    start: string | null;
    end: string | null;
  };
}

interface UserFormProps {
  disabled?: boolean;
  errors?: Record<string, string>;
  onErrorsChange?: (errors: Record<string, string>) => void;
  onSubmit: (values: UserFormValues) => void;
}

function defaultRole(assignableRoles: Role[]): Role {
  if (assignableRoles.includes("EMPLOYEE")) {
    return "EMPLOYEE";
  }
  return assignableRoles[0] ?? "EMPLOYEE";
}

export function UserForm({
  disabled = false,
  errors = {},
  onErrorsChange,
  onSubmit,
}: UserFormProps) {
  const { t, i18n } = useTranslation();
  const { role: viewerRole } = use(CurrentUser);
  const assignableRoles = getAssignableRoles(viewerRole);
  const roleOptions = getMembershipRoles(t).filter(({ value }) => (
    assignableRoles.includes(value)
  ));
  const { form, actions } = newUserPage();

  const [fullName, setFullName] = useState("");
  const [emailAddress, setEmailAddress] = useState("");
  const [role, setRole] = useState<Role>(defaultRole(assignableRoles));
  const [kind, setKind] = useState<string | null>("EMPLOYEE");
  const [position, setPosition] = useState("");
  const [emails, setEmails] = useState<string[]>([""]);
  const [contractStart, setContractStart] = useState("");
  const [contractEnd, setContractEnd] = useState("");
  const roleDescriptionKey = roleDescriptionKeys[role];

  function clearError(name: string) {
    if (onErrorsChange == null || errors[name] == null) {
      return;
    }
    const next = { ...errors };
    delete next[name];
    onErrorsChange(next);
  }

  function handleSubmit() {
    onSubmit({
      fullName: fullName.trim(),
      emailAddress: emailAddress.trim(),
      role,
      kind: emptyToNull(kind),
      position: emptyToNull(position),
      additionalEmailAddresses: emails.map(email => email.trim()).filter(email => email !== ""),
      contract: {
        start: formatDatetime(emptyToNull(contractStart)) ?? null,
        end: formatDatetime(emptyToNull(contractEnd)) ?? null,
      },
    });
  }

  return (
    <Form className={form()} errors={errors} onFormSubmit={handleSubmit}>
      <Field label={t("userForm.fields.fullName")} error={errors.fullName}>
        <TextField
          name="fullName"
          required
          value={fullName}
          disabled={disabled}
          onValueChange={(value) => {
            setFullName(value);
            clearError("fullName");
          }}
        />
      </Field>
      <Field label={t("userForm.fields.emailAddress")} error={errors.emailAddress}>
        <TextField
          name="emailAddress"
          type="email"
          required
          value={emailAddress}
          disabled={disabled}
          onValueChange={(value) => {
            setEmailAddress(value);
            clearError("emailAddress");
          }}
        />
      </Field>
      <Field label={t("userForm.fields.role")} error={errors.role}>
        <Select
          value={role}
          disabled={disabled}
          onValueChange={(value) => {
            if (value == null || !isUsersListRole(value)) {
              return;
            }
            setRole(value);
            clearError("role");
          }}
        >
          <SelectTrigger aria-label={t("userForm.fields.role")}>
            {(value: Role | null) => (
              value != null ? getMembershipRole(t, value) : null
            )}
          </SelectTrigger>
          <SelectPopup>
            {roleOptions.map(({ value, label }) => (
              <SelectItem key={value} value={value}>
                {label}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </Field>
      {roleDescriptionKey != null && (
        <Text size={2} color="faint">
          {t(`userForm.roleDescriptions.${roleDescriptionKey}`)}
        </Text>
      )}
      <Field label={t("userForm.fields.type")} error={errors.kind}>
        <Select
          value={kind}
          disabled={disabled}
          onValueChange={(value) => {
            setKind(value);
            clearError("kind");
          }}
        >
          <SelectTrigger aria-label={t("userForm.fields.type")}>
            {(value: string | null) => (
              value != null && isUsersListKind(value)
                ? t(`userForm.kinds.${value}`)
                : value
            )}
          </SelectTrigger>
          <SelectPopup>
            {peopleRoles.map(peopleKind => (
              <SelectItem key={peopleKind} value={peopleKind}>
                {t(`userForm.kinds.${peopleKind}`)}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </Field>
      <Field label={t("userForm.fields.position")} error={errors.position}>
        <TextField
          name="position"
          value={position}
          disabled={disabled}
          placeholder={t("userForm.fields.positionPlaceholder")}
          onValueChange={(value) => {
            setPosition(value);
            clearError("position");
          }}
        />
      </Field>
      <Field label={t("userForm.fields.additionalEmails")} error={errors.additionalEmailAddresses}>
        <UserEmailsField
          value={emails}
          disabled={disabled}
          onValueChange={(value) => {
            setEmails(value);
            clearError("additionalEmailAddresses");
          }}
        />
      </Field>
      <Field label={t("userForm.fields.contractStartDate")} error={errors.contractStart}>
        <DateField
          name="contractStart"
          value={contractStart}
          locale={i18n.language}
          disabled={disabled}
          nullable
          onValueChange={(next) => {
            setContractStart(next);
            clearError("contractStart");
          }}
        />
      </Field>
      <Field label={t("userForm.fields.contractEndDate")} error={errors.contractEnd}>
        <DateField
          name="contractEnd"
          value={contractEnd}
          locale={i18n.language}
          disabled={disabled}
          nullable
          onValueChange={(next) => {
            setContractEnd(next);
            clearError("contractEnd");
          }}
        />
      </Field>
      <div className={actions()}>
        <Button
          type="submit"
          variant="solid"
          color="neutral"
          highContrast
          loading={disabled}
        >
          {t("userForm.actions.create")}
        </Button>
      </div>
    </Form>
  );
}
