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
