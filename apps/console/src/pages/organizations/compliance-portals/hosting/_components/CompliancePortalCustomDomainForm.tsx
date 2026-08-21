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
import { LinkSimpleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePortalCustomDomainFormMutation } from "#/__generated__/core/CompliancePortalCustomDomainFormMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { DOMAIN_MAX_LENGTH, DOMAIN_PATTERN, normalizeDomain } from "../_lib/normalizeDomain";
import { customDomainForm, domainCard, hostingCard } from "../variants";

const createCustomDomainMutation = graphql`
  mutation CompliancePortalCustomDomainFormMutation($input: CreateCustomDomainInput!) {
    createCustomDomain(input: $input) {
      customDomain {
        id
        ...CompliancePortalDomainCardFragment
      }
    }
  }
`;

export function CompliancePortalCustomDomainForm() {
  const { t } = useTranslation("organizations/compliance-portals");
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const [domain, setDomain] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { frame, header, wash, fade, icon: iconSlot, body } = hostingCard({
    tone: "sand",
    wide: true,
  });
  const { copy, lead, subtitle } = domainCard();
  const { form, field, actions } = customDomainForm();

  const [createCustomDomain, isCreating]
    = useMutation<CompliancePortalCustomDomainFormMutation>(createCustomDomainMutation, {
      successMessage: t("customDomainForm.messages.created"),
      errorToast: t("customDomainForm.errors.create"),
    });

  function handleSubmit() {
    if (compliancePortalId == null) {
      return;
    }

    const normalized = normalizeDomain(domain);
    if (normalized.length > DOMAIN_MAX_LENGTH || !DOMAIN_PATTERN.test(normalized)) {
      setErrors({ domain: t("customDomainForm.fields.domainInvalid") });
      return;
    }

    void createCustomDomain({
      variables: {
        input: {
          compliancePortalId,
          domain: normalized,
        },
      },
      updater: (store, data) => {
        const newDomainId = data?.createCustomDomain?.customDomain?.id;
        if (!newDomainId) {
          return;
        }

        const compliancePortalRecord = store.get(compliancePortalId);
        const newDomainRecord = store.get(newDomainId);
        if (compliancePortalRecord && newDomainRecord) {
          compliancePortalRecord.setLinkedRecord(newDomainRecord, "customDomain");
        }
      },
    }).then(
      () => {
        setErrors({});
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Card variant="ghost" size={2} padding="none" className={frame()}>
      <Form
        className={form()}
        errors={errors}
        onFormSubmit={handleSubmit}
      >
        <div className={header()}>
          <div className={wash()} />
          <div className={fade()} />
          <div className={lead()}>
            <div className={iconSlot()}>
              <LinkSimpleIcon size={24} weight="duotone" />
            </div>
            <Text size={2} weight="medium" color="neutral" className={subtitle()}>
              {t("customDomainForm.title")}
            </Text>
          </div>
        </div>
        <div className={body()}>
          <div className={copy()}>
            <Field error={errors.domain} className={field()}>
              <TextField
                name="domain"
                required
                value={domain}
                aria-label={t("customDomainForm.fields.domain")}
                placeholder={t("customDomainForm.fields.domainPlaceholder")}
                onValueChange={(value) => {
                  setDomain(value);
                  setErrors({});
                }}
              />
            </Field>
            <Text size={1} color="faint">
              {t("customDomainForm.examples")}
            </Text>
          </div>
          <div className={actions()}>
            <Button
              type="submit"
              variant="solid"
              color="neutral"
              highContrast
              loading={isCreating}
            >
              {t("customDomainForm.actions.add")}
            </Button>
          </div>
        </div>
      </Form>
    </Card>
  );
}
