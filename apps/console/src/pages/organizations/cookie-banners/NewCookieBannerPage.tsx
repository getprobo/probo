// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Card,
  Field,
  Input,
  Label,
  Option,
  PageHeader,
  Select,
  useToast,
} from "@probo/ui";
import { type FormEvent, useState } from "react";
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
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();

  usePageTitle(__("New Cookie Banner"));

  const [createCookieBanner, isCreating]
    = useMutation<NewCookieBannerPageMutation>(createCookieBannerMutation);

  const [name, setName] = useState("");
  const [origin, setOrigin] = useState("");
  const [cookiePolicyUrl, setCookiePolicyUrl] = useState("");
  const [privacyPolicyUrl, setPrivacyPolicyUrl] = useState("");
  const [consentExpiryDays, setConsentExpiryDays] = useState("365");
  const [consentMode, setConsentMode] = useState("OPT_IN");

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
          consentMode: consentMode as "OPT_IN" | "OPT_OUT",
        },
      },
      onCompleted(data) {
        toast({
          title: __("Success"),
          description: __("Cookie banner created successfully"),
          variant: "success",
        });
        const bannerId = data.createCookieBanner.cookieBannerEdge.node.id;
        void navigate(`/organizations/${organizationId}/cookie-banners/${bannerId}`);
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to create cookie banner"), error as GraphQLError),
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
            label: __("Cookie Banners"),
            to: `/organizations/${organizationId}/cookie-banners`,
          },
          {
            label: __("New"),
          },
        ]}
      />
      <PageHeader
        title={__("Create Cookie Banner")}
        description={__(
          "Set up a new cookie consent banner with its origin URL and consent configuration.",
        )}
      />
      <Card padded asChild>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Field label={__("Name")}>
            <Input
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder={__("My Website")}
              required
            />
          </Field>

          <Field label={__("Origin URL")}>
            <Input
              value={origin}
              onChange={e => setOrigin(e.target.value)}
              placeholder="https://example.com"
              required
            />
          </Field>

          <Field label={__("Cookie Policy URL")}>
            <Input
              value={cookiePolicyUrl}
              onChange={e => setCookiePolicyUrl(e.target.value)}
              placeholder="https://example.com/cookies"
              required
            />
          </Field>

          <Field label={__("Privacy Policy URL")}>
            <Input
              value={privacyPolicyUrl}
              onChange={e => setPrivacyPolicyUrl(e.target.value)}
              placeholder="https://example.com/privacy"
            />
          </Field>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{__("Consent Expiry (days)")}</Label>
              <Input
                type="number"
                value={consentExpiryDays}
                onChange={e => setConsentExpiryDays(e.target.value)}
                min="1"
                required
              />
            </div>

            <div className="space-y-2">
              <Label>{__("Consent Mode")}</Label>
              <Select value={consentMode} onValueChange={setConsentMode}>
                <Option value="OPT_IN">{__("Opt-in")}</Option>
                <Option value="OPT_OUT">{__("Opt-out")}</Option>
              </Select>
            </div>
          </div>

          <Button type="submit" disabled={isCreating}>
            {isCreating ? __("Creating...") : __("Create Banner")}
          </Button>
        </form>
      </Card>
    </div>
  );
}
