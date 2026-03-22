import { useTranslate } from "@probo/i18n";
import { Button, Card, Field } from "@probo/ui";
import { Controller, useWatch } from "react-hook-form";
import { graphql, useFragment } from "react-relay";
import { useOutletContext } from "react-router";
import { z } from "zod";

import type { CookieBannerGraphNodeQuery$data } from "#/__generated__/core/CookieBannerGraphNodeQuery.graphql";
import type { CookieBannerAppearanceTabFragment$key } from "#/__generated__/core/CookieBannerAppearanceTabFragment.graphql";
import { useUpdateCookieBannerMutation } from "#/hooks/graph/CookieBannerGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { CookieBannerPreview } from "../_components/CookieBannerPreview";

const fragment = graphql`
  fragment CookieBannerAppearanceTabFragment on CookieBanner {
    id
    title
    description
    acceptAllLabel
    rejectAllLabel
    savePreferencesLabel
    canUpdate: permission(action: "core:cookie-banner:update")
    theme {
      primaryColor
      primaryTextColor
      secondaryColor
      secondaryTextColor
      backgroundColor
      textColor
      secondaryTextBodyColor
      borderColor
      fontFamily
      borderRadius
      position
      revisitPosition
    }
    categories(first: 50, orderBy: { field: RANK, direction: ASC }) {
      edges {
        node {
          id
          name
          description
          required
          rank
          cookies {
            name
            duration
            description
          }
        }
      }
    }
  }
`;

const hexColorRegex = /^#[0-9a-fA-F]{6}$/;

const schema = z.object({
  title: z.string().min(1, "Title is required"),
  description: z.string().min(1, "Description is required"),
  acceptAllLabel: z.string().min(1, "Label is required"),
  rejectAllLabel: z.string().min(1, "Label is required"),
  savePreferencesLabel: z.string().min(1, "Label is required"),
  primaryColor: z.string().regex(hexColorRegex, "Must be a valid hex color"),
  primaryTextColor: z.string().regex(hexColorRegex, "Must be a valid hex color"),
  backgroundColor: z.string().regex(hexColorRegex, "Must be a valid hex color"),
  textColor: z.string().regex(hexColorRegex, "Must be a valid hex color"),
  secondaryTextBodyColor: z.string().regex(hexColorRegex, "Must be a valid hex color"),
  borderColor: z.string().regex(hexColorRegex, "Must be a valid hex color"),
  borderRadius: z.coerce.number().min(0).max(24),
});

const lightPreset = {
  primaryColor: "#2563eb",
  primaryTextColor: "#ffffff",
  backgroundColor: "#ffffff",
  textColor: "#1a1a1a",
  secondaryTextBodyColor: "#4b5563",
  borderColor: "#e5e7eb",
};

const darkPreset = {
  primaryColor: "#3b82f6",
  primaryTextColor: "#ffffff",
  backgroundColor: "#1f2937",
  textColor: "#f9fafb",
  secondaryTextBodyColor: "#9ca3af",
  borderColor: "#374151",
};

function ColorField({
  label,
  value,
  onChange,
  disabled,
  error,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  error?: string;
}) {
  return (
    <div className="flex flex-col gap-[6px]">
      <label className="text-sm font-medium text-txt-primary">{label}</label>
      <div className="flex items-center gap-2">
        <input
          type="color"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className="h-9 w-9 shrink-0 cursor-pointer rounded-lg border border-border-mid p-0.5 disabled:cursor-not-allowed disabled:opacity-60"
        />
        <input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className="w-full rounded-[10px] border border-border-mid bg-secondary px-3 py-[6px] text-sm text-txt-primary hover:border-border-strong disabled:bg-transparent disabled:opacity-60"
        />
      </div>
      {error && <span className="mt-1 text-sm text-txt-danger">{error}</span>}
    </div>
  );
}

