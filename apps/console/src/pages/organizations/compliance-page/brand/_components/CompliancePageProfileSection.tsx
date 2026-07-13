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

import type { CompliancePageProfileSection_compliancePageFragment$key } from "#/__generated__/core/CompliancePageProfileSection_compliancePageFragment.graphql";
import { useUpdateCompliancePageMutation } from "#/hooks/graph/CompliancePageGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const compliancePageFragment = graphql`
  fragment CompliancePageProfileSection_compliancePageFragment on TrustCenter {
    id
    title
    description
    websiteUrl
    email
    headquarterAddress
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const profileSchema = z.object({
  title: z.string().min(1),
  description: z.string().optional(),
  websiteUrl: z.string().optional(),
  email: z.string().optional(),
  headquarterAddress: z.string().optional(),
});

type ProfileFormData = z.infer<typeof profileSchema>;

export function CompliancePageProfileSection(props: {
  compliancePageRef: CompliancePageProfileSection_compliancePageFragment$key;
}) {
  const { __ } = useTranslate();

  const { canUpdate, ...compliancePage } = useFragment(
    compliancePageFragment,
    props.compliancePageRef,
  );

  const [updateCompliancePage, isUpdating] = useUpdateCompliancePageMutation();

  const { formState, handleSubmit, register } = useFormWithSchema(profileSchema, {
    defaultValues: {
      title: compliancePage.title,
      description: compliancePage.description || "",
      websiteUrl: compliancePage.websiteUrl || "",
      email: compliancePage.email || "",
      headquarterAddress: compliancePage.headquarterAddress || "",
    },
  });

  const readOnly = formState.isSubmitting || !canUpdate;

  const onSubmit = handleSubmit(async (data: ProfileFormData) => {
    await updateCompliancePage({
      variables: {
        input: {
          trustCenterId: compliancePage.id,
          title: data.title,
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
          <h2 className="text-base font-medium">{__("General information")}</h2>
          <p className="text-sm text-txt-tertiary">
            {__("Description and contact details shown to visitors.")}
          </p>
        </div>
        {formState.isSubmitting && <Spinner />}
      </div>
      <Card padded className="space-y-4">
        <Field
          {...register("title")}
          readOnly={readOnly}
          name="title"
          label={__("Title")}
          placeholder={__("Your company or product name")}
        />
        <div>
          <Label>{__("Description")}</Label>
          <Textarea
            {...register("description")}
            readOnly={readOnly}
            name="description"
            placeholder={__("Brief description for visitors")}
            rows={3}
          />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field
            {...register("websiteUrl")}
            readOnly={readOnly}
            name="websiteUrl"
            type="url"
            label={__("Website URL")}
            placeholder={__("https://example.com")}
          />
          <Field
            {...register("email")}
            readOnly={readOnly}
            name="email"
            type="email"
            label={__("Email")}
            placeholder={__("contact@example.com")}
          />
        </div>
        <Field
          {...register("headquarterAddress")}
          readOnly={readOnly}
          name="headquarterAddress"
          label={__("Headquarter Address")}
          placeholder={__("123 Main St, City, Country")}
        />

        {formState.isDirty && canUpdate && (
          <div className="flex justify-end pt-6">
            <Button type="submit" disabled={formState.isSubmitting || isUpdating}>
              {formState.isSubmitting || isUpdating
                ? __("Updating...")
                : __("Save changes")}
            </Button>
          </div>
        )}
      </Card>
    </form>
  );
}
