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
import { countries, toFieldErrors } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Switch } from "@probo/ui/src/v2/Switch/Switch";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { BannerSettingsForm_cookieBanner$key } from "#/__generated__/core/BannerSettingsForm_cookieBanner.graphql";
import type { BannerSettingsFormMutation } from "#/__generated__/core/BannerSettingsFormMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { cookieBannerSettingsSection } from "../../../variants";

import { BannerLifecycleSection } from "./BannerLifecycleSection";

const LANGUAGES = ["en", "fr", "de", "es", "nl"] as const;

const LANGUAGE_KEYS = {
  en: "english",
  fr: "french",
  de: "german",
  es: "spanish",
  nl: "dutch",
} as const;

const bannerSettingsFormFragment = graphql`
  fragment BannerSettingsForm_cookieBanner on CookieBanner {
    id
    name
    origin
    cookiePolicyUrl
    privacyPolicyUrl
    consentExpiryDays
    defaultLanguage
    publisherCountryCode
    capabilities {
      resourceReporting
      corsless
      tcf
    }
    ...BannerLifecycleSection_cookieBanner
  }
`;

const updateBannerMutation = graphql`
  mutation BannerSettingsFormMutation($input: UpdateCookieBannerInput!) {
    updateCookieBanner(input: $input) {
      cookieBanner {
        id
        name
        cookiePolicyUrl
        privacyPolicyUrl
        consentExpiryDays
        defaultLanguage
        publisherCountryCode
        capabilities {
          resourceReporting
        }
        latestVersion {
          id
          version
          state
        }
      }
    }
  }
`;

interface BannerSettingsFormProps {
  cookieBannerKey: BannerSettingsForm_cookieBanner$key;
}

function isPublisherCountryCode(value: string): boolean {
  const code = value.trim().toUpperCase();
  return code === "AA" || (countries as readonly string[]).includes(code);
}

