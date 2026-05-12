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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useEffect } from "react";
import { useMutation, usePreloadedQuery, useQueryLoader } from "react-relay";
import { Link, useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { SignUpPageMutation } from "#/__generated__/iam/SignUpPageMutation.graphql";
import type { SignUpPageQuery } from "#/__generated__/iam/SignUpPageQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const signUpPageQuery = graphql`
  query SignUpPageQuery {
    signUpEnabled
  }
`;

const signUpMutation = graphql`
  mutation SignUpPageMutation($input: SignUpInput!) {
    signUp(input: $input) {
      identity {
        id
      }
    }
  }
`;

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
  fullName: z.string().min(2),
});

type FormData = z.infer<typeof schema>;

function SignUpPageContent(props: { queryRef: NonNullable<ReturnType<typeof useQueryLoader<SignUpPageQuery>>[0]> }) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();

  usePageTitle(__("Sign up"));

  const data = usePreloadedQuery<SignUpPageQuery>(signUpPageQuery, props.queryRef);

  const { register, handleSubmit, formState } = useFormWithSchema(schema, {
    defaultValues: {
      email: "",
      password: "",
      fullName: "",
    },
  });

  const [signUp] = useMutation<SignUpPageMutation>(signUpMutation);

  const onSubmit = (data: FormData) => {
    signUp({
      variables: {
        input: {
          email: data.email,
          password: data.password,
          fullName: data.fullName,
        },
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: __("Sign up failed"),
            description: formatError(__("Sign up failed"), e),
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __("Account created successfully"),
          variant: "success",
        });
        void navigate("/", { replace: true });
      },
      onError: (e) => {
        toast({
          title: __("Sign up failed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  if (!data.signUpEnabled) {
    return (
      <div className="space-y-6 w-full max-w-md mx-auto pt-8 text-center">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold">{__("Registration unavailable")}</h1>
          <p className="text-txt-tertiary">
            {__("New account registration is currently disabled. Please contact your administrator or reach out to Probo for assistance.")}
          </p>
        </div>

        <div>
          <Button
            variant="secondary"
            className="w-xs h-10 mx-auto"
            to="/auth/login"
          >
            {__("Back to login")}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Sign up")}</h1>
        <p className="text-txt-tertiary">
          {__("Enter your information to create an account")}
        </p>
      </div>

      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <Field
          label={__("Full Name")}
          type="text"
          placeholder={__("John Doe")}
          {...register("fullName")}
          required
          error={formState.errors.fullName?.message}
        />

        <Field
          label={__("Email")}
          type="email"
          placeholder={__("name@example.com")}
          {...register("email")}
          required
          error={formState.errors.email?.message}
        />

        <Field
          label={__("Password")}
          type="password"
          placeholder="••••••••"
          {...register("password")}
          required
          error={formState.errors.password?.message}
        />

        <Button type="submit" className="w-xs h-10 mx-auto mt-6" disabled={formState.isLoading}>
          {formState.isLoading
            ? __("Creating account...")
            : __("Sign up with email")}
        </Button>
      </form>

      <div className="text-center">
        <p className="text-sm text-txt-tertiary">
          {__("Already have an account?")}
          {" "}
          <Link
            to="/auth/login"
            className="underline text-txt-primary hover:text-txt-secondary"
          >
            {__("Log in here")}
          </Link>
        </p>
      </div>
    </div>
  );
}

export default function SignUpPage() {
  const [queryRef, loadQuery] = useQueryLoader<SignUpPageQuery>(signUpPageQuery);

  useEffect(() => {
    loadQuery({});
  }, [loadQuery]);

  if (!queryRef) return null;

  return <SignUpPageContent queryRef={queryRef} />;
}
