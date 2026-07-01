// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { Button, Card, Field, Label, Spinner, Textarea } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CompliancePageContactInfoSection_trustCenter$key } from "#/__generated__/core/CompliancePageContactInfoSection_trustCenter.graphql";
import type { CompliancePageContactInfoSection_updateMutation } from "#/__generated__/core/CompliancePageContactInfoSection_updateMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const fragment = graphql`
  fragment CompliancePageContactInfoSection_trustCenter on TrustCenter {
    id
    description
    websiteUrl
    email
    headquarterAddress
    canUpdate: permission(action: "core:trust-center:update")
  }
`;

const updateMutation = graphql`
  mutation CompliancePageContactInfoSection_updateMutation($input: UpdateTrustCenterBrandInput!) {
    updateTrustCenterBrand(input: $input) {
      trustCenter {
        id
        description
        websiteUrl
        email
        headquarterAddress
      }
    }
  }
`;

const contactInfoSchema = z.object({
  description: z.string().optional(),
  websiteUrl: z.string().optional(),
  email: z.string().optional(),
  headquarterAddress: z.string().optional(),
});

type ContactInfoFormData = z.infer<typeof contactInfoSchema>;

export function CompliancePageContactInfoSection(props: {
  trustCenterKey: CompliancePageContactInfoSection_trustCenter$key;
}) {
  const { trustCenterKey } = props;
  const { __ } = useTranslate();

  const trustCenter = useFragment(fragment, trustCenterKey);

  const [updateContactInfo, isUpdating] = useMutationWithToasts<CompliancePageContactInfoSection_updateMutation>(
    updateMutation,
    {
      successMessage: __("Compliance page contact information updated successfully"),
      errorMessage: __("Failed to update compliance page contact information"),
    },
  );

  const { formState, handleSubmit, register } = useFormWithSchema(
    contactInfoSchema,
    {
      defaultValues: {
        description: trustCenter.description || "",
        websiteUrl: trustCenter.websiteUrl || "",
        email: trustCenter.email || "",
        headquarterAddress: trustCenter.headquarterAddress || "",
      },
    },
  );

  const disabled = formState.isSubmitting || isUpdating || !trustCenter.canUpdate;

  const onSubmit = handleSubmit(async (data: ContactInfoFormData) => {
    await updateContactInfo({
      variables: {
        input: {
          trustCenterId: trustCenter.id,
          description: data.description || null,
          websiteUrl: data.websiteUrl || null,
          email: data.email || null,
          headquarterAddress: data.headquarterAddress || null,
        },
      },
    });
  });

  return (
    <form onSubmit={e => void onSubmit(e)} className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Contact information")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("These details are shown on your public compliance page.")}
          </p>
        </div>
        {(formState.isSubmitting || isUpdating) && <Spinner />}
      </div>

      <Card padded className="space-y-4">
        <div>
          <Label>{__("Description")}</Label>
          <Textarea
            {...register("description")}
            readOnly={disabled}
            name="description"
            placeholder={__("Brief description for your compliance page")}
            rows={3}
          />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field
            {...register("websiteUrl")}
            readOnly={disabled}
            name="websiteUrl"
            type="url"
            label={__("Website URL")}
            placeholder={__("https://example.com")}
          />
          <Field
            {...register("email")}
            readOnly={disabled}
            name="email"
            type="email"
            label={__("Email")}
            placeholder={__("contact@example.com")}
          />
        </div>
        <div>
          <Label>{__("Headquarter Address")}</Label>
          <Textarea
            {...register("headquarterAddress")}
            readOnly={disabled}
            name="headquarterAddress"
            placeholder={__("123 Main St, City, Country")}
          />
        </div>

        {formState.isDirty && trustCenter.canUpdate && (
          <div className="flex justify-end pt-2">
            <Button type="submit" disabled={disabled}>
              {formState.isSubmitting || isUpdating
                ? __("Updating...")
                : __("Update contact information")}
            </Button>
          </div>
        )}
      </Card>
    </form>
  );
}
