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
// OTHER TORTIOUS ACTION, ARISING FROM, OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { Card, Field, useToast } from "@probo/ui";
import type { FocusEvent } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalProfileSection_compliancePortalFragment$key } from "#/__generated__/core/CompliancePortalProfileSection_compliancePortalFragment.graphql";
import type { CompliancePortalProfileSection_updateMutation } from "#/__generated__/core/CompliancePortalProfileSection_updateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { z } from "#/lib/zod";

import { CompliancePortalVisualIdentitySection } from "./CompliancePortalVisualIdentitySection";

const compliancePortalFragment = graphql`
  fragment CompliancePortalProfileSection_compliancePortalFragment on CompliancePortal {
    id
    entityName
    description
    websiteUrl
    email
    headquarterAddress
    canUpdate: permission(action: "compliance-portal:portal:update")
    ...CompliancePortalVisualIdentitySection_compliancePortalFragment
  }
`;

const updateMutation = graphql`
  mutation CompliancePortalProfileSection_updateMutation($input: UpdateCompliancePortalInput!) {
    updateCompliancePortal(input: $input) {
      compliancePortal {
        id
        entityName
        description
        websiteUrl
        email
        headquarterAddress
      }
    }
  }
`;

type OptionalProfileField = "description" | "websiteUrl" | "email" | "headquarterAddress";

function normalizeOptional(value: string | null | undefined): string | null {
  const trimmed = value?.trim() ?? "";
  return trimmed.length > 0 ? trimmed : null;
}

export function CompliancePortalProfileSection({
  compliancePortalRef,
}: {
  compliancePortalRef: CompliancePortalProfileSection_compliancePortalFragment$key;
}) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { toast } = useToast();

  const compliancePortal = useFragment(
    compliancePortalFragment,
    compliancePortalRef,
  );
  const { canUpdate } = compliancePortal;

  const [updateCompliancePortal] = useMutation<CompliancePortalProfileSection_updateMutation>(
    updateMutation,
    {
      errorToast: t("brandPage.profile.errors.update"),
    },
  );

  function patch(
    field: "entityName" | OptionalProfileField,
    value: string | null,
  ) {
    void updateCompliancePortal({
      variables: {
        input: {
          compliancePortalId: compliancePortal.id,
          [field]: value,
        },
      },
    });
  }

  function handleEntityNameBlur(event: FocusEvent<HTMLInputElement>) {
    const next = event.currentTarget.value.trim();
    const current = compliancePortal.entityName;
    if (next.length === 0) {
      event.currentTarget.value = current;
      return;
    }
    if (next === current) {
      return;
    }
    patch("entityName", next);
  }

  // Blur-to-save never goes through form submission, so the browser's own
  // url/email constraint validation never runs.
  const fieldSchemas: Partial<Record<OptionalProfileField, {
    schema: z.ZodString;
    message: string;
  }>> = {
    websiteUrl: {
      schema: z.string().url(),
      message: t("externalUrls.validation.urlInvalid"),
    },
    email: {
      schema: z.string().email(),
      message: t("brandPage.profile.validation.emailInvalid"),
    },
  };

  function optionalBlurHandler(
    field: OptionalProfileField,
    current: string | null | undefined,
  ) {
    return (event: FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const next = normalizeOptional(event.currentTarget.value);
      if (next === normalizeOptional(current)) {
        return;
      }
      const validation = fieldSchemas[field];
      if (validation != null && next != null && !validation.schema.safeParse(next).success) {
        event.currentTarget.value = current ?? "";
        toast({
          title: validation.message,
          description: "",
          variant: "error",
        });
        return;
      }
      patch(field, next);
    };
  }

  const entityName = compliancePortal.entityName;
  const description = normalizeOptional(compliancePortal.description) ?? "";
  const websiteUrl = normalizeOptional(compliancePortal.websiteUrl) ?? "";
  const email = normalizeOptional(compliancePortal.email) ?? "";
  const headquarterAddress = normalizeOptional(compliancePortal.headquarterAddress) ?? "";

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-base font-medium">{t("brandPage.profile.title")}</h2>
        <p className="text-sm text-txt-tertiary">
          {t("brandPage.profile.description")}
        </p>
      </div>
      <Card padded className="space-y-4">
        <Field
          key={entityName}
          defaultValue={entityName}
          onBlur={handleEntityNameBlur}
          readOnly={!canUpdate}
          name="entityName"
          label={t("brandPage.profile.fields.entityName")}
          placeholder={t("brandPage.profile.fields.entityNamePlaceholder")}
        />
        <Field
          key={description}
          defaultValue={description}
          onBlur={optionalBlurHandler("description", compliancePortal.description)}
          readOnly={!canUpdate}
          name="description"
          type="textarea"
          label={t("brandPage.profile.fields.description")}
          placeholder={t("brandPage.profile.fields.descriptionPlaceholder")}
          rows={3}
        />
        <CompliancePortalVisualIdentitySection compliancePortalRef={compliancePortal} />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field
            key={websiteUrl}
            defaultValue={websiteUrl}
            onBlur={optionalBlurHandler("websiteUrl", compliancePortal.websiteUrl)}
            readOnly={!canUpdate}
            name="websiteUrl"
            type="url"
            label={t("brandPage.profile.fields.websiteUrl")}
            placeholder={t("externalUrls.fields.urlPlaceholder")}
          />
          <Field
            key={email}
            defaultValue={email}
            onBlur={optionalBlurHandler("email", compliancePortal.email)}
            readOnly={!canUpdate}
            name="email"
            type="email"
            label={t("brandPage.profile.fields.email")}
            placeholder={t("brandPage.profile.fields.emailPlaceholder")}
          />
        </div>
        <Field
          key={headquarterAddress}
          defaultValue={headquarterAddress}
          onBlur={optionalBlurHandler(
            "headquarterAddress",
            compliancePortal.headquarterAddress,
          )}
          readOnly={!canUpdate}
          name="headquarterAddress"
          label={t("brandPage.profile.fields.address")}
          placeholder={t("brandPage.profile.fields.addressPlaceholder")}
        />
      </Card>
    </div>
  );
}
