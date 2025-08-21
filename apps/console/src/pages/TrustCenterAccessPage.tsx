import { useTranslate } from "@probo/i18n";
import { useParams, useNavigate, useSearchParams } from "react-router";
import { useState, useEffect } from "react";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { PageError } from "/components/PageError";
import { buildEndpoint } from "/providers/RelayProviders";
import { IconClock, IconWarning } from "@probo/ui";

function TokenErrorPage({ error }: { error: string }) {
  const { __ } = useTranslate();

  const isExpiredToken = error.toLowerCase().includes('expired');

  return (
    <div className="min-h-screen bg-level-0 flex items-center justify-center p-4">
      <div className="max-w-md w-full text-center space-y-6">
        <div className="space-y-4">
          {isExpiredToken ? (
            <div className="space-y-3">
              <div className="inline-flex items-center justify-center w-16 h-16 bg-amber-100 rounded-full">
                <IconClock size={32} className="text-amber-600" />
              </div>
              <h1 className="text-2xl font-semibold text-txt-primary">
                {__("Access Link Expired")}
              </h1>
              <p className="text-txt-secondary">
                {__("This access link has expired. Trust center access links are valid for 7 days for security reasons.")}
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              <div className="inline-flex items-center justify-center w-16 h-16 bg-red-100 rounded-full">
                <IconWarning size={32} className="text-red-600" />
              </div>
              <h1 className="text-2xl font-semibold text-txt-primary">
                {__("Invalid Access Link")}
              </h1>
              <p className="text-txt-secondary">
                {__("This access link is not valid. It may have been revoked or the link might be incorrect.")}
              </p>
            </div>
          )}
        </div>

        <div className="bg-level-1 border border-border-low rounded-lg p-4 space-y-3">
          <h3 className="font-medium text-txt-primary">{__("What can you do?")}</h3>
          <ul className="text-sm text-txt-secondary space-y-2 text-left">
            <li className="flex items-start gap-2">
              <span className="text-primary-600 mt-1">•</span>
              <span>{__("Contact the person who sent you this link to request a new access invitation")}</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-primary-600 mt-1">•</span>
              <span>{__("Check if you received a newer email with an updated access link")}</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-primary-600 mt-1">•</span>
              <span>{__("Verify that you copied the entire link correctly from the email")}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}

export default function TrustCenterAccessPage() {
  const { __ } = useTranslate();
  const { slug } = useParams<{ slug: string }>();
  const [searchParams] = useSearchParams();
  const token = searchParams.get('authToken');
  const navigate = useNavigate();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!slug) {
      setError(__("Invalid trust center"));
      setLoading(false);
      return;
    }

    if (!token) {
      setError(__("Invalid or missing access token"));
      setLoading(false);
      return;
    }

    fetch(buildEndpoint('/api/trust/v1/auth/authenticate'), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ authToken: token }),
    })
    .then(async response => {
      if (!response.ok) {
        try {
          const errorData = await response.json();
          throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
        } catch {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
      }
      return response.json();
    })
    .then(data => {
      if (data.success) {
        navigate(`/trust/${slug}`);
      } else {
        setError(data.message || __("Authentication failed"));
        setLoading(false);
      }
    })
    .catch((error) => {
      setError(error.message || __("Authentication failed"));
      setLoading(false);
    });
  }, [slug, token, __, navigate]);

  if (loading) {
    return <PageSkeleton />;
  }

  if (error) {
    const isTokenError = error.toLowerCase().includes('token') ||
                        error.toLowerCase().includes('expired') ||
                        error.toLowerCase().includes('invalid');

    if (isTokenError) {
      return <TokenErrorPage error={error} />;
    }

    return <PageError error={error} />;
  }

  return <div>{__("Redirecting to trust center...")}</div>;
}
