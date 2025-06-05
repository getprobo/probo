import { Suspense, useEffect, useState, useCallback } from "react";
import { useParams, useNavigate, Link, Outlet } from "react-router";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
  ConnectionHandler,
} from "react-relay";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useToast } from "@/hooks/use-toast";
import type { FrameworkLayoutViewQuery as FrameworkLayoutViewQueryType } from "./__generated__/FrameworkLayoutViewQuery.graphql";
import type { FrameworkLayoutViewDeleteMutation } from "./__generated__/FrameworkLayoutViewDeleteMutation.graphql";
import { PageTemplate } from "@/components/PageTemplate";
import { FrameworkLayoutViewSkeleton } from "./FrameworkLayout";
import { ControlList } from "./FrameworkLayoutView/ControlList";
import { FrameworkLayoutViewExportAuditMutation } from "./__generated__/FrameworkLayoutViewExportAuditMutation.graphql";
import { Plus } from "lucide-react";

const FrameworkLayoutViewQuery = graphql`
  query FrameworkLayoutViewQuery($frameworkId: ID!) {
    node(id: $frameworkId) {
      id
      ... on Framework {
        name
        description
        ...ControlList_List
        firstControl: controls(
          first: 1
          orderBy: { field: SECTION_TITLE, direction: ASC }
        ) @connection(key: "FrameworkLayoutView_firstControl") {
          edges {
            node {
              id
              sectionTitle
              name
            }
          }
        }
      }
    }
  }
`;

const DeleteFrameworkMutation = graphql`
  mutation FrameworkLayoutViewDeleteMutation(
    $input: DeleteFrameworkInput!
    $connections: [ID!]!
  ) {
    deleteFramework(input: $input) {
      deletedFrameworkId @deleteEdge(connections: $connections)
    }
  }
`;

const exportAuditMutation = graphql`
  mutation FrameworkLayoutViewExportAuditMutation($input: ExportAuditInput!) {
    exportAudit(input: $input) {
      url
    }
  }
`;

function FrameworkLayoutViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<FrameworkLayoutViewQueryType>;
}) {
  const data = usePreloadedQuery(FrameworkLayoutViewQuery, queryRef);
  const framework = data.node;
  const navigate = useNavigate();
  const { organizationId } = useParams();
  const { toast } = useToast();
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const [exportAudit, _] =
    useMutation<FrameworkLayoutViewExportAuditMutation>(exportAuditMutation);

  // Setup delete mutation
  const [commitDeleteMutation] = useMutation<FrameworkLayoutViewDeleteMutation>(
    DeleteFrameworkMutation
  );

  const handleExportAudit = useCallback(() => {
    exportAudit({
      variables: {
        input: { frameworkId: framework.id },
      },
    });
  }, [exportAudit, framework.id]);

  const handleDeleteFramework = useCallback(() => {
    setIsDeleting(true);

    const connectionId = ConnectionHandler.getConnectionID(
      organizationId!,
      "FrameworkListView_frameworks"
    );

    commitDeleteMutation({
      variables: {
        input: {
          frameworkId: framework.id,
        },
        connections: [connectionId],
      },
      onCompleted: (_, errors) => {
        setIsDeleting(false);
        setIsDeleteDialogOpen(false);

        if (errors) {
          console.error("Error deleting framework:", errors);
          toast({
            title: "Error",
            description: "Failed to delete framework. Please try again.",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Framework deleted successfully.",
        });

        navigate(`/organizations/${organizationId}/frameworks`);
      },
      onError: (error) => {
        setIsDeleting(false);
        setIsDeleteDialogOpen(false);
        console.error("Error deleting framework:", error);
        toast({
          title: "Error",
          description: "Failed to delete framework. Please try again.",
          variant: "destructive",
        });
      },
    });
  }, [framework.id, organizationId, commitDeleteMutation, navigate, toast]);

  return (
    <PageTemplate
      title={framework.name ?? ""}
      description={framework.description || ""}
      actions={
        <div className="flex gap-4">
          <Button variant="secondary" asChild>
            <Link
              to={`/organizations/${organizationId}/frameworks/${framework.id}/controls/new`}
            >
              <Plus className="w-3 h-4 mr-2" />
              Create Control
            </Link>
          </Button>
          <Button variant="secondary" asChild>
            <Link
              to={`/organizations/${organizationId}/frameworks/${framework.id}/edit`}
            >
              Edit Framework
            </Link>
          </Button>
          <Button variant="secondary" onClick={() => handleExportAudit()}>
            Export Audit
          </Button>
          <Button
            variant="destructive"
            onClick={() => setIsDeleteDialogOpen(true)}
          >
            Delete Framework
          </Button>
        </div>
      }
    >
      <div className="flex flex-row items-start justify-start w-full h-[calc(100vh-216px)] overflow-hidden -ml-8">
        <ControlList fragmentKey={framework} />
        <Outlet />
      </div>

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Framework</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete the framework &quot;
              {framework.name}&quot;? This action cannot be undone.
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
              onClick={handleDeleteFramework}
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

export default function FrameworkLayoutView() {
  const { frameworkId } = useParams();
  const [queryRef, loadQuery] = useQueryLoader<FrameworkLayoutViewQueryType>(
    FrameworkLayoutViewQuery
  );

  useEffect(() => {
    loadQuery({ frameworkId: frameworkId! });
  }, [loadQuery, frameworkId]);

  if (!queryRef) {
    return <FrameworkLayoutViewSkeleton />;
  }

  return (
    <Suspense fallback={<FrameworkLayoutViewSkeleton />}>
      <FrameworkLayoutViewContent queryRef={queryRef} />
    </Suspense>
  );
}
