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

import { Form } from "@base-ui/react/form";
import { formatError, toFieldErrors } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { NewCookieBannerPageMutation } from "#/__generated__/core/NewCookieBannerPageMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { CookieBannerPageHeader } from "./_components/CookieBannerPageHeader";
import { cookieBannerInstallPath } from "./_lib/cookieBannerPaths";
import { cookieBannerPage, cookieBannerSettingsSection } from "./variants";

const createCookieBannerMutation = graphql`
  mutation NewCookieBannerPageMutation(
    $input: CreateCookieBannerInput!
    $connections: [ID!]!
  ) {
    createCookieBanner(input: $input) {
      cookieBannerEdge @prependEdge(connections: $connections) {
        node {
          id
          ...CookieBannerSwitcherListItem_cookieBanner
        }
      }
    }
  }
`;

export default function NewCookieBannerPage() {
  const { t } = useTranslation("organizations/cookie-banners");
  const toast = useToast();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const { card, fields, pair, actions } = cookieBannerSettingsSection();
  const [name, setName] = useState("");
  const [origin, setOrigin] = useState("");
  const [cookiePolicyUrl, setCookiePolicyUrl] = useState("");
  const [privacyPolicyUrl, setPrivacyPolicyUrl] = useState("");
  const [consentExpiryDays, setConsentExpiryDays] = useState("365");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const createError = t("newCookieBannerPage.errors.create");
  const [createCookieBanner, isCreating] = useMutation<NewCookieBannerPageMutation>(
    createCookieBannerMutation,
    {
      successMessage: t("newCookieBannerPage.messages.created"),
      errorToast: false,
    },
  );

  usePageTitle(t("newCookieBannerPage.pageTitle"));

  function clearFieldError(field: string) {
    setErrors((current) => {
      if (current[field] == null) {
        return current;
      }
      const next = { ...current };
      delete next[field];
      return next;
    });
  }

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "CookieBannerSwitcherMenu_cookieBanners",
  );

  function handleSubmit() {
    void createCookieBanner({
      variables: {
        input: {
          organizationId,
          name: name.trim(),
          origin: origin.trim(),
          cookiePolicyUrl: cookiePolicyUrl.trim(),
          privacyPolicyUrl: privacyPolicyUrl.trim() || undefined,
          consentExpiryDays: Number.parseInt(consentExpiryDays, 10),
        },
        connections: [connectionId],
      },
      onCompleted(_response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(fieldErrors);
          return;
        }
        if (payloadErrors != null && payloadErrors.length > 0) {
          toast.add({
            title: t("newCookieBannerPage.errors.title"),
            description: formatError(createError, payloadErrors),
            type: "error",
          });
        }
      },
      onError(error) {
        toast.add({
          title: t("newCookieBannerPage.errors.title"),
          description: formatError(createError, error),
          type: "error",
        });
      },
    }).then((response) => {
      const bannerId = response.createCookieBanner.cookieBannerEdge.node.id;
      void navigate(cookieBannerInstallPath(organizationId, bannerId));
    }).catch(() => {
      // Field errors stay on the form; other failures toast above.
    });
  }

  return (
    <div className={cookieBannerPage()}>
      <CookieBannerPageHeader
        title={t("newCookieBannerPage.title")}
        description={t("newCookieBannerPage.description")}
      />
      <Card size={2} variant="soft" className={card()}>
        <Form className={fields()} errors={errors} onFormSubmit={handleSubmit}>
          <div className={pair()}>
            <Field required label={t("newCookieBannerPage.fields.name")} error={errors.name}>
              <TextField
                name="name"
                required
                value={name}
                disabled={isCreating}
                placeholder={t("newCookieBannerPage.fields.namePlaceholder")}
                onValueChange={(value) => {
                  setName(value);
                  clearFieldError("name");
                }}
              />
            </Field>
            <Field
              required
              label={t("newCookieBannerPage.fields.consentExpiryDays")}
              error={errors.consentExpiryDays}
            >
              <TextField
                name="consentExpiryDays"
                type="number"
                required
                min={1}
                value={consentExpiryDays}
                disabled={isCreating}
                onValueChange={(value) => {
                  setConsentExpiryDays(value);
                  clearFieldError("consentExpiryDays");
                }}
              />
            </Field>
          </div>
          <Field required label={t("newCookieBannerPage.fields.origin")} error={errors.origin}>
            <TextField
              name="origin"
              required
              value={origin}
              disabled={isCreating}
              placeholder={t("newCookieBannerPage.fields.originPlaceholder")}
              onValueChange={(value) => {
                setOrigin(value);
                clearFieldError("origin");
              }}
            />
          </Field>
          <div className={pair()}>
            <Field
              required
              label={t("newCookieBannerPage.fields.cookiePolicyUrl")}
              error={errors.cookiePolicyUrl}
            >
              <TextField
                name="cookiePolicyUrl"
                type="url"
                required
                value={cookiePolicyUrl}
                disabled={isCreating}
                placeholder={t("newCookieBannerPage.fields.cookiePolicyUrlPlaceholder")}
                onValueChange={(value) => {
                  setCookiePolicyUrl(value);
                  clearFieldError("cookiePolicyUrl");
                }}
              />
            </Field>
            <Field
              label={t("newCookieBannerPage.fields.privacyPolicyUrl")}
              error={errors.privacyPolicyUrl}
            >
              <TextField
                name="privacyPolicyUrl"
                type="url"
                value={privacyPolicyUrl}
                disabled={isCreating}
                placeholder={t("newCookieBannerPage.fields.privacyPolicyUrlPlaceholder")}
                onValueChange={(value) => {
                  setPrivacyPolicyUrl(value);
                  clearFieldError("privacyPolicyUrl");
                }}
              />
            </Field>
          </div>
          <div className={actions()}>
            <Button
              type="submit"
              variant="solid"
              color="neutral"
              highContrast
              loading={isCreating}
            >
              {t("newCookieBannerPage.actions.create")}
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  );
}
