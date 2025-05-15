import { Suspense, useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import {
  graphql,
  useMutation,
  usePreloadedQuery,
  type PreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { HelpCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { type EditFrameworkViewUpdateFrameworkMutation } from "./__generated__/EditFrameworkViewUpdateFrameworkMutation.graphql";
import { type EditFrameworkViewQuery } from "./__generated__/EditFrameworkViewQuery.graphql";
import { PageTemplate } from "@/components/PageTemplate";
import { EditFrameworkViewSkeleton } from "./EditFrameworkPage";

const updateFrameworkMutation = graphql`
  mutation EditFrameworkViewUpdateFrameworkMutation(
    $input: UpdateFrameworkInput!
  ) {
    updateFramework(input: $input) {
      framework {
        id
        name
        description
      }
    }
  }
`;

const editFrameworkQuery = graphql`
  query EditFrameworkViewQuery($frameworkId: ID!) {
    node(id: $frameworkId) {
      ... on Framework {
        id
        name
        description
      }
    }
  }
`;

function EditableField({
  label,
  value,
  onChange,
  type = "text",
  helpText,
  required,
  multiline = false,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  helpText?: string;
  required?: boolean;
  multiline?: boolean;
}) {
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Label htmlFor={label} className="text-sm font-medium">
          {label}
          {required && <span className="text-red-500">*</span>}
        </Label>
        {helpText && (
          <div className="relative flex items-center">
            <HelpCircle className="h-4 w-4 text-tertiary" />
            <span className="sr-only">{helpText}</span>
          </div>
        )}
      </div>
      {multiline ? (
        <Textarea
          id={label}
          value={value}
          onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
            onChange(e.target.value)
          }
          className={cn(
            "w-full resize-none",
            required && !value && "border-red-500",
          )}
          placeholder={`Enter ${label.toLowerCase()}`}
          rows={4}
        />
      ) : (
        <Input
          id={label}
          type={type}
          value={value}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
            onChange(e.target.value)
          }
          className={cn("w-full", required && !value && "border-red-500")}
          placeholder={`Enter ${label.toLowerCase()}`}
        />
      )}
    </div>
  );
}

function EditFrameworkViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<EditFrameworkViewQuery>;
}) {
  const { organizationId, frameworkId } = useParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const data = usePreloadedQuery(editFrameworkQuery, queryRef);
  const [, setEditedFields] = useState<Set<string>>(new Set());
  const [formData, setFormData] = useState({
    name: "",
    description: "",
  });

  useEffect(() => {
    if (data.node) {
      setFormData({
        name: data.node.name || "",
        description: data.node.description || "",
      });
    }
  }, [data.node]);

  const [commit, isInFlight] =
    useMutation<EditFrameworkViewUpdateFrameworkMutation>(
      updateFrameworkMutation,
    );

  const handleFieldChange = (field: keyof typeof formData, value: unknown) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
    setEditedFields((prev) => new Set(prev).add(field));
  };

  const handleCancel = () => {
    navigate(`/organizations/${organizationId}/frameworks/${frameworkId}`);
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
          id: frameworkId!,
          name: formData.name,
          description: formData.description,
        },
      },
      onCompleted(data, errors) {
        if (errors) {
          toast({
            title: "Error",
            description: errors[0]?.message || "Failed to update framework",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Framework updated successfully",
        });

        navigate(`/organizations/${organizationId}/frameworks/${frameworkId}`);
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
      title="Update Framework"
      description="Update the framework details"
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
            helpText="Provide a detailed description of the framework"
          />

          <div className="flex justify-end gap-3">
            <Button type="button" variant="outline" onClick={handleCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled={isInFlight}>
              {isInFlight ? "Updating..." : "Update Framework"}
            </Button>
          </div>
        </form>
      </Card>
    </PageTemplate>
  );
}

export default function EditFrameworkView() {
  const { frameworkId } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<EditFrameworkViewQuery>(editFrameworkQuery);

  useEffect(() => {
    if (frameworkId) {
      loadQuery({ frameworkId });
    }
  }, [frameworkId, loadQuery]);

  if (!queryRef) {
    return <EditFrameworkViewSkeleton />;
  }

  return (
    <Suspense fallback={<EditFrameworkViewSkeleton />}>
      <EditFrameworkViewContent queryRef={queryRef} />
    </Suspense>
  );
}
