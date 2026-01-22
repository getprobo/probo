import { useTranslate } from "@probo/i18n";
import { useEffect, useMemo } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { useVendorFormFragment$key } from "#/__generated__/core/useVendorFormFragment.graphql";

import { useFormWithSchema } from "../useFormWithSchema";
import { useMutationWithToasts } from "../useMutationWithToasts";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string().optional().nullable(),
  category: z.string().nullish(),
  statusPageUrl: z.string().optional().nullable(),
  termsOfServiceUrl: z.string().optional().nullable(),
  privacyPolicyUrl: z.string().optional().nullable(),
  serviceLevelAgreementUrl: z.string().optional().nullable(),
  dataProcessingAgreementUrl: z.string().optional().nullable(),
  websiteUrl: z.string().optional().nullable(),
  legalName: z.string().optional().nullable(),
  headquarterAddress: z.string().optional().nullable(),
  certifications: z.array(z.string()),
  countries: z.array(z.string()),
  securityPageUrl: z.string().optional().nullable(),
  trustPageUrl: z.string().optional().nullable(),
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
      description: vendor.description || null,
      category: vendor.category || null,
      statusPageUrl: vendor.statusPageUrl || null,
      termsOfServiceUrl: vendor.termsOfServiceUrl || null,
      privacyPolicyUrl: vendor.privacyPolicyUrl || null,
      serviceLevelAgreementUrl: vendor.serviceLevelAgreementUrl || null,
      dataProcessingAgreementUrl: vendor.dataProcessingAgreementUrl || null,
      websiteUrl: vendor.websiteUrl || null,
      legalName: vendor.legalName || null,
      headquarterAddress: vendor.headquarterAddress || null,
      certifications: [...(vendor.certifications ?? [])],
      countries: [...(vendor.countries ?? [])],
      securityPageUrl: vendor.securityPageUrl || null,
      trustPageUrl: vendor.trustPageUrl || null,
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
          description: data.description || null,
          statusPageUrl: data.statusPageUrl || null,
          termsOfServiceUrl: data.termsOfServiceUrl || null,
          privacyPolicyUrl: data.privacyPolicyUrl || null,
          serviceLevelAgreementUrl: data.serviceLevelAgreementUrl || null,
          dataProcessingAgreementUrl: data.dataProcessingAgreementUrl || null,
          websiteUrl: data.websiteUrl || null,
          securityPageUrl: data.securityPageUrl || null,
          trustPageUrl: data.trustPageUrl || null,
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
