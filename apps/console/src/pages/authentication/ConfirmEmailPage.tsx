import { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router";
import { Helmet } from "react-helmet-async";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { Link } from "react-router";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import type { PayloadError } from "relay-runtime";
import type { ConfirmEmailPageMutation } from "./__generated__/ConfirmEmailPageMutation.graphql";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const ConfirmEmailMutation = graphql`
  mutation ConfirmEmailPageMutation($input: ConfirmEmailInput!) {
    confirmEmail(input: $input) {
      success
    }
  }
`;

export default function ConfirmEmailPage() {
  const [isLoading, setIsLoading] = useState(false);
  const [isConfirmed, setIsConfirmed] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [token, setToken] = useState<string>("");
  const location = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [commitMutation] =
    useMutation<ConfirmEmailPageMutation>(ConfirmEmailMutation);

  useEffect(() => {
    // Extract token from URL and prefill the form
    const searchParams = new URLSearchParams(location.search);
    const urlToken = searchParams.get("token");

    if (urlToken) {
      setToken(urlToken);
    }
  }, [location.search]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    if (!token.trim()) {
      setError("Please enter a confirmation token");
      setIsLoading(false);
      return;
    }

    try {
      commitMutation({
        variables: {
          input: {
            token: token.trim(),
          },
        },
        onCompleted: (response, errors: PayloadError[] | null) => {
          if (errors) {
            throw new Error(errors[0]?.message || "Failed to confirm email");
          }

          setIsConfirmed(true);
          toast({
            title: "Success",
            description: "Your email has been confirmed successfully",
          });

          setIsLoading(false);
        },
        onError: (err) => {
          setError(err.message || "Failed to confirm email. Please try again.");
          setIsLoading(false);
        },
      });
    } catch (error) {
      setError(
        error instanceof Error
          ? error.message
          : "Failed to confirm email. Please try again.",
      );
      setIsLoading(false);
    }
  };

  return (
    <>
      <Helmet>
        <title>Confirm Email - Probo</title>
      </Helmet>

      <div className="flex flex-col items-center justify-center min-h-[70vh] p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="text-2xl font-bold text-center">
              Email Confirmation
            </CardTitle>
            <CardDescription className="text-center">
              Confirm your email address to complete registration
            </CardDescription>
          </CardHeader>

          <CardContent>
            {isConfirmed ? (
              <div className="space-y-4 text-center">
                <p className="text-green-600 dark:text-green-400">
                  Your email has been confirmed successfully!
                </p>
                <Button onClick={() => navigate("/login")} className="w-full">
                  Proceed to Login
                </Button>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-4">
                {error && (
                  <div className="p-3 text-sm text-red-600 bg-red-50 dark:bg-red-900/20 dark:text-red-400 rounded-md">
                    {error}
                  </div>
                )}

                <div className="space-y-2">
                  <Label htmlFor="token">Confirmation Token</Label>
                  <Input
                    id="token"
                    type="text"
                    value={token}
                    onChange={(e) => setToken(e.target.value)}
                    placeholder="Enter your confirmation token"
                    disabled={isLoading}
                    required
                  />
                  <p className="text-xs text-secondary dark:text-tertiary">
                    The token has been automatically filled from the URL if
                    available
                  </p>
                </div>

                <Button type="submit" className="w-full" disabled={isLoading}>
                  {isLoading ? "Confirming..." : "Confirm Email"}
                </Button>
              </form>
            )}
          </CardContent>

          <CardFooter className="flex justify-center">
            {!isConfirmed && (
              <Link
                to="/login"
                className="text-sm text-secondary hover:text-primary dark:text-tertiary dark:hover:text-quaternary"
              >
                Back to Login
              </Link>
            )}
          </CardFooter>
        </Card>
      </div>
    </>
  );
}
