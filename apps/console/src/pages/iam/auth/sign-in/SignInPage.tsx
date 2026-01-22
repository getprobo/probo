import { useTranslate } from "@probo/i18n";
import { Button } from "@probo/ui";
import { Link } from "react-router";

export default function SignInPage() {
  const { __ } = useTranslate();

  return (
    <div className="space-y-4">
      <h1 className="text-center text-2xl font-bold">
        {__("Login to your account")}
      </h1>
      <p className="text-center text-txt-tertiary mt-1 mb-6">
        {__("Choose your login method")}
      </p>

      <Button className="w-full" to="/auth/password-login">
        {__("Login with Email")}
      </Button>

      <div className="relative my-6">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-border"></div>
        </div>
        <div className="relative flex justify-center">
          <span
            className="px-4 text-xs uppercase text-txt-secondary"
            style={{ backgroundColor: "var(--color-level-0)" }}
          >
            {__("Or")}
          </span>
        </div>
      </div>

      <Button variant="secondary" className="w-full" to="/auth/sso-login">
        {__("Login with SSO")}
      </Button>

      <div className="text-center mt-6 text-sm text-txt-secondary">
        {__("Don't have an account ?")}
        {" "}
        <Link to="/auth/register" className="underline hover:text-txt-primary">
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
