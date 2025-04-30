import { Suspense, useEffect, useState } from "react";
import {
  graphql,
  PreloadedQuery,
  useMutation,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { useParams, useNavigate } from "react-router";
import { PageTemplate } from "@/components/PageTemplate";
import { useToast } from "@/hooks/use-toast";
import { EditMeasureViewUpdateMeasureMutation } from "./__generated__/EditMeasureViewUpdateMeasureMutation.graphql";
import { EditableField } from "@/components/EditableField";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import { EditMeasureViewSkeleton } from "./EditMeasurePage";
import { EditMeasureViewQuery } from "./__generated__/EditMeasureViewQuery.graphql";

const editMeasureQuery = graphql`
  query EditMeasureViewQuery($measureId: ID!) {
    node(id: $measureId) {
      ... on Measure {
        id
        name
        description
        category
      }
    }
  }
`;

const updateMeasureMutation = graphql`
  mutation EditMeasureViewUpdateMeasureMutation($input: UpdateMeasureInput!) {
    updateMeasure(input: $input) {
      measure {
        id
        name
        description
      }
    }
  }
`;

function EditMeasureViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<EditMeasureViewQuery>;
}) {
  const { organizationId } = useParams<{
    organizationId: string;
  }>();
  const { measureId } = useParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { node: measure } = usePreloadedQuery(editMeasureQuery, queryRef);
  const [formData, setFormData] = useState({
    name: measure.name ?? "",
    description: measure.description ?? "",
    category: measure.category ?? ""
  });

  const [commit, isInFlight] =
    useMutation<EditMeasureViewUpdateMeasureMutation>(updateMeasureMutation);

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

    commit({
      variables: {
        input: {
          id: measureId!,
          name: formData.name,
          description: formData.description,
          category: formData.category
        },
      },
      onCompleted(data, errors) {
        if (errors) {
          toast({
            title: "Error",
            description: errors[0]?.message || "Failed to update measure",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Measure updated successfully",
        });

        navigate(`/organizations/${organizationId}/measures/${measureId}`);
      },
      onError(error) {
        toast({
          title: "Error",
          description: error.message || "Failed to update framework",
          variant: "destructive",
        });
      },
    });
  };

  return (
    <PageTemplate
      title="Edit Measure"
      description="Edit a measure. You will be able to link it to control and risks"
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
                navigate(`/organizations/${organizationId}/measures/${measureId}`)
              }
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isInFlight}>
              {isInFlight ? "Updating..." : "Update Measure"}
            </Button>
          </div>
        </form>
      </Card>
    </PageTemplate>
  );
}

export default function EditMeasureView() {
  const { measureId } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<EditMeasureViewQuery>(editMeasureQuery);

  useEffect(() => {
    if (measureId) {
      loadQuery({ measureId });
    }
  }, [measureId, loadQuery]);

  if (!queryRef) {
    return <EditMeasureViewSkeleton />;
  }

  return (
    <Suspense fallback={<EditMeasureViewSkeleton />}>
      <EditMeasureViewContent queryRef={queryRef} />
    </Suspense>
  );
}
