import { useTranslate } from "@probo/i18n";
import { useVendorForm } from "/hooks/forms/useVendorForm";
import type { useVendorFormFragment$key } from "/hooks/forms/__generated__/useVendorFormFragment.graphql";
import { useOutletContext } from "react-router";
import { Button, Card, Field, Input } from "@probo/ui";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useMemo } from "react";
import { usePageTitle } from "@probo/hooks";

export default function VendorOverviewTab() {
  const { vendor } = useOutletContext<{
    vendor: useVendorFormFragment$key & { name: string };
  }>();
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const {
    control,
    register,
    handleSubmit,
    formState: { errors, isDirty, isSubmitting },
  } = useVendorForm(vendor);

  const urls = useMemo(
    () =>
      [
        { name: "statusPageUrl", label: __("Status page URL") },
        { name: "termsOfServiceUrl", label: __("Terms of service URL") },
        { name: "privacyPolicyUrl", label: __("Privacy document URL") },
        {
          name: "serviceLevelAgreementUrl",
          label: __("Service level agreement URL"),
        },
        {
          name: "dataProcessingAgreementUrl",
          label: __("Data processing agreement URL"),
        },
        { name: "securityPageUrl", label: __("Security page URL") },
        { name: "trustPageUrl", label: __("Trust page URL") },
      ] as const,
    []
  );

  usePageTitle(vendor.name + " - " + __("Overview"));

  return (
    <form onSubmit={handleSubmit} className="space-y-12">
      {/* Vendor Details */}
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Vendor details")}</h2>
        <Card className="space-y-4" padded>
          <Field
            {...register("name")}
            label={__("Name")}
            type="text"
            error={errors.name?.message}
          />
          <Field
            {...register("description")}
            label={__("Description")}
            type="textarea"
            error={errors.description?.message}
          />
          <Field
            {...register("legalName")}
            label={__("Legal name")}
            type="text"
            error={errors.legalName?.message}
          />
          <Field
            {...register("headquarterAddress")}
            label={__("Headquarter address")}
            type="textarea"
            error={errors.headquarterAddress?.message}
          />
          <Field
            {...register("websiteUrl")}
            label={__("Website URL")}
            type="text"
            error={errors.websiteUrl?.message}
          />
        </Card>
      </div>

      {/* Ownership */}
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Ownership details`")}</h2>
        <Card className="space-y-4" padded>
          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="businessOwnerId"
            label={__("Business owner")}
            error={errors.businessOwnerId?.message}
          />
          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="securityOwnerId"
            label={__("Security owner")}
            error={errors.securityOwnerId?.message}
          />
        </Card>
      </div>

      {/* Links */}
      <div className="space-y-4 mb-4">
        <h2 className="text-base font-medium">{__("Links")}</h2>
        <Card className="divide-y divide-border-low">
          {urls.map((url) => (
            <div
              key={url.name}
              className="grid grid-cols-2 items-center divide-x divide-border-low"
            >
              <label
                className="p-4 text-sm font-medium text-txt-secondary"
                htmlFor={url.name}
              >
                {url.label}
              </label>
              <Input
                className="p-4 focus:bg-tertiary-pressed outline-none"
                id={url.name}
                key={url.name}
                {...register(url.name)}
                type="text"
                placeholder="https://..."
                variant="ghost"
              />
            </div>
          ))}
        </Card>
      </div>

      {/* Submit */}
      <div className="flex justify-end">
        {isDirty && (
          <Button type="submit" disabled={isSubmitting}>
            {__("Update vendor")}
          </Button>
        )}
      </div>
    </form>
  );
}
