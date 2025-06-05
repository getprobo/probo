import { Suspense, useCallback, useEffect, useState } from "react";
import { useParams, Link } from "react-router";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
  fetchQuery,
  useRelayEnvironment,
} from "react-relay";
import type { ShowRiskViewQuery } from "./__generated__/ShowRiskViewQuery.graphql";
import { PageTemplate } from "@/components/PageTemplate";
import { ShowRiskViewSkeleton } from "./ShowRiskPage";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import {
  Plus,
  Trash2,
  Search,
  Tag,
  Edit,
  ShieldCheck,
  Shield,
  Handshake,
  Ban,
  User,
} from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  ShowRiskViewOrganizationMeasuresQuery,
  ShowRiskViewOrganizationMeasuresQuery$data,
} from "./__generated__/ShowRiskViewOrganizationMeasuresQuery.graphql";
import { ShowRiskViewCreateRiskMeasureMappingMutation } from "./__generated__/ShowRiskViewCreateRiskMeasureMappingMutation.graphql";
import { ShowRiskViewDeleteRiskMeasureMappingMutation } from "./__generated__/ShowRiskViewDeleteRiskMeasureMappingMutation.graphql";
import {
  ShowRiskViewOrganizationDocumentsQuery,
  ShowRiskViewOrganizationDocumentsQuery$data,
} from "./__generated__/ShowRiskViewOrganizationDocumentsQuery.graphql";
import { ShowRiskViewCreateRiskDocumentMappingMutation } from "./__generated__/ShowRiskViewCreateRiskDocumentMappingMutation.graphql";
import { ShowRiskViewDeleteRiskDocumentMappingMutation } from "./__generated__/ShowRiskViewDeleteRiskDocumentMappingMutation.graphql";

const showRiskViewQuery = graphql`
  query ShowRiskViewQuery($riskId: ID!) {
    node(id: $riskId) {
      id
      ... on Risk {
        name
        description
        treatment
        owner {
          id
          fullName
        }
        inherentLikelihood
        inherentImpact
        residualLikelihood
        residualImpact
        note
        createdAt
        updatedAt
        measures(first: 100) @connection(key: "Risk__measures") {
          edges {
            node {
              id
              name
              description
              category
              createdAt
              state
            }
          }
        }
        documents(first: 100) @connection(key: "Risk__documents") {
          edges {
            node {
              id
              title
              createdAt
            }
          }
        }
        controls(first: 100) @connection(key: "Risk__controls") {
          edges {
            node {
              id
              sectionTitle
              name
              description
              createdAt
            }
          }
        }
      }
    }
  }
`;

// Add query to fetch all measures for the organization
const organizationMeasuresQuery = graphql`
  query ShowRiskViewOrganizationMeasuresQuery($organizationId: ID!) {
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

// Add query to fetch all documents for the organization
const organizationDocumentsQuery = graphql`
  query ShowRiskViewOrganizationDocumentsQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        documents(first: 100) @connection(key: "Organization__documents") {
          edges {
            node {
              id
              title
            }
          }
        }
      }
    }
  }
`;

// Add mutation to create risk-measure mapping
const createRiskMeasureMappingMutation = graphql`
  mutation ShowRiskViewCreateRiskMeasureMappingMutation(
    $input: CreateRiskMeasureMappingInput!
  ) {
    createRiskMeasureMapping(input: $input) {
      riskEdge {
        node {
          id
        }
      }
    }
  }
`;

// Add mutation to delete risk-measure mapping
const deleteRiskMeasureMappingMutation = graphql`
  mutation ShowRiskViewDeleteRiskMeasureMappingMutation(
    $input: DeleteRiskMeasureMappingInput!
  ) {
    deleteRiskMeasureMapping(input: $input) {
      deletedMeasureId
    }
  }
`;

// Add mutation to create risk-document mapping
const createRiskDocumentMappingMutation = graphql`
  mutation ShowRiskViewCreateRiskDocumentMappingMutation(
    $input: CreateRiskDocumentMappingInput!
  ) {
    createRiskDocumentMapping(input: $input) {
      riskEdge {
        node {
          id
        }
      }
    }
  }
`;

// Add mutation to delete risk-document mapping
const deleteRiskDocumentMappingMutation = graphql`
  mutation ShowRiskViewDeleteRiskDocumentMappingMutation(
    $input: DeleteRiskDocumentMappingInput!
  ) {
    deleteRiskDocumentMapping(input: $input) {
      deletedDocumentId
    }
  }
