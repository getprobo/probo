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
import { Button, Field, useToast } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { Link, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreatePasswordPageMutation } from "#/__generated__/iam/CreatePasswordPageMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const createPasswordMutation = graphql`
  mutation CreatePasswordPageMutation($input: ResetPasswordInput!) {
    resetPassword(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  password: z.string().min(8),
});

export default function CreatePasswordPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  usePageTitle(t("createPasswordPage.pageTitle"));

  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      password: "",
    },
  });

  const [createPassword, isCreatingPassword] = useMutation<CreatePasswordPageMutation>(createPasswordMutation);

  const onSubmit = (data: z.infer<typeof schema>) => {
    createPassword({
      variables: {
        input: {
          password: data.password,
          token: searchParams.get("token") ?? "",
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
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{t("createPasswordPage.title")}</h1>
        <p className="text-txt-tertiary">
          {t("createPasswordPage.description")}
        </p>
      </div>

      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <Field
          label={t("createPasswordPage.fields.password")}
          type="password"
          placeholder="••••••••"
          {...register("password")}
          required
          error={formState.errors.password?.message}
        />

        <Button type="submit" className="w-xs h-10 mx-auto mt-6" disabled={formState.isLoading || isCreatingPassword}>
          {t("createPasswordPage.actions.save")}
        </Button>
      </form>

      <div className="text-center">
        <p className="text-sm text-txt-tertiary">
          {t("createPasswordPage.alreadyHaveAccount")}
          {" "}
          <Link
            to="/auth/login"
            className="underline text-txt-primary hover:text-txt-secondary"
          >
            {t("createPasswordPage.actions.logIn")}
          </Link>
        </p>
      </div>
    </div>
  );
}
