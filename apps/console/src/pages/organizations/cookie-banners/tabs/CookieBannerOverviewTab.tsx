import { useTranslate } from "@probo/i18n";
import { Button, Card, Field } from "@probo/ui";
import { graphql, useFragment } from "react-relay";
import { useOutletContext } from "react-router";
import { z } from "zod";

import type { CookieBannerGraphNodeQuery$data } from "#/__generated__/core/CookieBannerGraphNodeQuery.graphql";
import type { CookieBannerOverviewTabFragment$key } from "#/__generated__/core/CookieBannerOverviewTabFragment.graphql";
import { useUpdateCookieBannerMutation } from "#/hooks/graph/CookieBannerGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const fragment = graphql`
  fragment CookieBannerOverviewTabFragment on CookieBanner {
    id
    name
    domain
    privacyPolicyUrl
    consentExpiryDays
    canUpdate: permission(action: "core:cookie-banner:update")
  }
`;

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  domain: z.string().min(1, "Domain is required"),
  privacyPolicyUrl: z.string().optional().nullable(),
  consentExpiryDays: z.number().min(1),
});

export default function CookieBannerOverviewTab() {
  const { banner } = useOutletContext<{
    banner: CookieBannerGraphNodeQuery$data["node"];
  }>();

  const { __ } = useTranslate();
  const data = useFragment<CookieBannerOverviewTabFragment$key>(
    fragment,
    banner,
  );

  const [mutate] = useUpdateCookieBannerMutation();

  const {
    register,
    handleSubmit,
    formState: { errors, isDirty, isSubmitting },
    reset,
  } = useFormWithSchema(schema, {
    defaultValues: {
      name: data.name,
      domain: data.domain,
      privacyPolicyUrl: data.privacyPolicyUrl || null,
      consentExpiryDays: data.consentExpiryDays,
    },
  });

  const isFormDisabled = isSubmitting || !data.canUpdate;

  const onSubmit = handleSubmit(async (formData) => {
    await mutate({
      variables: {
        input: {
          id: data.id,
          ...formData,
          privacyPolicyUrl: formData.privacyPolicyUrl || null,
        },
      },
    });
    reset(formData);
  });

  return (
    <form onSubmit={e => void onSubmit(e)} className="space-y-12">
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("General")}</h2>
        <Card className="space-y-4" padded>
          <Field
            {...register("name")}
            label={__("Name")}
            type="text"
            error={errors.name?.message}
            disabled={isFormDisabled}
          />
          <Field
            {...register("domain")}
            label={__("Domain")}
            type="text"
            error={errors.domain?.message}
            disabled={isFormDisabled}
          />
          <Field
            {...register("privacyPolicyUrl")}
            label={__("Privacy Policy URL")}
            type="text"
            error={errors.privacyPolicyUrl?.message}
            disabled={isFormDisabled}
          />
          <Field
            {...register("consentExpiryDays", { valueAsNumber: true })}
            label={__("Consent Expiry (days)")}
            type="number"
            error={errors.consentExpiryDays?.message}
            disabled={isFormDisabled}
          />
        </Card>
      </div>

      {isDirty && data.canUpdate && (
        <div className="flex justify-end">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? __("Updating...") : __("Update")}
          </Button>
        </div>
      )}
    </form>
  );
}
