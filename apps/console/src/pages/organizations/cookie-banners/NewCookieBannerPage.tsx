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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import {
  Breadcrumb,
  Button,
  Card,
  Field,
  Input,
  PageHeader,
  useToast,
} from "@probo/ui";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { NewCookieBannerPageMutation } from "#/__generated__/core/NewCookieBannerPageMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const createCookieBannerMutation = graphql`
  mutation NewCookieBannerPageMutation($input: CreateCookieBannerInput!) {
    createCookieBanner(input: $input) {
      cookieBannerEdge {
        node {
          id
        }
      }
    }
  }
`;

export default function NewCookieBannerPage() {
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();

  usePageTitle(t("newCookieBannerPage.pageTitle"));

  const [createCookieBanner, isCreating]
    = useMutation<NewCookieBannerPageMutation>(createCookieBannerMutation);

  const [name, setName] = useState("");
  const [origin, setOrigin] = useState("");
  const [cookiePolicyUrl, setCookiePolicyUrl] = useState("");
  const [privacyPolicyUrl, setPrivacyPolicyUrl] = useState("");
  const [consentExpiryDays, setConsentExpiryDays] = useState("365");

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    createCookieBanner({
      variables: {
        input: {
          organizationId,
          name,
          origin,
          cookiePolicyUrl,
          privacyPolicyUrl: privacyPolicyUrl || undefined,
          consentExpiryDays: parseInt(consentExpiryDays, 10),
        },
      },
      onCompleted(data) {
        toast({
          title: t("newCookieBannerPage.messages.successTitle"),
          description: t("newCookieBannerPage.messages.created"),
          variant: "success",
        });
        const bannerId = data.createCookieBanner.cookieBannerEdge.node.id;
        void navigate(`/organizations/${organizationId}/cookie-banners/${bannerId}`);
      },
      onError(error) {
        toast({
          title: t("newCookieBannerPage.errors.title"),
          description: formatError(t("newCookieBannerPage.errors.create"), error),
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("newCookieBannerPage.breadcrumbs.index"),
            to: `/organizations/${organizationId}/cookie-banners`,
          },
          {
            label: t("newCookieBannerPage.breadcrumbs.new"),
          },
        ]}
      />
      <PageHeader
        title={t("newCookieBannerPage.title")}
        description={t("newCookieBannerPage.description")}
      />
      <Card padded asChild>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Field label={t("newCookieBannerPage.fields.name")}>
            <Input
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder={t("newCookieBannerPage.fields.namePlaceholder")}
              required
            />
          </Field>

          <Field label={t("newCookieBannerPage.fields.origin")}>
            <Input
              value={origin}
              onChange={e => setOrigin(e.target.value)}
              placeholder={t("newCookieBannerPage.fields.originPlaceholder")}
              required
            />
          </Field>

          <Field label={t("newCookieBannerPage.fields.cookiePolicyUrl")}>
            <Input
              value={cookiePolicyUrl}
              onChange={e => setCookiePolicyUrl(e.target.value)}
              placeholder={t("newCookieBannerPage.fields.cookiePolicyUrlPlaceholder")}
              required
            />
          </Field>

          <Field label={t("newCookieBannerPage.fields.privacyPolicyUrl")}>
            <Input
              value={privacyPolicyUrl}
              onChange={e => setPrivacyPolicyUrl(e.target.value)}
              placeholder={t("newCookieBannerPage.fields.privacyPolicyUrlPlaceholder")}
            />
          </Field>

          <Field label={t("newCookieBannerPage.fields.consentExpiryDays")}>
            <Input
              type="number"
              value={consentExpiryDays}
              onChange={e => setConsentExpiryDays(e.target.value)}
              min="1"
              required
            />
          </Field>

          <Button type="submit" disabled={isCreating}>
            {isCreating ? t("newCookieBannerPage.actions.creating") : t("newCookieBannerPage.actions.create")}
          </Button>
        </form>
      </Card>
    </div>
  );
}
