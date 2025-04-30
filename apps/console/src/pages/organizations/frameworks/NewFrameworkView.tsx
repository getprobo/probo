import { Suspense, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { graphql, useMutation, ConnectionHandler } from "react-relay";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { HelpCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { NewFrameworkViewCreateFrameworkMutation } from "./__generated__/NewFrameworkViewCreateFrameworkMutation.graphql";
import { PageTemplate } from "@/components/PageTemplate";
import { NewFrameworkViewSkeleton } from "./NewFrameworkPage";

const createFrameworkMutation = graphql`
  mutation NewFrameworkViewCreateFrameworkMutation(
    $input: CreateFrameworkInput!
    $connections: [ID!]!
  ) {
    createFramework(input: $input) {
      frameworkEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
        }
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
            required && !value && "border-red-500"
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

function NewFrameworkViewContent() {
  const { organizationId } = useParams();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [formData, setFormData] = useState({
    name: "",
    description: "",
  });

  const [commit, isInFlight] =
    useMutation<NewFrameworkViewCreateFrameworkMutation>(
      createFrameworkMutation
    );

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
      "FrameworkListPage_frameworks"
    );

    commit({
      variables: {
        input: {
          organizationId: organizationId!,
          name: formData.name,
          description: formData.description,
        },
        connections: [connectionId],
      },
      onCompleted(data, errors) {
        if (errors) {
          toast({
            title: "Error",
            description: errors[0]?.message || "Failed to create framework",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Framework created successfully",
        });

        navigate(
          `/organizations/${organizationId}/frameworks/${data.createFramework.frameworkEdge.node.id}`
        );
      },
      onError(error) {
        toast({
          title: "Error",
          description: error.message || "Failed to create framework",
          variant: "destructive",
        });
      },
    });
  };

  return (
    <PageTemplate
      title="Create Framework"
      description="Create a new framework to organize your measures"
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
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                navigate(`/organizations/${organizationId}/frameworks`)
              }
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isInFlight}>
              {isInFlight ? "Creating..." : "Create Framework"}
            </Button>
          </div>
        </form>
      </Card>
    </PageTemplate>
  );
}

export default function NewFrameworkView() {
  return (
    <Suspense fallback={<NewFrameworkViewSkeleton />}>
      <NewFrameworkViewContent />
    </Suspense>
  );
}
