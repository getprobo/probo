import { useTranslate } from "@probo/i18n";
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
  __: (key: string) => string,
) => {
  switch (policy) {
    case "OFF":
      return __(
        "Your team members can't use single sign-on and must use their password",
      );
    case "REQUIRED":
      return __("Your team members must use single sign-on to log in");
    case "OPTIONAL":
    default:
      return __(
        "Your team members may use either single sign-on or their password to log in",
      );
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
  const { __ } = useTranslate();

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
            {__("Basic Configuration")}
          </h3>
          <div className="space-y-4">
            <div>
              <Field
                {...form.register("emailDomain")}
                label={__("Email Domain") + " *"}
                placeholder="example.com"
                disabled={isEditing}
                error={form.formState.errors.emailDomain?.message}
              />
              <p className="text-xs text-gray-600 mt-1">
                {isEditing
                  ? __("Email domain cannot be changed after creation")
                  : __(
                      "The email domain this SAML configuration applies to (e.g., example.com)",
                    )}
              </p>
            </div>
            {isEditing && (
              <div>
                <Label htmlFor="enforcementPolicy">
                  {__("Enforcement Policy") + " *"}
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
                        <Option value="OPTIONAL">{__("Optional")}</Option>
                        <Option value="REQUIRED">{__("Required")}</Option>
                        <Option value="OFF">{__("Off")}</Option>
                      </Select>
                    </div>
                  )}
                />
                {form.watch("enforcementPolicy") && (
                  <p className="text-xs text-gray-600 mt-2">
                    {getEnforcementPolicyLabel(
                      form.watch("enforcementPolicy"),
                      __,
                    )}
                  </p>
                )}
              </div>
            )}
          </div>
        </div>

        <div>
          <h3 className="text-base font-medium mb-4">
            {__("Identity Provider Configuration")}
          </h3>
          <div className="space-y-4">
            <Field
              {...form.register("idpEntityId")}
              label={__("IdP Entity ID") + " *"}
              placeholder="https://idp.example.com/metadata"
              error={form.formState.errors.idpEntityId?.message}
            />
            <Field
              {...form.register("idpSsoUrl")}
              label={__("IdP SSO URL") + " *"}
              placeholder="https://idp.example.com/sso"
              error={form.formState.errors.idpSsoUrl?.message}
            />
            <div>
              <Label htmlFor="idpCertificate">
                {__("IdP X.509 Certificate") + " *"}
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
            {__("Attribute Mapping")}
          </h3>
          <div className="space-y-4">
            <Field
              {...form.register("attributeMappings.email")}
              label={__("Email Attribute")}
              placeholder={defaultValues.attributeMappings.email}
              error={form.formState.errors.attributeMappings?.email?.message}
            />
            <Field
              {...form.register("attributeMappings.firstName")}
              label={__("First Name Attribute")}
              placeholder={defaultValues.attributeMappings.firstName}
              error={
                form.formState.errors.attributeMappings?.firstName?.message
              }
            />
            <Field
              {...form.register("attributeMappings.lastName")}
              label={__("Last Name Attribute")}
              placeholder={defaultValues.attributeMappings.lastName}
              error={form.formState.errors.attributeMappings?.lastName?.message}
            />
            <Field
              {...form.register("attributeMappings.role")}
              label={__("Role Attribute")}
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
                  {__("Enable automatic user signup via SAML")}
                </Label>
              </div>
            )}
          />
        </div>
      </DialogContent>
      <DialogFooter>
        <Button type="submit" disabled={disabled}>
          {isEditing ? __("Update Configuration") : __("Create Configuration")}
        </Button>
      </DialogFooter>
    </form>
  );
}
