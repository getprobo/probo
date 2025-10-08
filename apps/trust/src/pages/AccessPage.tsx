import { useTranslate } from "@probo/i18n";
import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { buildEndpoint } from "/providers/RelayProviders";
import { PageError } from "/components/PageError";
import { Spinner } from "@probo/ui";

/**
 * Page requested with an access token to authenticate the user for the Trust center
 */
export function AccessPage() {
  const { __ } = useTranslate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const navigate = useNavigate();

  const isValidRequest = !!token;
  const [error, setError] = useState<string | null>(() => {
    if (!token) {
      return __("Invalid access token");
    }
    return null;
  });

  // Initiate an authentication attempt
  useEffect(() => {
    if (!isValidRequest) {
      return;
    }
    fetch(buildEndpoint("/api/trust/v1/auth/authenticate"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        token,
      }),
    })
      .then((response) => {
        // For invalid response throw an error
        if (!response.ok) {
          const defaultMessage = `HTTP ${response.status}: ${response.statusText}`;
          return response
            .json()
            .then((json) => {
              throw new Error(json.message ?? defaultMessage);
            })
            .catch(() => {
              throw new Error(defaultMessage);
            });
        }
        return response.json();
      })
      .then((data) => {
        if (data.success) {
          navigate("/overview");
          return;
        }
        throw new Error(data.message ?? __("Authentication failed"));
      })
      .catch((error) => {
        setError(error.message);
      });
  }, [isValidRequest, token, __, navigate]);

  if (error) {
    return <PageError error={error} />;
  }

  return (
    <div className="p-4 text-center flex items-center justify-center gap-2">
      <Spinner size={16} />
      {__("Redirecting to trust center")}
    </div>
  );
}
