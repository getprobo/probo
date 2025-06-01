import { z } from "zod";
import { useFormWithSchema } from "../useFormWithSchema";
import { graphql } from "relay-runtime";
import { useFragment, useMutation } from "react-relay";
import type { useVendorFormFragment$key } from "./__generated__/useVendorFormFragment.graphql";
import { useMutationWithToasts } from "../useMutationWithToasts";
import { useTranslate } from "@probo/i18n";

const schema = z.object({
  name: z.string(),
  description: z.string(),
  statusPageUrl: z.string(),
  termsOfServiceUrl: z.string(),
  privacyPolicyUrl: z.string(),
  serviceLevelAgreementUrl: z.string(),
  dataProcessingAgreementUrl: z.string(),
  websiteUrl: z.string(),
  legalName: z.string(),
  headquarterAddress: z.string(),
  certifications: z.array(z.string()),
  securityPageUrl: z.string(),
  trustPageUrl: z.string(),
  businessOwnerId: z.string(),
  securityOwnerId: z.string(),
});

const vendorFormFragment = graphql`
  fragment useVendorFormFragment on Vendor {
    id
    name
    description
    statusPageUrl
    termsOfServiceUrl
    privacyPolicyUrl
    serviceLevelAgreementUrl
    dataProcessingAgreementUrl
    websiteUrl
    legalName
    headquarterAddress
    certifications
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
    errorMessage: __("Failed to update vendor. Please try again."),
  });

  const form = useFormWithSchema(schema, {
    defaultValues: {
      name: vendor.name,
      description: vendor.description ?? "",
      statusPageUrl: vendor.statusPageUrl ?? "",
      termsOfServiceUrl: vendor.termsOfServiceUrl ?? "",
      privacyPolicyUrl: vendor.privacyPolicyUrl ?? "",
      serviceLevelAgreementUrl: vendor.serviceLevelAgreementUrl ?? "",
      dataProcessingAgreementUrl: vendor.dataProcessingAgreementUrl ?? "",
      websiteUrl: vendor.websiteUrl ?? "",
      legalName: vendor.legalName ?? "",
      headquarterAddress: vendor.headquarterAddress ?? "",
      certifications: [...(vendor.certifications ?? [])],
      securityPageUrl: vendor.securityPageUrl ?? "",
      trustPageUrl: vendor.trustPageUrl ?? "",
      businessOwnerId: vendor.businessOwner?.id,
      securityOwnerId: vendor.securityOwner?.id,
    },
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

  return {
    ...form,
    handleSubmit,
  };
}
