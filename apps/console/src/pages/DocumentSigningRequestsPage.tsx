import { useEffect, useState } from "react";
import { useSearchParams } from "react-router";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Markdown,
  IconCheckmark1,
  IconCircleProgress,
} from "@probo/ui";
import { ProgressBar } from "../components/documentSigning/ProgressBar";
import { buildEndpoint } from "/providers/RelayProviders";

type Document = {
  document_version_id: string;
  title: string;
  content: string;
  signed?: boolean;
};

type DocumentSigningResponse = {
  documents: Document[];
  requesterName: string;
  requesterOrganization: string;
};

export default function DocumentSigningRequestsPage() {
  const { __ } = useTranslate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("authToken");

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [signingData, setSigningData] =
    useState<DocumentSigningResponse | null>(null);
  const [currentDocIndex, setCurrentDocIndex] = useState(0);

  useEffect(() => {
    if (!token) {
      setError(
        __("Missing signing token. Please check your URL and try again.")
      );
      setLoading(false);
      return;
    }

    async function fetchDocuments() {
      try {
        const response = await fetch(
          buildEndpoint("/api/console/v1/documents/signing-requests"),
          {
            method: "GET",
            headers: {
              Authorization: `Bearer ${token}`,
              "Content-Type": "application/json",
            },
          }
        );

        if (!response.ok) {
          throw new Error(__("Failed to fetch signing documents"));
        }

        const documents: Document[] = await response.json();

        const enhancedDocuments = documents.map((doc) => ({
          ...doc,
          signed: false,
        }));

        setSigningData({
          documents: enhancedDocuments,
          requesterName: "Requester",
          requesterOrganization: "Organization",
        });
      } catch (err) {
        setError(
          err instanceof Error ? err.message : __("An unknown error occurred")
        );
      } finally {
        setLoading(false);
      }
    }

    fetchDocuments();
  }, [token, __]);

  const handleSignDocument = async () => {
    if (!signingData || !token) return;

    const docToSign = signingData.documents[currentDocIndex];

    try {
      const response = await fetch(
        buildEndpoint(
          `/api/console/v1/documents/signing-requests/${docToSign.document_version_id}/sign`
        ),
        {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
        }
      );

      if (!response.ok) {
        throw new Error(__("Failed to sign document"));
      }

      const updatedDocs = [...signingData.documents];
      updatedDocs[currentDocIndex] = {
        ...updatedDocs[currentDocIndex],
        signed: true,
      };

      setSigningData({
        ...signingData,
        documents: updatedDocs,
      });

      if (currentDocIndex < updatedDocs.length - 1) {
        setCurrentDocIndex(currentDocIndex + 1);
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : __("Failed to sign document")
      );
    }
  };

  const handleNextDocument = () => {
    if (signingData && currentDocIndex < signingData.documents.length - 1) {
      setCurrentDocIndex(currentDocIndex + 1);
    }
  };

  const getSignedCount = () => {
    if (!signingData) return 0;
    return signingData.documents.filter((doc) => doc.signed).length;
  };

  const getProgressPercentage = () => {
    if (!signingData || signingData.documents.length === 0) return 0;
    return (getSignedCount() / signingData.documents.length) * 100;
  };

  if (loading) {
    return (
      <>
        <title>{__("Loading Signing Requests")}</title>
        <div className="flex justify-center items-center min-h-screen">
          <Card padded className="w-full max-w-3xl">
            <div className="flex items-center gap-4">
              <IconCircleProgress size={24} className="text-txt-accent" />
              <div>
                <h1 className="text-xl font-semibold">
                  {__("Loading Signing Requests")}
                </h1>
                <p className="text-txt-tertiary">
                  {__("Please wait while we fetch your documents...")}
                </p>
              </div>
            </div>
          </Card>
        </div>
      </>
    );
  }

  if (error) {
    return (
      <>
        <title>{__("Error")}</title>
        <div className="flex justify-center items-center min-h-screen">
          <Card padded className="w-full max-w-3xl">
            <h1 className="text-xl font-semibold text-red-600 mb-2">
              {__("Error")}
            </h1>
            <p className="text-txt-tertiary mb-4">{error}</p>
            <Button onClick={() => window.location.reload()}>
              {__("Try Again")}
            </Button>
          </Card>
        </div>
      </>
    );
  }

  if (!signingData || signingData.documents.length === 0) {
    return (
      <>
        <title>{__("No Documents to Sign")}</title>
        <div className="flex justify-center items-center min-h-screen">
          <Card padded className="w-full max-w-3xl">
            <h1 className="text-xl font-semibold mb-2">
              {__("No Documents to Sign")}
            </h1>
            <p className="text-txt-tertiary">
              {__(
                "There are no documents requiring your signature at this time."
              )}
            </p>
          </Card>
        </div>
      </>
    );
  }

  const currentDoc = signingData.documents[currentDocIndex];
  const isLastDocument = currentDocIndex === signingData.documents.length - 1;
  const allSigned = getSignedCount() === signingData.documents.length;

  return (
    <>
      <title>{__("Document Signing")}</title>
      <div className="container mx-auto py-10 space-y-6">
        <div className="space-y-4">
          <h1 className="text-3xl font-bold">
            {__("Document Signing Request")}
          </h1>
          <p className="text-txt-tertiary">
            {__("From")} {signingData.requesterName} {__("at")}{" "}
            {signingData.requesterOrganization}
          </p>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm text-txt-tertiary">
                {getSignedCount()} {__("of")} {signingData.documents.length}{" "}
                {__("documents signed")}
              </span>
              <span className="text-sm font-medium">
                {Math.round(getProgressPercentage())}%
              </span>
            </div>
            <ProgressBar value={getProgressPercentage()} className="h-2" />
          </div>
        </div>

        <Card padded>
          <div className="space-y-4">
            <div>
              <h2 className="text-xl font-semibold">{currentDoc.title}</h2>
              <p className="text-txt-tertiary">
                {__("Document")} {currentDocIndex + 1} {__("of")}{" "}
                {signingData.documents.length}
              </p>
            </div>

            <div className="border rounded-md p-4 min-h-[400px] bg-bg-tertiary">
              <Markdown content={currentDoc.content} />
            </div>

            <div className="flex justify-between items-center">
              {currentDoc.signed ? (
                <div className="flex items-center gap-4">
                  <div className="flex items-center gap-2 text-green-600 font-medium">
                    <IconCheckmark1 size={16} />
                    {__("Signed")}
                  </div>
                  {!isLastDocument && (
                    <Button onClick={handleNextDocument}>
                      {__("Next Document")}
                    </Button>
                  )}
                </div>
              ) : (
                <Button onClick={handleSignDocument}>
                  {__("Sign Document")}
                </Button>
              )}

              {allSigned && (
                <div className="text-green-600 font-medium">
                  {__("All documents have been signed")}
                </div>
              )}
            </div>
          </div>
        </Card>
      </div>
    </>
  );
}
