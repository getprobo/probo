import { Suspense, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { graphql, useMutation, ConnectionHandler } from "react-relay";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { HelpCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { NewPeopleViewCreatePeopleMutation } from "./__generated__/NewPeopleViewCreatePeopleMutation.graphql";
import { PageTemplate } from "@/components/PageTemplate";
import { NewPeopleViewSkeleton } from "./NewPeoplePage";

const createPeopleMutation = graphql`
  mutation NewPeopleViewCreatePeopleMutation(
    $input: CreatePeopleInput!
    $connections: [ID!]!
  ) {
    createPeople(input: $input) {
      peopleEdge @prependEdge(connections: $connections) {
        node {
          id
          fullName
          primaryEmailAddress
          additionalEmailAddresses
          kind
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
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  helpText?: string;
  required?: boolean;
}) {
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <HelpCircle className="h-4 w-4 text-tertiary" />
        <Label className="text-sm">{label}</Label>
      </div>
      <div className="space-y-2">
        <Input
          type={type}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          required={required}
        />
        {helpText && <p className="text-sm text-secondary">{helpText}</p>}
      </div>
    </div>
  );
}

function NewPeopleViewContent() {
  const navigate = useNavigate();
  const { organizationId } = useParams();

  const [createPeople] =
    useMutation<NewPeopleViewCreatePeopleMutation>(createPeopleMutation);
  const { toast } = useToast();
  const [formData, setFormData] = useState({
    fullName: "",
    primaryEmailAddress: "",
    additionalEmailAddresses: [] as string[],
    kind: "EMPLOYEE" as "EMPLOYEE" | "CONTRACTOR" | "SERVICE_ACCOUNT",
  });

  const handleFieldChange = (field: keyof typeof formData, value: unknown) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    createPeople({
      variables: {
        connections: [
          ConnectionHandler.getConnectionID(
            organizationId!,
            "PeopleListView_peoples",
            {
              orderBy: {
                direction: "ASC",
                field: "FULL_NAME",
              },
            }
          ),
          ConnectionHandler.getConnectionID(
            organizationId!,
            "PeopleSelector_organization_peoples",
            {
              orderBy: {
                direction: "ASC",
                field: "FULL_NAME",
              },
            }
          ),
          ConnectionHandler.getConnectionID(
            organizationId!,
            "MeasureView_peoples",
            {
              orderBy: {
                direction: "ASC",
                field: "FULL_NAME",
              },
            }
          ),
        ],
        input: {
          organizationId: organizationId!,
          fullName: formData.fullName,
          primaryEmailAddress: formData.primaryEmailAddress,
          additionalEmailAddresses: formData.additionalEmailAddresses,
          kind: formData.kind,
        },
      },
      onCompleted: () => {
        toast({
          title: "Success",
          description: "Person created successfully",
          variant: "default",
        });
        navigate(`/organizations/${organizationId}/people`);
      },
      onError: (error) => {
        toast({
          title: "Error",
          description: error.message || "Failed to create person",
          variant: "destructive",
        });
      },
    });
  };

  return (
    <PageTemplate
      title="Create Person"
      description="Add a new person interacting with organization"
    >
      <form onSubmit={handleSubmit}>
        <div className="max-w-2xl space-y-6">
          <EditableField
            label="Full Name"
            value={formData.fullName}
            onChange={(value) => handleFieldChange("fullName", value)}
            required
          />

          <EditableField
            label="Primary Email"
            value={formData.primaryEmailAddress}
            type="email"
            onChange={(value) =>
              handleFieldChange("primaryEmailAddress", value)
            }
            required
          />

          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <HelpCircle className="h-4 w-4 text-tertiary" />
              <Label className="text-sm">Additional Email Addresses</Label>
            </div>
            <div className="space-y-2">
              {formData.additionalEmailAddresses.map((email, index) => (
                <div key={index} className="flex gap-2">
                  <Input
                    type="email"
                    value={email}
                    onChange={(e) => {
                      const newEmails = [...formData.additionalEmailAddresses];
                      newEmails[index] = e.target.value;
                      handleFieldChange("additionalEmailAddresses", newEmails);
                    }}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      const newEmails =
                        formData.additionalEmailAddresses.filter(
                          (_, i) => i !== index
                        );
                      handleFieldChange("additionalEmailAddresses", newEmails);
                    }}
                  >
                    Remove
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  handleFieldChange("additionalEmailAddresses", [
                    ...formData.additionalEmailAddresses,
                    "",
                  ]);
                }}
              >
                Add Email
              </Button>
            </div>
          </div>

          <Card className="p-6">
            <div className="space-y-4">
              <div className="space-y-2">
                <h2 className="text-lg font-medium">Additional Information</h2>
                <p className="text-sm text-secondary">
                  Additional details about the person
                </p>
              </div>

              <div className="space-y-4">
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <HelpCircle className="h-4 w-4 text-tertiary" />
                    <Label className="text-sm">Kind</Label>
                  </div>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => handleFieldChange("kind", "EMPLOYEE")}
                      className={cn(
                        "rounded-full px-4 py-1 text-sm transition-colors",
                        formData.kind === "EMPLOYEE"
                          ? "bg-blue-100 text-blue-900 ring-2 ring-blue-600 ring-offset-2"
                          : "bg-secondary-bg text-primary hover:bg-h-secondary-bg"
                      )}
                    >
                      Employee
                    </button>
                    <button
                      type="button"
                      onClick={() => handleFieldChange("kind", "CONTRACTOR")}
                      className={cn(
                        "rounded-full px-4 py-1 text-sm transition-colors",
                        formData.kind === "CONTRACTOR"
                          ? "bg-purple-100 text-purple-900 ring-2 ring-purple-600 ring-offset-2"
                          : "bg-secondary-bg text-primary hover:bg-h-secondary-bg"
                      )}
                    >
                      Contractor
                    </button>
                    <button
                      type="button"
                      onClick={() =>
                        handleFieldChange("kind", "SERVICE_ACCOUNT")
                      }
                      className={cn(
                        "rounded-full px-4 py-1 text-sm transition-colors",
                        formData.kind === "SERVICE_ACCOUNT"
                          ? "bg-green-100 text-green-900 ring-2 ring-green-600 ring-offset-2"
                          : "bg-secondary-bg text-primary hover:bg-h-secondary-bg"
                      )}
                    >
                      Service Account
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </Card>
        </div>
        <div className="fixed bottom-6 right-6 flex gap-2">
          <Button type="button" variant="outline" onClick={() => navigate(-1)}>
            Cancel
          </Button>
          <Button
            type="submit"
            className="bg-primary text-invert hover:bg-primary/90"
          >
            Create Person
          </Button>
        </div>
      </form>
    </PageTemplate>
  );
}

export default function NewPeopleView() {
  return (
    <Suspense fallback={<NewPeopleViewSkeleton />}>
      <NewPeopleViewContent />
    </Suspense>
  );
}
