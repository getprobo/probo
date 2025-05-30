import { Dialog, DialogContent } from "@/components/ui/dialog";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { format } from "date-fns";
import { Button } from "@/components/ui/button";
import { useState, useEffect } from "react";
import {
  graphql,
  useFragment,
  useMutation,
  useQueryLoader,
  usePreloadedQuery,
  ConnectionHandler,
} from "react-relay";
import type {
  SignaturesModal_documentVersions$data,
  SignaturesModal_documentVersions$key,
} from "./__generated__/SignaturesModal_documentVersions.graphql";
import type { SignaturesModalRequestSignatureMutation } from "./__generated__/SignaturesModalRequestSignatureMutation.graphql";
import type { SignaturesModalOrganizationQuery } from "./__generated__/SignaturesModalOrganizationQuery.graphql";
import { useToast } from "@/hooks/use-toast";
import { useParams } from "react-router";
import { Loader2, CheckCircle2, Clock } from "lucide-react";
import { PreloadedQuery } from "react-relay";

export const documentVersionsFragment = graphql`
  fragment SignaturesModal_documentVersions on Document {
    title
    documentVersions: versions(first: 10) {
      edges {
        node {
          id
          version
          status
          publishedAt
          updatedAt
          publishedBy {
            fullName
          }
          signatures(first: 100)
            @connection(key: "SignaturesModal_documentVersions_signatures") {
            edges {
              node {
                id
                state
                signedAt
                requestedAt
                signedBy {
                  fullName
                  id
                }
                requestedBy {
                  fullName
                }
              }
            }
          }
        }
      }
    }
  }
`;

const requestSignatureMutation = graphql`
  mutation SignaturesModalRequestSignatureMutation(
    $input: RequestSignatureInput!
    $connections: [ID!]!
  ) {
    requestSignature(input: $input) {
      documentVersionSignatureEdge @prependEdge(connections: $connections) {
        node {
          id
          state
          signedAt
          requestedAt
          signedBy {
            fullName
            id
          }
          requestedBy {
            fullName
          }
        }
      }
    }
  }
`;

const organizationQuery = graphql`
  query SignaturesModalOrganizationQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        peoples(first: 100, orderBy: { direction: ASC, field: FULL_NAME }) {
          edges {
            node {
              id
              fullName
              primaryEmailAddress
              kind
            }
          }
        }
      }
    }
  }
`;

interface SignaturesModalProps {
  isOpen: boolean;
  onClose: () => void;
  documentRef: SignaturesModal_documentVersions$key;
  owner?: {
    fullName: string;
  } | null;
}

