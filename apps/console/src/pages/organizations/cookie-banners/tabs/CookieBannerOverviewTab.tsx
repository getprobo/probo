import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card, Field, useToast } from "@probo/ui";
import { Controller } from "react-hook-form";
import { graphql, useFragment, useMutation } from "react-relay";
import { useOutletContext } from "react-router";
import { z } from "zod";

import type { CookieBannerDetailPageQuery$data } from "#/__generated__/core/CookieBannerDetailPageQuery.graphql";
import type { CookieBannerOverviewTabFragment$key } from "#/__generated__/core/CookieBannerOverviewTabFragment.graphql";
import type { CookieBannerOverviewTabUpdateMutation } from "#/__generated__/core/CookieBannerOverviewTabUpdateMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const fragment = graphql`
  fragment CookieBannerOverviewTabFragment on CookieBanner {
    id
    name
    domain
    privacyPolicyUrl
    consentExpiryDays
    consentMode
    canUpdate: permission(action: "core:cookie-banner:update")
    analytics {
      totalRecords
      acceptAllCount
      rejectAllCount
      customizeCount
      acceptCategoryCount
      gpcCount
    }
  }
`;

const updateCookieBannerMutation = graphql`
  mutation CookieBannerOverviewTabUpdateMutation(
    $input: UpdateCookieBannerInput!
  ) {
    updateCookieBanner(input: $input) {
      cookieBanner {
        id
        name
        domain
        privacyPolicyUrl
        consentExpiryDays
        consentMode
        version
        updatedAt
      }
    }
  }
`;

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  domain: z.string().min(1, "Domain is required"),
  privacyPolicyUrl: z.string().optional().nullable(),
  consentExpiryDays: z.number().min(1),
  consentMode: z.enum(["OPT_IN", "OPT_OUT"]),
});

function AnalyticsCard({
  analytics,
}: {
  analytics: {
    totalRecords: number;
    acceptAllCount: number;
    rejectAllCount: number;
    customizeCount: number;
    acceptCategoryCount: number;
    gpcCount: number;
  };
}) {
  const { __ } = useTranslate();

  const rate =
    analytics.totalRecords > 0
      ? Math.round(
          (analytics.acceptAllCount / analytics.totalRecords) * 100,
        )
      : 0;

  return (
    <Card padded>
      <div className="grid grid-cols-2 gap-6 sm:grid-cols-3 lg:grid-cols-6">
        <div>
          <div className="text-2xl font-semibold">{analytics.totalRecords}</div>
          <div className="text-sm text-txt-secondary">{__("Total Records")}</div>
        </div>
        <div>
          <div className="text-2xl font-semibold text-txt-success">
            {analytics.acceptAllCount}
          </div>
          <div className="text-sm text-txt-secondary">{__("Accept All")}</div>
        </div>
        <div>
          <div className="text-2xl font-semibold text-txt-danger">
            {analytics.rejectAllCount}
          </div>
          <div className="text-sm text-txt-secondary">{__("Reject All")}</div>
        </div>
        <div>
          <div className="text-2xl font-semibold text-txt-warning">
            {analytics.customizeCount}
          </div>
          <div className="text-sm text-txt-secondary">{__("Customize")}</div>
        </div>
        <div>
          <div className="text-2xl font-semibold">
            {analytics.acceptCategoryCount}
          </div>
          <div className="text-sm text-txt-secondary">{__("Accept Category")}</div>
        </div>
        <div>
          <div className="text-2xl font-semibold">{rate}%</div>
          <div className="text-sm text-txt-secondary">{__("Acceptance Rate")}</div>
        </div>
      </div>
    </Card>
  );
}

export default function CookieBannerOverviewTab() {
  const { banner } = useOutletContext<{
    banner: CookieBannerDetailPageQuery$data["node"];
  }>();

  const { __ } = useTranslate();
  const { toast } = useToast();
  const data = useFragment<CookieBannerOverviewTabFragment$key>(
    fragment,
    banner,
  );

  const [mutate] = useMutation<CookieBannerOverviewTabUpdateMutation>(updateCookieBannerMutation);

  const {
    control,
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
      consentMode: data.consentMode,
    },
  });

  const isFormDisabled = isSubmitting || !data.canUpdate;

  const onSubmit = handleSubmit((formData) => {
    mutate({
      variables: {
        input: {
          id: data.id,
          ...formData,
          privacyPolicyUrl: formData.privacyPolicyUrl || null,
        },
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Cookie banner updated successfully."),
          variant: "success",
        });
        reset(formData);
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to update cookie banner"), error as GraphQLError),
          variant: "error",
        });
      },
    });
  });

  return (
    <div className="space-y-12">
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Consent Analytics")}</h2>
        <AnalyticsCard analytics={data.analytics} />
      </div>

      <form onSubmit={(e) => void onSubmit(e)} className="space-y-12">
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

        <div className="space-y-4">
          <h2 className="text-base font-medium">{__("Consent Mode")}</h2>
          <Card className="space-y-4" padded>
            <Controller
              name="consentMode"
              control={control}
              render={({ field }) => (
                <div className="space-y-3">
                  <label className="flex items-start gap-3 cursor-pointer">
                    <input
                      type="radio"
                      value="OPT_IN"
                      checked={field.value === "OPT_IN"}
                      onChange={field.onChange}
                      disabled={isFormDisabled}
                      className="mt-1"
                    />
                    <div>
                      <div className="font-medium">{__("Opt-in (recommended)")}</div>
                      <div className="text-sm text-txt-secondary">
                        {__(
                          "Visitors must actively consent before non-essential cookies are set. Required by GDPR.",
                        )}
                      </div>
                    </div>
                  </label>
                  <label className="flex items-start gap-3 cursor-pointer">
                    <input
                      type="radio"
                      value="OPT_OUT"
                      checked={field.value === "OPT_OUT"}
                      onChange={field.onChange}
                      disabled={isFormDisabled}
                      className="mt-1"
                    />
                    <div>
                      <div className="font-medium">{__("Opt-out")}</div>
                      <div className="text-sm text-txt-secondary">
                        {__(
                          "All cookies are active by default. Visitors can opt out. Common for US-based sites.",
                        )}
                      </div>
                    </div>
                  </label>
                </div>
              )}
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
    </div>
  );
}
