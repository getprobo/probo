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
import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CreatePasswordPageMutation } from "#/__generated__/iam/CreatePasswordPageMutation.graphql";

const createPasswordMutation = graphql`
  mutation CreatePasswordPageMutation($input: ResetPasswordInput!) {
    resetPassword(input: $input) {
      success
    }
  }
`;

export default function CreatePasswordPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  usePageTitle(t("createPasswordPage.pageTitle"));

  const [createPassword, isCreatingPassword] = useMutation<CreatePasswordPageMutation>(createPasswordMutation);
  const token = searchParams.get("token")?.trim() ?? "";

  const onSubmit = (password: string) => {
    if (token === "") {
      toast({
        title: t("createPasswordPage.errors.creationFailed"),
        description: t("createPasswordPage.errors.missingToken"),
        variant: "error",
      });
      return;
    }

    createPassword({
      variables: {
        input: {
          password,
          token,
        },
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: t("createPasswordPage.errors.creationFailed"),
            description: formatError(t("createPasswordPage.errors.creationFailed"), e),
            variant: "error",
          });
          return;
        }

        toast({
          title: t("common.success"),
          description: t("createPasswordPage.messages.created"),
          variant: "success",
        });

        searchParams.delete("token");
        void navigate({
          pathname: "/auth/password-login",
          search: "?" + searchParams.toString(),
        }, {
          replace: true,
        });
      },
      onError: (e) => {
        toast({
          title: t("createPasswordPage.errors.creationFailed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("createPasswordPage.title")}
        </Heading>
        <Text size={2} align="center" className="block">
          {t("createPasswordPage.description")}
        </Text>
      </div>

      <Form
        className="flex flex-col gap-5"
        onFormSubmit={(values) => {
          onSubmit(String(values.password ?? ""));
        }}
      >
        <Field.Root name="password" className="flex flex-col gap-1.5">
          <Field.Label className="text-1 font-medium text-sand-12">
            {t("createPasswordPage.fields.password")}
          </Field.Label>
          <TextField
            type="password"
            name="password"
            required
            minLength={8}
            placeholder="••••••••"
          />
          <Field.Error className="text-1 text-red-11" />
        </Field.Root>

        <Button
          type="submit"
          variant="solid"
          color="neutral"
          highContrast
          size={3}
          className="w-full"
          loading={isCreatingPassword}
        >
          {t("createPasswordPage.actions.save")}
        </Button>
      </Form>

      <Text align="center" size={2} className="block">
        {t("createPasswordPage.alreadyHaveAccount")}
        {" "}
        <Link to="/auth/login">
          {t("createPasswordPage.actions.logIn")}
        </Link>
      </Text>
    </div>
  );
}
