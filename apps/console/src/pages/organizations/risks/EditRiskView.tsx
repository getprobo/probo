"use client";

import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import {
  graphql,
  useMutation,
  usePreloadedQuery,
  useQueryLoader,
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
import { Suspense } from "react";
import { EditRiskViewSkeleton } from "./EditRiskPage";
import PeopleSelector from "@/components/PeopleSelector";
import { User } from "lucide-react";
import type { EditRiskViewQuery } from "./__generated__/EditRiskViewQuery.graphql";
import type { EditRiskViewUpdateRiskMutation } from "./__generated__/EditRiskViewUpdateRiskMutation.graphql";
import type { RiskTreatment } from "./__generated__/EditRiskViewUpdateRiskMutation.graphql";

// Query to get risk details
const editRiskViewQuery = graphql`
  query EditRiskViewQuery($riskId: ID!, $organizationId: ID!) {
    risk: node(id: $riskId) {
      ... on Risk {
        id
        name
        description
        category
        inherentLikelihood
        inherentImpact
        residualLikelihood
        residualImpact
        treatment
        owner {
          id
          fullName
        }
      }
    }
    organization: node(id: $organizationId) {
      ...PeopleSelector_organization
    }
  }
`;

// Mutation to update risk
const updateRiskMutation = graphql`
  mutation EditRiskViewUpdateRiskMutation($input: UpdateRiskInput!) {
    updateRisk(input: $input) {
      risk {
        id
        name
        description
        category
        inherentLikelihood
        inherentImpact
        residualLikelihood
        residualImpact
        treatment
        updatedAt
        owner {
          id
          fullName
        }
      }
    }
  }
`;

// Component to handle the actual editing once data is loaded
function EditRiskViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<EditRiskViewQuery>;
}) {
  const navigate = useNavigate();
  const { organizationId, riskId } = useParams<{
    organizationId: string;
    riskId: string;
  }>();
  const { toast } = useToast();

  const data = usePreloadedQuery(editRiskViewQuery, queryRef);
  const risk = data.risk;

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState("");
  const [inherentLikelihood, setInherentLikelihood] = useState<number>(3);
  const [inherentImpact, setInherentImpact] = useState<number>(3);
  const [residualLikelihood, setResidualLikelihood] = useState<number>(3);
  const [residualImpact, setResidualImpact] = useState<number>(3);
  const [treatment, setTreatment] = useState<RiskTreatment>("MITIGATED");
  const [ownerId, setOwnerId] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [updateRisk, isInFlight] =
    useMutation<EditRiskViewUpdateRiskMutation>(updateRiskMutation);

  // Initialize form with risk data
  useEffect(() => {
    if (risk) {
      setName(risk.name || "");
      setDescription(risk.description || "");
      setCategory(risk.category || "");
      // Set values directly as integers
      setInherentLikelihood(risk.inherentLikelihood || 3);
      setInherentImpact(risk.inherentImpact || 3);
      setResidualLikelihood(risk.residualLikelihood || 3);
      setResidualImpact(risk.residualImpact || 3);
      setTreatment(risk.treatment || "MITIGATED");
      setOwnerId(risk.owner?.id || null);
    }
  }, [risk]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast({
        title: "Error",
        description: "Please enter a name for the risk.",
        variant: "destructive",
      });
      return;
    }

    setIsSubmitting(true);

    const input = {
      id: riskId!,
      name,
      description,
      category,
      inherentLikelihood,
      inherentImpact,
      residualLikelihood,
      residualImpact,
      treatment,
      ownerId: ownerId || undefined,
    };

    updateRisk({
      variables: {
        input,
      },
      onCompleted: (response, errors) => {
        setIsSubmitting(false);
        if (errors) {
          console.error("Error updating risk:", errors);
          toast({
            title: "Error",
            description: "Failed to update risk. Please try again.",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Risk updated successfully!",
        });

        navigate(`/organizations/${organizationId}/risks/${riskId}`);
      },
      onError: (error) => {
        setIsSubmitting(false);
        console.error("Error updating risk:", error);
        toast({
          title: "Error",
          description: "Failed to update risk. Please try again.",
          variant: "destructive",
        });
      },
    });
  };

  return (
    <PageTemplate
      title="Edit Risk"
      description="Update the details of this risk"
    >
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Risk Details</CardTitle>
          <CardDescription>Update the details of this risk</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="category">Category</Label>
              <Input
                id="category"
                placeholder="Risk category"
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                required
              />
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
              </Label>
              <PeopleSelector
                organizationRef={data.organization}
                selectedPersonId={ownerId}
                onSelect={setOwnerId}
                placeholder="Select risk owner (optional)"
                required={false}
              />
            </div>

            <Separator />

            <div>
              <h3 className="text-lg font-medium mb-2">
                Initial Risk Assessment
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="inherentLikelihood">
                    Inherent Likelihood (1-5)
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
                  <Label htmlFor="inherentImpact">Inherent Impact (1-5)</Label>
                  <Select
                    value={inherentImpact.toString()}
                    onValueChange={(value) =>
                      setInherentImpact(parseInt(value))
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

            <Separator />

            <div className="space-y-2">
              <Label htmlFor="treatment">Treatment Strategy</Label>
              <Select
                value={treatment}
                onValueChange={(value: RiskTreatment) => setTreatment(value)}
              >
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

            <Separator />

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

            <Separator />

            <div className="flex justify-end gap-3">
              <Button
                type="button"
                variant="outline"
                onClick={() =>
                  navigate(`/organizations/${organizationId}/risks/${riskId}`)
                }
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isInFlight || isSubmitting}>
                {isInFlight || isSubmitting ? "Updating..." : "Update Risk"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </PageTemplate>
  );
}

// Main component that loads the query
export function EditRiskView() {
  const { riskId, organizationId } = useParams<{
    riskId: string;
    organizationId: string;
  }>();
  const [queryRef, loadQuery] =
    useQueryLoader<EditRiskViewQuery>(editRiskViewQuery);

  useEffect(() => {
    if (riskId && organizationId) {
      loadQuery({ riskId, organizationId });
    }
  }, [loadQuery, riskId, organizationId]);

  if (!queryRef) {
    return <EditRiskViewSkeleton />;
  }

  return (
    <Suspense fallback={<EditRiskViewSkeleton />}>
      <EditRiskViewContent queryRef={queryRef} />
    </Suspense>
  );
}

export default EditRiskView;
