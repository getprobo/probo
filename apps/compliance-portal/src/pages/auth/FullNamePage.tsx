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

import { Field } from "@base-ui/react/field";
import { Form } from "@base-ui/react/form";
import { Toast } from "@base-ui/react/toast";
import type { GraphQLError } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import { useSafeContinueUrl } from "#/lib/auth/useSafeContinueUrl";
import { useMutation } from "#/lib/relay/useMutation";

import type { FullNamePageMutation } from "./__generated__/FullNamePageMutation.graphql";

const updateFullNameMutation = graphql`
  mutation FullNamePageMutation($input: UpdateFullNameInput!) {
    updateFullName(input: $input) {
      success
    }
  }
`;

// Post-sign-in gate collecting the user's display name when the backend reports
// FULL_NAME_REQUIRED, then forwarding to the validated continue URL.
export default function FullNamePage() {
  const { t } = useTranslation();
  const toast = Toast.useToastManager();
  const safeContinueUrl = useSafeContinueUrl();

  const [updateFullName, isUpdating] = useMutation<FullNamePageMutation>(
    updateFullNameMutation,
    { errorToast: false },
  );

  const handleSubmit = (fullName: string) => {
    void updateFullName({
      variables: { input: { fullName } },
      onCompleted: (_response, errors) => {
        const code = (errors?.[0] as GraphQLError | undefined)?.extensions?.code;

        if (code === "ALREADY_AUTHENTICATED" || (errors == null || errors.length === 0)) {
          window.location.href = safeContinueUrl;
          return;
        }

        toast.add({ title: t("auth.errors.fullNameFailed"), type: "error" });
      },
      onError: () => {
        toast.add({ title: t("auth.errors.fullNameFailed"), type: "error" });
      },
    }).catch(() => {});
  };

  return (
    <div className="flex flex-col gap-6">
      <Heading level={1} size={5} align="center">
        {t("auth.fullName.title")}
      </Heading>

      <Form
        className="flex flex-col gap-6"
        onFormSubmit={(values) => {
          handleSubmit(String(values.fullName ?? ""));
        }}
      >
        <Field.Root name="fullName" className="flex flex-col gap-1.5">
          <Field.Label className="text-1 font-medium text-sand-12">
            {t("auth.fullName.label")}
          </Field.Label>
          <TextField
            type="text"
            name="fullName"
            required
            minLength={2}
            placeholder={t("auth.fullName.placeholder")}
          />
          <Field.Error className="text-1 text-red-11" match="valueMissing">
            {t("auth.fullName.required")}
          </Field.Error>
          <Field.Error className="text-1 text-red-11" match="tooShort">
            {t("auth.fullName.tooShort")}
          </Field.Error>
        </Field.Root>

        <Button type="submit" variant="solid" color="neutral" highContrast loading={isUpdating}>
          {t("auth.fullName.submit")}
        </Button>
      </Form>
    </div>
  );
}
