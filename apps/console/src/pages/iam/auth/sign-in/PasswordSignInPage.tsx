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

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Field, IconChevronLeft, useToast } from "@probo/ui";
import { type FormEventHandler, useState } from "react";
import { useMutation } from "react-relay";
import { Link, matchPath, useLocation } from "react-router";
import { graphql } from "relay-runtime";

import type { PasswordSignInPageMutation } from "#/__generated__/iam/PasswordSignInPageMutation.graphql";
import { usePostAuthRedirectUrl } from "#/hooks/usePostAuthRedirectUrl";

const signInMutation = graphql`
  mutation PasswordSignInPageMutation($input: SignInInput!) {
    signIn(input: $input) {
      session {
        id
      }
    }
  }
`;

function hasInvalidCredentialsError(
  error: GraphQLError | GraphQLError[] | null | undefined,
) {
  const errors = Array.isArray(error)
    ? error
    : (error?.source?.errors ?? (error ? [error] : []));

  return errors.some(error => error.extensions?.code === "INVALID_CREDENTIALS");
}

function forgotPasswordSearch(locationSearch: string, email: string) {
  const searchParams = new URLSearchParams(locationSearch);
  searchParams.set("email", email);

  return `?${searchParams.toString()}`;
}

export default function PasswordSignInPage() {
  const location = useLocation();
  const postAuthRedirectUrl = usePostAuthRedirectUrl();

  const { __ } = useTranslate();
  const { toast } = useToast();
  const [failedSignInEmail, setFailedSignInEmail] = useState<
    string | null
  >(null);

  const [signIn, isSigningIn]
    = useMutation<PasswordSignInPageMutation>(signInMutation);

  const handlePasswordLogin: FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const emailValue = formData.get("email") ? (formData.get("email") as string).toString() : "";
    const passwordValue = formData.get("password") ? (formData.get("password") as string).toString() : "";

    if (!emailValue || !passwordValue) return;

    setFailedSignInEmail(null);

    const match = matchPath(
      { path: "/organizations/:organizationId", caseSensitive: false, end: false },
      new URL(postAuthRedirectUrl, window.location.origin).pathname,
    );

    signIn({
      variables: {
        input: {
          email: emailValue,
          password: passwordValue,
          // Assume when signing in
          organizationId: match && match.params.organizationId,
        },
      },
      onCompleted: (_, error) => {
        if (error) {
          if (hasInvalidCredentialsError(error)) {
            setFailedSignInEmail(emailValue);
          }

          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to login"),
              error,
            ),
            variant: "error",
          });
          return;
        }

        window.location.href = postAuthRedirectUrl;
      },
      onError: (e) => {
        toast({
          title: __("Error"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <form className="space-y-6 w-full max-w-md mx-auto pt-4" onSubmit={handlePasswordLogin}>
      <Link
        to={{ pathname: "/auth/login", search: location.search }}
        className="flex items-center gap-2 text-txt-secondary hover:text-txt-primary transition-colors mb-4"
      >
        <IconChevronLeft size={20} />
        <span className="text-sm">{__("Back")}</span>
      </Link>

      <h1 className="text-center text-2xl font-bold">
        {__("Sign in with password")}
      </h1>
      <p className="text-center text-txt-tertiary mt-1 mb-6">
        {__("Enter your email and password")}
      </p>

      <div className="space-y-4">
        <Field
          required
          placeholder={__("Email")}
          name="email"
          type="email"
          label={__("Email")}
          autoFocus
        />

        <Field
          required
          placeholder={__("Password")}
          name="password"
          type="password"
          label={__("Password")}
        />
      </div>

      {failedSignInEmail && (
        <div
          role="alert"
          className="space-y-3 rounded-lg border border-border-low bg-subtle p-4 text-sm text-txt-secondary"
        >
          <p>
            {__(
              "Check your credentials, or set a password if you were invited and have not set one yet.",
            )}
          </p>
          <Button
            variant="secondary"
            className="w-full h-10"
            to={{
              pathname: "/auth/forgot-password",
              search: forgotPasswordSearch(location.search, failedSignInEmail),
            }}
          >
            {__("Set or reset password")}
          </Button>
        </div>
      )}

      <Button className="w-xs h-10 mx-auto mt-6" disabled={isSigningIn}>
        {isSigningIn ? __("Logging in...") : __("Login")}
      </Button>

      <div className="text-center mt-6 text-sm text-txt-secondary">
        {__("Don't have an account ?")}
        {" "}
        <Link to={{ pathname: "/auth/register", search: location.search }} className="underline hover:text-txt-primary">
          {__("Register")}
        </Link>
      </div>

      <div className="text-center text-sm text-txt-secondary">
        {__("Forgot password?")}
        {" "}
        <Link
          to="/auth/forgot-password"
          className="underline hover:text-txt-primary"
        >
          {__("Reset password")}
        </Link>
      </div>
    </form>
  );
}