export function SignaturesModal({
  isOpen,
  onClose,
  documentRef,
  owner,
}: SignaturesModalProps) {
  const data = useFragment<SignaturesModal_documentVersions$key>(
    documentVersionsFragment,
    documentRef,
  );

  const versionNodes =
    data?.documentVersions?.edges?.map((edge) => edge.node) || [];
  const publishedVersions = versionNodes
    .filter((v) => v.status === "PUBLISHED")
    .sort((a, b) => b.version - a.version);

  const [selectedVersion, setSelectedVersion] = useState<number>(
    publishedVersions[0]?.version || 0,
  );

  const selectedVersionData = versionNodes.find(
    (v) => v.version === selectedVersion,
  );

  const { organizationId } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<SignaturesModalOrganizationQuery>(organizationQuery);
  const { toast } = useToast();
  const [isRequesting, setIsRequesting] = useState<string | null>(null);

  const [commitRequestSignature] =
    useMutation<SignaturesModalRequestSignatureMutation>(
      requestSignatureMutation,
    );

  useEffect(() => {
    if (
      isOpen &&
      organizationId &&
      selectedVersionData?.status === "PUBLISHED"
    ) {
      loadQuery({ organizationId });
    }
  }, [isOpen, organizationId, loadQuery, selectedVersionData]);

  const formatDateTime = (dateString?: string | null) => {
    if (!dateString) return "N/A";
    const date = new Date(dateString);
    return format(date, "h:mm a • MMM d, yyyy");
  };

  const signatures =
    selectedVersionData?.signatures?.edges?.map((edge) => edge.node) || [];

  const handleRequestSignature = (personId: string) => {
    if (!selectedVersionData) return;

    setIsRequesting(personId);

    commitRequestSignature({
      variables: {
        input: {
          documentVersionId: selectedVersionData.id,
          signatoryId: personId,
        },
        connections: [
          ConnectionHandler.getConnectionID(
            selectedVersionData.id,
            "SignaturesModal_documentVersions_signatures",
          ),
        ],
      },
      onCompleted: () => {
        toast({
          title: "Success",
          description: "Signature request sent successfully",
        });
        setIsRequesting(null);
      },
      onError: (error) => {
        toast({
          title: "Error",
          description: error.message,
          variant: "destructive",
        });
        setIsRequesting(null);
      },
    });
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-[1080px] h-[744px] p-0 overflow-hidden flex flex-col">
        <div className="flex flex-1 overflow-hidden">
          {/* Version History Sidebar */}
          <div className="w-[316px] border-r border-solid-b h-full overflow-y-auto">
            <h2 className="text-lg font-semibold p-6 pb-4">Version history</h2>
            <div className="overflow-y-auto">
              {versionNodes.length > 0 ? (
                versionNodes.map((version) => {
                  const isPublished = version.status === "PUBLISHED";
                  const isSelected = selectedVersion === version.version;

                  return (
                    <div
                      key={version.id}
                      className={`flex items-center gap-3 p-6 py-4 ${
                        isPublished
                          ? "cursor-pointer"
                          : "cursor-not-allowed opacity-50"
                      } ${isSelected ? "bg-slate-50" : ""}`}
                      onClick={() =>
                        isPublished && setSelectedVersion(version.version)
                      }
                    >
                      <Avatar className="h-10 w-10">
                        <AvatarImage
                          src=""
                          alt={
                            version.publishedBy?.fullName ||
                            owner?.fullName ||
                            ""
                          }
                        />
                        <AvatarFallback>
                          {(
                            version.publishedBy?.fullName ||
                            owner?.fullName ||
                            ""
                          ).charAt(0)}
                        </AvatarFallback>
                      </Avatar>
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">
                            Version {version.version}
                          </span>
                          {isPublished ? (
                            <span className="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-800">
                              Published
                            </span>
                          ) : (
                            <span className="text-xs px-2 py-0.5 rounded-full bg-yellow-100 text-yellow-800">
                              Draft
                            </span>
                          )}
                        </div>
                        <div className="text-sm text-tertiary">
                          {version.publishedBy?.fullName ||
                            owner?.fullName ||
                            "Unknown"}{" "}
                          •{" "}
                          {formatDateTime(
                            version.publishedAt || version.updatedAt,
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })
              ) : (
                <div className="p-6 text-center text-tertiary">
                  No versions available
                </div>
              )}
            </div>
          </div>

          {/* Main Content Area */}
          <div className="flex-1 flex flex-col overflow-hidden">
            {/* Section header */}
            <div className="p-6 border-b border-solid-b">
              <h2 className="text-2xl font-semibold mb-2">
                {selectedVersionData?.status === "PUBLISHED"
                  ? "Signatures & Requests"
                  : "Signatures"}
              </h2>
              {selectedVersionData?.status === "PUBLISHED" && (
                <p className="text-sm text-muted-foreground">
                  Click request to ask for a signature from people in your
                  organization
                </p>
              )}
            </div>

            {/* Content area with unified people and signatures list */}
            <div className="flex-1 overflow-y-auto p-6">
              {selectedVersionData ? (
                selectedVersionData.status === "PUBLISHED" && queryRef ? (
                  <PeopleAndSignaturesList
                    queryRef={queryRef}
                    onRequestSignature={handleRequestSignature}
                    requestingId={isRequesting}
                    existingSignatures={signatures}
                    formatDateTime={formatDateTime}
                  />
                ) : (
                  <div className="space-y-4">
                    {signatures.length > 0 ? (
                      <div className="divide-y divide-border">
                        {signatures.map((signature) => (
                          <div key={signature.id} className="py-4">
                            <div className="flex justify-between items-center">
                              <div className="flex items-center gap-3">
                                <Avatar className="h-10 w-10">
                                  <AvatarFallback>
                                    {signature.signedBy?.fullName?.charAt(0) ||
                                      signature.requestedBy?.fullName?.charAt(
                                        0,
                                      )}
                                  </AvatarFallback>
                                </Avatar>
                                <div>
                                  <div className="font-medium">
                                    {signature.signedBy?.fullName ||
                                      signature.requestedBy?.fullName}
                                  </div>
                                  <div className="text-sm text-tertiary">
                                    {signature.state === "SIGNED"
                                      ? `Signed on ${formatDateTime(signature.signedAt)}`
                                      : `Requested on ${formatDateTime(signature.requestedAt)}`}
                                  </div>
                                </div>
                              </div>
                              <span
                                className={`text-xs px-2 py-0.5 rounded-full ${
                                  signature.state === "SIGNED"
                                    ? "bg-green-100 text-green-800"
                                    : "bg-yellow-100 text-yellow-800"
                                }`}
                              >
                                {signature.state === "SIGNED"
                                  ? "Signed"
                                  : "Pending"}
                              </span>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-center py-8">
                        <p className="text-tertiary">
                          No signatures available for this version
                        </p>
                        {selectedVersionData.status !== "PUBLISHED" && (
                          <p className="text-sm text-muted-foreground mt-2">
                            Only published versions can have signatures
                          </p>
                        )}
                      </div>
                    )}
                  </div>
                )
              ) : (
                <div className="text-center text-tertiary py-8">
                  No version selected
                </div>
              )}

              {selectedVersionData?.status === "PUBLISHED" && !queryRef && (
                <div className="py-8 text-center">
                  <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2" />
                  <p className="text-tertiary">Loading...</p>
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="flex justify-end p-4 border-t border-solid-b mt-auto">
          <Button variant="outline" onClick={onClose} className="mr-2">
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// Combined people and signatures list component
function PeopleAndSignaturesList({
  queryRef,
  onRequestSignature,
  requestingId,
  existingSignatures,
  formatDateTime,
}: {
  queryRef: PreloadedQuery<SignaturesModalOrganizationQuery>;
  onRequestSignature: (personId: string) => void;
  requestingId: string | null;
  existingSignatures: Array<
    SignaturesModal_documentVersions$data["documentVersions"]["edges"][0]["node"]["signatures"]["edges"][0]["node"]
  >;
  formatDateTime: (date?: string | null) => string;
}) {
  const data = usePreloadedQuery<SignaturesModalOrganizationQuery>(
    organizationQuery,
    queryRef,
  );

  const people = data?.organization?.peoples?.edges || [];

  if (!data?.organization) {
    return <div className="text-center py-8">Organization not found</div>;
  }

  if (people.length === 0) {
    return (
      <div className="text-center py-8">
        No people available to request signatures from
      </div>
    );
  }

  // Create lookup for signatures by person ID
  const signaturesByPersonId = new Map();
  existingSignatures.forEach((sig) => {
    if (sig.signedBy?.id) {
      signaturesByPersonId.set(sig.signedBy.id, sig);
    }
  });

  return (
    <div className="divide-y divide-border">
      {people.map((edge: any) => {
        const person = edge.node;
        const signature = signaturesByPersonId.get(person.id);
        const isRequesting = requestingId === person.id;

        return (
          <div key={person.id} className="py-4">
            <div className="flex justify-between items-center">
              <div className="flex items-center gap-3">
                <Avatar className="h-10 w-10">
                  <AvatarFallback>{person.fullName.charAt(0)}</AvatarFallback>
                </Avatar>
                <div>
                  <div className="font-medium">{person.fullName}</div>
                  <div className="text-sm text-tertiary">
                    {signature ? (
                      signature.state === "SIGNED" ? (
                        <div className="flex items-center gap-1 text-green-700">
                          <CheckCircle2 className="h-3.5 w-3.5" />
                          <span>
                            Signed on {formatDateTime(signature.signedAt)}
                          </span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1 text-amber-700">
                          <Clock className="h-3.5 w-3.5" />
                          <span>
                            Requested on {formatDateTime(signature.requestedAt)}
                          </span>
                        </div>
                      )
                    ) : (
                      person.primaryEmailAddress
                    )}
                  </div>
                </div>
              </div>

              {signature ? (
                <span
                  className={`text-xs px-2 py-0.5 rounded-full ${
                    signature.state === "SIGNED"
                      ? "bg-green-100 text-green-800"
                      : "bg-yellow-100 text-yellow-800"
                  }`}
                >
                  {signature.state === "SIGNED" ? "Signed" : "Pending"}
                </span>
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onRequestSignature(person.id)}
                  disabled={isRequesting}
                >
                  {isRequesting ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Requesting...
                    </>
                  ) : (
                    "Request Signature"
                  )}
                </Button>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
