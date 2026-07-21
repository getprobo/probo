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

import {
  Button,
  Checkbox,
  DialogContent,
  DialogFooter,
  Field,
  Label,
  Option,
  Select,
  Textarea,
} from "@probo/ui";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const defaultValues: SAMLConfigurationFormData = {
  emailDomain: "",
  enforcementPolicy: "OPTIONAL" as const,
  idpEntityId: "",
  idpSsoUrl: "",
  idpCertificate: "",
  attributeMappings: {
    email: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
    firstName:
      "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
    lastName: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
    role: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
  },
  autoSignupEnabled: false,
};

const getEnforcementPolicyLabel = (
  policy: string,
  t: (key: string) => string,
) => {
  switch (policy) {
    case "OFF":
      return t("samlConfigurationForm.enforcementDescriptions.off");
    case "REQUIRED":
      return t("samlConfigurationForm.enforcementDescriptions.required");
    case "OPTIONAL":
    default:
      return t("samlConfigurationForm.enforcementDescriptions.optional");
  }
};

const samlConfigSchema = z.object({
  emailDomain: z
    .string()
    .min(1, "Email domain is required")
    .regex(
      /^[a-z0-9.-]+\.[a-z]{2,}$/i,
      "Must be a valid domain (e.g., example.com)",
    ),
  enforcementPolicy: z.enum(["OFF", "OPTIONAL", "REQUIRED"]),
  spCertificate: z.string().optional(),
  spPrivateKey: z.string().optional(),
  idpEntityId: z.string().min(1, "IdP Entity ID is required"),
  idpSsoUrl: z.string().url("IdP SSO URL must be a valid URL"),
  idpCertificate: z.string().min(1, "IdP Certificate is required"),
  attributeMappings: z.object({
    email: z.string().optional(),
    firstName: z.string().optional(),
    lastName: z.string().optional(),
    role: z.string().optional(),
  }),
  autoSignupEnabled: z.boolean().default(false),
});

export type SAMLConfigurationFormData = z.infer<typeof samlConfigSchema>;

export function SAMLConfigurationForm(props: {
  isEditing?: boolean;
  disabled: boolean;
  initialValues?: SAMLConfigurationFormData;
  onSubmit: (data: SAMLConfigurationFormData) => Promise<void>;
}) {
  const {
    disabled,
    initialValues = defaultValues,
    isEditing,
    onSubmit,
  } = props;
  const { t } = useTranslation();

  const form = useFormWithSchema(samlConfigSchema, {
    defaultValues: initialValues,
  });

  return (
    <form
      onSubmit={(e) => {
        void form.handleSubmit(onSubmit)(e);
        form.reset(form.getValues());
      }}
    >
      <DialogContent padded className="space-y-6">
        <div>
          <h3 className="text-base font-medium mb-4">
            {t("samlConfigurationForm.sections.basic")}
          </h3>
          <div className="space-y-4">
            <div>
              <Field
                {...form.register("emailDomain")}
                label={t("samlConfigurationForm.fields.emailDomain")}
                placeholder="example.com"
                disabled={isEditing}
                error={form.formState.errors.emailDomain?.message}
              />
              <p className="text-xs text-gray-600 mt-1">
                {isEditing
                  ? t("samlConfigurationForm.fields.emailDomainLocked")
                  : t("samlConfigurationForm.fields.emailDomainHelp")}
              </p>
            </div>
            {isEditing && (
              <div>
                <Label htmlFor="enforcementPolicy">
                  {t("samlConfigurationForm.fields.enforcementPolicy")}
                </Label>
                <Controller
                  control={form.control}
                  name="enforcementPolicy"
                  render={({ field }) => (
                    <div className="mt-2">
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                      >
                        <Option value="OPTIONAL">{t("samlConfigurationForm.enforcement.optional")}</Option>
                        <Option value="REQUIRED">{t("samlConfigurationForm.enforcement.required")}</Option>
                        <Option value="OFF">{t("samlConfigurationForm.enforcement.off")}</Option>
                      </Select>
                    </div>
                  )}
                />
                {form.watch("enforcementPolicy") && (
                  <p className="text-xs text-gray-600 mt-2">
                    {getEnforcementPolicyLabel(
                      form.watch("enforcementPolicy"),
                      t,
                    )}
                  </p>
                )}
              </div>
            )}
          </div>
        </div>

        <div>
          <h3 className="text-base font-medium mb-4">
            {t("samlConfigurationForm.sections.identityProvider")}
          </h3>
          <div className="space-y-4">
            <Field
              {...form.register("idpEntityId")}
              label={t("samlConfigurationForm.fields.idpEntityId")}
              placeholder="https://idp.example.com/metadata"
              error={form.formState.errors.idpEntityId?.message}
            />
            <Field
              {...form.register("idpSsoUrl")}
              label={t("samlConfigurationForm.fields.idpSsoUrl")}
              placeholder="https://idp.example.com/sso"
              error={form.formState.errors.idpSsoUrl?.message}
            />
            <div>
              <Label htmlFor="idpCertificate">
                {t("samlConfigurationForm.fields.idpCertificate")}
              </Label>
              <Textarea
                {...form.register("idpCertificate")}
                id="idpCertificate"
                rows={6}
                placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                className="font-mono text-sm"
              />
              {form.formState.errors.idpCertificate && (
                <p className="text-sm text-red-600 mt-1">
                  {form.formState.errors.idpCertificate.message}
                </p>
              )}
            </div>
          </div>
        </div>

        <div>
          <h3 className="text-base font-medium mb-4">
            {t("samlConfigurationForm.sections.attributeMapping")}
          </h3>
          <div className="space-y-4">
            <Field
              {...form.register("attributeMappings.email")}
              label={t("samlConfigurationForm.fields.emailAttribute")}
              placeholder={defaultValues.attributeMappings.email}
              error={form.formState.errors.attributeMappings?.email?.message}
            />
            <Field
              {...form.register("attributeMappings.firstName")}
              label={t("samlConfigurationForm.fields.firstNameAttribute")}
              placeholder={defaultValues.attributeMappings.firstName}
              error={
                form.formState.errors.attributeMappings?.firstName?.message
              }
            />
            <Field
              {...form.register("attributeMappings.lastName")}
              label={t("samlConfigurationForm.fields.lastNameAttribute")}
              placeholder={defaultValues.attributeMappings.lastName}
              error={form.formState.errors.attributeMappings?.lastName?.message}
            />
            <Field
              {...form.register("attributeMappings.role")}
              label={t("samlConfigurationForm.fields.roleAttribute")}
              placeholder={defaultValues.attributeMappings.role}
              error={form.formState.errors.attributeMappings?.role?.message}
            />
          </div>
        </div>

        <div>
          <Controller
            control={form.control}
            name="autoSignupEnabled"
            render={({ field }) => (
              <div className="flex items-center gap-2">
                <Checkbox
                  checked={field.value ?? false}
                  onChange={field.onChange}
                />
                <Label htmlFor="autoSignupEnabled" className="cursor-pointer">
                  {t("samlConfigurationForm.fields.autoSignupEnabled")}
                </Label>
              </div>
            )}
          />
        </div>
      </DialogContent>
      <DialogFooter>
        <Button type="submit" disabled={disabled}>
          {isEditing ? t("samlConfigurationForm.actions.update") : t("samlConfigurationForm.actions.create")}
        </Button>
      </DialogFooter>
    </form>
  );
}
