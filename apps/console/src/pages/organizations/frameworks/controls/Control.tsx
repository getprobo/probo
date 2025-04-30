import { Button } from "@/components/ui/button";
import { DialogHeader, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { useToast } from "@/hooks/use-toast";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from "@radix-ui/react-dialog";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@radix-ui/react-select";
import { Search, Loader2, X, LinkIcon } from "lucide-react";
import { useState, useEffect, useCallback } from "react";
import {
  fetchQuery,
  graphql,
  useFragment,
  useMutation,
  useRelayEnvironment,
} from "react-relay";
import { useParams, Link } from "react-router";
import {
  ControlLinkedMeasuresQuery$data,
  ControlLinkedMeasuresQuery,
} from "./__generated__/ControlLinkedMeasuresQuery.graphql";
import {
  ControlOrganizationMeasuresQuery$data,
  ControlOrganizationMeasuresQuery,
} from "./__generated__/ControlOrganizationMeasuresQuery.graphql";
import { ControlFragment_Control$key } from "./__generated__/ControlFragment_Control.graphql";
import {
  ControlLinkedPoliciesQuery$data,
  ControlLinkedPoliciesQuery,
} from "./__generated__/ControlLinkedPoliciesQuery.graphql";
import {
  ControlOrganizationPoliciesQuery$data,
  ControlOrganizationPoliciesQuery,
} from "./__generated__/ControlOrganizationPoliciesQuery.graphql";

const controlFragment = graphql`
  fragment ControlFragment_Control on Control {
    id
    description
    name
    referenceId
  }
`;

// New query to fetch linked measures
const linkedMeasuresQuery = graphql`
  query ControlLinkedMeasuresQuery($controlId: ID!) {
    control: node(id: $controlId) {
      id
      ... on Control {
        measures(first: 100) @connection(key: "Control__measures") {
          edges {
            node {
              id
              name
              description
              category
              state
            }
          }
        }
      }
    }
  }
`;

// Query to fetch all measures for the organization
const organizationMeasuresQuery = graphql`
  query ControlOrganizationMeasuresQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        measures(first: 100) @connection(key: "Organization__measures") {
          edges {
            node {
              id
              name
              description
              category
              state
            }
          }
        }
      }
    }
  }
`;

// Query to fetch linked policies
const linkedPoliciesQuery = graphql`
  query ControlLinkedPoliciesQuery($controlId: ID!) {
    control: node(id: $controlId) {
      id
      ... on Control {
        policies(first: 100) @connection(key: "Control__policies") {
          edges {
            node {
              id
              title
              description
              currentPublishedVersion
              createdAt
              updatedAt
              owner {
                id
                fullName
              }
            }
          }
        }
      }
    }
  }
`;

// Query to fetch all policies for the organization
const organizationPoliciesQuery = graphql`
  query ControlOrganizationPoliciesQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        policies(first: 100) @connection(key: "Organization__policies") {
          edges {
            node {
              id
              title
              description
              currentPublishedVersion
              createdAt
              updatedAt
              owner {
                id
                fullName
              }
            }
          }
        }
      }
    }
  }
`;

// Mutation to create a mapping between a control and a measure
const createMeasureMappingMutation = graphql`
  mutation ControlCreateMeasureMappingMutation(
    $input: CreateControlMeasureMappingInput!
  ) {
    createControlMeasureMapping(input: $input) {
      success
    }
  }
`;

// Mutation to delete a mapping between a control and a measure
const deleteMeasureMappingMutation = graphql`
  mutation ControlDeleteMeasureMappingMutation(
    $input: DeleteControlMeasureMappingInput!
  ) {
    deleteControlMeasureMapping(input: $input) {
      success
    }
  }
`;

// Mutation to create a mapping between a control and a policy
const createPolicyMappingMutation = graphql`
  mutation ControlCreatePolicyMappingMutation(
    $input: CreateControlPolicyMappingInput!
  ) {
    createControlPolicyMapping(input: $input) {
      success
    }
  }
`;

// Mutation to delete a mapping between a control and a policy
const deletePolicyMappingMutation = graphql`
  mutation ControlDeletePolicyMappingMutation(
    $input: DeleteControlPolicyMappingInput!
  ) {
    deleteControlPolicyMapping(input: $input) {
      success
    }
  }
`;

export function Control({
  controlKey,
}: {
  controlKey: ControlFragment_Control$key;
}) {
  const { organizationId /* frameworkId */ } = useParams<{
    organizationId: string;
    frameworkId: string;
  }>();
  const control = useFragment(controlFragment, controlKey);
  const { toast } = useToast();
  const environment = useRelayEnvironment();

  // State for measure mapping
  const [isMeasureMappingDialogOpen, setIsMeasureMappingDialogOpen] =
    useState(false);
  const [linkedMeasuresData, setLinkedMeasuresData] =
    useState<ControlLinkedMeasuresQuery$data | null>(null);
  const [organizationMeasuresData, setOrganizationMeasuresData] =
    useState<ControlOrganizationMeasuresQuery$data | null>(null);
  const [measureSearchQuery, setMeasureSearchQuery] = useState("");
  const [isLoadingMeasures, setIsLoadingMeasures] = useState(false);
  const [isLinkingMeasure, setIsLinkingMeasure] = useState(false);
  const [isUnlinkingMeasure, setIsUnlinkingMeasure] = useState(false);
  const [categoryFilter, setCategoryFilter] = useState<string | null>(null);

  // Policy state
  const [isPolicyMappingDialogOpen, setIsPolicyMappingDialogOpen] =
    useState(false);
  const [linkedPoliciesData, setLinkedPoliciesData] =
    useState<ControlLinkedPoliciesQuery$data | null>(null);
  const [organizationPoliciesData, setOrganizationPoliciesData] =
    useState<ControlOrganizationPoliciesQuery$data | null>(null);
  const [policySearchQuery, setPolicySearchQuery] = useState("");
  const [isLoadingPolicies, setIsLoadingPolicies] = useState(false);
  const [isLinkingPolicy, setIsLinkingPolicy] = useState(false);
  const [isUnlinkingPolicy, setIsUnlinkingPolicy] = useState(false);

  // Create mutation hooks
  const [commitCreateMeasureMapping] = useMutation(createMeasureMappingMutation);
  const [commitDeleteMeasureMapping] = useMutation(deleteMeasureMappingMutation);
  const [commitCreatePolicyMapping] = useMutation(createPolicyMappingMutation);
  const [commitDeletePolicyMapping] = useMutation(deletePolicyMappingMutation);

  // Load initial linked measures data
  useEffect(() => {
    if (control.id) {
      setIsLoadingMeasures(true);
      fetchQuery<ControlLinkedMeasuresQuery>(environment, linkedMeasuresQuery, {
        controlId: control.id,
      }).subscribe({
        next: (data) => {
          setLinkedMeasuresData(data);
          setIsLoadingMeasures(false);
        },
        error: (error: Error) => {
          console.error("Error loading initial measures:", error);
          setIsLoadingMeasures(false);
        },
      });
    }
  }, [control.id, environment]);

  // Load measures data
  const loadMeasuresData = useCallback(() => {
    if (!organizationId || !control.id) return;

    setIsLoadingMeasures(true);

    // Fetch all measures for the organization
    fetchQuery<ControlOrganizationMeasuresQuery>(
      environment,
      organizationMeasuresQuery,
      {
        organizationId,
      }
    ).subscribe({
      next: (data) => {
        setOrganizationMeasuresData(data);
      },
      complete: () => {
        // Fetch linked measures for this control
        fetchQuery<ControlLinkedMeasuresQuery>(environment, linkedMeasuresQuery, {
          controlId: control.id,
        }).subscribe({
          next: (data) => {
            setLinkedMeasuresData(data);
            setIsLoadingMeasures(false);
          },
          error: (error: Error) => {
            console.error("Error fetching linked measures:", error);
            setIsLoadingMeasures(false);
            toast({
              title: "Error",
              description: "Failed to load linked measures.",
              variant: "destructive",
            });
          },
        });
      },
      error: (error: Error) => {
        console.error("Error fetching organization measures:", error);
        setIsLoadingMeasures(false);
        toast({
          title: "Error",
          description: "Failed to load measures.",
          variant: "destructive",
        });
      },
    });
  }, [control.id, environment, organizationId, toast]);

  // Helper functions
  const getMeasures = useCallback(() => {
    if (!organizationMeasuresData?.organization?.measures?.edges) return [];
    return organizationMeasuresData.organization.measures.edges.map(
      (edge) => edge.node
    );
  }, [organizationMeasuresData]);

  const getLinkedMeasures = useCallback(() => {
    if (!linkedMeasuresData?.control?.measures?.edges) return [];
    return linkedMeasuresData.control.measures.edges.map((edge) => edge.node);
  }, [linkedMeasuresData]);

  const isMeasureLinked = useCallback(
    (measureId: string) => {
      const linkedMeasures = getLinkedMeasures();
      return linkedMeasures.some((measure) => measure.id === measureId);
    },
    [getLinkedMeasures]
  );

  const getMeasureCategories = useCallback(() => {
    const measures = getMeasures();
    const categories = new Set<string>();

    measures.forEach((measure) => {
      if (measure.category) {
        categories.add(measure.category);
      }
    });

    return Array.from(categories).sort();
  }, [getMeasures]);

  const filteredMeasures = useCallback(() => {
    const measures = getMeasures();
    if (!measureSearchQuery && !categoryFilter) return measures;

    return measures.filter((measure) => {
      // Filter by search query
      const matchesSearch =
        !measureSearchQuery ||
        measure.name.toLowerCase().includes(measureSearchQuery.toLowerCase()) ||
        (measure.description &&
          measure.description
            .toLowerCase()
            .includes(measureSearchQuery.toLowerCase()));

      // Filter by category
      const matchesCategory =
        !categoryFilter ||
        categoryFilter === "all" ||
        measure.category === categoryFilter;

      return matchesSearch && matchesCategory;
    });
  }, [categoryFilter, getMeasures, measureSearchQuery]);

  // Handle link/unlink functions
  const handleLinkMeasure = useCallback(
    (measureId: string) => {
      if (!control.id) return;

      setIsLinkingMeasure(true);

      commitCreateMeasureMapping({
        variables: {
          input: {
            controlId: control.id,
            measureId: measureId,
          },
        },
        onCompleted: (_, errors) => {
          setIsLinkingMeasure(false);

          if (errors) {
            console.error("Error linking measure:", errors);
            toast({
              title: "Error",
              description: "Failed to link measure. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh linked measures data
          fetchQuery<ControlLinkedMeasuresQuery>(
            environment,
            linkedMeasuresQuery,
            {
              controlId: control.id,
            }
          ).subscribe({
            next: (data) => {
              setLinkedMeasuresData(data);
            },
            error: (error: Error) => {
              console.error("Error refreshing linked measures:", error);
            },
          });

          toast({
            title: "Success",
            description: "Measure successfully linked to control.",
          });
        },
        onError: (error) => {
          setIsLinkingMeasure(false);
          console.error("Error linking measure:", error);
          toast({
            title: "Error",
            description: "Failed to link measure. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [commitCreateMeasureMapping, control.id, environment, toast]
  );

  const handleUnlinkMeasure = useCallback(
    (measureId: string) => {
      if (!control.id) return;

      setIsUnlinkingMeasure(true);

      commitDeleteMeasureMapping({
        variables: {
          input: {
            controlId: control.id,
            measureId: measureId,
          },
        },
        onCompleted: (_, errors) => {
          setIsUnlinkingMeasure(false);

          if (errors) {
            console.error("Error unlinking measure:", errors);
            toast({
              title: "Error",
              description: "Failed to unlink measure. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh linked measures data
          fetchQuery<ControlLinkedMeasuresQuery>(
            environment,
            linkedMeasuresQuery,
            {
              controlId: control.id,
            }
          ).subscribe({
            next: (data) => {
              setLinkedMeasuresData(data);
            },
            error: (error: Error) => {
              console.error("Error refreshing linked measures:", error);
            },
          });

          toast({
            title: "Success",
            description: "Measure successfully unlinked from control.",
          });
        },
        onError: (error) => {
          setIsUnlinkingMeasure(false);
          console.error("Error unlinking measure:", error);
          toast({
            title: "Error",
            description: "Failed to unlink measure. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [commitDeleteMeasureMapping, control.id, environment, toast]
  );

  const handleOpenMeasureMappingDialog = useCallback(() => {
    loadMeasuresData();
    setIsMeasureMappingDialogOpen(true);
  }, [loadMeasuresData]);

  // Load initial linked policies data
  useEffect(() => {
    if (control.id) {
      setIsLoadingPolicies(true);
      fetchQuery<ControlLinkedPoliciesQuery>(environment, linkedPoliciesQuery, {
        controlId: control.id,
      }).subscribe({
        next: (data) => {
          setLinkedPoliciesData(data);
          setIsLoadingPolicies(false);
        },
        error: (error: Error) => {
          console.error("Error loading initial policies:", error);
          setIsLoadingPolicies(false);
        },
      });
    }
  }, [control.id, environment]);

  // Load policies data
  const loadPoliciesData = useCallback(() => {
    if (!organizationId || !control.id) return;

    setIsLoadingPolicies(true);

    // Fetch all policies for the organization
    fetchQuery<ControlOrganizationPoliciesQuery>(
      environment,
      organizationPoliciesQuery,
      {
        organizationId,
      }
    ).subscribe({
      next: (data) => {
        setOrganizationPoliciesData(data);
      },
      complete: () => {
        // Fetch linked policies for this control
        fetchQuery<ControlLinkedPoliciesQuery>(
          environment,
          linkedPoliciesQuery,
          {
            controlId: control.id,
          }
        ).subscribe({
          next: (data) => {
            setLinkedPoliciesData(data);
            setIsLoadingPolicies(false);
          },
          error: (error: Error) => {
            console.error("Error fetching linked policies:", error);
            setIsLoadingPolicies(false);
            toast({
              title: "Error",
              description: "Failed to load linked policies.",
              variant: "destructive",
            });
          },
        });
      },
      error: (error: Error) => {
        console.error("Error fetching organization policies:", error);
        setIsLoadingPolicies(false);
        toast({
          title: "Error",
          description: "Failed to load policies.",
          variant: "destructive",
        });
      },
    });
  }, [control.id, environment, organizationId, toast]);

  // Policy helper functions
  const getPolicies = useCallback(() => {
    if (!organizationPoliciesData?.organization?.policies?.edges) return [];
    return organizationPoliciesData.organization.policies.edges.map(
      (edge) => edge.node
    );
  }, [organizationPoliciesData]);

  const getLinkedPolicies = useCallback(() => {
    if (!linkedPoliciesData?.control?.policies?.edges) return [];
    return linkedPoliciesData.control.policies.edges.map((edge) => edge.node);
  }, [linkedPoliciesData]);

  const isPolicyLinked = useCallback(
    (policyId: string) => {
      const linkedPolicies = getLinkedPolicies();
      return linkedPolicies.some((policy) => policy.id === policyId);
    },
    [getLinkedPolicies]
  );

  const filteredPolicies = useCallback(() => {
    const policies = getPolicies();
    if (!policySearchQuery) return policies;

    return policies.filter((policy) => {
      return (
        !policySearchQuery ||
        policy.title?.toLowerCase().includes(policySearchQuery.toLowerCase()) ||
        (policy.description &&
          policy.description
            .toLowerCase()
            .includes(policySearchQuery.toLowerCase()))
      );
    });
  }, [getPolicies, policySearchQuery]);

  // Policy link/unlink handlers
  const handleLinkPolicy = useCallback(
    (policyId: string) => {
      if (!control.id) return;

      setIsLinkingPolicy(true);

      commitCreatePolicyMapping({
        variables: {
          input: {
            controlId: control.id,
            policyId: policyId,
          },
        },
        onCompleted: (_, errors) => {
          setIsLinkingPolicy(false);

          if (errors) {
            console.error("Error linking policy:", errors);
            toast({
              title: "Error",
              description: "Failed to link policy. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh linked policies data
          fetchQuery<ControlLinkedPoliciesQuery>(
            environment,
            linkedPoliciesQuery,
            {
              controlId: control.id,
            }
          ).subscribe({
            next: (data) => {
              setLinkedPoliciesData(data);
            },
            error: (error: Error) => {
              console.error("Error refreshing linked policies:", error);
            },
          });

          toast({
            title: "Success",
            description: "Policy successfully linked to control.",
          });
        },
        onError: (error) => {
          setIsLinkingPolicy(false);
          console.error("Error linking policy:", error);
          toast({
            title: "Error",
            description: "Failed to link policy. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [commitCreatePolicyMapping, control.id, environment, toast]
  );

  const handleUnlinkPolicy = useCallback(
    (policyId: string) => {
      if (!control.id) return;

      setIsUnlinkingPolicy(true);

      commitDeletePolicyMapping({
        variables: {
          input: {
            controlId: control.id,
            policyId: policyId,
          },
        },
        onCompleted: (_, errors) => {
          setIsUnlinkingPolicy(false);

          if (errors) {
            console.error("Error unlinking policy:", errors);
            toast({
              title: "Error",
              description: "Failed to unlink policy. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh linked policies data
          fetchQuery<ControlLinkedPoliciesQuery>(
            environment,
            linkedPoliciesQuery,
            {
              controlId: control.id,
            }
          ).subscribe({
            next: (data) => {
              setLinkedPoliciesData(data);
            },
            error: (error: Error) => {
              console.error("Error refreshing linked policies:", error);
            },
          });

          toast({
            title: "Success",
            description: "Policy successfully unlinked from control.",
          });
        },
        onError: (error) => {
          setIsUnlinkingPolicy(false);
          console.error("Error unlinking policy:", error);
          toast({
            title: "Error",
            description: "Failed to unlink policy. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [commitDeletePolicyMapping, control.id, environment, toast]
  );

  const handleOpenPolicyMappingDialog = useCallback(() => {
    loadPoliciesData();
    setIsPolicyMappingDialogOpen(true);
  }, [loadPoliciesData]);

  const formatState = (state: string | undefined): string => {
    if (!state) return "Unknown";

    switch (state) {
      case "NOT_STARTED":
        return "Not Started";
      case "IN_PROGRESS":
        return "In Progress";
      case "IMPLEMENTED":
        return "Implemented";
      case "NOT_APPLICABLE":
        return "Not Applicable";
      default:
        return state;
    }
  };

  const getStateColor = (state: string | undefined): string => {
    if (!state) return "bg-secondary-bg text-secondary";

    switch (state) {
      case "NOT_STARTED":
        return "bg-secondary-bg text-secondary";
      case "IN_PROGRESS":
        return "bg-info-bg text-info";
      case "IMPLEMENTED":
        return "bg-success-bg text-success";
      case "NOT_APPLICABLE":
        return "bg-warning-bg text-warning";
      default:
        return "bg-secondary-bg text-secondary";
    }
  };

  return (
    <div className="w-auto p-5 flex items-start gap-5">
      <div className="font-mono text-lg px-1 py-0.25 rounded-sm bg-active-bg border-mid-b border font-bold">
        {control.referenceId}
      </div>
      <div className="flex-1">
        <h2 className="text-2xl font-medium">{control.name}</h2>

        {/* Control Description */}
        {control.description && (
          <div className="mt-4 text-tertiary">{control.description}</div>
        )}

        {/* Security Measures Section */}
        <div className="mt-8">
          {/* Measure Mapping Dialog */}
          <Dialog
            open={isMeasureMappingDialogOpen}
            onOpenChange={setIsMeasureMappingDialogOpen}
          >
            <DialogContent className="max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
              <DialogHeader>
                <DialogTitle>Link Security Measures to Control</DialogTitle>
                <DialogDescription>
                  Search and select security measures to link to this control.
                  This helps track which security measures address this control.
                </DialogDescription>
              </DialogHeader>

              <div className="flex items-center space-x-4 mb-4">
                <div className="flex-1">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-tertiary" />
                    <Input
                      placeholder="Search security measures by name or description..."
                      value={measureSearchQuery}
                      onChange={(e) => setMeasureSearchQuery(e.target.value)}
                      className="w-full pl-10"
                    />
                  </div>
                </div>
                <div className="w-[200px]">
                  <Select
                    value={categoryFilter || "all"}
                    onValueChange={(value) =>
                      setCategoryFilter(value === "all" ? null : value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All categories" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All categories</SelectItem>
                      {getMeasureCategories().map((category) => (
                        <SelectItem key={category} value={category}>
                          {category}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="flex-1 overflow-hidden">
                {isLoadingMeasures ? (
                  <div className="flex items-center justify-center h-full">
                    <Loader2 className="w-8 h-8 animate-spin text-info" />
                    <span className="ml-2">Loading security measures...</span>
                  </div>
                ) : (
                  <div className="max-h-[50vh] overflow-y-auto pr-2">
                    {filteredMeasures().length === 0 ? (
                      <div className="text-center py-8 text-secondary">
                        No security measures found. Try adjusting your search or
                        select a different category.
                      </div>
                    ) : (
                      <table className="w-full bg-level-1">
                        <thead className="sticky top-0 bg-white">
                          <tr className="border-b text-left text-sm text-secondary bg-invert-bg">
                            <th className="py-3 px-4 font-medium">Name</th>
                            <th className="py-3 px-4 font-medium">State</th>
                            <th className="py-3 px-4 font-medium text-right">
                              Actions
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          {filteredMeasures().map((measure) => {
                            const isLinked = isMeasureLinked(measure.id);
                            return (
                              <tr
                                key={measure.id}
                                className="border-b hover:bg-invert-bg"
                              >
                                <td className="py-3 px-4">
                                  <div className="font-medium">
                                    {measure.name}
                                  </div>
                                  {measure.description && (
                                    <div className="text-xs text-secondary line-clamp-1 mt-0.5">
                                      {measure.description}
                                    </div>
                                  )}
                                </td>
                                <td className="py-3 px-4">
                                  <div
                                    className={`px-2 py-0.5 rounded-full text-xs ${getStateColor(
                                      measure.state
                                    )} inline-block`}
                                  >
                                    {formatState(measure.state)}
                                  </div>
                                </td>
                                <td className="py-3 px-4 text-right whitespace-nowrap">
                                  {isLinked ? (
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      onClick={() =>
                                        handleUnlinkMeasure(measure.id)
                                      }
                                      disabled={isUnlinkingMeasure}
                                      className="text-xs h-7 text-danger border-danger-b hover:bg-h-danger-bg"
                                    >
                                      {isUnlinkingMeasure ? (
                                        <Loader2 className="w-4 h-4 animate-spin" />
                                      ) : (
                                        <X className="w-4 h-4" />
                                      )}
                                      <span className="ml-1">Unlink</span>
                                    </Button>
                                  ) : (
                                    <Button
                                      variant="secondary"
                                      size="sm"
                                      onClick={() =>
                                        handleLinkMeasure(measure.id)
                                      }
                                      disabled={isLinkingMeasure}
                                      className="text-xs h-7  text-info border-info-b hover:bg-h-info-bg"
                                    >
                                      {isLinkingMeasure ? (
                                        <Loader2 className="w-4 h-4 animate-spin" />
                                      ) : (
                                        <LinkIcon className="w-4 h-4" />
                                      )}
                                      <span className="ml-1">Link</span>
                                    </Button>
                                  )}
                                </td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    )}
                  </div>
                )}
              </div>

              <DialogFooter className="mt-4">
                <Button onClick={() => setIsMeasureMappingDialogOpen(false)}>
                  Close
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Linked Measures List */}
          <div>
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-xl font-medium text-secondary">
                Security measures
              </h3>
              <Button
                variant="outline"
                size="sm"
                className="flex items-center gap-1"
                onClick={handleOpenMeasureMappingDialog}
              >
                <LinkIcon className="w-4 h-4" />
                <span>Link Security Measures</span>
              </Button>
            </div>

            {isLoadingMeasures ? (
              <div className="flex items-center justify-center h-24">
                <Loader2 className="w-6 h-6 animate-spin text-info" />
                <span className="ml-2">Loading security measures...</span>
              </div>
            ) : linkedMeasuresData?.control?.measures?.edges &&
              linkedMeasuresData.control.measures.edges.length > 0 ? (
              <div className="overflow-x-auto border rounded-md">
                <table className="w-full">
                  <thead>
                    <tr className="border-b text-left text-sm text-secondary bg-invert-bg">
                      <th className="py-3 px-4 font-medium">Name</th>
                      <th className="py-3 px-4 font-medium">State</th>
                      <th className="py-3 px-4 font-medium text-right">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {getLinkedMeasures().map((measure) => (
                      <tr
                        key={measure.id}
                        className="border-b hover:bg-invert-bg"
                      >
                        <td className="py-3 px-4">
                          <div className="font-medium">{measure.name}</div>
                          {measure.description && (
                            <div className="text-xs text-secondary line-clamp-1 mt-0.5">
                              {measure.description}
                            </div>
                          )}
                        </td>
                        <td className="py-3 px-4">
                          <div
                            className={`px-2 py-0.5 rounded-full text-xs ${getStateColor(
                              measure.state
                            )} inline-block`}
                          >
                            {formatState(measure.state)}
                          </div>
                        </td>
                        <td className="py-3 px-4 text-right whitespace-nowrap">
                          <div className="flex gap-2 justify-end">
                            <Button
                              variant="outline"
                              size="sm"
                              asChild
                              className="text-xs h-7"
                            >
                              <Link
                                to={`/organizations/${organizationId}/measures/${measure.id}`}
                              >
                                View
                              </Link>
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleUnlinkMeasure(measure.id)}
                              disabled={isUnlinkingMeasure}
                              className="text-xs h-7 text-danger border-danger-b hover:bg-h-danger-bg"
                            >
                              Unlink
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-8 text-secondary border rounded-md">
                No security measures linked to this control yet. Click
                &quot;Link Security Measures&quot; to connect some.
              </div>
            )}
          </div>
        </div>

        {/* Policies Section */}
        <div className="mt-8">
          {/* Policy Mapping Dialog */}
          <Dialog
            open={isPolicyMappingDialogOpen}
            onOpenChange={setIsPolicyMappingDialogOpen}
          >
            <DialogContent className="max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
              <DialogHeader>
                <DialogTitle>Link Policies to Control</DialogTitle>
                <DialogDescription>
                  Search and select policies to link to this control. This helps
                  track which policies address this control.
                </DialogDescription>
              </DialogHeader>

              <div className="flex items-center space-x-4 mb-4">
                <div className="flex-1">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-tertiary" />
                    <Input
                      placeholder="Search policies by name or content..."
                      value={policySearchQuery}
                      onChange={(e) => setPolicySearchQuery(e.target.value)}
                      className="w-full pl-10"
                    />
                  </div>
                </div>
              </div>

              <div className="flex-1 overflow-hidden">
                {isLoadingPolicies ? (
                  <div className="flex items-center justify-center h-full">
                    <Loader2 className="w-8 h-8 animate-spin text-info" />
                    <span className="ml-2">Loading policies...</span>
                  </div>
                ) : (
                  <div className="max-h-[50vh] overflow-y-auto pr-2">
                    {filteredPolicies().length === 0 ? (
                      <div className="text-center py-8 text-secondary">
                        No policies found. Try adjusting your search.
                      </div>
                    ) : (
                      <table className="w-full bg-level-1">
                        <thead className="sticky top-0 bg-white">
                          <tr className="border-b text-left text-sm text-secondary bg-invert-bg">
                            <th className="py-3 px-4 font-medium">Name</th>
                            <th className="py-3 px-4 font-medium">Review Date</th>
                            <th className="py-3 px-4 font-medium text-right">
                              Actions
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          {filteredPolicies().map((policy) => {
                            const isLinked = isPolicyLinked(policy.id);
                            return (
                              <tr
                                key={policy.id}
                                className="border-b hover:bg-invert-bg"
                              >
                                <td className="py-3 px-4">
                                  <div className="font-medium">
                                    {policy.title}
                                  </div>
                                  {policy.description && (
                                    <div className="text-xs text-secondary line-clamp-1 mt-0.5">
                                      {policy.description}
                                    </div>
                                  )}
                                </td>
                                <td className="py-3 px-4">
                                  {policy.updatedAt
                                    ? new Date(
                                        policy.updatedAt
                                      ).toLocaleDateString()
                                    : "Not set"}
                                </td>
                                <td className="py-3 px-4 text-right whitespace-nowrap">
                                  {isLinked ? (
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      onClick={() =>
                                        handleUnlinkPolicy(policy.id)
                                      }
                                      disabled={isUnlinkingPolicy}
                                      className="text-xs h-7 text-danger border-danger-b hover:bg-h-danger-bg"
                                    >
                                      {isUnlinkingPolicy ? (
                                        <Loader2 className="w-4 h-4 animate-spin" />
                                      ) : (
                                        <X className="w-4 h-4" />
                                      )}
                                      <span className="ml-1">Unlink</span>
                                    </Button>
                                  ) : (
                                    <Button
                                      variant="secondary"
                                      size="sm"
                                      onClick={() =>
                                        handleLinkPolicy(policy.id)
                                      }
                                      disabled={isLinkingPolicy}
                                      className="text-xs h-7  text-info border-info-b hover:bg-h-info-bg"
                                    >
                                      {isLinkingPolicy ? (
                                        <Loader2 className="w-4 h-4 animate-spin" />
                                      ) : (
                                        <LinkIcon className="w-4 h-4" />
                                      )}
                                      <span className="ml-1">Link</span>
                                    </Button>
                                  )}
                                </td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    )}
                  </div>
                )}
              </div>

              <DialogFooter className="mt-4">
                <Button onClick={() => setIsPolicyMappingDialogOpen(false)}>
                  Close
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Linked Policies List */}
          <div>
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-xl font-medium text-secondary">Policies</h3>
              <Button
                variant="outline"
                size="sm"
                className="flex items-center gap-1"
                onClick={handleOpenPolicyMappingDialog}
              >
                <LinkIcon className="w-4 h-4" />
                <span>Link Policies</span>
              </Button>
            </div>

            {isLoadingPolicies ? (
              <div className="flex items-center justify-center h-24">
                <Loader2 className="w-6 h-6 animate-spin text-info" />
                <span className="ml-2">Loading policies...</span>
              </div>
            ) : linkedPoliciesData?.control?.policies?.edges &&
              linkedPoliciesData.control.policies.edges.length > 0 ? (
              <div className="overflow-x-auto border rounded-md">
                <table className="w-full">
                  <thead>
                    <tr className="border-b text-left text-sm text-secondary bg-invert-bg">
                      <th className="py-3 px-4 font-medium">Name</th>
                      <th className="py-3 px-4 font-medium">Review Date</th>
                      <th className="py-3 px-4 font-medium text-right">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {getLinkedPolicies().map((policy) => (
                      <tr
                        key={policy.id}
                        className="border-b hover:bg-invert-bg"
                      >
                        <td className="py-3 px-4">
                          <div className="font-medium">{policy.title}</div>
                          {policy.description && (
                            <div className="text-xs text-secondary line-clamp-1 mt-0.5">
                              {policy.description}
                            </div>
                          )}
                        </td>
                        <td className="py-3 px-4">
                          {policy.updatedAt
                            ? new Date(policy.updatedAt).toLocaleDateString()
                            : "Not set"}
                        </td>
                        <td className="py-3 px-4 text-right whitespace-nowrap">
                          <div className="flex gap-2 justify-end">
                            <Button
                              variant="outline"
                              size="sm"
                              asChild
                              className="text-xs h-7"
                            >
                              <Link
                                to={`/organizations/${organizationId}/policies/${policy.id}`}
                              >
                                View
                              </Link>
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleUnlinkPolicy(policy.id)}
                              disabled={isUnlinkingPolicy}
                              className="text-xs h-7 text-danger border-danger-b hover:bg-h-danger-bg"
                            >
                              Unlink
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-8 text-secondary border rounded-md">
                No policies linked to this control yet. Click &quot;Link
                Policies&quot; to connect some.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
