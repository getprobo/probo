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

import { Button, Card, Field, Label, Spinner, Textarea } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalProfileSection_compliancePortalFragment$key } from "#/__generated__/core/CompliancePortalProfileSection_compliancePortalFragment.graphql";
import { useUpdateCompliancePortalMutation } from "#/hooks/graph/CompliancePortalGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { z } from "#/lib/zod";

const compliancePortalFragment = graphql`
  fragment CompliancePortalProfileSection_compliancePortalFragment on CompliancePortal {
    id
    entityName
    description
    websiteUrl
    email
    headquarterAddress
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const profileSchema = z.object({
  entityName: z.string().min(1),
  description: z.string().optional(),
  websiteUrl: z.string().optional(),
  email: z.string().optional(),
  headquarterAddress: z.string().optional(),
});

type ProfileFormData = z.infer<typeof profileSchema>;

export function CompliancePortalProfileSection(props: {
  compliancePortalRef: CompliancePortalProfileSection_compliancePortalFragment$key;
}) {
  const { t } = useTranslation("organizations/compliance-portals");

  const { canUpdate, ...compliancePortal } = useFragment(
    compliancePortalFragment,
    props.compliancePortalRef,
  );

  const [updateCompliancePortal, isUpdating] = useUpdateCompliancePortalMutation();

  const { formState, handleSubmit, register } = useFormWithSchema(profileSchema, {
    defaultValues: {
      entityName: compliancePortal.entityName,
      description: compliancePortal.description || "",
      websiteUrl: compliancePortal.websiteUrl || "",
      email: compliancePortal.email || "",
      headquarterAddress: compliancePortal.headquarterAddress || "",
    },
  });

  const readOnly = formState.isSubmitting || !canUpdate;

  const onSubmit = handleSubmit(async (data: ProfileFormData) => {
    await updateCompliancePortal({
      variables: {
        input: {
          compliancePortalId: compliancePortal.id,
          entityName: data.entityName,
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
          <h2 className="text-base font-medium">{t("brandPage.profile.title")}</h2>
          <p className="text-sm text-txt-tertiary">
            {t("brandPage.profile.description")}
          </p>
        </div>
        {formState.isSubmitting && <Spinner />}
      </div>
      <Card padded className="space-y-4">
        <Field
          {...register("entityName")}
          readOnly={readOnly}
          name="entityName"
          label={t("brandPage.profile.fields.entityName")}
          placeholder={t("brandPage.profile.fields.entityNamePlaceholder")}
        />
        <div>
          <Label>{t("brandPage.profile.fields.description")}</Label>
          <Textarea
            {...register("description")}
            readOnly={readOnly}
            name="description"
            placeholder={t("brandPage.profile.fields.descriptionPlaceholder")}
            rows={3}
          />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field
            {...register("websiteUrl")}
            readOnly={readOnly}
            name="websiteUrl"
            type="url"
            label={t("brandPage.profile.fields.websiteUrl")}
            placeholder={t("externalUrls.fields.urlPlaceholder")}
          />
          <Field
            {...register("email")}
            readOnly={readOnly}
            name="email"
            type="email"
            label={t("brandPage.profile.fields.email")}
            placeholder={t("brandPage.profile.fields.emailPlaceholder")}
          />
        </div>
        <Field
          {...register("headquarterAddress")}
          readOnly={readOnly}
          name="headquarterAddress"
          label={t("brandPage.profile.fields.address")}
          placeholder={t("brandPage.profile.fields.addressPlaceholder")}
        />

        {formState.isDirty && canUpdate && (
          <div className="flex justify-end pt-6">
            <Button type="submit" disabled={formState.isSubmitting || isUpdating}>
              {formState.isSubmitting || isUpdating
                ? t("brandPage.actions.updating")
                : t("externalUrls.actions.save")}
            </Button>
          </div>
        )}
      </Card>
    </form>
  );
}
