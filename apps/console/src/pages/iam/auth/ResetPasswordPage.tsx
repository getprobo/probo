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

import type { ResetPasswordPageMutation } from "#/__generated__/iam/ResetPasswordPageMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const resetPasswordMutation = graphql`
  mutation ResetPasswordPageMutation($input: ResetPasswordInput!) {
    resetPassword(input: $input) {
      success
    }
  }
`;

const schema = z
  .object({
    password: z.string().min(8),
    confirmPassword: z.string().min(8),
  })
  .refine(data => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

export default function ResetPasswordPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");

  usePageTitle(t("resetPasswordPage.pageTitle"));

  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      password: "",
      confirmPassword: "",
    },
  });

  const [resetPassword] = useMutation<ResetPasswordPageMutation>(
    resetPasswordMutation,
  );

  const onSubmit = handleSubmit((data) => {
    if (!token) {
      toast({
        title: t("resetPasswordPage.errors.resetFailed"),
        description: t("resetPasswordPage.errors.invalidToken"),
        variant: "error",
      });
      return;
    }

    resetPassword({
      variables: {
        input: {
          password: data.password,
          token,
        },
      },
      onError: (e: Error) => {
        toast({
          title: t("resetPasswordPage.errors.resetFailed"),
          description: e.message,
          variant: "error",
        });
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: t("resetPasswordPage.errors.resetFailed"),
            description: formatError(
              t("resetPasswordPage.errors.reset"),
              e,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("common.success"),
          description: t("resetPasswordPage.messages.reset"),
          variant: "success",
        });
        void navigate("/auth/login", { replace: true });
      },
    });
  });

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{t("resetPasswordPage.title")}</h1>
        <p className="text-txt-tertiary">
          {t("resetPasswordPage.description")}
        </p>
      </div>

      <form onSubmit={e => void onSubmit(e)} className="space-y-4">
        <Field
          label={t("resetPasswordPage.fields.newPassword")}
          type="password"
          placeholder="••••••••"
          {...register("password")}
          required
          error={formState.errors.password?.message}
        />

        <Field
          label={t("resetPasswordPage.fields.confirmPassword")}
          type="password"
          placeholder="••••••••"
          {...register("confirmPassword")}
          required
          error={formState.errors.confirmPassword?.message}
        />

        <Button type="submit" className="w-xs h-10 mx-auto mt-6" disabled={formState.isLoading}>
          {formState.isLoading
            ? t("resetPasswordPage.actions.resetting")
            : t("resetPasswordPage.actions.reset")}
        </Button>
      </form>

      <div className="text-center">
        <p className="text-sm text-txt-tertiary">
          {t("resetPasswordPage.rememberPassword")}
          {" "}
          <Link
            to="/auth/login"
            className="underline text-txt-primary hover:text-txt-secondary"
          >
            {t("resetPasswordPage.actions.logIn")}
          </Link>
        </p>
      </div>
    </div>
  );
}
