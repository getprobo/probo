import { useTranslate } from "@probo/i18n";
import { Button, Field, IconChevronLeft, useToast } from "@probo/ui";
import type { FormEventHandler } from "react";
import { useState } from "react";
import { Link, useSearchParams } from "react-router";
import { buildEndpoint } from "/providers/RelayProviders";

export default function LoginPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();

  const authMethod = searchParams.get("method");
  const initialMode = authMethod === "password" ? "password" : authMethod === "sso" ? "sso" : "default";

  const [mode, setMode] = useState<"default" | "password" | "sso">(initialMode);
  const [isLoading, setIsLoading] = useState(false);
  const [isChecking, setIsChecking] = useState(false);

  const handlePasswordLogin: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const emailValue = formData.get("email")?.toString();
    const passwordValue = formData.get("password")?.toString();

    if (!emailValue || !passwordValue) return;

    setIsLoading(true);

    try {
      const res = await fetch(buildEndpoint("/auth/login"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email: emailValue, password: passwordValue }),
      });

      if (!res.ok) {
        const error = await res.json();
        throw new Error(error.message || __("Failed to login"));
      }

      window.location.href = "/";
    } catch (e: any) {
      toast({
        title: __("Error"),
        description: e.message as string,
        variant: "error",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleSSOLogin: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const emailValue = formData.get("email")?.toString();

    if (!emailValue) return;

    setIsChecking(true);

    try {
      const res = await fetch(buildEndpoint("/auth/check-sso"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email: emailValue }),
      });

      if (!res.ok) {
        const error = await res.json();
        throw new Error(error.message || __("SSO not available for this email domain"));
      }

      const data = await res.json();

      if (data.ssoAvailable && data.samlConfigId) {
        window.location.href = buildEndpoint(
          `/auth/saml/login/${data.samlConfigId}`
        );
      } else {
        throw new Error(__("SSO not available for this email domain"));
      }
    } catch (e: any) {
      toast({
        title: __("Error"),
        description: e.message as string,
        variant: "error",
      });
    } finally {
      setIsChecking(false);
    }
  };

  const handleBack = () => {
    setMode("default");
  };

  if (mode === "default") {
    return (
      <div className="space-y-4">
        <h1 className="text-center text-2xl font-bold">
          {__("Login to your account")}
        </h1>
        <p className="text-center text-txt-tertiary mt-1 mb-6">
          {__("Choose your login method")}
        </p>

        <Button
          className="w-full"
          onClick={() => setMode("password")}
        >
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

        <Button
          variant="secondary"
          className="w-full"
          onClick={() => setMode("sso")}
        >
          {__("Login with SSO")}
        </Button>

        <div className="text-center mt-6 text-sm text-txt-secondary">
          {__("Don't have an account ?")}{" "}
          <Link to="/authentication/register" className="underline hover:text-txt-primary">
            {__("Register")}
          </Link>
        </div>

        <div className="text-center text-sm text-txt-secondary">
          {__("Forgot password?")}{" "}
          <Link
            to="/authentication/forgot-password"
            className="underline hover:text-txt-primary"
          >
            {__("Reset password")}
          </Link>
        </div>
      </div>
    );
  }

  if (mode === "password") {
    return (
      <form className="space-y-4" onSubmit={handlePasswordLogin}>
        <button
          type="button"
          onClick={handleBack}
          className="flex items-center gap-2 text-txt-secondary hover:text-txt-primary transition-colors mb-4"
        >
          <IconChevronLeft size={20} />
          <span className="text-sm">{__("Back")}</span>
        </button>

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

        <Button className="w-full" disabled={isLoading}>
          {isLoading ? __("Logging in...") : __("Login")}
        </Button>

        <div className="text-center mt-6 text-sm text-txt-secondary">
          {__("Don't have an account ?")}{" "}
          <Link to="/authentication/register" className="underline hover:text-txt-primary">
            {__("Register")}
          </Link>
        </div>

        <div className="text-center text-sm text-txt-secondary">
          {__("Forgot password?")}{" "}
          <Link
            to="/authentication/forgot-password"
            className="underline hover:text-txt-primary"
          >
            {__("Reset password")}
          </Link>
        </div>
      </form>
    );
  }

  return (
    <form className="space-y-4" onSubmit={handleSSOLogin}>
      <button
        type="button"
        onClick={handleBack}
        className="flex items-center gap-2 text-txt-secondary hover:text-txt-primary transition-colors mb-4"
      >
        <IconChevronLeft size={20} />
        <span className="text-sm">{__("Back")}</span>
      </button>

      <h1 className="text-center text-2xl font-bold">
        {__("Login with SSO")}
      </h1>
      <p className="text-center text-txt-tertiary mt-1 mb-6">
        {__("Enter your work email to continue with SSO")}
      </p>

      <Field
        required
        placeholder={__("Work Email")}
        name="email"
        type="email"
        label={__("Work Email")}
        autoFocus
      />

      <Button className="w-full" disabled={isChecking}>
        {isChecking ? __("Checking...") : __("Continue with SSO")}
      </Button>

      <div className="text-center mt-6 text-sm text-txt-secondary">
        {__("Don't have an account ?")}{" "}
        <Link to="/authentication/register" className="underline hover:text-txt-primary">
          {__("Register")}
        </Link>
      </div>
    </form>
  );
}
