// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { useEffect, useMemo } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { useThirdPartyFormFragment$key } from "#/__generated__/core/useThirdPartyFormFragment.graphql";

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

const thirdPartyFormFragment = graphql`
  fragment useThirdPartyFormFragment on ThirdParty {
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

const thirdPartyUpdateQuery = graphql`
  mutation useThirdPartyFormMutation($input: UpdateThirdPartyInput!) {
    updateThirdParty(input: $input) {
      thirdParty {
        ...useThirdPartyFormFragment
      }
    }
  }
`;

export function useThirdPartyForm(thirdPartyKey: useThirdPartyFormFragment$key) {
  const thirdParty = useFragment(thirdPartyFormFragment, thirdPartyKey);
  const { __ } = useTranslate();

  const [mutate] = useMutationWithToasts(thirdPartyUpdateQuery, {
    successMessage: __("Third party updated successfully."),
    errorMessage: __("Failed to update third party"),
  });

  const defaultValues = useMemo(
    () => ({
      name: thirdParty.name,
      description: thirdParty.description || null,
      category: thirdParty.category || null,
      statusPageUrl: thirdParty.statusPageUrl || null,
      termsOfServiceUrl: thirdParty.termsOfServiceUrl || null,
      privacyPolicyUrl: thirdParty.privacyPolicyUrl || null,
      serviceLevelAgreementUrl: thirdParty.serviceLevelAgreementUrl || null,
      dataProcessingAgreementUrl: thirdParty.dataProcessingAgreementUrl || null,
      websiteUrl: thirdParty.websiteUrl || null,
      legalName: thirdParty.legalName || null,
      headquarterAddress: thirdParty.headquarterAddress || null,
      certifications: [...(thirdParty.certifications ?? [])],
      countries: [...(thirdParty.countries ?? [])],
      securityPageUrl: thirdParty.securityPageUrl || null,
      trustPageUrl: thirdParty.trustPageUrl || null,
      businessOwnerId: thirdParty.businessOwner?.id,
      securityOwnerId: thirdParty.securityOwner?.id,
    }),
    [thirdParty],
  );

  const form = useFormWithSchema(schema, {
    defaultValues,
  });

  const handleSubmit = form.handleSubmit((data) => {
    return mutate({
      variables: {
        input: {
          id: thirdParty.id,
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