export default function CookieBannerAppearanceTab() {
  const { banner } = useOutletContext<{
    banner: CookieBannerGraphNodeQuery$data["node"];
  }>();

  const { __ } = useTranslate();
  const data = useFragment<CookieBannerAppearanceTabFragment$key>(
    fragment,
    banner,
  );

  const [mutate] = useUpdateCookieBannerMutation();

  const {
    control,
    register,
    handleSubmit,
    formState: { errors, isDirty, isSubmitting },
    reset,
    setValue,
  } = useFormWithSchema(schema, {
    defaultValues: {
      title: data.title,
      description: data.description,
      acceptAllLabel: data.acceptAllLabel,
      rejectAllLabel: data.rejectAllLabel,
      savePreferencesLabel: data.savePreferencesLabel,
      primaryColor: data.theme.primaryColor,
      primaryTextColor: data.theme.primaryTextColor,
      backgroundColor: data.theme.backgroundColor,
      textColor: data.theme.textColor,
      secondaryTextBodyColor: data.theme.secondaryTextBodyColor,
      borderColor: data.theme.borderColor,
      borderRadius: data.theme.borderRadius,
    },
  });

  const watchedValues = useWatch({ control });

  const categories = (data.categories?.edges ?? []).map((edge) => ({
    id: edge.node.id,
    name: edge.node.name,
    description: edge.node.description,
    required: edge.node.required,
    rank: edge.node.rank,
    cookies: edge.node.cookies.map((c) => ({
      name: c.name,
      duration: c.duration,
      description: c.description,
    })),
  }));

  const isFormDisabled = isSubmitting || !data.canUpdate;

  const applyPreset = (preset: typeof lightPreset) => {
    for (const [key, value] of Object.entries(preset)) {
      setValue(key as keyof typeof preset, value, { shouldDirty: true });
    }
  };

  const themeForPreview = {
    primary_color: watchedValues.primaryColor ?? data.theme.primaryColor,
    primary_text_color: watchedValues.primaryTextColor ?? data.theme.primaryTextColor,
    background_color: watchedValues.backgroundColor ?? data.theme.backgroundColor,
    text_color: watchedValues.textColor ?? data.theme.textColor,
    secondary_text_body_color: watchedValues.secondaryTextBodyColor ?? data.theme.secondaryTextBodyColor,
    border_color: watchedValues.borderColor ?? data.theme.borderColor,
    border_radius: watchedValues.borderRadius ?? data.theme.borderRadius,
  };

  const onSubmit = handleSubmit(async (formData) => {
    const {
      primaryColor,
      primaryTextColor,
      backgroundColor,
      textColor,
      secondaryTextBodyColor,
      borderColor,
      borderRadius,
      ...contentData
    } = formData;

    await mutate({
      variables: {
        input: {
          id: data.id,
          ...contentData,
          theme: {
            primaryColor,
            primaryTextColor,
            backgroundColor,
            textColor,
            secondaryTextBodyColor,
            borderColor,
            borderRadius,
          },
        },
      },
    });
    reset(formData);
  });

  return (
    <div className="grid grid-cols-1 gap-8 xl:grid-cols-2">
    <form onSubmit={e => void onSubmit(e)} className="space-y-12">
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Banner Content")}</h2>
        <Card className="space-y-4" padded>
          <Field
            {...register("title")}
            label={__("Title")}
            type="text"
            error={errors.title?.message}
            disabled={isFormDisabled}
          />
          <Field
            {...register("description")}
            label={__("Description")}
            type="textarea"
            error={errors.description?.message}
            disabled={isFormDisabled}
          />
        </Card>
      </div>

      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Button Labels")}</h2>
        <Card className="space-y-4" padded>
          <Field
            {...register("acceptAllLabel")}
            label={__("Accept All Label")}
            type="text"
            error={errors.acceptAllLabel?.message}
            disabled={isFormDisabled}
          />
          <Field
            {...register("rejectAllLabel")}
            label={__("Reject All Label")}
            type="text"
            error={errors.rejectAllLabel?.message}
            disabled={isFormDisabled}
          />
          <Field
            {...register("savePreferencesLabel")}
            label={__("Save Preferences Label")}
            type="text"
            error={errors.savePreferencesLabel?.message}
            disabled={isFormDisabled}
          />
        </Card>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-medium">{__("Theme")}</h2>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => applyPreset(lightPreset)}
              disabled={isFormDisabled}
              className="rounded-lg border border-border-mid bg-white px-3 py-1.5 text-xs font-medium text-txt-primary hover:bg-tertiary-hover disabled:opacity-60"
            >
              {__("Light")}
            </button>
            <button
              type="button"
              onClick={() => applyPreset(darkPreset)}
              disabled={isFormDisabled}
              className="rounded-lg border border-border-mid bg-gray-800 px-3 py-1.5 text-xs font-medium text-white hover:bg-gray-700 disabled:opacity-60"
            >
              {__("Dark")}
            </button>
          </div>
        </div>
        <Card className="space-y-4" padded>
          <div className="grid grid-cols-2 gap-4">
            <Controller
              name="primaryColor"
              control={control}
              render={({ field }) => (
                <ColorField
                  label={__("Primary Color")}
                  value={field.value}
                  onChange={field.onChange}
                  disabled={isFormDisabled}
                  error={errors.primaryColor?.message}
                />
              )}
            />
            <Controller
              name="primaryTextColor"
              control={control}
              render={({ field }) => (
                <ColorField
                  label={__("Primary Text Color")}
                  value={field.value}
                  onChange={field.onChange}
                  disabled={isFormDisabled}
                  error={errors.primaryTextColor?.message}
                />
              )}
            />
            <Controller
              name="backgroundColor"
              control={control}
              render={({ field }) => (
                <ColorField
                  label={__("Background Color")}
                  value={field.value}
                  onChange={field.onChange}
                  disabled={isFormDisabled}
                  error={errors.backgroundColor?.message}
                />
              )}
            />
            <Controller
              name="textColor"
              control={control}
              render={({ field }) => (
                <ColorField
                  label={__("Text Color")}
                  value={field.value}
                  onChange={field.onChange}
                  disabled={isFormDisabled}
                  error={errors.textColor?.message}
                />
              )}
            />
            <Controller
              name="secondaryTextBodyColor"
              control={control}
              render={({ field }) => (
                <ColorField
                  label={__("Secondary Text Color")}
                  value={field.value}
                  onChange={field.onChange}
                  disabled={isFormDisabled}
                  error={errors.secondaryTextBodyColor?.message}
                />
              )}
            />
            <Controller
              name="borderColor"
              control={control}
              render={({ field }) => (
                <ColorField
                  label={__("Border Color")}
                  value={field.value}
                  onChange={field.onChange}
                  disabled={isFormDisabled}
                  error={errors.borderColor?.message}
                />
              )}
            />
          </div>

          <Field
            {...register("borderRadius", { valueAsNumber: true })}
            label={__("Border Radius")}
            type="number"
            error={errors.borderRadius?.message}
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

    <div className="space-y-4">
      <h2 className="text-base font-medium">{__("Preview")}</h2>
      <Card className="overflow-hidden" padded>
        <CookieBannerPreview
          title={watchedValues.title ?? ""}
          description={watchedValues.description ?? ""}
          acceptAllLabel={watchedValues.acceptAllLabel ?? ""}
          rejectAllLabel={watchedValues.rejectAllLabel ?? ""}
          savePreferencesLabel={watchedValues.savePreferencesLabel ?? ""}
          categories={categories}
          theme={themeForPreview}
        />
      </Card>
    </div>
    </div>
  );
}
