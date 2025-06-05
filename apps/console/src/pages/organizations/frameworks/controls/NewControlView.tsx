import { useState } from "react";
import { graphql, useMutation } from "react-relay";
import { useNavigate, useParams } from "react-router";
import { useToast } from "@/hooks/use-toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Loader2 } from "lucide-react";
import { NewControlViewCreateControlMutation } from "./__generated__/NewControlViewCreateControlMutation.graphql";

const createControlMutation = graphql`
  mutation NewControlViewCreateControlMutation($input: CreateControlInput!) {
    createControl(input: $input) {
      controlEdge {
        node {
          id
          name
          description
          sectionTitle
        }
      }
    }
  }
`;

export default function NewControlView() {
  const { organizationId, frameworkId } = useParams<{
    organizationId: string;
    frameworkId: string;
  }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [isLoading, setIsLoading] = useState(false);
  const [commitCreateControl] = useMutation<NewControlViewCreateControlMutation>(
    createControlMutation
  );

  const [formData, setFormData] = useState({
    name: "",
    description: "",
    sectionTitle: "",
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    commitCreateControl({
      variables: {
        input: {
          organizationId: organizationId!,
          frameworkId: frameworkId!,
          name: formData.name,
          description: formData.description,
          sectionTitle: formData.sectionTitle,
        },
      },
      onCompleted: (response, errors) => {
        setIsLoading(false);

        if (errors) {
          console.error("Error creating control:", errors);
          toast({
            title: "Error",
            description: "Failed to create control. Please try again.",
            variant: "destructive",
          });
          return;
        }

        const controlId = response.createControl.controlEdge.node.id;
        toast({
          title: "Success",
          description: "Control created successfully.",
        });

        navigate(
          `/organizations/${organizationId}/frameworks/${frameworkId}/controls/${controlId}?t=${Date.now()}`,
          { replace: true }
        );
      },
      onError: (error) => {
        setIsLoading(false);
        console.error("Error creating control:", error);
        toast({
          title: "Error",
          description: "Failed to create control. Please try again.",
          variant: "destructive",
        });
      },
      updater: (store) => {
        const payload = store.getRootField('createControl');
        const newControl = payload.getLinkedRecord('controlEdge').getLinkedRecord('node');

        const framework = store.get(frameworkId!);
        if (framework) {
          const controls = framework.getLinkedRecord('controls');
          if (controls) {
            const edges = controls.getLinkedRecords('edges') || [];
            const newEdge = store.create(
              'edge',
              'ControlEdge'
            ).setLinkedRecord(newControl, 'node');
            controls.setLinkedRecords([newEdge, ...edges], 'edges');
          }
        }
      }
    });
  };

  return (
    <div className="container py-10">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl font-bold mb-8">Create Control</h1>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="block text-sm font-medium mb-2">Section Title</label>
            <Input
              value={formData.sectionTitle}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, sectionTitle: e.target.value }))
              }
              required
              placeholder="e.g. CTRL-001"
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
              placeholder="Enter control name"
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
              placeholder="Describe the control"
            />
          </div>

          <div className="flex justify-end gap-4">
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                navigate(
                  `/organizations/${organizationId}/frameworks/${frameworkId}`
                )
              }
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                  Creating...
                </>
              ) : (
                "Create Control"
              )}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
