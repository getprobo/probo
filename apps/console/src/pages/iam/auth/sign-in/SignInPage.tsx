import { useTranslate } from "@probo/i18n";
import { Button } from "@probo/ui";
import { Link, useLocation } from "react-router";

export default function SignInPage() {
  const { __ } = useTranslate();

  const location = useLocation();

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <h1 className="text-center text-2xl font-bold">
        {__("Login to your account")}
      </h1>
      <p className="text-center text-txt-tertiary mt-1 mb-6">
        {__("Choose your login method")}
      </p>

      <Button
        className="w-xs h-10 mx-auto"
        to={{ pathname: "/auth/password-login", search: location.search }}
      >
        {__("Login with Email")}
      </Button>

      <div className="relative my-6 w-full">
        <div className="w-xs border-t border-border-mid mx-auto" />
        <span
          className="px-4 text-xs uppercase text-txt-secondary bg-level-0 absolute top-0 left-1/2 -translate-1/2"
        >
          {__("Or")}
        </span>
      </div>

      <Button
        variant="secondary"
        className="w-xs h-10 mx-auto"
        to={{ pathname: "/auth/sso-login", search: location.search }}
      >
        {__("Login with SSO")}
      </Button>

      <div className="text-center mt-6 text-sm text-txt-secondary">
        {__("Don't have an account ?")}
        {" "}
        <Link
          to={{ pathname: "/auth/register", search: location.search }}
          className="underline hover:text-txt-primary"
        >
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
    </div>
  );
}
