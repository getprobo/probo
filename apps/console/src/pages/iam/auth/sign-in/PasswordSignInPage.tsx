import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Field, IconChevronLeft, useToast } from "@probo/ui";
import type { FormEventHandler } from "react";
import { useMutation } from "react-relay";
import { Link, matchPath, useLocation, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { PasswordSignInPageMutation } from "#/__generated__/iam/PasswordSignInPageMutation.graphql";

const signInMutation = graphql`
  mutation PasswordSignInPageMutation($input: SignInInput!) {
    signIn(input: $input) {
      session {
        id
      }
    }
  }
`;

export default function PasswordSignInPage() {
  const location = useLocation();
  const [searchParams] = useSearchParams();

  const { __ } = useTranslate();
  const { toast } = useToast();

  const [signIn, isSigningIn]
    = useMutation<PasswordSignInPageMutation>(signInMutation);

  const handlePasswordLogin: FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const emailValue = formData.get("email") ? (formData.get("email") as string).toString() : "";
    const passwordValue = formData.get("password") ? (formData.get("password") as string).toString() : "";

    if (!emailValue || !passwordValue) return;

    const continueUrlParam = searchParams.get("continue");
    let safeContinueUrl: URL;
    if (continueUrlParam) {
      let continueUrl: URL;
      try {
        continueUrl = new URL(continueUrlParam, window.location.origin);
      } catch {
        continueUrl = new URL(window.location.origin);
      }
      safeContinueUrl = new URL(continueUrl.pathname + continueUrl.search, window.location.origin);
    } else {
      safeContinueUrl = new URL(window.location.origin);
    }

    const match = matchPath(
      { path: "/organizations/:organizationId", caseSensitive: false, end: false },
      safeContinueUrl.pathname,
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
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to login"),
              error as GraphQLError,
            ),
            variant: "error",
          });
          window.location.href = "/";
          return;
        }

        window.location.href = safeContinueUrl.href;
      },
      onError: (e) => {
        toast({
          title: __("Error"),
          description: e.message,
          variant: "error",
        });
        window.location.href = "/";
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
        {__("Login with Email")}
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
