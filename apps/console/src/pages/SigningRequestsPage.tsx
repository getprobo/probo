import { useEffect, useState } from "react";
import { useSearchParams } from "react-router";
import { Helmet } from "react-helmet-async";
import { buildEndpoint } from "../utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import ReactMarkdown from "react-markdown";

type Document = {
  policy_version_id: string;
  title: string;
  content: string;
  signed?: boolean;
};

type SigningResponse = {
  documents: Document[];
  requesterName: string;
  requesterOrganization: string;
};

export default function SigningRequestsPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [signingData, setSigningData] = useState<SigningResponse | null>(null);
  const [currentDocIndex, setCurrentDocIndex] = useState(0);
  
  // Fetch documents to sign using the token
  useEffect(() => {
    if (!token) {
      setError("Missing signing token. Please check your URL and try again.");
      setLoading(false);
      return;
    }

    async function fetchDocuments() {
      try {
        const response = await fetch(buildEndpoint("/api/console/v1/policies/signing-requests"), {
          method: "GET",
          headers: {
            "Authorization": `Bearer ${token}`,
            "Content-Type": "application/json",
          },
        });

        if (!response.ok) {
          throw new Error("Failed to fetch signing documents");
        }

        const documents: Document[] = await response.json();
        // Transform the API response to match our internal structure
        const enhancedDocuments = documents.map(doc => ({
          ...doc,
          signed: false
        }));
        
        setSigningData({
          documents: enhancedDocuments,
          requesterName: "Requester", // Default values as they're not in the API response
          requesterOrganization: "Organization"
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : "An unknown error occurred");
      } finally {
        setLoading(false);
      }
    }

    fetchDocuments();
  }, [token]);

  // Handle document signing
  const handleSignDocument = async () => {
    if (!signingData || !token) return;
    
    const docToSign = signingData.documents[currentDocIndex];
    
    try {
      const response = await fetch(buildEndpoint(`/api/console/v1/policies/signing-requests/${docToSign.policy_version_id}/sign`), {
        method: "POST",
        headers: {
          "Authorization": `Bearer ${token}`,
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        throw new Error("Failed to sign document");
      }

      // Update local state
      const updatedDocs = [...signingData.documents];
      updatedDocs[currentDocIndex] = {
        ...updatedDocs[currentDocIndex],
        signed: true,
      };

      setSigningData({
        ...signingData,
        documents: updatedDocs,
      });

      // Move to next document if available
      if (currentDocIndex < updatedDocs.length - 1) {
        setCurrentDocIndex(currentDocIndex + 1);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to sign document");
    }
  };

  // Handle going to next document
  const handleNextDocument = () => {
    if (signingData && currentDocIndex < signingData.documents.length - 1) {
      setCurrentDocIndex(currentDocIndex + 1);
    }
  };

  // Calculate progress
  const getSignedCount = () => {
    if (!signingData) return 0;
    return signingData.documents.filter(doc => doc.signed).length;
  };

  const getProgressPercentage = () => {
    if (!signingData || signingData.documents.length === 0) return 0;
    return (getSignedCount() / signingData.documents.length) * 100;
  };

  // Loading state
  if (loading) {
    return (
      <div className="container mx-auto py-10 flex justify-center">
        <Card className="w-full max-w-3xl">
          <CardHeader>
            <CardTitle>Loading Signing Requests</CardTitle>
            <CardDescription>Please wait while we fetch your documents...</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="container mx-auto py-10 flex justify-center">
        <Card className="w-full max-w-3xl">
          <CardHeader>
            <CardTitle>Error</CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
          <CardFooter>
            <Button onClick={() => window.location.reload()}>Try Again</Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  // No documents or missing data
  if (!signingData || signingData.documents.length === 0) {
    return (
      <div className="container mx-auto py-10 flex justify-center">
        <Card className="w-full max-w-3xl">
          <CardHeader>
            <CardTitle>No Documents to Sign</CardTitle>
            <CardDescription>There are no documents requiring your signature at this time.</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  // Current document
  const currentDoc = signingData.documents[currentDocIndex];
  const isLastDocument = currentDocIndex === signingData.documents.length - 1;
  const allSigned = getSignedCount() === signingData.documents.length;

  return (
    <div className="container mx-auto py-10">
      <Helmet>
        <title>Document Signing - Probo</title>
      </Helmet>
      
      <div className="mb-8">
        <h1 className="text-2xl font-bold mb-2">Document Signing Request</h1>
        <p className="text-muted-foreground">
          From {signingData.requesterName} at {signingData.requesterOrganization}
        </p>
        
        <div className="mt-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-muted-foreground">
              {getSignedCount()} of {signingData.documents.length} documents signed
            </span>
            <span className="text-sm font-medium">{Math.round(getProgressPercentage())}%</span>
          </div>
          <Progress value={getProgressPercentage()} className="h-2" />
        </div>
      </div>
      
      <Card className="w-full">
        <CardHeader>
          <CardTitle>{currentDoc.title}</CardTitle>
          <CardDescription>
            Document {currentDocIndex + 1} of {signingData.documents.length}
          </CardDescription>
        </CardHeader>
        
        <CardContent>
          <div className="border rounded-md p-4 min-h-[400px] bg-muted/20">
            <div className="prose prose-olive max-w-none">
              <ReactMarkdown>{currentDoc.content}</ReactMarkdown>
            </div>
          </div>
        </CardContent>
        
        <CardFooter className="flex justify-between">
          {currentDoc.signed ? (
            <div className="flex items-center gap-2">
              <div className="text-sm text-green-600 font-medium">✓ Signed</div>
              {!isLastDocument && (
                <Button onClick={handleNextDocument}>
                  Next Document
                </Button>
              )}
            </div>
          ) : (
            <Button onClick={handleSignDocument}>
              Sign Document
            </Button>
          )}
          
          {allSigned && (
            <div className="text-green-600 font-medium">
              All documents have been signed
            </div>
          )}
        </CardFooter>
      </Card>
    </div>
  );
} 