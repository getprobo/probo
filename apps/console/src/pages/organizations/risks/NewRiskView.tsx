import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import {
  ConnectionHandler,
  graphql,
  useMutation,
  useQueryLoader,
  usePreloadedQuery,
  PreloadedQuery,
} from "react-relay";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { PageTemplate } from "@/components/PageTemplate";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import PeopleSelector from "@/components/PeopleSelector";
import { User, Loader2 } from "lucide-react";
import { Suspense } from "react";
import type { NewRiskViewQuery } from "./__generated__/NewRiskViewQuery.graphql";

interface RiskTemplate {
  name: string;
  description: string;
  variations: {
    context: string;
    impact: number;
    likelihood: number;
    recommendedTreatment: string;
  }[];
  category: string;
}

const newRiskQuery = graphql`
  query NewRiskViewQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ...PeopleSelector_organization
    }
  }
`;

const createRiskMutation = graphql`
  mutation NewRiskViewCreateRiskMutation(
    $input: CreateRiskInput!
    $connections: [ID!]!
  ) {
    createRisk(input: $input) {
      riskEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          category
          inherentLikelihood
          inherentImpact
          residualLikelihood
          residualImpact
          treatment
          createdAt
          updatedAt
        }
      }
    }
  }
`;

function NewRiskForm({
  queryRef,
}: {
  queryRef: PreloadedQuery<NewRiskViewQuery>;
}) {
  const navigate = useNavigate();
  const { organizationId } = useParams<{ organizationId: string }>();
  const data = usePreloadedQuery<NewRiskViewQuery>(newRiskQuery, queryRef);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [inherentLikelihood, setInherentLikelihood] = useState<number>(3);
  const [inherentImpact, setInherentImpact] = useState<number>(3);
  const [residualLikelihood, setResidualLikelihood] = useState<number>(3);
  const [residualImpact, setResidualImpact] = useState<number>(3);
  const [treatment, setTreatment] = useState<string>("MITIGATED");
  const [ownerId, setOwnerId] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<string>("");
  const [selectedCategory, setSelectedCategory] = useState<string>("");
  const [riskTemplates, setRiskTemplates] = useState<RiskTemplate[]>([]);
  const [ownerError, setOwnerError] = useState<string>("");
  const [isLoadingTemplates, setIsLoadingTemplates] = useState(true);
  const { toast } = useToast();

  const [createRisk, isInFlight] = useMutation(createRiskMutation);

  // Get unique categories from risk templates
  const categories = Array.from(new Set(riskTemplates.map(template => template.category)));

  // Filter risks by selected category
  const filteredRisks = riskTemplates.filter(template => 
    !selectedCategory || template.category === selectedCategory
  ).map(template => ({
    ...template,
    originalIndex: riskTemplates.findIndex(t => t.name === template.name && t.description === template.description)
  }));

  // Handle category selection
  const selectCategory = (category: string) => {
    setSelectedCategory(category);
    // Only reset template if we're changing categories
    if (selectedCategory !== category) {
      setSelectedTemplate("");
    }
    // Focus on the risk dropdown after a short delay to ensure it's rendered
    setTimeout(() => {
      const selectTrigger = document.getElementById('template');
      if (selectTrigger) {
        selectTrigger.focus();
      }
    }, 0);
  };

  useEffect(() => {
    const loadRiskTemplates = async () => {
      try {
        setIsLoadingTemplates(true);
        const response = await fetch("/data/risks/risks.json");
        const data = await response.json();
        setRiskTemplates(data);
      } catch (error) {
        console.error("Error loading risk templates:", error);
        toast({
          title: "Error",
          description: "Failed to load risk templates. Please try again.",
          variant: "destructive",
        });
      } finally {
        setIsLoadingTemplates(false);
      }
    };

    loadRiskTemplates();
  }, [toast]);

  // Handle template selection and prefill form
  const handleTemplateChange = (templateId: string) => {
    setSelectedTemplate(templateId);

    if (templateId === "none") {
      // Clear form if "Select a template" is chosen
      setName("");
      setDescription("");
      setInherentLikelihood(3);
      setInherentImpact(3);
      setResidualLikelihood(3);
      setResidualImpact(3);
      setTreatment("MITIGATED");
      return;
    }

    const selectedIndex = parseInt(templateId);
    const template = riskTemplates[selectedIndex];
    if (template) {
      setName(template.name);
      setDescription(template.description);
      setTreatment("MITIGATED");
    }
  };

  // Clear error when owner is selected
  const handleOwnerSelect = (id: string | null) => {
    setOwnerId(id);
    setOwnerError("");
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    let hasError = false;

    if (!name.trim()) {
      toast({
        title: "Error",
        description: "Please enter a name for the risk.",
        variant: "destructive",
      });
      hasError = true;
    }

    if (!ownerId) {
      setOwnerError("Please select a risk owner");
      hasError = true;
    }

    if (hasError) {
      return;
    }

    setIsSubmitting(true);

    const input = {
      organizationId: organizationId!,
      name,
      description,
      category: selectedCategory,
      inherentLikelihood,
      inherentImpact,
      residualLikelihood,
      residualImpact,
      treatment,
      ownerId: ownerId || undefined,
    };

    createRisk({
      variables: {
        input,
        connections: [
          ConnectionHandler.getConnectionID(
            organizationId!,
            "ListRiskView_risks"
          ),
        ],
      },
      onCompleted: (response, errors) => {
        setIsSubmitting(false);
        if (errors) {
          console.error("Error creating risk:", errors);
          toast({
            title: "Error",
            description: "Failed to create risk. Please try again.",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Risk created successfully!",
        });

        navigate(`/organizations/${organizationId}/risks`);
      },
      onError: (error) => {
        setIsSubmitting(false);
        console.error("Error creating risk:", error);
        toast({
          title: "Error",
          description: "Failed to create risk. Please try again.",
          variant: "destructive",
        });
      },
    });
  };

  return (
    <PageTemplate
      title="Create Risk"
      description="Add a new risk to your organization"
    >
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Risk Details</CardTitle>
          <CardDescription>
            Enter the details of the risk you want to add to your organization
            or select from a template.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="template">Risk Template</Label>
                <div className="flex flex-wrap gap-2 mb-2">
                  {isLoadingTemplates ? (
                    <div className="flex items-center gap-2 text-tertiary">
                      <Loader2 className="h-4 w-4 animate-spin" />
                      <span>Loading categories...</span>
                    </div>
                  ) : (
                    categories.map((category) => (
                      <Button
                        key={category}
                        variant={selectedCategory === category ? "default" : "outline"}
                        size="sm"
                        onClick={() => selectCategory(category)}
                        className="rounded-full"
                      >
                        {category}
                      </Button>
                    ))
                  )}
                </div>
                <Select
                  value={selectedTemplate}
                  onValueChange={handleTemplateChange}
                >
                  <SelectTrigger id="template">
                    <SelectValue placeholder="Select a risk template" />
                  </SelectTrigger>
                  <SelectContent className="max-h-[300px] overflow-y-auto">
                    <SelectItem value="none">Select a template</SelectItem>
                    {filteredRisks.map((template) => (
                      <SelectItem 
                        key={template.originalIndex} 
                        value={template.originalIndex.toString()}
                      >
                        {template.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-sm text-tertiary">
                  Select a risk template to prefill the form or create a custom risk.
                </p>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="Risk name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Describe the risk in detail"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={5}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="owner" className="flex items-center gap-2">
                <User className="h-4 w-4" />
                Risk Owner
                <span className="text-destructive">*</span>
              </Label>
              <PeopleSelector
                organizationRef={data.organization}
                selectedPersonId={ownerId}
                onSelect={handleOwnerSelect}
                placeholder="Select risk owner"
                required={true}
              />
              {ownerError && (
                <p className="text-sm text-destructive mt-1">Please select a risk owner</p>
              )}
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="inherentLikelihood">
                  Initial Likelihood (1-5)
                </Label>
                <Select
                  value={inherentLikelihood.toString()}
                  onValueChange={(value) =>
                    setInherentLikelihood(parseInt(value))
                  }
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select likelihood level" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1">1 - Improbable</SelectItem>
                    <SelectItem value="2">2 - Remote</SelectItem>
                    <SelectItem value="3">3 - Occasional</SelectItem>
                    <SelectItem value="4">4 - Probable</SelectItem>
                    <SelectItem value="5">5 - Frequent</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="inherentImpact">Initial Impact (1-5)</Label>
                <Select
                  value={inherentImpact.toString()}
                  onValueChange={(value) => setInherentImpact(parseInt(value))}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select impact level" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1">1 - Negligible</SelectItem>
                    <SelectItem value="2">2 - Low</SelectItem>
                    <SelectItem value="3">3 - Moderate</SelectItem>
                    <SelectItem value="4">4 - Significant</SelectItem>
                    <SelectItem value="5">5 - Catastrophic</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="treatment">Treatment Strategy</Label>
              <Select value={treatment} onValueChange={setTreatment}>
                <SelectTrigger id="treatment">
                  <SelectValue placeholder="Select a treatment strategy" />
                </SelectTrigger>
                <SelectContent className="max-h-[300px] overflow-y-auto">
                  <SelectItem value="AVOIDED">Avoid</SelectItem>
                  <SelectItem value="MITIGATED">Mitigate</SelectItem>
                  <SelectItem value="TRANSFERRED">Transfer</SelectItem>
                  <SelectItem value="ACCEPTED">Accept</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-tertiary">
                Choose how you plan to address this risk
              </p>
            </div>

            <Separator className="my-4" />

            <div>
              <h3 className="text-lg font-medium mb-2">Residual Risk</h3>
              <p className="text-sm text-tertiary mb-4">
                Estimate the risk after treatment measures have been applied
              </p>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="residualLikelihood">
                    Residual Likelihood (1-5)
                  </Label>
                  <Select
                    value={residualLikelihood.toString()}
                    onValueChange={(value) =>
                      setResidualLikelihood(parseInt(value))
                    }
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select likelihood level" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">1 - Improbable</SelectItem>
                      <SelectItem value="2">2 - Remote</SelectItem>
                      <SelectItem value="3">3 - Occasional</SelectItem>
                      <SelectItem value="4">4 - Probable</SelectItem>
                      <SelectItem value="5">5 - Frequent</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="residualImpact">Residual Impact (1-5)</Label>
                  <Select
                    value={residualImpact.toString()}
                    onValueChange={(value) =>
                      setResidualImpact(parseInt(value))
                    }
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select impact level" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">1 - Negligible</SelectItem>
                      <SelectItem value="2">2 - Low</SelectItem>
                      <SelectItem value="3">3 - Moderate</SelectItem>
                      <SelectItem value="4">4 - Significant</SelectItem>
                      <SelectItem value="5">5 - Catastrophic</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>

            <div className="flex justify-end gap-3">
              <Button
                type="button"
                variant="outline"
                onClick={() =>
                  navigate(`/organizations/${organizationId}/risks`)
                }
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isInFlight || isSubmitting}>
                {isInFlight || isSubmitting ? "Creating..." : "Create Risk"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </PageTemplate>
  );
}

export default function NewRiskView() {
  const { organizationId } = useParams<{ organizationId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<NewRiskViewQuery>(newRiskQuery);

  useEffect(() => {
    if (organizationId) {
      loadQuery({ organizationId });
    }
  }, [organizationId, loadQuery]);

  if (!queryRef) {
    return <div>Loading...</div>;
  }

  return (
    <Suspense fallback={<div>Loading...</div>}>
      <NewRiskForm queryRef={queryRef} />
    </Suspense>
  );
}
