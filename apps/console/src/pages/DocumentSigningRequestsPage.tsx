import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  IconCircleCheck,
  IconCircleProgress,
  Logo,
  Spinner,
} from "@probo/ui";
import { clsx } from "clsx";
import { useEffect, useState } from "react";
import { useSearchParams } from "react-router";
import { useWindowSize } from "usehooks-ts";

import { PDFPreview } from "../components/documents/PDFPreview";

type Document = {
  document_version_id: string;
  title: string;
  signed?: boolean;
  organization_name: string;
};

type DocumentSigningResponse = {
  documents: Document[];
  organizationName: string;
};

export default function DocumentSigningRequestsPage() {
  const { __ } = useTranslate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");

  const [loading, setLoading] = useState(true);
  const [signing, setSigning] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [signingData, setSigningData]
    = useState<DocumentSigningResponse | null>(null);
  const [currentDocIndex, setCurrentDocIndex] = useState(0);
  const [showAllDocuments, setShowAllDocuments] = useState(false);

  const { width } = useWindowSize();
  const isMobile = width < 1100;
  const isDesktop = !isMobile;

  useEffect(() => {
    document.body.style.setProperty("overflow", "hidden");
    return () => {
      document.body.style.removeProperty("overflow");
    };
  }, []);

  useEffect(() => {
    if (!token) {
      setError(
        __("Missing signing token. Please check your URL and try again."),
      );
      setLoading(false);
      return;
    }

    async function fetchDocuments() {
      try {
        const response = await fetch(
          "/api/console/v1/documents/signing-requests",
          {
            method: "GET",
            headers: {
              "Authorization": `Bearer ${token}`,
              "Content-Type": "application/json",
            },
          },
        );

        if (!response.ok) {
          throw new Error(__("Failed to fetch signing documents"));
        }

        const documents = (await response.json()) as Document[];

        const enhancedDocuments = documents.map(doc => ({
          ...doc,
          signed: false,
        }));

        // Extract organization name from the first document
        const organizationName
          = documents.length > 0
            ? documents[0].organization_name || "Organization"
            : "Organization";

        setSigningData({
          documents: enhancedDocuments,
          organizationName,
        });
      } catch (err) {
        setError(
          err instanceof Error ? err.message : __("An unknown error occurred"),
        );
      } finally {
        setLoading(false);
      }
    }

    void fetchDocuments();
  }, [token, __]);

  const handleSignDocument = async () => {
    if (!signingData || !token) return;

    const docToSign = signingData.documents[currentDocIndex];

    setSigning(true);
    try {
      const response = await fetch(
        `/api/console/v1/documents/signing-requests/${docToSign.document_version_id}/sign`,
        {
          method: "POST",
          headers: {
            "Authorization": `Bearer ${token}`,
            "Content-Type": "application/json",
          },
        },
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

      // Collapse the document list when signing
      setShowAllDocuments(false);

      if (currentDocIndex < updatedDocs.length - 1) {
        setCurrentDocIndex(currentDocIndex + 1);
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : __("Failed to sign document"),
      );
    } finally {
      setSigning(false);
    }
  };

  const handleNextDocument = () => {
    if (signingData && currentDocIndex < signingData.documents.length - 1) {
      setCurrentDocIndex(currentDocIndex + 1);
    }
  };

  const getSignedCount = () => {
    if (!signingData) return 0;
    return signingData.documents.filter(doc => doc.signed).length;
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
                "There are no documents requiring your signature at this time.",
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

  // Build PDF URL with watermark
  const pdfUrl = token
    ? `${`/api/console/v1/documents/signing-requests/${currentDoc.document_version_id}/pdf`}?token=${encodeURIComponent(token)}`
    : null;

  return (
    <>
      <title>{__("Document Signing")}</title>
      <div className="fixed inset-0 bg-level-2 z-100 flex flex-col lg:h-screen">
        <header className="flex items-center h-12 justify-between border-b border-border-solid px-4 flex-none">
          <Logo />
        </header>
        <div className="grid lg:grid-cols-2 min-h-0 h-full">
          <div className="max-w-[440px] mx-auto py-20 overflow-y-auto scrollbar-hide">
            <h1 className="text-2xl font-semibold mb-6">
              {sprintf(
                __("%s requests your signature"),
                signingData.organizationName,
              )}
            </h1>
            {allSigned
              ? (
                  <p className="text-txt-secondary text-base">
                    {__("You have successfully signed all documents. You can now close this page.")}
                  </p>
                )
              : (
                  <>
                    <p className="text-txt-secondary text-base mb-4">
                      {__("Please review and sign the following documents:")}
                    </p>
                    <Card className="mb-6 overflow-hidden">
                      <div className="divide-y divide-border-solid">
                        {(() => {
                          const renderDocumentItem = (
                            doc: Document,
                            index: number,
                          ) => (
                            <div
                              key={doc.document_version_id}
                              className={clsx(
                                "flex items-center gap-3 py-3 px-4 transition-colors",
                                index === currentDocIndex
                                  ? "bg-blue-50 border-l-4 border-blue-500"
                                  : "bg-transparent hover:bg-level-1",
                              )}
                            >
                              <div className="flex items-center justify-center w-8 h-8 rounded-full bg-level-2 flex-shrink-0">
                                {doc.signed
                                  ? (
                                      <IconCircleCheck
                                        size={20}
                                        className="text-txt-success"
                                      />
                                    )
                                  : (
                                      <span className="text-sm font-semibold text-txt-tertiary">
                                        {index + 1}
                                      </span>
                                    )}
                              </div>
                              <div className="flex-1 min-w-0">
                                <p
                                  className={clsx(
                                    "text-sm font-medium truncate",
                                    doc.signed
                                      ? "text-txt-tertiary"
                                      : "text-txt-primary",
                                  )}
                                >
                                  {doc.title}
                                </p>
                              </div>
                              <div className="flex-shrink-0">
                                <span
                                  className={clsx(
                                    "inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium",
                                    doc.signed
                                      ? "bg-green-100 text-green-800"
                                      : index === currentDocIndex
                                        ? "bg-blue-100 text-blue-800"
                                        : "bg-gray-100 text-gray-700",
                                  )}
                                >
                                  {doc.signed
                                    ? __("Signed")
                                    : index === currentDocIndex
                                      ? __("In review")
                                      : __("Waiting signature")}
                                </span>
                              </div>
                            </div>
                          );

                          const totalDocs = signingData.documents.length;

                          if (totalDocs <= 4) {
                            return signingData.documents.map((doc, index) =>
                              renderDocumentItem(doc, index),
                            );
                          }

                          if (showAllDocuments) {
                            return (
                              <>
                                {signingData.documents.map((doc, index) =>
                                  renderDocumentItem(doc, index),
                                )}
                                <button
                                  onClick={() => setShowAllDocuments(false)}
                                  className="w-full py-3 px-4 text-sm text-txt-secondary hover:bg-level-1 transition-colors text-left flex items-center gap-2"
                                >
                                  <span className="text-txt-tertiary">•••</span>
                                  {__("Show less")}
                                </button>
                              </>
                            );
                          }

                          // Always show current document in collapsed view with two "show more" buttons
                          const firstDoc = signingData.documents[0];
                          const currentIsFirst = currentDocIndex === 0;
                          const currentIsLast = currentDocIndex === totalDocs - 1;

                          // Calculate hidden docs before and after current
                          const hiddenBeforeCurrent = currentIsFirst
                            ? 0
                            : currentDocIndex - 1;
                          const hiddenAfterCurrent = currentIsLast
                            ? 0
                            : totalDocs - currentDocIndex - 2;

                          return (
                            <>
                              {/* First document */}
                              {renderDocumentItem(firstDoc, 0)}

                              {/* Show more button for documents BEFORE current (signed documents) */}
                              {hiddenBeforeCurrent > 0 && (
                                <button
                                  onClick={() => setShowAllDocuments(true)}
                                  className="w-full py-3 px-4 text-sm text-txt-secondary hover:bg-level-1 transition-colors text-left flex items-center gap-2"
                                >
                                  <span className="text-txt-tertiary">•••</span>
                                  {sprintf(
                                    __("Show %s more documents"),
                                    hiddenBeforeCurrent,
                                  )}
                                </button>
                              )}

                              {/* Current document (if not first) */}
                              {!currentIsFirst
                                && renderDocumentItem(currentDoc, currentDocIndex)}

                              {/* Show more button for documents AFTER current (upcoming documents) */}
                              {hiddenAfterCurrent > 0 && (
                                <button
                                  onClick={() => setShowAllDocuments(true)}
                                  className="w-full py-3 px-4 text-sm text-txt-secondary hover:bg-level-1 transition-colors text-left flex items-center gap-2"
                                >
                                  <span className="text-txt-tertiary">•••</span>
                                  {sprintf(
                                    __("Show %s more documents"),
                                    hiddenAfterCurrent,
                                  )}
                                </button>
                              )}
                            </>
                          );
                        })()}
                      </div>
                    </Card>
                    <p className="text-txt-secondary text-sm mb-6">
                      {__("Please review the document carefully before signing.")}
                    </p>
                  </>
                )}
            {isMobile && pdfUrl && (
              <Button variant="secondary" asChild className="my-6 w-full">
                <a target="_blank" rel="noopener noreferrer" href={pdfUrl}>
                  {__("View document")}
                </a>
              </Button>
            )}
            {!currentDoc.signed && !allSigned && (
              <>
                <Button
                  onClick={() => void handleSignDocument()}
                  className="h-10 w-full"
                  icon={signing ? Spinner : undefined}
                  disabled={signing}
                >
                  {__("I acknowledge and agree")}
                </Button>
                <p className="text-xs text-txt-tertiary mt-2">
                  {__(
                    "By clicking 'I acknowledge and agree', your digital signature will be recorded.",
                  )}
                </p>
              </>
            )}
            {currentDoc.signed && !isLastDocument && (
              <Button onClick={handleNextDocument} className="h-10 w-full mt-4">
                {__("Next Document")}
              </Button>
            )}
            <a
              href="https://www.getprobo.com/"
              className={clsx(
                "flex gap-1 text-sm font-medium text-txt-tertiary items-center w-max mx-auto",
                isMobile ? "mt-15" : "mt-30",
              )}
            >
              Powered by
              {" "}
              <Logo withPicto className="h-6" />
            </a>
          </div>
          {isDesktop && (
            <div className="bg-subtle h-full border-l border-border-solid min-h-0">
              {pdfUrl && <PDFPreview src={pdfUrl} name={currentDoc.title} />}
            </div>
          )}
        </div>
      </div>
    </>
  );
}
