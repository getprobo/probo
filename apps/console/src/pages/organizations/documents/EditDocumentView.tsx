import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import {
  graphql,
  useMutation,
  usePreloadedQuery,
  useQueryLoader,
  PreloadedQuery,
} from "react-relay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { toast } from "@/hooks/use-toast";
import { Suspense } from "react";
import type { EditDocumentViewQuery } from "./__generated__/EditDocumentViewQuery.graphql";
import type { EditDocumentViewMutation as EditDocumentViewMutationType } from "./__generated__/EditDocumentViewMutation.graphql";
import { PageTemplate } from "@/components/PageTemplate";
import { EditDocumentViewSkeleton } from "./EditDocumentPage";

const editDocumentViewQuery = graphql`
  query EditDocumentViewQuery(
    $documentId: ID!
    $organizationId: ID!
    $documentVersionId: ID!
  ) {
    documentVersion: node(id: $documentVersionId) {
      id
      ... on DocumentVersion {
        content
      }
    }

    document: node(id: $documentId) {
      id
      ... on Document {
        title
        owner {
          id
          fullName
        }
      }
    }
    organization: node(id: $organizationId) {
      ...PeopleSelector_organization
    }
  }
`;

const UpdateDocumentMutation = graphql`
  mutation EditDocumentViewMutation($input: UpdateDocumentVersionInput!) {
    updateDocumentVersion(input: $input) {
      documentVersion {
        id
        content
      }
    }
  }
`;

function EditDocumentViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<EditDocumentViewQuery>;
}) {
  const navigate = useNavigate();
  const { organizationId, documentId, versionId } = useParams();
  const data = usePreloadedQuery<EditDocumentViewQuery>(
    editDocumentViewQuery,
    queryRef,
  );

  const [content, setContent] = useState(data.documentVersion.content || "");

  const [updateDocument, isSubmitting] =
    useMutation<EditDocumentViewMutationType>(UpdateDocumentMutation);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    updateDocument({
      variables: {
        input: {
          documentVersionId: data.documentVersion.id,
          content,
        },
      },
      onCompleted: (_, errors) => {
        if (errors) {
          console.error("Error updating document:", errors);
          toast({
            title: "Error",
            description: "Failed to update document. Please try again.",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Document updated successfully.",
        });

        navigate(`/organizations/${organizationId}/documents/${documentId}`);
      },
      onError: (error) => {
        console.error("Error updating document:", error);
        toast({
          title: "Error",
          description: "Failed to update document. Please try again.",
        });
      },
    });
  };

  return (
    <PageTemplate title="Update Document" description="Update an existing document">
      <form onSubmit={handleSubmit}>
        <div className="grid gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Document Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="content">Document Content</Label>
                <div className="min-h-[500px]">
                  <Textarea
                    id="content"
                    placeholder="Enter document description"
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    required
                    rows={20}
                    className="min-h-[500px] resize-y"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <div className="flex justify-end gap-4">
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                navigate(
                  `/organizations/${organizationId}/documents/${documentId}`,
                )
              }
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Updating..." : "Update Document"}
            </Button>
          </div>
        </div>
      </form>
    </PageTemplate>
  );
}

export default function EditDocumentView() {
  const { organizationId, documentId, versionId } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<EditDocumentViewQuery>(editDocumentViewQuery);

  useEffect(() => {
    loadQuery({
      organizationId: organizationId!,
      documentId: documentId!,
      documentVersionId: versionId!,
    });
  }, [organizationId, documentId, versionId]);

  if (!queryRef) {
    return <EditDocumentViewSkeleton />;
  }

  return (
    <Suspense fallback={<EditDocumentViewSkeleton />}>
      <EditDocumentViewContent queryRef={queryRef} />
    </Suspense>
  );
}
