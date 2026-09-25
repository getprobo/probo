// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import { DialogContent, DialogFooter } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";

import { newSamlSsoPage } from "../variants";

const EMAIL_DOMAIN_PATTERN = "^[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$";

const defaultAttributeMappings = {
  email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
  firstName: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
  lastName: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
  role: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
};

export type SAMLEnforcementPolicy = "OFF" | "OPTIONAL" | "REQUIRED";

export interface SAMLConfigurationFormData {
  emailDomain: string;
  enforcementPolicy: SAMLEnforcementPolicy;
  idpEntityId: string;
  idpSsoUrl: string;
  idpCertificate: string;
  attributeMappings: {
    email?: string;
    firstName?: string;
    lastName?: string;
    role?: string;
  };
  autoSignupEnabled: boolean;
}

const defaultValues: SAMLConfigurationFormData = {
  emailDomain: "",
  enforcementPolicy: "OPTIONAL",
  idpEntityId: "",
  idpSsoUrl: "",
  idpCertificate: "",
  attributeMappings: defaultAttributeMappings,
  autoSignupEnabled: false,
};

function isEnforcementPolicy(value: unknown): value is SAMLEnforcementPolicy {
  return value === "OFF" || value === "OPTIONAL" || value === "REQUIRED";
}

function emptyToUndefined(value: string): string | undefined {
  const trimmed = value.trim();
  return trimmed === "" ? undefined : trimmed;
}

interface SAMLConfigurationFormProps {
  isEditing?: boolean;
  disabled: boolean;
  variant?: "dialog" | "page";
  initialValues?: SAMLConfigurationFormData;
  onSubmit: (data: SAMLConfigurationFormData) => void | Promise<void>;
}

