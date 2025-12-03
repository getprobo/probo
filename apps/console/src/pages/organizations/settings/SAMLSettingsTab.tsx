import { useState, useEffect, use } from "react";
import { useOutletContext } from "react-router";
import { useFragment, graphql } from "react-relay";
import { Controller } from "react-hook-form";
import { z } from "zod";
import {
  Breadcrumb,
  Button,
  Card,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Label,
  Option,
  Select,
  Table,
  Tbody,
  Td,
  Textarea,
  Th,
  Thead,
  Tr,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import {
  useCreateSAMLConfigurationMutation,
  useUpdateSAMLConfigurationMutation,
  useDeleteSAMLConfigurationMutation,
  useEnableSAMLMutation,
  useDisableSAMLMutation,
  useInitiateDomainVerificationMutation,
  useVerifyDomainMutation,
} from "/hooks/graph/SAMLConfigurationGraph";
import type { SAMLSettingsTabFragment$key } from "./__generated__/SAMLSettingsTabFragment.graphql";
import { PermissionsContext } from "/providers/PermissionsContext";

const samlSettingsTabFragment = graphql`
  fragment SAMLSettingsTabFragment on Organization {
    id
    name
    samlConfigurations {
      id
      enabled
      emailDomain
      enforcementPolicy
      domainVerified
      domainVerificationToken
      domainVerifiedAt
      spEntityId
      spAcsUrl
      spMetadataUrl
      testLoginUrl
      idpEntityId
      idpSsoUrl
      idpCertificate
      idpMetadataUrl
      attributeEmail
      attributeFirstname
      attributeLastname
      attributeRole
      autoSignupEnabled
    }
  }
`;

const initiateSchema = z.object({
  emailDomain: z.string().min(1, "Email domain is required").regex(/^[a-z0-9.-]+\.[a-z]{2,}$/i, "Must be a valid domain (e.g., example.com)"),
});

const samlConfigSchema = z.object({
  emailDomain: z.string().min(1, "Email domain is required").regex(/^[a-z0-9.-]+\.[a-z]{2,}$/i, "Must be a valid domain (e.g., example.com)"),
  enforcementPolicy: z.enum(["OFF", "OPTIONAL", "REQUIRED"]),
  spCertificate: z.string().optional(),
  spPrivateKey: z.string().optional(),
  idpEntityId: z.string().min(1, "IdP Entity ID is required"),
  idpSsoUrl: z.string().url("IdP SSO URL must be a valid URL"),
  idpCertificate: z.string().min(1, "IdP Certificate is required"),
  idpMetadataUrl: z.string().url("IdP Metadata URL must be a valid URL").optional().or(z.literal("")),
  attributeEmail: z.string().optional(),
  attributeFirstname: z.string().optional(),
  attributeLastname: z.string().optional(),
  attributeRole: z.string().optional(),
  autoSignupEnabled: z.boolean().default(false),
});

type OutletContext = {
  organization: SAMLSettingsTabFragment$key;
};

type SetupStep = "initiate" | "verify" | "configure";

export default function SAMLSettingsTab() {
  const { __ } = useTranslate();
  const { organization: organizationKey } = useOutletContext<OutletContext>();
  const { isAuthorized } = use(PermissionsContext);
  const organization = useFragment(samlSettingsTabFragment, organizationKey);
  const configs = organization.samlConfigurations;

  const dialogRef = useDialogRef();
  const [editingConfig, setEditingConfig] = useState<Partial<typeof configs[0]> & {id: string} | null>(null);
  const [currentStep, setCurrentStep] = useState<SetupStep>("initiate");
  const [dnsRecord, setDnsRecord] = useState<string>("");

  const [createMutation, isCreating] = useCreateSAMLConfigurationMutation();
  const [updateMutation, isUpdating] = useUpdateSAMLConfigurationMutation();
  const [deleteMutation] = useDeleteSAMLConfigurationMutation();
  const [enableMutation, isEnabling] = useEnableSAMLMutation();
  const [disableMutation, isDisabling] = useDisableSAMLMutation();
  const [initiateDomainMutation, isInitiating] = useInitiateDomainVerificationMutation();
  const [verifyDomainMutation, isVerifying] = useVerifyDomainMutation();

  const confirm = useConfirm();

  const handleOpenModal = (config?: typeof configs[0]) => {
    setEditingConfig(config || null);
    if (config) {
      if (!config.domainVerified) {
        setCurrentStep("verify");
        setDnsRecord(`probo-verification=${config.domainVerificationToken}`);
      } else {
        setCurrentStep("configure");
      }
    } else {
      setCurrentStep("initiate");
    }
    dialogRef.current?.open();
  };

  const handleCloseModal = () => {
    setEditingConfig(null);
    setCurrentStep("initiate");
    setDnsRecord("");
    dialogRef.current?.close();
  };

  const initiateForm = useFormWithSchema(initiateSchema, {
    defaultValues: {
      emailDomain: editingConfig?.emailDomain || "",
    },
  });

  const form = useFormWithSchema(samlConfigSchema, {
    defaultValues: editingConfig
      ? {
          emailDomain: editingConfig.emailDomain || "",
          enforcementPolicy: editingConfig.enforcementPolicy || "OPTIONAL",
          idpEntityId: editingConfig.idpEntityId || "",
          idpSsoUrl: editingConfig.idpSsoUrl || "",
          idpCertificate: editingConfig.idpCertificate || "",
          idpMetadataUrl: editingConfig.idpMetadataUrl || "",
          attributeEmail: editingConfig.attributeEmail || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
          attributeFirstname: editingConfig.attributeFirstname || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
          attributeLastname: editingConfig.attributeLastname || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
          attributeRole: editingConfig.attributeRole || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
          autoSignupEnabled: editingConfig.autoSignupEnabled || false,
        }
      : {
          emailDomain: "",
          enforcementPolicy: "OPTIONAL",
          idpEntityId: "",
          idpSsoUrl: "",
          idpCertificate: "",
          idpMetadataUrl: "",
          attributeEmail: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
          attributeFirstname: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
          attributeLastname: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
          attributeRole: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
          autoSignupEnabled: false,
        },
  });

  useEffect(() => {
    if (editingConfig) {
      form.reset({
        emailDomain: editingConfig.emailDomain || "",
        enforcementPolicy: editingConfig.enforcementPolicy || "OPTIONAL",
        idpEntityId: editingConfig.idpEntityId || "",
        idpSsoUrl: editingConfig.idpSsoUrl || "",
        idpCertificate: editingConfig.idpCertificate || "",
        idpMetadataUrl: editingConfig.idpMetadataUrl || "",
        attributeEmail: editingConfig.attributeEmail || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
        attributeFirstname: editingConfig.attributeFirstname || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
        attributeLastname: editingConfig.attributeLastname || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
        attributeRole: editingConfig.attributeRole || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
        autoSignupEnabled: editingConfig.autoSignupEnabled || false,
      });
      initiateForm.reset({
        emailDomain: editingConfig.emailDomain || "",
      });
    } else {
      form.reset({
        emailDomain: "",
        enforcementPolicy: "OPTIONAL",
        idpEntityId: "",
        idpSsoUrl: "",
        idpCertificate: "",
        idpMetadataUrl: "",
        attributeEmail: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
        attributeFirstname: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
        attributeLastname: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
        attributeRole: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
        autoSignupEnabled: false,
      });
      initiateForm.reset({
        emailDomain: "",
      });
    }
  }, [editingConfig, form, initiateForm]);

  const handleInitiateDomain = initiateForm.handleSubmit((data) => {
    initiateDomainMutation({
      variables: {
        input: {
          organizationId: organization.id,
          emailDomain: data.emailDomain,
        },
      },
      onCompleted: (response) => {
        setDnsRecord(response.initiateDomainVerification.dnsRecord);
        setEditingConfig(response.initiateDomainVerification.samlConfiguration);
        setCurrentStep("verify");
      },
    });
  });

  const handleVerifyDomain = () => {
    if (!editingConfig) return;
    verifyDomainMutation({
      variables: {
        input: {
          id: editingConfig.id,
        },
      },
      onCompleted: (response) => {
        if (response.verifyDomain.verified) {
          setCurrentStep("configure");
        }
      },
    });
  };

  const onSubmit = form.handleSubmit((data) => {
    if (editingConfig) {
      updateMutation({
        variables: {
          input: {
            id: editingConfig.id,
            enforcementPolicy: data.enforcementPolicy,
            idpEntityId: data.idpEntityId,
            idpSsoUrl: data.idpSsoUrl,
            idpCertificate: data.idpCertificate,
            idpMetadataUrl: data.idpMetadataUrl || null,
            attributeEmail: data.attributeEmail || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
            attributeFirstname: data.attributeFirstname || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
            attributeLastname: data.attributeLastname || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
            attributeRole: data.attributeRole || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
            autoSignupEnabled: data.autoSignupEnabled || false,
          },
        },
        onCompleted: () => {
          handleCloseModal();
        },
      });
    } else {
      createMutation({
        variables: {
          input: {
            organizationId: organization.id,
            emailDomain: data.emailDomain,
            enforcementPolicy: data.enforcementPolicy,
            idpEntityId: data.idpEntityId,
            idpSsoUrl: data.idpSsoUrl,
            idpCertificate: data.idpCertificate,
            idpMetadataUrl: data.idpMetadataUrl || null,
            attributeEmail: data.attributeEmail || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
            attributeFirstname: data.attributeFirstname || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
            attributeLastname: data.attributeLastname || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
            attributeRole: data.attributeRole || "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
            autoSignupEnabled: data.autoSignupEnabled || false,
          },
        },
        onCompleted: () => {
          handleCloseModal();
        },
      });
    }
  });

  const handleToggleEnabled = (config: typeof configs[0]) => {
    if (config.enabled) {
      confirm(
        async () => {
          disableMutation({
            variables: {
              input: {
                id: config.id,
              },
            },
          });
        },
        {
          title: __("Disable SAML"),
          message: __(
            "Are you sure you want to disable SAML authentication for " + config.emailDomain + "?"
          ),
          label: __("Disable"),
          variant: "danger",
        }
      );
    } else {
      enableMutation({
        variables: {
          input: {
            id: config.id,
          },
        },
      });
    }
  };

  const handleDelete = (config: typeof configs[0]) => {
    confirm(
      async () => {
        deleteMutation({
          variables: {
            input: {
              id: config.id,
            },
          },
          onCompleted: () => {
            handleCloseModal();
          },
        });
      },
      {
        title: __("Delete SAML Configuration"),
        message: __(
          "Are you sure you want to delete the SAML configuration for " + config.emailDomain + "? This action cannot be undone."
        ),
        label: __("Delete"),
        variant: "danger",
      }
    );
  };

  const [copiedUrl, setCopiedUrl] = useState<string | null>(null);

  const handleCopy = (url: string) => {
    navigator.clipboard.writeText(url);
    setCopiedUrl(url);
    setTimeout(() => setCopiedUrl(null), 2000);
  };

  const getEnforcementPolicyLabel = (policy: string) => {
    switch (policy) {
      case "OFF":
        return __("Your team members can't use single sign-on and must use their password");
      case "REQUIRED":
        return __("Your team members must use single sign-on to log in");
      case "OPTIONAL":
      default:
        return __("Your team members may use either single sign-on or their password to log in");
    }
  };

  return (
    <>
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <h2 className="text-base font-medium">{__("SAML Single Sign-On")}</h2>
          {isAuthorized("Organization", "createSAMLConfiguration") && (
            <Button onClick={() => handleOpenModal()}>
              {__("Add Configuration")}
            </Button>
          )}
        </div>

        {configs.length === 0 ? (
          <Card padded>
            <div className="text-center py-12">
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                {__("No SAML Configurations")}
              </h3>
              <p className="text-gray-600 mb-6">
                {__("Set up SAML 2.0 single sign-on for your organization by adding a configuration for each email domain.")}
              </p>
              {isAuthorized("Organization", "createSAMLConfiguration") && (
                <Button onClick={() => handleOpenModal()}>
                  {__("Add Your First Configuration")}
                </Button>
              )}
            </div>
          </Card>
        ) : (
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Email Domain")}</Th>
                <Th>{__("Domain Status")}</Th>
                <Th>{__("SAML Status")}</Th>
                <Th>{__("Enforcement")}</Th>
                <Th>{__("SSO URL")}</Th>
                <Th></Th>
              </Tr>
            </Thead>
            <Tbody>
              {configs.map((config) => (
                <Tr key={config.id}>
                  <Td>
                    <button
                      onClick={() => handleOpenModal(config)}
                      className="font-semibold text-blue-600 hover:text-blue-800"
                    >
                      {config.emailDomain}
                    </button>
                  </Td>
                  <Td>
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        config.domainVerified
                          ? "bg-green-100 text-green-800"
                          : "bg-yellow-100 text-yellow-800"
                      }`}
                    >
                      {config.domainVerified ? __("Verified") : __("Pending Verification")}
                    </span>
                  </Td>
                  <Td>
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        config.enabled
                          ? "bg-green-100 text-green-800"
                          : "bg-gray-100 text-gray-800"
                      }`}
                    >
                      {config.enabled ? __("Enabled") : __("Disabled")}
                    </span>
                  </Td>
                  <Td>{config.enforcementPolicy}</Td>
                  <Td>
                    {config.domainVerified && config.enabled ? (
                      <button
                        onClick={() => handleCopy(config.testLoginUrl)}
                        className="text-blue-600 hover:text-blue-800"
                      >
                        {copiedUrl === config.testLoginUrl ? __("Copied!") : __("Copy URL")}
                      </button>
                    ) : (
                      <span className="text-gray-400">—</span>
                    )}
                  </Td>
                  <Td width={180} className="text-end">
                    <div className="flex gap-2 justify-end">
                      {config.domainVerified ? (
                        <>
                          {isAuthorized("SAMLConfiguration", "updateSAMLConfiguration") && (
                            <Button
                              variant={config.enabled ? "danger" : "primary"}
                              onClick={() => handleToggleEnabled(config)}
                              disabled={isEnabling || isDisabling}
                            >
                              {config.enabled ? __("Disable") : __("Enable")}
                            </Button>
                          )}
                          {isAuthorized("SAMLConfiguration", "updateSAMLConfiguration") && (
                            <Button
                              variant="secondary"
                              onClick={() => handleOpenModal(config)}
                            >
                              {__("Edit")}
                            </Button>
                          )}
                        </>
                      ) : (
                        <>
                          {isAuthorized("Organization", "verifyDomain") && (
                            <Button
                              variant="primary"
                              onClick={() => handleOpenModal(config)}
                            >
                              {__("Verify Domain")}
                            </Button>
                          )}
                          {isAuthorized("Organization", "deleteOrganization") && (
                            <Button
                              variant="danger"
                              onClick={() => handleDelete(config)}
                            >
                              {__("Delete")}
                            </Button>
                          )}
                        </>
                      )}
                    </div>
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        )}
      </div>

      <Dialog
        ref={dialogRef}
        onClose={handleCloseModal}
        title={
          <Breadcrumb
            items={[
              __("SAML Settings"),
              currentStep === "initiate" && __("Register Domain"),
              currentStep === "verify" && __("Verify Domain"),
              currentStep === "configure" && (editingConfig?.domainVerified ? __("Configure SAML") : __("Configure SAML")),
            ].filter(Boolean) as string[]}
          />
        }
      >
        {currentStep === "initiate" && (
          <form onSubmit={handleInitiateDomain}>
            <DialogContent padded className="space-y-6">
              <div>
                <h3 className="text-base font-medium mb-4">
                  {__("Register Your Domain")}
                </h3>
                <p className="text-sm text-gray-600 mb-4">
                  {__("To set up SAML SSO, you must first register and verify ownership of your email domain.")}
                </p>
                <div className="space-y-4">
                  <Field
                    {...initiateForm.register("emailDomain")}
                    label={__("Email Domain") + " *"}
                    placeholder="example.com"
                    error={initiateForm.formState.errors.emailDomain?.message}
                  />
                  <p className="text-xs text-gray-600">
                    {__("The email domain this SAML configuration will apply to (e.g., example.com)")}
                  </p>
                </div>
              </div>
            </DialogContent>
            <DialogFooter>
              <Button type="submit" disabled={isInitiating}>
                {__("Next: Verify Domain")}
              </Button>
            </DialogFooter>
          </form>
        )}

        {currentStep === "verify" && (
          <>
            <DialogContent padded className="space-y-6">
              <div>
                <h3 className="text-base font-medium mb-4">
                  {__("Verify Domain Ownership")}
                </h3>
                <p className="text-sm text-gray-600 mb-4">
                  {__("Add the following TXT record to your domain's DNS configuration to verify ownership:")}
                </p>
                <div className="bg-gray-50 rounded-lg p-4 mb-4">
                  <div className="space-y-2">
                    <div>
                      <span className="font-semibold text-sm">{__("Host/Name:")}</span>
                      <code className="ml-2 bg-white px-2 py-1 rounded text-sm">@</code>
                      <span className="ml-2 text-xs text-gray-600">{__("or use your domain name")}</span>
                    </div>
                    <div>
                      <span className="font-semibold text-sm">{__("Type:")}</span>
                      <code className="ml-2 bg-white px-2 py-1 rounded text-sm">TXT</code>
                    </div>
                    <div>
                      <span className="font-semibold text-sm">{__("Value:")}</span>
                      <div className="mt-1 flex items-center gap-2">
                        <code className="flex-1 bg-white px-2 py-1 rounded text-sm break-all font-mono">
                          {dnsRecord}
                        </code>
                        <Button
                          type="button"
                          variant="secondary"
                          onClick={() => handleCopy(dnsRecord)}
                        >
                          {copiedUrl === dnsRecord ? __("Copied!") : __("Copy")}
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <p className="text-sm text-blue-800">
                    <strong>{__("Note:")}</strong> {__("DNS changes may take up to 48 hours to propagate, but typically complete within a few minutes.")}
                  </p>
                </div>
              </div>
            </DialogContent>
            <DialogFooter>
              <Button onClick={handleVerifyDomain} disabled={isVerifying}>
                {__("Verify and Continue")}
              </Button>
            </DialogFooter>
          </>
        )}

        {currentStep === "configure" && (
          <form onSubmit={onSubmit}>
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
                    disabled={!!editingConfig}
                    error={form.formState.errors.emailDomain?.message}
                  />
                  <p className="text-xs text-gray-600 mt-1">
                    {editingConfig
                      ? __("Email domain cannot be changed after creation")
                      : __("The email domain this SAML configuration applies to (e.g., example.com)")}
                  </p>
                </div>

                <div>
                  <Label htmlFor="enforcementPolicy">{__("Enforcement Policy") + " *"}</Label>
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
                      {getEnforcementPolicyLabel(form.watch("enforcementPolicy"))}
                    </p>
                  )}
                </div>
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
                <Field
                  {...form.register("idpMetadataUrl")}
                  label={__("IdP Metadata URL (Optional)")}
                  placeholder="https://idp.example.com/metadata.xml"
                  error={form.formState.errors.idpMetadataUrl?.message}
                />
              </div>
            </div>

            <div>
              <h3 className="text-base font-medium mb-4">
                {__("Attribute Mapping")}
              </h3>
              <div className="space-y-4">
                <Field
                  {...form.register("attributeEmail")}
                  label={__("Email Attribute")}
                  placeholder="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
                  error={form.formState.errors.attributeEmail?.message}
                />
                <Field
                  {...form.register("attributeFirstname")}
                  label={__("First Name Attribute")}
                  placeholder="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
                  error={form.formState.errors.attributeFirstname?.message}
                />
                <Field
                  {...form.register("attributeLastname")}
                  label={__("Last Name Attribute")}
                  placeholder="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
                  error={form.formState.errors.attributeLastname?.message}
                />
                <Field
                  {...form.register("attributeRole")}
                  label={__("Role Attribute")}
                  placeholder="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role"
                  error={form.formState.errors.attributeRole?.message}
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
              <Button type="submit" disabled={isCreating || isUpdating}>
                {editingConfig?.domainVerified ? __("Update Configuration") : __("Create Configuration")}
              </Button>
            </DialogFooter>
          </form>
        )}
      </Dialog>
    </>
  );
}
