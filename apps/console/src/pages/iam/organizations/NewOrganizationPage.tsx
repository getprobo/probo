// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import { MembershipRequiredError } from "@probo/relay";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { useLocation, useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { NewOrganizationPageMutation } from "#/__generated__/iam/NewOrganizationPageMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { newOrganizationPage } from "./variants";

const createOrganizationMutation = graphql`
  mutation NewOrganizationPageMutation($input: CreateOrganizationInput!) {
    createOrganization(input: $input) {
      organization {
        id
      }
    }
  }
`;

function NewOrganizationPageInner() {
  const location = useLocation();
  const navigate = useNavigate();
  const toast = useToast();
  const { t } = useTranslation();
  const slots = newOrganizationPage();
  const nameHelpId = useId();
  const [name, setName] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  usePageTitle(t("newOrganizationPage.title"));

  const createError = t("newOrganizationPage.errors.create");
  const [createOrganization, isCreating] = useMutation<NewOrganizationPageMutation>(
    createOrganizationMutation,
    {
      successMessage: t("newOrganizationPage.messages.created"),
      errorToast: false,
    },
  );

  function handleSubmit() {
    void createOrganization({
      variables: { input: { name: name.trim() } },
      onCompleted(_response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(fieldErrors);
          return;
        }
        if (payloadErrors != null && payloadErrors.length > 0) {
          toast.add({
            title: t("common.error"),
            description: formatError(createError, payloadErrors),
            type: "error",
          });
        }
      },
      onError(error) {
        if (error instanceof MembershipRequiredError) {
          toast.add({
            title: t("noOrganizationAccess.title"),
            description: t("noOrganizationAccess.description"),
            type: "error",
          });
          return;
        }
        toast.add({
          title: t("common.error"),
          description: formatError(createError, error),
          type: "error",
        });
      },
    }).then((response) => {
      const organizationId = response.createOrganization?.organization?.id;
      if (organizationId != null) {
        void navigate(`/organizations/${organizationId}`);
      }
    }).catch(() => {
      // Field errors stay on the form; other failures toast above.
    });
  }

  return (
    <main className={slots.root()}>
      <Heading level={1} size={6} weight="medium" highContrast>
        {t("newOrganizationPage.title")}
      </Heading>
      <Card padding={2} variant="soft">
        <Form className={slots.form()} errors={errors} onFormSubmit={handleSubmit}>
          <Text size={2} color="faint">
            {t("newOrganizationPage.description")}
          </Text>
          <div className={slots.field()}>
            <Field required label={t("newOrganizationPage.fields.name")} error={errors.name}>
              <TextField
                size={2}
                name="name"
                required
                aria-describedby={nameHelpId}
                value={name}
                disabled={isCreating}
                placeholder={t("newOrganizationPage.fields.name")}
                onValueChange={(value) => {
                  setName(value);
                  setErrors({});
                }}
              />
            </Field>
            <Text id={nameHelpId} size={1} color="faint">
              {t("newOrganizationPage.fields.nameHelp")}
            </Text>
          </div>
          <div className={slots.actions()}>
            <Button type="submit" size={2} loading={isCreating}>
              {t("newOrganizationPage.actions.create")}
            </Button>
            <ButtonLink
              to={(location.state as { from?: string } | null)?.from ?? "/"}
              size={2}
              variant="soft"
              color="neutral"
            >
              {t("newOrganizationPage.actions.back")}
            </ButtonLink>
          </div>
        </Form>
      </Card>
    </main>
  );
}

export default function NewOrganizationPage() {
  return (
    <IAMRelayProvider>
      <NewOrganizationPageInner />
    </IAMRelayProvider>
  );
}