export function SAMLConfigurationForm({
  isEditing = false,
  disabled,
  variant = "dialog",
  initialValues = defaultValues,
  onSubmit,
}: SAMLConfigurationFormProps) {
  const { t } = useTranslation();
  const { form, section, fields, check, actions } = newSamlSsoPage();

  const [emailDomain, setEmailDomain] = useState(initialValues.emailDomain);
  const [enforcementPolicy, setEnforcementPolicy] = useState<SAMLEnforcementPolicy>(
    initialValues.enforcementPolicy,
  );
  const [idpEntityId, setIdpEntityId] = useState(initialValues.idpEntityId);
  const [idpSsoUrl, setIdpSsoUrl] = useState(initialValues.idpSsoUrl);
  const [idpCertificate, setIdpCertificate] = useState(initialValues.idpCertificate);
  const [emailAttribute, setEmailAttribute] = useState(
    initialValues.attributeMappings.email ?? defaultAttributeMappings.email,
  );
  const [firstNameAttribute, setFirstNameAttribute] = useState(
    initialValues.attributeMappings.firstName ?? defaultAttributeMappings.firstName,
  );
  const [lastNameAttribute, setLastNameAttribute] = useState(
    initialValues.attributeMappings.lastName ?? defaultAttributeMappings.lastName,
  );
  const [roleAttribute, setRoleAttribute] = useState(
    initialValues.attributeMappings.role ?? defaultAttributeMappings.role,
  );
  const [autoSignupEnabled, setAutoSignupEnabled] = useState(
    initialValues.autoSignupEnabled,
  );

  function handleSubmit() {
    void onSubmit({
      emailDomain: emailDomain.trim(),
      enforcementPolicy,
      idpEntityId: idpEntityId.trim(),
      idpSsoUrl: idpSsoUrl.trim(),
      idpCertificate: idpCertificate.trim(),
      attributeMappings: {
        email: emptyToUndefined(emailAttribute),
        firstName: emptyToUndefined(firstNameAttribute),
        lastName: emptyToUndefined(lastNameAttribute),
        role: emptyToUndefined(roleAttribute),
      },
      autoSignupEnabled,
    });
  }

  const submit = (
    <Button
      type="submit"
      variant="solid"
      color="neutral"
      highContrast
      loading={disabled}
    >
      {isEditing
        ? t("samlConfigurationForm.actions.update")
        : t("samlSsoPage.actions.add")}
    </Button>
  );

  function sectionBlock(title: string | null, children: ReactNode) {
    const content = <div className={fields()}>{children}</div>;
    const card = variant === "page"
      ? <Card variant="soft" size={2}>{content}</Card>
      : content;

    if (title == null) {
      return card;
    }

    return (
      <section className={section()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {title}
        </Heading>
        {card}
      </section>
    );
  }

  const body = (
    <>
      {sectionBlock(null, (
        <>
          <Field required label={t("samlConfigurationForm.fields.emailDomain")}>
            <TextField
              name="emailDomain"
              required
              pattern={EMAIL_DOMAIN_PATTERN}
              size={2}
              value={emailDomain}
              disabled={disabled || isEditing}
              placeholder="example.com"
              onValueChange={setEmailDomain}
            />
          </Field>
          <Text size={1} color="faint">
            {isEditing
              ? t("samlConfigurationForm.fields.emailDomainLocked")
              : t("samlConfigurationForm.fields.emailDomainHelp")}
          </Text>
          <Field required label={t("samlConfigurationForm.fields.enforcementPolicy")}>
            <Select
              value={enforcementPolicy}
              disabled={disabled}
              onValueChange={(value) => {
                if (isEnforcementPolicy(value)) {
                  setEnforcementPolicy(value);
                }
              }}
            >
              <SelectTrigger
                size={2}
                aria-label={t("samlConfigurationForm.fields.enforcementPolicy")}
              >
                {(value: SAMLEnforcementPolicy | null) => (
                  value != null
                    ? t(`samlConfigurationForm.enforcement.${value.toLowerCase()}`)
                    : null
                )}
              </SelectTrigger>
              <SelectPopup>
                <SelectItem value="OPTIONAL">
                  {t("samlConfigurationForm.enforcement.optional")}
                </SelectItem>
                <SelectItem value="REQUIRED">
                  {t("samlConfigurationForm.enforcement.required")}
                </SelectItem>
                {isEditing && (
                  <SelectItem value="OFF">
                    {t("samlConfigurationForm.enforcement.off")}
                  </SelectItem>
                )}
              </SelectPopup>
            </Select>
          </Field>
          <Text size={1} color="faint">
            {t(`samlConfigurationForm.enforcementDescriptions.${enforcementPolicy.toLowerCase()}`)}
          </Text>
          <label className={check()}>
            <Checkbox
              checked={autoSignupEnabled}
              disabled={disabled}
              onCheckedChange={(checked) => {
                setAutoSignupEnabled(checked === true);
              }}
            />
            <Text size={2}>
              {t("samlConfigurationForm.fields.autoSignupEnabled")}
            </Text>
          </label>
        </>
      ))}
      {sectionBlock(t("samlConfigurationForm.sections.identityProvider"), (
        <>
          <Field required label={t("samlConfigurationForm.fields.idpEntityId")}>
            <TextField
              name="idpEntityId"
              required
              size={2}
              value={idpEntityId}
              disabled={disabled}
              placeholder="https://idp.example.com/metadata"
              onValueChange={setIdpEntityId}
            />
          </Field>
          <Field required label={t("samlConfigurationForm.fields.idpSsoUrl")}>
            <TextField
              name="idpSsoUrl"
              type="url"
              required
              size={2}
              value={idpSsoUrl}
              disabled={disabled}
              placeholder="https://idp.example.com/sso"
              onValueChange={setIdpSsoUrl}
            />
          </Field>
          <Field required label={t("samlConfigurationForm.fields.idpCertificate")}>
            <Textarea
              name="idpCertificate"
              required
              rows={6}
              value={idpCertificate}
              disabled={disabled}
              placeholder={"-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"}
              className="font-mono"
              onChange={(event) => {
                setIdpCertificate(event.target.value);
              }}
            />
          </Field>
        </>
      ))}
      {sectionBlock(t("samlConfigurationForm.sections.attributeMapping"), (
        <>
          <Field label={t("samlConfigurationForm.fields.emailAttribute")}>
            <TextField
              name="emailAttribute"
              size={2}
              value={emailAttribute}
              disabled={disabled}
              placeholder={defaultAttributeMappings.email}
              onValueChange={setEmailAttribute}
            />
          </Field>
          <Field label={t("samlConfigurationForm.fields.firstNameAttribute")}>
            <TextField
              name="firstNameAttribute"
              size={2}
              value={firstNameAttribute}
              disabled={disabled}
              placeholder={defaultAttributeMappings.firstName}
              onValueChange={setFirstNameAttribute}
            />
          </Field>
          <Field label={t("samlConfigurationForm.fields.lastNameAttribute")}>
            <TextField
              name="lastNameAttribute"
              size={2}
              value={lastNameAttribute}
              disabled={disabled}
              placeholder={defaultAttributeMappings.lastName}
              onValueChange={setLastNameAttribute}
            />
          </Field>
          <Field label={t("samlConfigurationForm.fields.roleAttribute")}>
            <TextField
              name="roleAttribute"
              size={2}
              value={roleAttribute}
              disabled={disabled}
              placeholder={defaultAttributeMappings.role}
              onValueChange={setRoleAttribute}
            />
          </Field>
        </>
      ))}
    </>
  );

  function wrapFields(content: ReactNode) {
    if (variant === "page") {
      return (
        <>
          {content}
          <div className={actions()}>{submit}</div>
        </>
      );
    }

    return (
      <>
        <DialogContent padded>
          <div className={form()}>{content}</div>
        </DialogContent>
        <DialogFooter>{submit}</DialogFooter>
      </>
    );
  }

  return (
    <Form className={form()} onFormSubmit={handleSubmit}>
      {wrapFields(body)}
    </Form>
  );
}
