import { Suspense, useEffect, useState } from "react";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
  ConnectionHandler,
} from "react-relay";
import { Link, useParams, useNavigate } from "react-router";
import { Button } from "@/components/ui/button";
import {
  Plus,
  ChevronDown,
  MoreHorizontal,
  Trash2,
  Eye,
  X,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useToast } from "@/hooks/use-toast";
import { format } from "date-fns";
import type { PolicyListViewQuery, PolicyListViewQuery$data } from "./__generated__/PolicyListViewQuery.graphql";
import type { PolicyListViewDeleteMutation } from "./__generated__/PolicyListViewDeleteMutation.graphql";
import type { PolicyListViewCreateMutation } from "./__generated__/PolicyListViewCreateMutation.graphql";
import type { PolicyListViewSendSigningNotificationsMutation } from "./__generated__/PolicyListViewSendSigningNotificationsMutation.graphql";
import { PageTemplate } from "@/components/PageTemplate";
import { PolicyListViewSkeleton } from "./PolicyListPage";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import PeopleSelector from "@/components/PeopleSelector";
import type { PeopleSelector_organization$key } from "@/components/__generated__/PeopleSelector_organization.graphql";
import { Textarea } from "@/components/ui/textarea";

const policyListViewQuery = graphql`
  query PolicyListViewQuery($organizationId: ID!) {
    viewer {
      user {
        id
      }
    }

    organization: node(id: $organizationId) {
      ... on Organization {
        ...PeopleSelector_organization
        policies(first: 50, orderBy: {field: TITLE, direction: ASC}) @connection(key: "PolicyListView_policies") {
          edges {
            node {
              id
              title
              description
              currentPublishedVersion
              createdAt
              updatedAt
              owner {
                id
                fullName
              }
              versions(first: 1) {
                edges {
                  node {
                    id
                    status
                    updatedAt
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;

const DeletePolicyMutation = graphql`
  mutation PolicyListViewDeleteMutation(
    $input: DeletePolicyInput!
    $connections: [ID!]!
  ) {
    deletePolicy(input: $input) {
      deletedPolicyId @deleteEdge(connections: $connections)
    }
  }
`;

const createPolicyMutation = graphql`
  mutation PolicyListViewCreateMutation(
    $input: CreatePolicyInput!
    $connections: [ID!]!
  ) {
    createPolicy(input: $input) {
      policyEdge @prependEdge(connections: $connections) {
        node {
          id
          title
          description
          createdAt
          updatedAt
          owner {
            id
            fullName
          }
        }
      }
    }
  }
`;

const sendSigningNotificationsMutation = graphql`
  mutation PolicyListViewSendSigningNotificationsMutation($input: SendSigningNotificationsInput!) {
    sendSigningNotifications(input: $input) {
      success
    }
  }
