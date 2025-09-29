import { useTranslate } from "@probo/i18n";
import { useEffect, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router";
import { buildEndpoint } from "/providers/RelayProviders";
import { PageError } from "/components/PageError";
import { Spinner } from "@probo/ui";

/**
 * Page requested with an access token to authenticate the user for the Trust center
 */
export function AccessPage() {
  const { __ } = useTranslate();
  const { slug } = useParams<{ slug: string }>();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const navigate = useNavigate();

  const isValidRequest = !!(slug && token);
  const [error, setError] = useState<string | null>(() => {
    if (!slug) {
      return __("Invalid trust center");
    }
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
          navigate(`/trust/${slug}`);
          return;
        }
        throw new Error(data.message ?? __("Authentication failed"));
      })
      .catch((error) => {
        setError(error.message);
      });
  }, [isValidRequest, slug, token, __, navigate]);

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
