import { Suspense, useEffect, useState, useCallback, useRef } from "react";
import { useParams, useNavigate } from "react-router";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
  useFragment,
} from "react-relay";
import {
  Clock,
  Download,
  Edit,
  Trash2,
  MoreHorizontal,
  FileSignature,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { PageTemplate } from "@/components/PageTemplate";
import type { ShowPolicyViewQuery } from "./__generated__/ShowPolicyViewQuery.graphql";
import { type ShowPolicyViewPublishMutation } from "./__generated__/ShowPolicyViewPublishMutation.graphql";
import { type ShowPolicyViewCreateDraftMutation } from "./__generated__/ShowPolicyViewCreateDraftMutation.graphql";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ShowPolicyViewSkeleton } from "./ShowPolicyPage";
import { format } from "date-fns";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { SignaturesModal } from "./SignaturesModal";
import { VersionHistoryModal } from "./VersionHistoryModal";
import rehypeRaw from "rehype-raw";
import { createRoot } from "react-dom/client";
import { policyVersionsFragment } from "./SignaturesModal";
import type { SignaturesModal_policyVersions$key } from "./__generated__/SignaturesModal_policyVersions.graphql";

const policyViewQuery = graphql`
  query ShowPolicyViewQuery($policyId: ID!, $organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        name
      }
    }

    node(id: $policyId) {
      id
      ... on Policy {
        title
        description
        createdAt
        updatedAt
        currentPublishedVersion
        owner {
          id
          fullName
          primaryEmailAddress
        }

        ...SignaturesModal_policyVersions
        ...VersionHistoryModal_policyVersions

        latestVersion: versions(first: 1) {
          edges {
            node {
              id
              version
              status
              content
              changelog
              publishedAt
              publishedBy {
                fullName
              }
              createdAt
              updatedAt
            }
          }
        }
      }
    }
  }
`;

const publishPolicyVersionMutation = graphql`
  mutation ShowPolicyViewPublishMutation($input: PublishPolicyVersionInput!) {
    publishPolicyVersion(input: $input) {
      policy {
        id
        currentPublishedVersion
      }
      policyVersion {
        id
        status
        publishedAt
        publishedBy {
          fullName
        }
      }
    }
  }
`;

const createDraftPolicyVersionMutation = graphql`
  mutation ShowPolicyViewCreateDraftMutation(
    $input: CreateDraftPolicyVersionInput!
  ) {
    createDraftPolicyVersion(input: $input) {
      policyVersionEdge {
        node {
          id
          version
          status
        }
      }
    }
  }
`;

function ShowPolicyContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<ShowPolicyViewQuery>;
}) {
  const data = usePreloadedQuery<ShowPolicyViewQuery>(
    policyViewQuery,
    queryRef,
  );
  const policy = data.node;
  const { organizationId } = useParams();

  const navigate = useNavigate();
  const { toast } = useToast();
  const [queryRef2, loadQuery] =
    useQueryLoader<ShowPolicyViewQuery>(policyViewQuery);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isVersionHistoryOpen, setIsVersionHistoryOpen] = useState(false);
  const [isSignaturesModalOpen, setIsSignaturesModalOpen] = useState(false);
  const printContentRef = useRef<HTMLDivElement>(null);

  const [publishDraft, isPublishInFlight] =
    useMutation<ShowPolicyViewPublishMutation>(publishPolicyVersionMutation);
  const [createDraft, isCreateDraftInFlight] =
    useMutation<ShowPolicyViewCreateDraftMutation>(
      createDraftPolicyVersionMutation,
    );

  const latestVersionEdge = policy.latestVersion?.edges[0];
  const latestVersionNode = latestVersionEdge?.node;

  const isDraft = latestVersionNode?.status === "DRAFT";

  useEffect(() => {
    // No need to update selectedVersion state since we're using the VersionHistoryModal component
  }, []);

  // Handle delete policy
  const handleDeletePolicy = useCallback(() => {
    setIsDeleteDialogOpen(true);
  }, []);

  // Confirm delete policy
  const confirmDeletePolicy = useCallback(() => {
    setIsDeleting(true);

    setTimeout(() => {
      toast({
        title: "Policy deleted",
        description: "The policy has been deleted successfully",
      });
      setIsDeleting(false);
      setIsDeleteDialogOpen(false);
      navigate(`/organizations/${organizationId}/policies`);
    }, 1000);
  }, [toast, navigate, organizationId]);

  // Navigate to publish flow
  const handlePublish = useCallback(() => {
    if (!policy.id) return;

    publishDraft({
      variables: {
        input: {
          policyId: policy.id,
        },
      },
      onCompleted: (_, errors) => {
        if (errors) {
          toast({
            title: "Error publishing policy",
            description: errors[0]?.message || "An unknown error occurred",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Policy published",
          description: `The policy has been published successfully`,
        });

        // Reload the query to refresh the data
        loadQuery({ policyId: policy.id, organizationId: organizationId! });
      },
      onError: (error) => {
        toast({
          title: "Error publishing policy",
          description: error.message || "An unknown error occurred",
          variant: "destructive",
        });
      },
    });
  }, [policy.id, publishDraft, toast, loadQuery]);

  // Open version history modal
  const handleVersionHistoryClick = useCallback(() => {
    setIsVersionHistoryOpen(true);
  }, []);

  // Restore version
  const handleRestoreVersion = useCallback(
    (versionNumber: number) => {
      // Here you would implement the logic to restore a version
      toast({
        title: "Version restored",
        description: `Version ${versionNumber} has been restored`,
      });
      setIsVersionHistoryOpen(false);
      // Reload the query to refresh the data
      if (policy.id) {
        loadQuery({ policyId: policy.id, organizationId: organizationId! });
      }
    },
    [policy.id, loadQuery, toast],
  );

  // Handle edit policy
  const handleEditPolicy = useCallback(() => {
    if (!policy.id) return;

    if (latestVersionNode?.status === "PUBLISHED") {
      // Create a new draft version first
      createDraft({
        variables: {
          input: {
            policyID: policy.id,
          },
        },
        onCompleted: (response, errors) => {
          if (errors) {
            toast({
              title: "Error creating draft",
              description: errors[0]?.message || "An unknown error occurred",
              variant: "destructive",
            });
            return;
          }

          const newDraftId =
            response.createDraftPolicyVersion.policyVersionEdge.node.id;
          navigate(
            `/organizations/${organizationId}/policies/${policy.id}/versions/${newDraftId}/edit`,
          );
        },
        onError: (error) => {
          toast({
            title: "Error creating draft",
            description: error.message || "An unknown error occurred",
            variant: "destructive",
          });
        },
      });
    } else {
      // Navigate directly to edit if it's already a draft
      navigate(
        `/organizations/${organizationId}/policies/${policy.id}/versions/${latestVersionNode?.id}/edit`,
      );
    }
  }, [
    policy.id,
    latestVersionNode,
    createDraft,
    navigate,
    organizationId,
    toast,
  ]);

  // Format date helper
  const formatDate = (dateString?: string) => {
    if (!dateString) return "N/A";
    const date = new Date(dateString);
    return format(date, "MMM d, yyyy");
  };

  // Get policy signatures data
  const policyData = useFragment<SignaturesModal_policyVersions$key>(
    policyVersionsFragment,
    policy as unknown as SignaturesModal_policyVersions$key
  );

  // Handle PDF download
  const handlePrintPDF = useCallback(() => {
    // Create a new window for printing
    const printWindow = window.open('', '_blank');
    if (!printWindow) {
      toast({
        title: "Error",
        description: "Unable to open print window. Please check your browser settings.",
        variant: "destructive",
      });
      return;
    }

    // Get version nodes and signatures from the fragment data we have
    const versionNodes = policyData?.policyVersions?.edges?.map(edge => edge.node) || [];
    const currentVersion = versionNodes.find(v => v.version === latestVersionNode?.version);
    const signatures = currentVersion?.signatures?.edges?.map(edge => edge.node) || [];

    // Create temporary div to render and capture markdown
    const tempDiv = document.createElement('div');
    const root = createRoot(tempDiv);

    // Render the markdown content to HTML
    root.render(
      <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}>
        {latestVersionNode?.content || 'No content available'}
      </ReactMarkdown>
    );

    // Give React a moment to render the content
    setTimeout(() => {
      const renderedMarkdown = tempDiv.innerHTML;

      // Prepare signatures HTML
      let signaturesHtml = '<p>No signatures yet</p>';
      if (signatures.length > 0) {
        signaturesHtml = `
          <table style="width: 100%; border-collapse: collapse; margin-top: 10px;">
            <thead>
              <tr>
                <th style="text-align: left; padding: 8px; border: 1px solid #ddd; background-color: #f2f2f2;">Name</th>
                <th style="text-align: left; padding: 8px; border: 1px solid #ddd; background-color: #f2f2f2;">Status</th>
                <th style="text-align: left; padding: 8px; border: 1px solid #ddd; background-color: #f2f2f2;">Date</th>
              </tr>
            </thead>
            <tbody>
              ${signatures.map(sig => `
                <tr>
                  <td style="padding: 8px; border: 1px solid #ddd;">${sig.signedBy?.fullName || 'Unknown'}</td>
                  <td style="padding: 8px; border: 1px solid #ddd;">${sig.state === 'SIGNED' ? 'Signed' : 'Pending'}</td>
                  <td style="padding: 8px; border: 1px solid #ddd;">${sig.signedAt ? formatDate(sig.signedAt) : 'Not yet signed'}</td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        `;
      }

      // Create the print document content with the policy document data
      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
          <head>
            <title>${policy.title || 'Policy Document'}</title>
            <style>
              body { font-family: system-ui, -apple-system, sans-serif; margin: 40px; }
              .header { margin-bottom: 30px; }
              .content { margin-bottom: 40px; }
              .footer { border-top: 1px solid #ccc; padding-top: 20px; }
              .metadata { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-bottom: 30px; background-color: #f9f9f9; padding: 15px; border-radius: 4px; }
              .metadata-item { margin-bottom: 5px; }
              .signature-title { margin-top: 20px; font-weight: bold; font-size: 1.2em; margin-bottom: 10px; }
              h1 { font-size: 24px; margin-bottom: 20px; }
              .company-name { font-size: 16px; color: #555; margin-bottom: 5px; }
              
              /* Markdown styles */
              .content h1 { font-size: 1.8em; margin-top: 1em; margin-bottom: 0.5em; }
              .content h2 { font-size: 1.5em; margin-top: 1em; margin-bottom: 0.5em; }
              .content h3 { font-size: 1.3em; margin-top: 1em; margin-bottom: 0.5em; }
              .content h4 { font-size: 1.2em; margin-top: 1em; margin-bottom: 0.5em; }
              .content h5 { font-size: 1.1em; margin-top: 1em; margin-bottom: 0.5em; }
              .content h6 { font-size: 1em; margin-top: 1em; margin-bottom: 0.5em; }
              .content p { margin-bottom: 1em; line-height: 1.6; }
              .content ul, .content ol { margin-bottom: 1em; padding-left: 2em; }
              .content li { margin-bottom: 0.5em; }
              .content blockquote { border-left: 4px solid #ccc; padding-left: 1em; margin-left: 0; }
              .content pre { background: #f4f4f4; padding: 1em; overflow-x: auto; }
              .content code { background: #f4f4f4; padding: 0.2em 0.4em; }
              .content table { border-collapse: collapse; width: 100%; margin-bottom: 1em; }
              .content th, .content td { border: 1px solid #ccc; padding: 0.5em; }
              .content th { background: #f4f4f4; }
              
              @media print {
                body { margin: 0; }
                .page-break { page-break-after: always; }
                /* Hide browser default headers and footers */
                @page {
                  margin: 0;
                  size: auto;
                }
                html {
                  background-color: #FFFFFF;
                  margin: 0px;
                }
                body {
                  margin: 40px;
                  -webkit-print-color-adjust: exact !important;
                  print-color-adjust: exact !important;
                }
              }
            </style>
          </head>
          <body>
            <div class="header">
              <h1>${policy.title || 'Policy Document'}</h1>
              <div class="metadata">
                <div class="metadata-item"><strong>Version:</strong> ${latestVersionNode?.version || 'N/A'}</div>
                <div class="metadata-item"><strong>Status:</strong> ${latestVersionNode?.status === 'PUBLISHED' ? 'Published' : 'Draft'}</div>
                <div class="metadata-item"><strong>Published Date:</strong> ${latestVersionNode?.status === 'PUBLISHED' ? formatDate(latestVersionNode.publishedAt || '') : 'Not yet published'}</div>
                <div class="metadata-item"><strong>Owner:</strong> ${policy.owner?.fullName || 'Unknown'}</div>
                <div class="metadata-item"><strong>Last Modified:</strong> ${formatDate(latestVersionNode?.updatedAt)}</div>
                <div class="metadata-item"><strong>Company:</strong> ${data.organization.name}</div>
              </div>
            </div>
            <div class="content">
              ${renderedMarkdown}
            </div>
            <div class="footer">
              <div class="signature-title">Signatory</div>
              ${signaturesHtml}
            </div>
          </body>
        </html>
      `);

      // Trigger print
      printWindow.document.close();
      printWindow.focus();

      // Use setTimeout to give the browser time to load any resources
      setTimeout(() => {
        printWindow.print();
        // Optional: close the window after printing
        // printWindow.close();
      }, 500);

      // Clean up
      root.unmount();
    }, 100);

  }, [policy, latestVersionNode, formatDate, toast, policyData, data.organization.name!]);

  return (
    <PageTemplate
      title={policy.title!}
      actions={
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            className="rounded-full h-9 px-3 gap-1.5 shadow-sm border-[#022A0214] bg-white text-[#141E12]"
            onClick={handlePrintPDF}
          >
            <Download className="h-4 w-4" />
            <span className="font-medium">Download PDF</span>
          </Button>

          <Button
            variant="outline"
            size="sm"
            className="rounded-full h-9 px-3 gap-1.5 shadow-sm border-[#022A0214] bg-white text-[#141E12]"
            onClick={handleVersionHistoryClick}
          >
            <Clock className="h-4 w-4" />
            <span className="font-medium">Version history</span>
          </Button>

          <Button
            variant="outline"
            size="sm"
            className="rounded-full h-9 px-3 gap-1.5 shadow-sm border-[#022A0214] bg-white text-[#141E12]"
            onClick={() => setIsSignaturesModalOpen(true)}
          >
            <FileSignature className="h-4 w-4" />
            <span className="font-medium">Signature history</span>
          </Button>

          {isDraft && (
            <Button
              variant="default"
              size="sm"
              className="rounded-full h-9 px-3 gap-1.5 shadow-sm"
              onClick={handlePublish}
              disabled={isPublishInFlight}
            >
              <span className="font-medium">Publish version</span>
            </Button>
          )}

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="rounded-full h-9 w-9 shadow-sm border-[#022A0214] bg-white text-[#141E12]"
              >
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem asChild>
                <div onClick={handleEditPolicy}>
                  <Edit className="mr-2 h-4 w-4" />
                  {latestVersionNode?.status === "PUBLISHED"
                    ? "Create new draft"
                    : "Edit draft policy"}
                </div>
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={handleDeletePolicy}
                className="text-danger focus:text-danger focus:bg-danger-bg"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete policy
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      }
    >
      <div className="space-y-4">
        {latestVersionNode ? (
          <>
            <div className="bg-gray-50 rounded-lg border border-solid-b shadow-sm p-6 mb-4">
              <div className="grid grid-cols-2 gap-y-3">
                <div>
                  <span className="font-medium">Document Title:</span> {policy.title}
                </div>
                <div>
                  <span className="font-medium">Version:</span> {latestVersionNode.version || "N/A"}
                </div>
                <div>
                  <span className="font-medium">Status:</span> {latestVersionNode.status === "PUBLISHED" ? "Published" : "Draft"}
                </div>
                <div>
                  <span className="font-medium">Published Date:</span> {latestVersionNode.status === "PUBLISHED" ? formatDate(latestVersionNode.publishedAt || "") : "Not yet published"}
                </div>
                <div>
                  <span className="font-medium">Owner:</span> {policy.owner?.fullName || "Unknown"}
                </div>
                <div>
                  <span className="font-medium">Last Modified:</span> {formatDate(latestVersionNode.updatedAt)}
                </div>
              </div>
            </div>

            <div className="bg-white rounded-lg border border-solid-b shadow-sm p-6">
              {latestVersionNode.changelog && (
                <div className="mb-6 p-4 bg-level-1 rounded-md border border-solid-b">
                  <div className="text-xs text-tertiary uppercase font-medium mb-1">
                    Change summary
                  </div>
                  <div className="text-sm text-secondary">
                    {latestVersionNode.changelog}
                  </div>
                </div>
              )}

              <div className="prose prose-olive max-w-none">
                <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}>
                  {latestVersionNode.content || "No content available"}
                </ReactMarkdown>
              </div>

              <div className="mt-8 pt-4 border-t border-solid-b text-xs text-tertiary flex justify-between items-center">
                <div>
                  {latestVersionNode.status === "PUBLISHED"
                    ? `Published on ${formatDate(latestVersionNode.publishedAt || "")}${latestVersionNode.publishedBy ? ` by ${latestVersionNode.publishedBy.fullName}` : ""}`
                    : `Last modified on ${formatDate(latestVersionNode.updatedAt)} by ${policy.owner?.fullName || "Unknown"}`}
                </div>
              </div>
            </div>
          </>
        ) : (
          <div className="bg-white rounded-lg border border-solid-b shadow-sm p-6 text-center text-tertiary">
            No content available
          </div>
        )}
      </div>

      {/* Signatures Modal */}
      <SignaturesModal
        isOpen={isSignaturesModalOpen}
        onClose={() => setIsSignaturesModalOpen(false)}
        policyRef={policy}
        owner={policy.owner}
      />

      {/* Version History Modal */}
      <VersionHistoryModal
        isOpen={isVersionHistoryOpen}
        onClose={() => setIsVersionHistoryOpen(false)}
        policyRef={policy}
        onRestoreVersion={handleRestoreVersion}
      />

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Policy</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete the policy "{policy.title}"? This
              action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteDialogOpen(false)}
              disabled={isDeleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDeletePolicy}
              disabled={isDeleting}
            >
              {isDeleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageTemplate>
  );
}

export default function ShowPolicyView() {
  const [queryRef, loadQuery] =
    useQueryLoader<ShowPolicyViewQuery>(policyViewQuery);
  const { policyId, organizationId } = useParams();

  useEffect(() => {
    loadQuery({ policyId: policyId!, organizationId: organizationId! });
  }, [loadQuery, policyId, organizationId]);

  if (!queryRef) {
    return <ShowPolicyViewSkeleton />;
  }

  return (
    <Suspense fallback={<ShowPolicyViewSkeleton />}>
      <ShowPolicyContent queryRef={queryRef} />
    </Suspense>
  );
}