export function BannerSettingsForm({ cookieBannerKey }: BannerSettingsFormProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const banner = useFragment(bannerSettingsFormFragment, cookieBannerKey);
  const { root, card, block, intro, fields, field, pair, toggle, toggleCopy, actions } = cookieBannerSettingsSection();
  const [name, setName] = useState(banner.name);
  const [cookiePolicyUrl, setCookiePolicyUrl] = useState(banner.cookiePolicyUrl);
  const [privacyPolicyUrl, setPrivacyPolicyUrl] = useState(banner.privacyPolicyUrl ?? "");
  const [consentExpiryDays, setConsentExpiryDays] = useState(String(banner.consentExpiryDays));
  const [defaultLanguage, setDefaultLanguage] = useState(banner.defaultLanguage);
  const [publisherCountryCode, setPublisherCountryCode] = useState(banner.publisherCountryCode);
  const [resourceReportingEnabled, setResourceReportingEnabled] = useState(
    banner.capabilities.resourceReporting,
  );
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [updateBanner, isUpdating] = useMutation<BannerSettingsFormMutation>(
    updateBannerMutation,
    {
      successMessage: t("bannerSettingsForm.messages.updated"),
      errorToast: t("bannerSettingsForm.errors.update"),
    },
  );

  function handleSubmit() {
    if (banner.capabilities.tcf && !isPublisherCountryCode(publisherCountryCode)) {
      setErrors({
        publisherCountryCode: t("bannerSettingsForm.errors.publisherCountryCode"),
      });
      return;
    }

    void updateBanner({
      variables: {
        input: {
          cookieBannerId: banner.id,
          name: name.trim(),
          cookiePolicyUrl: cookiePolicyUrl.trim(),
          privacyPolicyUrl: privacyPolicyUrl.trim() || undefined,
          consentExpiryDays: Number.parseInt(consentExpiryDays, 10),
          defaultLanguage,
          publisherCountryCode: banner.capabilities.tcf
            ? (publisherCountryCode.trim().toUpperCase() || "AA")
            : undefined,
          capabilities: { resourceReporting: resourceReportingEnabled },
        },
      },
      onCompleted(response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(fieldErrors);
          return;
        }
        if (payloadErrors != null && payloadErrors.length > 0) {
          return;
        }
        const saved = response.updateCookieBanner.cookieBanner;
        setName(saved.name);
        setCookiePolicyUrl(saved.cookiePolicyUrl);
        setPrivacyPolicyUrl(saved.privacyPolicyUrl ?? "");
        setPublisherCountryCode(saved.publisherCountryCode);
      },
    }).catch(() => {
      // Field errors are mapped in onCompleted; other failures toast.
    });
  }

  return (
    <section className={root()}>
      <Card size={2} variant="soft" className={card()}>
        <BannerLifecycleSection cookieBannerKey={banner} />
        <div className={block()}>
          <div className={intro()}>
            <Heading level={2} size={4} weight="medium" highContrast>
              {t("bannerSettingsForm.title")}
            </Heading>
          </div>
          <Form className={fields()} errors={errors} onFormSubmit={handleSubmit}>
            <div className={pair()}>
              <Field required label={t("bannerSettingsForm.fields.name")} error={errors.name}>
                <TextField
                  name="name"
                  required
                  value={name}
                  disabled={isUpdating}
                  onValueChange={(value) => {
                    setName(value);
                    setErrors({});
                  }}
                />
              </Field>
              <Field
                required
                label={t("bannerSettingsForm.fields.consentExpiryDays")}
                error={errors.consentExpiryDays}
              >
                <TextField
                  name="consentExpiryDays"
                  type="number"
                  required
                  min={1}
                  value={consentExpiryDays}
                  disabled={isUpdating}
                  onValueChange={(value) => {
                    setConsentExpiryDays(value);
                    setErrors({});
                  }}
                />
              </Field>
            </div>
            {!banner.capabilities.corsless && (
              <Field label={t("bannerSettingsForm.fields.origin")}>
                <TextField name="origin" value={banner.origin} disabled />
              </Field>
            )}
            <div className={pair()}>
              <Field
                required
                label={t("bannerSettingsForm.fields.cookiePolicyUrl")}
                error={errors.cookiePolicyUrl}
              >
                <TextField
                  name="cookiePolicyUrl"
                  type="url"
                  required
                  value={cookiePolicyUrl}
                  disabled={isUpdating}
                  onValueChange={(value) => {
                    setCookiePolicyUrl(value);
                    setErrors({});
                  }}
                />
              </Field>
              <Field
                label={t("bannerSettingsForm.fields.privacyPolicyUrl")}
                error={errors.privacyPolicyUrl}
              >
                <TextField
                  name="privacyPolicyUrl"
                  type="url"
                  value={privacyPolicyUrl}
                  disabled={isUpdating}
                  onValueChange={(value) => {
                    setPrivacyPolicyUrl(value);
                    setErrors({});
                  }}
                />
              </Field>
            </div>
            <div className={pair()}>
              <Field
                required
                label={t("bannerSettingsForm.fields.defaultLanguage")}
                error={errors.defaultLanguage}
              >
                <Select
                  name="defaultLanguage"
                  value={defaultLanguage}
                  disabled={isUpdating}
                  onValueChange={(value: string | null) => {
                    if (value != null) {
                      setDefaultLanguage(value);
                      setErrors({});
                    }
                  }}
                >
                  <SelectTrigger size={2}>
                    {(value: string | null) => (
                      value != null
                        ? t(`bannerSettingsForm.languages.${LANGUAGE_KEYS[value as keyof typeof LANGUAGE_KEYS]}`)
                        : ""
                    )}
                  </SelectTrigger>
                  <SelectPopup>
                    {LANGUAGES.map(code => (
                      <SelectItem key={code} value={code}>
                        {t(`bannerSettingsForm.languages.${LANGUAGE_KEYS[code]}`)}
                      </SelectItem>
                    ))}
                  </SelectPopup>
                </Select>
              </Field>
              {banner.capabilities.tcf && (
                <div className={field()}>
                  <Field
                    required
                    label={t("bannerSettingsForm.fields.publisherCountryCode")}
                    error={errors.publisherCountryCode}
                  >
                    <TextField
                      name="publisherCountryCode"
                      required
                      minLength={2}
                      maxLength={2}
                      pattern="[A-Za-z]{2}"
                      value={publisherCountryCode}
                      disabled={isUpdating}
                      onValueChange={(value) => {
                        setPublisherCountryCode(value.toUpperCase());
                        setErrors({});
                      }}
                    />
                  </Field>
                  <Text size={1} color="faint">
                    {t("bannerSettingsForm.fields.publisherCountryCodeHelp")}
                  </Text>
                </div>
              )}
            </div>
            <div className={toggle()}>
              <Switch
                checked={resourceReportingEnabled}
                disabled={isUpdating}
                aria-label={t("bannerSettingsForm.fields.resourceReportingEnabled")}
                onCheckedChange={(checked) => {
                  setResourceReportingEnabled(checked);
                }}
              />
              <div className={toggleCopy()}>
                <Text size={2} weight="medium" highContrast>
                  {t("bannerSettingsForm.fields.resourceReportingEnabled")}
                </Text>
                <Text size={1} color="faint">
                  {t("bannerSettingsForm.fields.resourceReportingEnabledHelp")}
                </Text>
              </div>
            </div>
            <div className={actions()}>
              <Button
                type="submit"
                variant="solid"
                color="neutral"
                highContrast
                loading={isUpdating}
              >
                {t("bannerSettingsForm.actions.save")}
              </Button>
            </div>
          </Form>
        </div>
      </Card>
    </section>
  );
}