`;
function PolicyTableRow({
  policy,
  organizationId,
}: {
  policy: NonNullable<PolicyListViewQuery$data["organization"]["policies"]>["edges"][0]["node"];
  organizationId: string;
}) {
  const navigate = useNavigate();
  const { toast } = useToast();
  const [deletePolicy, isDeleting] = useMutation<PolicyListViewDeleteMutation>(DeletePolicyMutation);

  const latestVersion = policy.versions?.edges[0]?.node;
  const status = latestVersion?.status || "DRAFT";

  const handleDeletePolicy = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (window.confirm("Are you sure you want to delete this policy? This action cannot be undone.")) {
      deletePolicy({
        variables: {
          input: {
            policyId: policy.id,
          },
          connections: [
            ConnectionHandler.getConnectionID(
              organizationId,
              "PolicyListView_policies"
            ),
          ],
        },
        onCompleted: (_, errors) => {
          if (errors) {
            console.error("Error deleting policy:", errors);
            toast({
              title: "Error",
              description: "Failed to delete policy. Please try again.",
              variant: "destructive",
            });
            return;
          }
          toast({
            title: "Success",
            description: "Policy deleted successfully.",
          });
        },
        onError: (error) => {
          console.error("Error deleting policy:", error);
          toast({
            title: "Error",
            description: "Failed to delete policy. Please try again.",
            variant: "destructive",
          });
        },
      });
    }
  };

  const handleView = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(`/organizations/${organizationId}/policies/${policy.id}`);
  };

  return (
    <tr 
      className="border-t border-solid-b hover:bg-subtle-bg cursor-pointer"
      onClick={() => {
        navigate(`/organizations/${organizationId}/policies/${policy.id}`);
      }}
    >
      <td className="py-4 px-6">
        <div className="flex flex-col">
          <span className="font-medium text-primary">{policy.title}</span>
          <span className="text-sm text-tertiary">
            {policy.description || "No description provided"}
          </span>
        </div>
      </td>
      <td className="py-4 px-6">
        <span className="text-sm text-primary">
          {format(new Date(policy.updatedAt), "MMM d, yyyy")}
        </span>
      </td>
      <td className="py-4 px-6">
        <Badge
          className={`font-medium border-0 py-0 px-[6px] h-5 text-xs rounded-md ${
            status === "PUBLISHED" 
              ? "bg-success-bg text-success" 
              : "bg-secondary-bg text-tertiary"
          }`}
        >
          {status === "PUBLISHED" ? "Published" : "Draft"}
        </Badge>
      </td>
      <td className="py-4 px-6 text-right">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={handleView}>
              <Eye className="mr-2 h-4 w-4" />
              View
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={handleDeletePolicy}
              className="text-danger focus:text-danger focus:bg-danger-bg"
              disabled={isDeleting}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              {isDeleting ? "Deleting..." : "Delete"}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </td>
    </tr>
  );
}

function CreatePolicyModal({
  open,
  onOpenChange,
  organizationId,
  organizationRef,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organizationId: string;
  organizationRef: PeopleSelector_organization$key;
}) {
  const { toast } = useToast();
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [ownerId, setOwnerId] = useState<string | null>(null);

  const [createPolicy, isCreating] = useMutation<PolicyListViewCreateMutation>(createPolicyMutation);

  // Reset form fields
  const resetForm = () => {
    setTitle("");
    setContent("");
    setOwnerId(null);
  };

  const handleCreatePolicy = () => {
    if (!title.trim()) {
      toast({
        title: "Error",
        description: "Please enter a policy title.",
        variant: "destructive",
      });
      return;
    }

    if (!ownerId) {
      toast({
        title: "Error",
        description: "Please select an owner for the policy.",
        variant: "destructive",
      });
      return;
    }
    
    const input = {
      organizationId,
      title,
      content,
      ownerId,
    };

    createPolicy({
      variables: {
        input,
        connections: [
          ConnectionHandler.getConnectionID(
            organizationId,
            "PolicyListView_policies",
            {orderBy: {field: "TITLE", direction: "ASC"}}
          ),
        ],
      },
      onCompleted: (response, errors) => {
        if (errors) {
          console.error("Error creating policy:", errors);
          toast({
            title: "Error",
            description: "Failed to create policy. Please try again.",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Policy created successfully!",
        });

        resetForm();
        onOpenChange(false);
      },
      onError: (error) => {
        console.error("Error creating policy:", error);
        toast({
          title: "Error",
          description: "Failed to create policy. Please try again.",
          variant: "destructive",
        });
      },
    });
  };

  // Reset form when modal closes
  useEffect(() => {
    if (!open) {
      resetForm();
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[1080px] h-[700px] max-h-[700px] p-0 gap-0 flex flex-col">
        <DialogTitle className="sr-only">Create new policy</DialogTitle>
        <DialogDescription className="sr-only">Form to create a new policy with title, content, and owner information.</DialogDescription>
        <div className="flex justify-between items-center py-2 px-4 border-b border-solid-b h-[40px]">
          <div className="flex items-center gap-1 text-sm">
            <span className="text-tertiary">Policies</span>
            <ChevronDown className="h-3 w-3 text-quaternary rotate-[270deg]" />
            <span className="font-medium">New policy</span>
          </div>
        </div>

        <div className="flex flex-1 overflow-hidden">
          <div className="flex-1 p-6 overflow-y-auto">
            <div className="mb-6">
              <h1 
                className={`text-4xl leading-tight font-bold outline-none focus:outline-none ${!title ? 'text-gray-400' : 'text-black'}`}
                contentEditable
                suppressContentEditableWarning
                onInput={(e) => {
                  const newText = e.currentTarget.textContent || "";
                  setTitle(newText);
                }}
                style={{ WebkitTapHighlightColor: 'transparent' }}
                onClick={(e) => {
                  if (!title) {
                    e.currentTarget.textContent = '';
                  }
                }}
                onFocus={(e) => {
                  if (e.currentTarget.textContent === "Enter policy title...") {
                    e.currentTarget.textContent = '';
                  }
                }}
                onBlur={(e) => {
                  if (!e.currentTarget.textContent?.trim()) {
                    e.currentTarget.textContent = "Enter policy title...";
                    setTitle("");
                  }
                }}
                ref={(el) => {
                  if (el && !el.textContent) {
                    el.textContent = title || "Enter policy title...";
                  }
                }}
              >
              </h1>
            </div>
            <Textarea
              placeholder="This Privacy Policy outlines how NovaSoft collects, uses, and protects personal information provided by users of its services. By accessing or using our platform, you agree to the collection and use of information in accordance with this policy..."
              className="min-h-[300px] border-none resize-none p-0 focus-visible:ring-0 focus-visible:ring-offset-0"
              value={content}
              onChange={(e) => setContent(e.target.value)}
            />
          </div>
          <div className="w-[420px] bg-[rgba(5,77,5,0.03)] p-6 flex flex-col gap-4">
            <h3 className="font-medium text-base">Properties</h3>
            
            <div className="flex flex-col gap-4">
              <div className="flex justify-between items-center py-2 border-t border-[rgba(2,42,2,0.08)]">
                <span className="text-sm font-medium text-tertiary">Status</span>
                <div className="bg-[rgba(0,39,0,0.05)] py-1.5 px-2 rounded-lg">
                  <span className="text-sm font-medium">Draft</span>
                </div>
              </div>
              
              <div className="flex justify-between items-center py-2 border-t border-[rgba(2,42,2,0.08)]">
                <span className="text-sm font-medium text-tertiary">Owner</span>
                <div className="relative">
                  <PeopleSelector
                    organizationRef={organizationRef}
                    selectedPersonId={ownerId}
                    onSelect={setOwnerId}
                    placeholder="Select owner"
                  />
                </div>
              </div>
              
              <div className="flex justify-between items-center py-2 border-t border-[rgba(2,42,2,0.08)]">
                <span className="text-sm font-medium text-tertiary">Review date</span>
                <Button size="icon" variant="outline" className="h-8 w-8 rounded-full bg-[rgba(0,39,0,0.05)] border-none">
                  <Plus className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter className="flex justify-end items-center gap-3 px-4 py-2 border-t border-solid-b h-[60px]">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="h-9"
          >
            Cancel
          </Button>
          <Button
            onClick={handleCreatePolicy}
            disabled={isCreating}
            className="h-9"
          >
            {isCreating ? "Creating..." : "Create policy"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function PolicyListViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<PolicyListViewQuery>;
}) {
  const data = usePreloadedQuery<PolicyListViewQuery>(
    policyListViewQuery,
    queryRef
  );
  const { organizationId } = useParams();
  const policies =
    data.organization?.policies?.edges.map((edge) => edge?.node) ?? [];
  
  const [createModalOpen, setCreateModalOpen] = useState(false);

  const [sendSigningNotifications, _] = useMutation<PolicyListViewSendSigningNotificationsMutation>(sendSigningNotificationsMutation);

  const handleOpenModal = () => {
    setCreateModalOpen(true);
  };

  const handleCloseModal = (open: boolean) => {
    setCreateModalOpen(open);
  };

  const handleSendSigningNotifications = () => {
    sendSigningNotifications({
      variables: {
        input: {
          organizationId: organizationId!,
        },
      },
    });
  };

  return (
    <PageTemplate
      title="Policies"
      actions={
        <div>
          <Button onClick={handleOpenModal}>
            <Plus className="mr-2 h-4 w-4" />
            New policy
        </Button>
        <Button onClick={handleSendSigningNotifications}>
            Send signing notifications
          </Button>
        </div>
      }
    >
      {/* Policy table */}
      <div className="rounded-lg border border-solid-b overflow-hidden bg-level-1">
        <table className="w-full">
          <thead>
            <tr className="bg-level-1 text-left">
              <th className="py-3 px-6 text-xs font-medium text-tertiary border-b border-low-b">
                <div className="flex items-center gap-1">
                  Policy
                  <ChevronDown className="h-3 w-3 text-quaternary" />
                </div>
              </th>
              <th className="py-3 px-6 text-xs font-medium text-tertiary border-b border-low-b">
                <div className="flex items-center gap-1">
                  Last update
                  <ChevronDown className="h-3 w-3 text-quaternary" />
                </div>
              </th>
              <th className="py-3 px-6 text-xs font-medium text-tertiary border-b border-low-b">
                <div className="flex items-center gap-1">
                  Status
                  <ChevronDown className="h-3 w-3 text-quaternary" />
                </div>
              </th>
              <th className="py-3 px-6 text-right text-xs font-medium text-tertiary border-b border-low-b"></th>
            </tr>
          </thead>
          <tbody className="bg-level-1">
            {policies.length > 0 ? (
              policies.map((policy: any) => (
                <PolicyTableRow
                  key={policy.id}
                  policy={policy}
                  organizationId={organizationId!}
                />
              ))
            ) : (
              <tr>
                <td colSpan={4} className="py-12 text-center">
                  <div className="flex flex-col items-center gap-2">
                    <h3 className="text-lg font-medium">No policies found</h3>
                    <p className="text-tertiary">
                      Create your first policy to get started
                    </p>
                    <Button onClick={handleOpenModal}>
                      <Plus className="mr-2 h-4 w-4" />
                      Create policy
                    </Button>
                  </div>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <CreatePolicyModal
        open={createModalOpen}
        onOpenChange={handleCloseModal}
        organizationId={organizationId!}
        organizationRef={data.organization}
      />
    </PageTemplate>
  );
}

export default function PolicyListView() {
  const [queryRef, loadQuery] =
    useQueryLoader<PolicyListViewQuery>(policyListViewQuery);

  const { organizationId } = useParams();

  useEffect(() => {
    loadQuery({ organizationId: organizationId! });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <PolicyListViewSkeleton />;
  }

  return (
    <Suspense fallback={<PolicyListViewSkeleton />}>
      {<PolicyListViewContent queryRef={queryRef} />}
    </Suspense>
  );
}
