import { useTranslate } from "@probo/i18n";
import { Button, Field, IconChevronLeft, useToast } from "@probo/ui";
import type { FormEventHandler } from "react";
import { useMutation } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";
import type { PasswordSignInPageMutation } from "/__generated__/iam/PasswordSignInPageMutation.graphql";

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
  const { __ } = useTranslate();

  const { toast } = useToast();

  const [signIn, isSigningIn] =
    useMutation<PasswordSignInPageMutation>(signInMutation);

  const handlePasswordLogin: FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const emailValue = formData.get("email")?.toString();
    const passwordValue = formData.get("password")?.toString();

    if (!emailValue || !passwordValue) return;

    signIn({
      variables: {
        input: {
          email: emailValue,
          password: passwordValue,
        },
      },
      onCompleted: () => {
        window.location.href = "/";
      },
      onError: (e) => {
        toast({
          title: __("Error"),
          description: e instanceof Error ? e.message : __("Failed to login"),
          variant: "error",
        });
      },
    });
  };

  return (
    <form className="space-y-4" onSubmit={handlePasswordLogin}>
      <Link
        to="/auth/login"
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

      <Button className="w-full" disabled={isSigningIn}>
        {isSigningIn ? __("Logging in...") : __("Login")}
      </Button>

      <div className="text-center mt-6 text-sm text-txt-secondary">
        {__("Don't have an account ?")}{" "}
        <Link to="/auth/register" className="underline hover:text-txt-primary">
          {__("Register")}
        </Link>
      </div>

      <div className="text-center text-sm text-txt-secondary">
        {__("Forgot password?")}{" "}
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
