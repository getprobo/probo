import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { Button, Field, useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useMutation } from "@tanstack/react-query";
import { buildEndpoint } from "/providers/RelayProviders";

interface RegisterData {
  email: string;
  password: string;
  fullName: string;
}

export default function RegisterPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fullName, setFullName] = useState("");
  const { __ } = useTranslate();
  const { toast } = useToast();
  const navigate = useNavigate();

  const registerUser = async (data: RegisterData) => {
    const response = await fetch(
      buildEndpoint("/api/console/v1/auth/register"),
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(data),
      }
    );

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return {
        success: false,
        error: errorData.message || __("Registration failed"),
      };
    }

    return { success: true, data: await response.json() };
  };

  const registerMutation = useMutation({
    mutationFn: registerUser,
    onSuccess: (result) => {
      if (result.success) {
        toast({
          title: __("Success"),
          description: __("Account created successfully"),
          variant: "success",
        });
        navigate("/", { replace: true });
      } else {
        toast({
          title: __("Error"),
          description: result.error,
          variant: "error",
        });
      }
    },
    onError: () => {
      toast({
        title: __("Error"),
        description: __("Registration failed"),
        variant: "error",
      });
    },
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!email || !password || !fullName) {
      toast({
        title: __("Error"),
        description: __("Full name, email, and password are required"),
        variant: "error",
      });
      return;
    }

    registerMutation.mutate({
      email,
      password,
      fullName,
    });
  };

  return (
    <>
      <title>{__("Sign up")} - Probo</title>

      <div className="space-y-6 w-full max-w-md mx-auto">
        <div className="space-y-2 text-center">
          <h1 className="text-3xl font-bold">{__("Sign up")}</h1>
          <p className="text-txt-tertiary">
            {__("Enter your information to create an account")}
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <Field
            label={__("Full Name")}
            type="text"
            placeholder={__("John Doe")}
            value={fullName}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              setFullName(e.target.value)
            }
            required
          />

          <Field
            label={__("Email")}
            type="email"
            placeholder={__("name@example.com")}
            value={email}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              setEmail(e.target.value)
            }
            required
          />

          <Field
            label={__("Password")}
            type="password"
            placeholder="••••••••"
            value={password}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              setPassword(e.target.value)
            }
            required
          />

          <Button
            type="submit"
            className="w-full"
            disabled={registerMutation.isPending}
          >
            {registerMutation.isPending
              ? __("Creating account...")
              : __("Sign up with email")}
          </Button>
        </form>

        <div className="text-center">
          <p className="text-sm text-txt-tertiary">
            {__("Already have an account?")}{" "}
            <Link
              to="/auth/login"
              className="underline text-txt-primary hover:text-txt-secondary"
            >
              {__("Log in here")}
            </Link>
          </p>
        </div>
      </div>
    </>
  );
}
