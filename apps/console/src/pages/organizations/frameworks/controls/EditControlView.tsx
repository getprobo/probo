import { Suspense, useEffect, useState } from "react";
import {
  graphql,
  PreloadedQuery,
  useMutation,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { useParams, useNavigate } from "react-router";
import { useToast } from "@/hooks/use-toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Loader2 } from "lucide-react";
import { EditControlViewSkeleton } from "./EditControlPage";
import { EditControlViewQuery } from "./__generated__/EditControlViewQuery.graphql";

const editControlViewQuery = graphql`
  query EditControlViewQuery($controlId: ID!) {
    control: node(id: $controlId) {
      ... on Control {
        id
        name
        description
        sectionTitle
      }
    }
  }
`;

const updateControlMutation = graphql`
  mutation EditControlViewUpdateControlMutation($input: UpdateControlInput!) {
    updateControl(input: $input) {
      control {
        id
        name
        description
        sectionTitle
      }
    }
  }
`;

function EditControlViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<EditControlViewQuery>;
}) {
  const { organizationId, frameworkId } = useParams<{
    organizationId: string;
    frameworkId: string;
  }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  const data = usePreloadedQuery(editControlViewQuery, queryRef);
  const [isLoading, setIsLoading] = useState(false);
  const [commitUpdateControl] = useMutation(updateControlMutation);

  if (!data.control) {
    return <EditControlViewSkeleton />;
  }

  const [formData, setFormData] = useState({
    name: data.control.name || "",
    description: data.control.description || "",
    sectionTitle: data.control.sectionTitle || "",
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    commitUpdateControl({
      variables: {
        input: {
          id: data.control.id,
          name: formData.name,
          description: formData.description,
          sectionTitle: formData.sectionTitle,
        },
      },
      onCompleted: (_, errors) => {
        setIsLoading(false);

        if (errors) {
          console.error("Error updating control:", errors);
          toast({
            title: "Error",
            description: "Failed to update control. Please try again.",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Control updated successfully.",
        });

        navigate(`/organizations/${organizationId}/frameworks/${frameworkId}/controls/${data.control.id}`);
      },
      onError: (error) => {
        setIsLoading(false);
        console.error("Error updating control:", error);
        toast({
          title: "Error",
          description: "Failed to update control. Please try again.",
          variant: "destructive",
        });
      },
    });
  };

  return (
    <div className="container py-10">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl font-bold mb-8">Edit Control</h1>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="block text-sm font-medium mb-2">Section Title</label>
            <Input
              value={formData.sectionTitle}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, sectionTitle: e.target.value }))
              }
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Name</label>
            <Input
              value={formData.name}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, name: e.target.value }))
              }
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Description</label>
            <Textarea
              value={formData.description}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, description: e.target.value }))
              }
              required
              rows={5}
            />
          </div>

          <div className="flex justify-end gap-4">
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                navigate(
                  `/organizations/${organizationId}/frameworks/${frameworkId}/controls/${data.control.id}`
                )
              }
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                  Saving...
                </>
              ) : (
                "Save Changes"
              )}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function EditControlView({ controlId }: { controlId?: string }) {
  const { controlId: controlIdParam } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<EditControlViewQuery>(editControlViewQuery);

  useEffect(() => {
    loadQuery({ controlId: (controlId ?? controlIdParam)! });
  }, [loadQuery, controlId, controlIdParam]);

  if (!queryRef) {
    return <EditControlViewSkeleton />;
  }

  return (
    <Suspense fallback={<EditControlViewSkeleton />}>
      <EditControlViewContent queryRef={queryRef} />
    </Suspense>
  );
}
