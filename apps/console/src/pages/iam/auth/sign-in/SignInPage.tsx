import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import type { FormEventHandler } from "react";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { Link, matchPath, useLocation } from "react-router";
import { graphql } from "relay-runtime";

import type { SignInPageMutation } from "#/__generated__/iam/SignInPageMutation.graphql";
import type { SignInPageQuery } from "#/__generated__/iam/SignInPageQuery.graphql";
import { useSafeContinueUrl } from "#/hooks/useSafeContinueUrl";

import { Divider } from "./_components/Divider";
import { OIDCButton } from "./_components/OIDCButton";

const signInMutation = graphql`
  mutation SignInPageMutation($input: SignInInput!) {
    signIn(input: $input) {
      session {
        id
      }
    }
  }
`;

export const signInPageQuery = graphql`
  query SignInPageQuery {
    oidcProviders {
      ...OIDCButtonFragment
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<SignInPageQuery>;
};

export default function SignInPage(props: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const location = useLocation();
  const safeContinueUrl = useSafeContinueUrl();

  const data = usePreloadedQuery<SignInPageQuery>(signInPageQuery, props.queryRef);

  const [signIn, isSigningIn]
    = useMutation<SignInPageMutation>(signInMutation);

  const handleSubmit: FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const email = (formData.get("email") as string) ?? "";
    const password = (formData.get("password") as string) ?? "";

    if (!email || !password) return;

    const match = matchPath(
      {
        path: "/organizations/:organizationId",
        caseSensitive: false,
        end: false,
      },
      safeContinueUrl.pathname,
    );

    signIn({
      variables: {
        input: {
          email,
          password,
          organizationId: match?.params.organizationId ?? null,
        },
      },
      onCompleted: (_, error) => {
        if (error) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to sign in"),
              error as GraphQLError,
            ),
            variant: "error",
          });
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
      },
    });
  };

  return (
    <div className="w-full max-w-sm mx-auto pt-8">
      <h1 className="text-2xl font-bold">
        {__("Sign in to your account")}
      </h1>

      <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
        <Field
          required
          name="email"
          type="email"
          label={__("Email")}
          autoFocus
        />

        <div>
          <div className="flex items-center justify-between mb-1">
            <label className="text-sm font-medium" htmlFor="password">
              {__("Password")}
            </label>
            <Link
              to="/auth/forgot-password"
              className="text-sm text-txt-secondary hover:text-txt-primary"
            >
              {__("Forgot your password?")}
            </Link>
          </div>
          <Field
            required
            name="password"
            id="password"
            type="password"
          />
        </div>

        <Button className="w-full h-10" disabled={isSigningIn}>
          {isSigningIn ? __("Signing in...") : __("Sign in")}
        </Button>
      </form>

      <div className="mt-6 space-y-4">
        <Divider>{__("Or")}</Divider>

        {data.oidcProviders.map((providerRef, index) => (
          <OIDCButton key={index} providerRef={providerRef} />
        ))}

        <Button
          variant="secondary"
          className="w-full h-10"
          to={{ pathname: "/auth/sso-login", search: location.search }}
        >
          {__("Sign in with SSO")}
        </Button>
      </div>

      <p className="mt-8 text-center text-sm text-txt-secondary">
        {__("New to Probo?")}
        {" "}
        <Link
          to={{ pathname: "/auth/register", search: location.search }}
          className="underline hover:text-txt-primary"
        >
          {__("Create account")}
        </Link>
      </p>
    </div>
  );
}
