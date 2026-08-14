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

import { useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { useThirdPartyProfileFormFragment$key } from "#/__generated__/core/useThirdPartyProfileFormFragment.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { z } from "#/lib/zod";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string().optional().nullable(),
  category: z.string().nullish(),
  websiteUrl: z.string().optional().nullable(),
  legalName: z.string().optional().nullable(),
  headquarterAddress: z.string().optional().nullable(),
  countries: z.array(z.string()),
});

const thirdPartyProfileFormFragment = graphql`
  fragment useThirdPartyProfileFormFragment on ThirdParty {
    id
    name
    description
    category
    websiteUrl
    legalName
    headquarterAddress
    countries
  }
`;

const thirdPartyProfileUpdateMutation = graphql`
  mutation useThirdPartyProfileFormMutation($input: UpdateThirdPartyInput!) {
    updateThirdParty(input: $input) {
      thirdParty {
        ...useThirdPartyProfileFormFragment
      }
    }
  }
`;

export function useThirdPartyProfileForm(
  thirdPartyKey: useThirdPartyProfileFormFragment$key,
) {
  const thirdParty = useFragment(thirdPartyProfileFormFragment, thirdPartyKey);
  const { t } = useTranslation();

  const [mutate] = useMutationWithToasts(thirdPartyProfileUpdateMutation, {
    successMessage: t("thirdPartyProfilePage.messages.updated"),
    errorMessage: t("thirdPartyProfilePage.messages.updateError"),
  });

  const defaultValues = useMemo(
    () => ({
      name: thirdParty.name,
      description: thirdParty.description || null,
      category: thirdParty.category || null,
      websiteUrl: thirdParty.websiteUrl || null,
      legalName: thirdParty.legalName || null,
      headquarterAddress: thirdParty.headquarterAddress || null,
      countries: [...(thirdParty.countries ?? [])],
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
          name: data.name,
          category: data.category,
          countries: data.countries,
          description: data.description || null,
          websiteUrl: data.websiteUrl || null,
          legalName: data.legalName || null,
          headquarterAddress: data.headquarterAddress || null,
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
