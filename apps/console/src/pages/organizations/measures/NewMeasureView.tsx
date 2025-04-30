import { useState } from "react";
import { ConnectionHandler, graphql, useMutation } from "react-relay";
import { useParams, useNavigate } from "react-router";
import { PageTemplate } from "@/components/PageTemplate";
import { useToast } from "@/hooks/use-toast";
import { NewMeasureViewCreateMeasureMutation } from "./__generated__/NewMeasureViewCreateMeasureMutation.graphql";
import { EditableField } from "@/components/EditableField";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

const createMeasureMutation = graphql`
  mutation NewMeasureViewCreateMeasureMutation(
    $input: CreateMeasureInput!
    $connections: [ID!]!
  ) {
    createMeasure(input: $input) {
      measureEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
        }
      }
    }
  }
`;

export default function NewMeasureView() {
  const { organizationId } = useParams<{
    organizationId: string;
  }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    category: ""
  });

  const [commit, isInFlight] =
    useMutation<NewMeasureViewCreateMeasureMutation>(createMeasureMutation);

  const handleFieldChange = (field: keyof typeof formData, value: unknown) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.name || !formData.description) {
      toast({
        title: "Validation Error",
        description: "Please fill in all required fields.",
        variant: "destructive",
      });
      return;
    }

    const connectionId = ConnectionHandler.getConnectionID(
      organizationId!,
      "MeasureListView_measures"
    );

    commit({
      variables: {
        input: {
          organizationId: organizationId!,
          name: formData.name,
          description: formData.description,
          category: formData.category
        },
        connections: [connectionId],
      },
      onCompleted(data, errors) {
        if (errors) {
          toast({
            title: "Error",
            description: errors[0]?.message || "Failed to create measure",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Measure created successfully",
        });

        navigate(
          `/organizations/${organizationId}/measures/${data.createMeasure.measureEdge.node.id}`
        );
      },
      onError(error) {
        toast({
          title: "Error",
          description: error.message || "Failed to create measure",
          variant: "destructive",
        });
      },
    });
  };

  return (
    <PageTemplate
      title="New Measure"
      description="Add a new measure. You will be able to link it to control and risks"
    >
      <Card className="max-w-2xl">
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          <EditableField
            label="Name"
            value={formData.name}
            onChange={(value) => handleFieldChange("name", value)}
            required
          />

          <EditableField
            label="Description"
            value={formData.description}
            onChange={(value) => handleFieldChange("description", value)}
            required
            multiline
            helpText="Provide a detailed description of the measure"
          />

          <EditableField
            label="Category"
            value={formData.category}
            onChange={(value) => handleFieldChange("category", value)}
            required
          />

          <div className="flex justify-end gap-3">
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                navigate(`/organizations/${organizationId}/measures`)
              }
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isInFlight}>
              {isInFlight ? "Creating..." : "Create Measure"}
            </Button>
          </div>
        </form>
      </Card>
    </PageTemplate>
  );
}
