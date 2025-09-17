import { z } from "zod";
import { useFormWithSchema } from "../useFormWithSchema";
import { graphql } from "relay-runtime";
import { useFragment } from "react-relay";
import type { useVendorFormFragment$key } from "./__generated__/useVendorFormFragment.graphql";
import { useMutationWithToasts } from "../useMutationWithToasts";
import { useTranslate } from "@probo/i18n";
import { useEffect, useMemo } from "react";

const schema = z.object({
  name: z.string(),
  description: z.string(),
  category: z.string().nullish(),
  statusPageUrl: z.string(),
  termsOfServiceUrl: z.string(),
  privacyPolicyUrl: z.string(),
  serviceLevelAgreementUrl: z.string(),
  dataProcessingAgreementUrl: z.string(),
  websiteUrl: z.string(),
  legalName: z.string(),
  headquarterAddress: z.string(),
  certifications: z.array(z.string()),
  countries: z.array(z.string()),
  securityPageUrl: z.string(),
  trustPageUrl: z.string(),
  businessOwnerId: z.string().nullish(),
  securityOwnerId: z.string().nullish(),
});

const vendorFormFragment = graphql`
  fragment useVendorFormFragment on Vendor {
    id
    name
    description
    category
    statusPageUrl
    termsOfServiceUrl
    privacyPolicyUrl
    serviceLevelAgreementUrl
    dataProcessingAgreementUrl
    websiteUrl
    legalName
    headquarterAddress
    certifications
    countries
    securityPageUrl
    trustPageUrl
    businessOwner {
      id
    }
    securityOwner {
      id
    }
  }
`;

const vendorUpdateQuery = graphql`
  mutation useVendorFormMutation($input: UpdateVendorInput!) {
    updateVendor(input: $input) {
      vendor {
        ...useVendorFormFragment
      }
    }
  }
`;

export function useVendorForm(vendorKey: useVendorFormFragment$key) {
  const vendor = useFragment(vendorFormFragment, vendorKey);
  const { __ } = useTranslate();

  const [mutate] = useMutationWithToasts(vendorUpdateQuery, {
    successMessage: __("Vendor updated successfully."),
    errorMessage: __("Failed to update vendor"),
  });

  const defaultValues = useMemo(
    () => ({
      name: vendor.name,
      description: vendor.description ?? "",
      category: vendor.category ?? null,
      statusPageUrl: vendor.statusPageUrl ?? "",
      termsOfServiceUrl: vendor.termsOfServiceUrl ?? "",
      privacyPolicyUrl: vendor.privacyPolicyUrl ?? "",
      serviceLevelAgreementUrl: vendor.serviceLevelAgreementUrl ?? "",
      dataProcessingAgreementUrl: vendor.dataProcessingAgreementUrl ?? "",
      websiteUrl: vendor.websiteUrl ?? "",
      legalName: vendor.legalName ?? "",
      headquarterAddress: vendor.headquarterAddress ?? "",
      certifications: [...(vendor.certifications ?? [])],
      countries: [...(vendor.countries ?? [])],
      securityPageUrl: vendor.securityPageUrl ?? "",
      trustPageUrl: vendor.trustPageUrl ?? "",
      businessOwnerId: vendor.businessOwner?.id,
      securityOwnerId: vendor.securityOwner?.id,
    }),
    [vendor],
  );

  const form = useFormWithSchema(schema, {
    defaultValues,
  });

  const handleSubmit = form.handleSubmit((data) => {
    return mutate({
      variables: {
        input: {
          id: vendor.id,
          ...data,
        },
      },
    }).then(() => {
      form.reset(data);
    });
  });

  useEffect(() => {
    form.reset(defaultValues, { keepDirty: true });
  }, [defaultValues, form]);

  return {
    ...form,
    handleSubmit,
  };
}