`;

// Replace the current getRiskSeverity function with the functions from ListRiskView.tsx
// Helper function to convert risk score to risk level
const calculateRiskLevel = (score: number): string => {
  if (score >= 15) return "High";
  if (score >= 8) return "Medium";
  return "Low";
};

// Helper function to get color for risk level that matches the matrix colors
const getRiskLevelColor = (
  level: string
): { backgroundColor: string; color: string } => {
  switch (level) {
    case "High":
      return { backgroundColor: "#ef4444", color: "#ffffff" }; // red-500
    case "Medium":
      return { backgroundColor: "#fcd34d", color: "#000000" }; // yellow-300
    default:
      return { backgroundColor: "#22c55e", color: "#ffffff" }; // green-500
  }
};

function getRiskSeverity(likelihood: number, impact: number) {
  const score = likelihood * impact;
  const level = calculateRiskLevel(score);
  return {
    level: level,
    score: score,
    style: getRiskLevelColor(level),
  };
}

// Keep the getLikelihoodLabel and getImpactLabel functions as they are
function getLikelihoodLabel(likelihood: number): {
  label: string;
  style: React.CSSProperties;
} {
  if (likelihood === 5)
    return {
      label: "Frequent (5)",
      style: {
        backgroundColor: "#fff",
        color: "black",
        border: "1px solid #e2e8f0",
      },
    };
  if (likelihood === 4)
    return {
      label: "Probable (4)",
      style: {
        backgroundColor: "#fff",
        color: "black",
        border: "1px solid #e2e8f0",
      },
    };
  if (likelihood === 3)
    return {
      label: "Occasional (3)",
      style: {
        backgroundColor: "#fff",
        color: "black",
        border: "1px solid #e2e8f0",
      },
    };
  if (likelihood === 2)
    return {
      label: "Remote (2)",
      style: {
        backgroundColor: "#fff",
        color: "black",
        border: "1px solid #e2e8f0",
      },
    };
  return {
    label: "Improbable (1)",
    style: {
      backgroundColor: "#fff",
      color: "black",
      border: "1px solid #e2e8f0",
    },
  };
}

function getImpactLabel(impact: number): {
  label: string;
  style: React.CSSProperties;
} {
  if (impact === 5)
    return {
      label: "Catastrophic (5)",
      style: {
        backgroundColor: "#fff",
        color: "black",
        border: "1px solid #e2e8f0",
      },
    };
  if (impact === 4)
    return {
      label: "Significant (4)",
      style: {
        backgroundColor: "#fff",
        color: "black",
        border: "1px solid #e2e8f0",
      },
    };
  if (impact === 3)
    return {
      label: "Moderate (3)",
      style: {
        backgroundColor: "#fff",
        color: "black",
        border: "1px solid #e2e8f0",
      },
    };
  if (impact === 2)
    return {
      label: "Low (2)",
      style: {
        backgroundColor: "#fff",
        color: "black",
        border: "1px solid #e2e8f0",
      },
    };
  return {
    label: "Negligible (1)",
    style: {
      backgroundColor: "#fff",
      color: "black",
      border: "1px solid #e2e8f0",
    },
  };
}

function ShowRiskViewContent({
  queryRef,
  loadQuery,
}: {
  queryRef: PreloadedQuery<ShowRiskViewQuery>;
  loadQuery: (variables: { riskId: string }) => void;
}) {
  const data = usePreloadedQuery<ShowRiskViewQuery>(
    showRiskViewQuery,
    queryRef
  );
  const { toast } = useToast();
  const environment = useRelayEnvironment();
  const { organizationId } = useParams<{ organizationId: string }>();

  // Cast the node to Risk type without using any
  const risk = data.node;
  const severity = getRiskSeverity(
    risk.inherentLikelihood!,
    risk.inherentImpact!
  );

  // Fix typing for measures
  const measures = risk.measures?.edges?.map((edge) => edge.node) || [];
  const residualSeverity = getRiskSeverity(
    risk.residualLikelihood!,
    risk.residualImpact!
  );

  // Fix typing for documents
  const documents = risk.documents?.edges?.map((edge) => edge.node) || [];

  // Fix typing for controls
  const controls = risk.controls?.edges?.map((edge) => edge.node) || [];

  // Add state for measure mapping dialog
  const [isMeasureDialogOpen, setIsMeasureDialogOpen] = useState(false);
  const [organizationMeasuresData, setOrganizationMeasuresData] =
    useState<ShowRiskViewOrganizationMeasuresQuery$data | null>(null);
  const [measureSearchQuery, setMeasureSearchQuery] = useState("");
  const [categoryFilter, setCategoryFilter] = useState<string | null>(null);
  const [isLoadingMeasures, setIsLoadingMeasures] = useState(false);
  const [linkingMeasures, setLinkingMeasures] = useState<
    Record<string, boolean>
  >({});
  const [unlinkingMeasures, setUnlinkingMeasures] = useState<
    Record<string, boolean>
  >({});

  // Add state for document mapping dialog
  const [isDocumentDialogOpen, setIsDocumentDialogOpen] = useState(false);
  const [organizationDocumentsData, setOrganizationDocumentsData] =
    useState<ShowRiskViewOrganizationDocumentsQuery$data | null>(null);
  const [documentSearchQuery, setDocumentSearchQuery] = useState("");
  const [isLoadingDocuments, setIsLoadingDocuments] = useState(false);
  const [linkingDocuments, setLinkingDocuments] = useState<
    Record<string, boolean>
  >({});
  const [unlinkingDocuments, setUnlinkingDocuments] = useState<
    Record<string, boolean>
  >({});

  // Setup mutation hooks
  const [createRiskMeasureMapping] =
    useMutation<ShowRiskViewCreateRiskMeasureMappingMutation>(
      createRiskMeasureMappingMutation
    );
  const [deleteRiskMeasureMapping] =
    useMutation<ShowRiskViewDeleteRiskMeasureMappingMutation>(
      deleteRiskMeasureMappingMutation
    );

  // Setup document mutation hooks
  const [createRiskDocumentMapping] =
    useMutation<ShowRiskViewCreateRiskDocumentMappingMutation>(
      createRiskDocumentMappingMutation
    );
  const [deleteRiskDocumentMapping] =
    useMutation<ShowRiskViewDeleteRiskDocumentMappingMutation>(
      deleteRiskDocumentMappingMutation
    );

  // Clear filters when dialog closes
  useEffect(() => {
    if (!isMeasureDialogOpen) {
      setMeasureSearchQuery("");
      setCategoryFilter(null);
      setLinkingMeasures({});
      setUnlinkingMeasures({});
    }
  }, [isMeasureDialogOpen]);

  // Clear filters when document dialog closes
  useEffect(() => {
    if (!isDocumentDialogOpen) {
      setDocumentSearchQuery("");
      setLinkingDocuments({});
      setUnlinkingDocuments({});
    }
  }, [isDocumentDialogOpen]);

  // Load measures data when needed
  const loadMeasuresData = useCallback(() => {
    if (!organizationId || !risk.id) return;

    setIsLoadingMeasures(true);

    // Fetch all measures for the organization
    fetchQuery<ShowRiskViewOrganizationMeasuresQuery>(
      environment,
      organizationMeasuresQuery,
      {
        organizationId,
      }
    ).subscribe({
      next: (data) => {
        setOrganizationMeasuresData(data);
        setIsLoadingMeasures(false);
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
  }, [risk.id, environment, organizationId, toast]);

  // Load documents data when needed
  const loadDocumentsData = useCallback(() => {
    if (!organizationId || !risk.id) return;

    setIsLoadingDocuments(true);

    // Fetch all documents for the organization
    fetchQuery<ShowRiskViewOrganizationDocumentsQuery>(
      environment,
      organizationDocumentsQuery,
      {
        organizationId,
      }
    ).subscribe({
      next: (data) => {
        setOrganizationDocumentsData(data);
        setIsLoadingDocuments(false);
      },
      error: (error: Error) => {
        console.error("Error fetching organization documents:", error);
        setIsLoadingDocuments(false);
        toast({
          title: "Error",
          description: "Failed to load documents.",
          variant: "destructive",
        });
      },
    });
  }, [risk.id, environment, organizationId, toast]);

  // Helper functions
  const getMeasures = useCallback(() => {
    if (!organizationMeasuresData?.organization?.measures?.edges) return [];
    return organizationMeasuresData.organization.measures.edges.map(
      (edge) => edge.node
    );
  }, [organizationMeasuresData]);

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
  }, [getMeasures, measureSearchQuery, categoryFilter]);

  // Helper functions for documents
  const getDocuments = useCallback(() => {
    if (!organizationDocumentsData?.organization?.documents?.edges) return [];
    return organizationDocumentsData.organization.documents.edges.map(
      (edge) => edge.node
    );
  }, [organizationDocumentsData]);

  const filteredDocuments = useCallback(() => {
    const documents = getDocuments();
    if (!documentSearchQuery) return documents;

    return documents.filter((document) => {
      return (
        !documentSearchQuery ||
        document.title.toLowerCase().includes(documentSearchQuery.toLowerCase())
      );
    });
  }, [getDocuments, documentSearchQuery]);

  // Handle linking a measure to this risk
  const handleLinkMeasure = useCallback(
    (
      measure: NonNullable<
        NonNullable<
          ShowRiskViewOrganizationMeasuresQuery$data["organization"]
        >["measures"]
      >["edges"][0]["node"]
    ) => {
      if (!risk.id) return;

      // Track this specific measure as linking
      setLinkingMeasures((prev) => ({ ...prev, [measure.id]: true }));

      createRiskMeasureMapping({
        variables: {
          input: {
            riskId: risk.id,
            measureId: measure.id,
          },
        },
        onCompleted: (_, errors) => {
          setLinkingMeasures((prev) => ({
            ...prev,
            [measure.id]: false,
          }));

          if (errors && errors.length > 0) {
            console.error("Error linking measure:", errors);
            toast({
              title: "Error",
              description: "Failed to link measure. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh main query data using the exact same pattern as in ControlView.tsx
          fetchQuery(environment, showRiskViewQuery, {
            riskId: risk.id,
          }).subscribe({
            next: () => {
              // Force reload the view data to reflect changes
              loadQuery({ riskId: risk.id });
            },
            error: (error: Error) => {
              console.error("Error refreshing risk data:", error);
            },
          });

          toast({
            title: "Success",
            description: `Linked measure "${measure.name}" to this risk.`,
          });
        },
        onError: (error) => {
          setLinkingMeasures((prev) => ({
            ...prev,
            [measure.id]: false,
          }));
          console.error("Error linking measure:", error);
          toast({
            title: "Error",
            description: "Failed to link measure. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [risk.id, createRiskMeasureMapping, toast, environment, loadQuery]
  );

  // Handle unlinking a measure from this risk
  const handleUnlinkMeasure = useCallback(
    (
      measure: NonNullable<
        NonNullable<
          ShowRiskViewOrganizationMeasuresQuery$data["organization"]
        >["measures"]
      >["edges"][0]["node"]
    ) => {
      if (!risk.id) return;

      // Track this specific measure as unlinking
      setUnlinkingMeasures((prev) => ({ ...prev, [measure.id]: true }));

      deleteRiskMeasureMapping({
        variables: {
          input: {
            riskId: risk.id,
            measureId: measure.id,
          },
        },
        onCompleted: (_, errors) => {
          setUnlinkingMeasures((prev) => ({
            ...prev,
            [measure.id]: false,
          }));

          if (errors && errors.length > 0) {
            console.error("Error unlinking measure:", errors);
            toast({
              title: "Error",
              description: "Failed to unlink measure. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh main query data using the exact same pattern as in ControlView.tsx
          fetchQuery(environment, showRiskViewQuery, {
            riskId: risk.id,
          }).subscribe({
            next: () => {
              // Force reload the view data to reflect changes
              loadQuery({ riskId: risk.id });
            },
            error: (error: Error) => {
              console.error("Error refreshing risk data:", error);
            },
          });

          toast({
            title: "Success",
            description: `Unlinked measure "${measure.name}" from this risk.`,
          });
        },
        onError: (error) => {
          setUnlinkingMeasures((prev) => ({
            ...prev,
            [measure.id]: false,
          }));
          console.error("Error unlinking measure:", error);
          toast({
            title: "Error",
            description: "Failed to unlink measure. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [risk.id, deleteRiskMeasureMapping, toast, environment, loadQuery]
  );

  // Handle linking a document to this risk
  const handleLinkDocument = useCallback(
    (
      document: NonNullable<
        NonNullable<
          ShowRiskViewOrganizationDocumentsQuery$data["organization"]
        >["documents"]
      >["edges"][0]["node"]
    ) => {
      if (!risk.id) return;

      // Track this specific document as linking
      setLinkingDocuments((prev) => ({ ...prev, [document.id]: true }));

      createRiskDocumentMapping({
        variables: {
          input: {
            riskId: risk.id,
            documentId: document.id,
          },
        },
        onCompleted: (_, errors) => {
          setLinkingDocuments((prev) => ({
            ...prev,
            [document.id]: false,
          }));

          if (errors && errors.length > 0) {
            console.error("Error linking document:", errors);
            toast({
              title: "Error",
              description: "Failed to link document. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh main query data using the exact same pattern as in ControlView.tsx
          fetchQuery(environment, showRiskViewQuery, {
            riskId: risk.id,
          }).subscribe({
            next: () => {
              // Force reload the view data to reflect changes
              loadQuery({ riskId: risk.id });
            },
            error: (error: Error) => {
              console.error("Error refreshing risk data:", error);
            },
          });

          toast({
            title: "Success",
            description: `Linked document "${document.title}" to this risk.`,
          });
        },
        onError: (error) => {
          setLinkingDocuments((prev) => ({
            ...prev,
            [document.id]: false,
          }));
          console.error("Error linking document:", error);
          toast({
            title: "Error",
            description: "Failed to link document. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [risk.id, createRiskDocumentMapping, toast, environment, loadQuery]
  );

  // Handle unlinking a document from this risk
  const handleUnlinkDocument = useCallback(
    (
      document: NonNullable<
        NonNullable<
          ShowRiskViewOrganizationDocumentsQuery$data["organization"]
        >["documents"]
      >["edges"][0]["node"]
    ) => {
      if (!risk.id) return;

      // Track this specific document as unlinking
      setUnlinkingDocuments((prev) => ({ ...prev, [document.id]: true }));

      deleteRiskDocumentMapping({
        variables: {
          input: {
            riskId: risk.id,
            documentId: document.id,
          },
        },
        onCompleted: (_, errors) => {
          setUnlinkingDocuments((prev) => ({
            ...prev,
            [document.id]: false,
          }));

          if (errors && errors.length > 0) {
            console.error("Error unlinking document:", errors);
            toast({
              title: "Error",
              description: "Failed to unlink document. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh main query data using the exact same pattern as in ControlView.tsx
          fetchQuery(environment, showRiskViewQuery, {
            riskId: risk.id,
          }).subscribe({
            next: () => {
              // Force reload the view data to reflect changes
              loadQuery({ riskId: risk.id });
            },
            error: (error: Error) => {
              console.error("Error refreshing risk data:", error);
            },
          });

          toast({
            title: "Success",
            description: `Unlinked document "${document.title}" from this risk.`,
          });
        },
        onError: (error) => {
          setUnlinkingDocuments((prev) => ({
            ...prev,
            [document.id]: false,
          }));
          console.error("Error unlinking document:", error);
          toast({
            title: "Error",
            description: "Failed to unlink document. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [risk.id, deleteRiskDocumentMapping, toast, environment, loadQuery]
  );

  return (
    <PageTemplate
      title={risk.name ?? ""}
      description={risk.description || ""}
      actions={
        <Button variant="outline" asChild>
          <Link to={`/organizations/${organizationId}/risks/${risk.id}/edit`}>
            <Edit className="mr-2 h-4 w-4" />
            Edit Risk
          </Link>
        </Button>
      }
    >
      <div className="space-y-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle>Risk Overview</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-8">
              {/* Risk Management Overview */}
              <div className="flex flex-col md:flex-row gap-5">
                <div className="flex items-center">
                  {risk.treatment === "MITIGATED" && (
                    <Shield className="h-6 w-6 mr-3 text-blue-500" />
                  )}
                  {risk.treatment === "TRANSFERRED" && (
                    <Handshake className="h-6 w-6 mr-3 text-purple-500" />
                  )}
                  {risk.treatment === "AVOIDED" && (
                    <Ban className="h-6 w-6 mr-3 text-red-500" />
                  )}
                  {risk.treatment === "ACCEPTED" && (
                    <ShieldCheck className="h-6 w-6 mr-3 text-green-500" />
                  )}
                  <div>
                    <p className="text-xs text-slate-500 uppercase">
                      Treatment
                    </p>
                    <p className="text-base font-medium capitalize mt-1">
                      {risk.treatment ? risk.treatment.toLowerCase() : "N/A"}
                    </p>
                  </div>
                </div>

                <div className="flex items-center">
                  <User className="h-6 w-6 mr-3 text-slate-500" />
                  <div>
                    <p className="text-xs text-slate-500 uppercase">
                      Risk Owner
                    </p>
                    <p className="text-base font-medium mt-1 truncate">
                      {risk.owner ? (
                        <Link
                          to={`/organizations/${organizationId}/people/${risk.owner.id}`}
                          className="text-blue-600 hover:underline"
                        >
                          {risk.owner.fullName}
                        </Link>
                      ) : (
                        "Unassigned"
                      )}
                    </p>
                  </div>
                </div>

                <div className="flex items-center">
                  <div>
                    <p className="text-xs text-slate-500 uppercase">
                      Last Updated
                    </p>
                    <p className="text-base font-medium mt-1">
                      {risk.updatedAt
                        ? new Date(risk.updatedAt).toLocaleDateString("en-US", {
                            year: "numeric",
                            month: "long",
                            day: "numeric",
                          })
                        : "N/A"}
                    </p>
                  </div>
                </div>
              </div>

              {/* Risk Assessment */}
              <div>
                <h3 className="text-sm font-medium mb-6">Risk Assessment</h3>

                <div className="flex flex-col space-y-8">
                  {/* Initial Assessment */}
                  <div>
                    <h4 className="text-sm font-medium mb-5 flex items-center">
                      <span className="w-3 h-3 rounded-full bg-orange-500 mr-2"></span>
                      Initial Assessment
                    </h4>

                    <div className="flex items-center">
                      <div
                        className="px-5 py-2 rounded-md font-medium text-sm flex justify-center items-center whitespace-nowrap"
                        style={
                          getLikelihoodLabel(risk.inherentLikelihood!).style
                        }
                      >
                        {getLikelihoodLabel(risk.inherentLikelihood!).label}
                      </div>

                      <span className="mx-4 text-xl">×</span>

                      <div
                        className="px-5 py-2 rounded-md font-medium text-sm flex justify-center items-center whitespace-nowrap"
                        style={getImpactLabel(risk.inherentImpact!).style}
                      >
                        {getImpactLabel(risk.inherentImpact!).label}
                      </div>

                      <span className="mx-4 text-xl">=</span>

                      <div
                        className="px-5 py-2 rounded-md font-medium text-sm flex justify-center items-center whitespace-nowrap"
                        style={severity.style}
                      >
                        {severity.level} (
                        {risk.inherentLikelihood! * risk.inherentImpact!})
                      </div>
                    </div>
                  </div>

                  {/* Residual Assessment */}
                  <div>
                    <h4 className="text-sm font-medium mb-5 flex items-center">
                      <span className="w-3 h-3 rounded-full bg-green-500 mr-2"></span>
                      After Treatment
                    </h4>

                    <div className="flex items-center">
                      <div
                        className="px-5 py-2 rounded-md font-medium text-sm flex justify-center items-center whitespace-nowrap"
                        style={
                          getLikelihoodLabel(risk.residualLikelihood!).style
                        }
                      >
                        {getLikelihoodLabel(risk.residualLikelihood!).label}
                      </div>

                      <span className="mx-4 text-xl">×</span>

                      <div
                        className="px-5 py-2 rounded-md font-medium text-sm flex justify-center items-center whitespace-nowrap"
                        style={getImpactLabel(risk.residualImpact!).style}
                      >
                        {getImpactLabel(risk.residualImpact!).label}
                      </div>

                      <span className="mx-4 text-xl">=</span>

                      <div
                        className="px-5 py-2 rounded-md font-medium text-sm flex justify-center items-center whitespace-nowrap"
                        style={residualSeverity.style}
                      >
                        {residualSeverity.level} (
                        {risk.residualLikelihood! * risk.residualImpact!})
                      </div>
                    </div>
                  </div>
                </div>

                {/* Risk Reduction Summary */}
                {(risk.inherentLikelihood !== risk.residualLikelihood ||
                  risk.inherentImpact !== risk.residualImpact) && (
                  <div className="mt-6 text-center pt-4 border-t">
                    <p className="text-sm text-slate-600">
                      Risk severity reduced from{" "}
                      <span className="font-semibold">
                        {severity.level} (
                        {risk.inherentLikelihood! * risk.inherentImpact!})
                      </span>{" "}
                      to{" "}
                      <span className="font-semibold">
                        {residualSeverity.level} (
                        {risk.residualLikelihood! * risk.residualImpact!})
                      </span>
                      {risk.treatment === "MITIGATED" &&
                        risk.measures?.edges &&
                        risk.measures.edges.length > 0 &&
                        ` with ${risk.measures.edges.length} ${
                          risk.measures.edges.length === 1
                            ? "measure"
                            : "measures"
                        }`}
                    </p>
                  </div>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        {risk.note && (
          <Card>
            <CardHeader className="pb-2">
              <CardTitle>Note</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="prose prose-sm max-w-none">{risk.note}</div>
            </CardContent>
          </Card>
        )}

        <Tabs defaultValue="measures" className="w-full">
          <TabsList>
            <TabsTrigger value="measures">Measures</TabsTrigger>
            <TabsTrigger value="documents">Documents</TabsTrigger>
            <TabsTrigger value="controls">Controls</TabsTrigger>
          </TabsList>

          <TabsContent value="measures" className="space-y-4">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">Risk Measures</h2>
              <Button
                onClick={() => {
                  setIsMeasureDialogOpen(true);
                  loadMeasuresData();
                }}
              >
                <Plus className="mr-2 h-4 w-4" />
                Link Measure
              </Button>
            </div>
            {risk.measures?.edges && risk.measures.edges.length > 0 ? (
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-full">Measure</TableHead>
                      <TableHead className="w-20">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {risk.measures?.edges.map(({ node: measure }) => (
                      <TableRow key={measure.id}>
                        <TableCell>
                          <Link
                            to={`/organizations/${organizationId}/measures/${measure.id}`}
                            className="font-medium text-blue-600 hover:underline"
                          >
                            {measure.name}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleUnlinkMeasure(measure)}
                            disabled={unlinkingMeasures[measure.id] || false}
                            title="Unlink measure"
                          >
                            <Trash2 className="h-4 w-4 text-danger" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            ) : (
              <div className="text-center py-10 text-secondary">
                <p>No measures associated with this risk.</p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="documents" className="space-y-4">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">Risk Documents</h2>
              <div className="flex space-x-2">
                <Button
                  onClick={() => {
                    setIsDocumentDialogOpen(true);
                    loadDocumentsData();
                  }}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Link Document
                </Button>
              </div>
            </div>
            {documents.length > 0 ? (
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-full">Document</TableHead>
                      <TableHead className="w-20">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {documents.map((document) => (
                      <TableRow key={document.id}>
                        <TableCell>
                          <Link
                            to={`/organizations/${organizationId}/documents/${document.id}`}
                            className="font-medium text-blue-600 hover:underline"
                          >
                            {document.title}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleUnlinkDocument(document)}
                            disabled={unlinkingDocuments[document.id] || false}
                            title="Unlink document"
                          >
                            <Trash2 className="h-4 w-4 text-danger" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            ) : (
              <div className="text-center py-10 text-secondary">
                <p>No documents associated with this risk.</p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="controls" className="space-y-4">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">Related Controls</h2>
            </div>
            {controls.length > 0 ? (
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-full">Control</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {controls.map((control) => (
                      <TableRow key={control.id}>
                        <TableCell>
                          <Link
                            to={`/organizations/${organizationId}/controls/${control.id}`}
                            className="font-medium text-blue-600 hover:underline"
                          >
                            {control.sectionTitle} - {control.name}
                          </Link>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            ) : (
              <div className="text-center py-10 text-secondary">
                <ShieldCheck className="mx-auto h-10 w-10 mb-2 opacity-30" />
                <p>No controls associated with this risk.</p>
              </div>
            )}
          </TabsContent>
        </Tabs>

        {/* Dialog for linking measures */}
        <Dialog
          open={isMeasureDialogOpen}
          onOpenChange={setIsMeasureDialogOpen}
        >
          <DialogContent className="max-w-3xl">
            <DialogHeader>
              <div className="flex justify-between items-center">
                <div>
                  <DialogTitle>Manage Risk Measures</DialogTitle>
                  <DialogDescription>
                    Link or unlink measures to manage this risk.
                  </DialogDescription>
                </div>
              </div>
            </DialogHeader>

            <div className="flex flex-col space-y-4">
              <div className="flex flex-row space-x-2">
                <div className="flex-1">
                  <div className="relative">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-tertiary" />
                    <Input
                      placeholder="Search measures..."
                      className="pl-8"
                      value={measureSearchQuery}
                      onChange={(e) => setMeasureSearchQuery(e.target.value)}
                    />
                  </div>
                </div>
                <Select
                  value={categoryFilter || "all"}
                  onValueChange={(value) =>
                    setCategoryFilter(value === "all" ? null : value)
                  }
                >
                  <SelectTrigger className="w-[180px]">
                    <SelectValue placeholder="Filter by category" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Categories</SelectItem>
                    {getMeasureCategories().map((category) => (
                      <SelectItem key={category} value={category}>
                        {category}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="border rounded-md max-h-96 overflow-y-auto">
                {isLoadingMeasures ? (
                  <div className="p-4 text-center">Loading measures...</div>
                ) : filteredMeasures().length === 0 ? (
                  <div className="p-4 text-center">No measures found.</div>
                ) : (
                  <div className="divide-y">
                    {filteredMeasures().map((measure) => {
                      // For each render, recalculate linked status directly against the current risk data
                      const isLinked = risk.measures?.edges?.some(
                        (edge) => edge.node.id === measure.id
                      );
                      const isLinking = linkingMeasures[measure.id] || false;
                      const isUnlinking =
                        unlinkingMeasures[measure.id] || false;

                      return (
                        <div
                          key={measure.id}
                          className="relative p-4 hover:bg-blue-50 transition-colors duration-150"
                        >
                          <div className="flex justify-between items-center">
                            <div className="flex items-center gap-2">
                              <h3 className="font-medium">{measure.name}</h3>
                              {measure.category && (
                                <Badge
                                  variant="outline"
                                  className="text-xs font-normal"
                                >
                                  <Tag className="h-3 w-3 mr-1" />
                                  {measure.category}
                                </Badge>
                              )}
                            </div>
                            {isLinked ? (
                              <Button
                                size="sm"
                                variant="outline"
                                disabled={isUnlinking}
                                onClick={() => handleUnlinkMeasure(measure)}
                                className="text-red-600 border-red-200 hover:bg-red-50"
                              >
                                {isUnlinking ? "Unlinking..." : "Unlink"}
                              </Button>
                            ) : (
                              <Button
                                size="sm"
                                disabled={isLinking}
                                onClick={() => handleLinkMeasure(measure)}
                              >
                                {isLinking ? "Linking..." : "Link"}
                              </Button>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsMeasureDialogOpen(false)}
              >
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Dialog for linking documents */}
        <Dialog open={isDocumentDialogOpen} onOpenChange={setIsDocumentDialogOpen}>
          <DialogContent className="max-w-3xl">
            <DialogHeader>
              <div className="flex justify-between items-center">
                <div>
                  <DialogTitle>Manage Risk Documents</DialogTitle>
                  <DialogDescription>
                    Link or unlink documents to manage this risk.
                  </DialogDescription>
                </div>
              </div>
            </DialogHeader>

            <div className="flex flex-col space-y-4">
              <div className="flex flex-row space-x-2">
                <div className="flex-1">
                  <div className="relative">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-tertiary" />
                    <Input
                      placeholder="Search documents..."
                      className="pl-8"
                      value={documentSearchQuery}
                      onChange={(e) => setDocumentSearchQuery(e.target.value)}
                    />
                  </div>
                </div>
              </div>

              <div className="border rounded-md max-h-96 overflow-y-auto">
                {isLoadingDocuments ? (
                  <div className="p-4 text-center">Loading documents...</div>
                ) : filteredDocuments().length === 0 ? (
                  <div className="p-4 text-center">No documents found.</div>
                ) : (
                  <div className="divide-y">
                    {filteredDocuments().map((document) => {
                      // For each render, recalculate linked status directly against the current risk data
                      const isLinked = documents.some(
                        (riskDocument) => riskDocument.id === document.id
                      );
                      const isLinking = linkingDocuments[document.id] || false;
                      const isUnlinking = unlinkingDocuments[document.id] || false;

                      return (
                        <div
                          key={document.id}
                          className="relative p-4 hover:bg-blue-50 transition-colors duration-150"
                        >
                          <div className="flex justify-between items-center">
                            <div className="flex items-center gap-2">
                              <h3 className="font-medium">{document.title}</h3>
                            </div>
                            {isLinked ? (
                              <Button
                                size="sm"
                                variant="outline"
                                disabled={isUnlinking}
                                onClick={() => handleUnlinkDocument(document)}
                                className="text-red-600 border-red-200 hover:bg-red-50"
                              >
                                {isUnlinking ? "Unlinking..." : "Unlink"}
                              </Button>
                            ) : (
                              <Button
                                size="sm"
                                disabled={isLinking}
                                onClick={() => handleLinkDocument(document)}
                              >
                                {isLinking ? "Linking..." : "Link"}
                              </Button>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsDocumentDialogOpen(false)}
              >
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </PageTemplate>
  );
}

export default function ShowRiskView() {
  const { riskId } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<ShowRiskViewQuery>(showRiskViewQuery);

  useEffect(() => {
    loadQuery({ riskId: riskId! });
  }, [loadQuery, riskId]);

  if (!queryRef) {
    return <ShowRiskViewSkeleton />;
  }

  return (
    <Suspense fallback={<ShowRiskViewSkeleton />}>
      <ShowRiskViewContent queryRef={queryRef} loadQuery={loadQuery} />
    </Suspense>
  );
}
