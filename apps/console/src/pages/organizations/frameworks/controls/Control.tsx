import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { useToast } from "@/hooks/use-toast";
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
import { useParams, Link, useNavigate } from "react-router";
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
  ControlLinkedDocumentsQuery$data,
  ControlLinkedDocumentsQuery,
} from "./__generated__/ControlLinkedDocumentsQuery.graphql";
import {
  ControlOrganizationDocumentsQuery$data,
  ControlOrganizationDocumentsQuery,
} from "./__generated__/ControlOrganizationDocumentsQuery.graphql";

const controlFragment = graphql`
  fragment ControlFragment_Control on Control {
    id
    description
    name
    sectionTitle
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

// Query to fetch linked documents
const linkedDocumentsQuery = graphql`
  query ControlLinkedDocumentsQuery($controlId: ID!) {
    control: node(id: $controlId) {
      id
      ... on Control {
        documents(first: 100) @connection(key: "Control__documents") {
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

// Query to fetch all documents for the organization
const organizationDocumentsQuery = graphql`
  query ControlOrganizationDocumentsQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        documents(first: 100) @connection(key: "Organization__documents") {
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
      controlEdge {
        node {
          id
        }
      }
    }
  }
`;

// Mutation to delete a mapping between a control and a measure
const deleteMeasureMappingMutation = graphql`
  mutation ControlDeleteMeasureMappingMutation(
    $input: DeleteControlMeasureMappingInput!
  ) {
    deleteControlMeasureMapping(input: $input) {
      deletedMeasureId
    }
  }
`;

// Mutation to create a mapping between a control and a document
const createDocumentMappingMutation = graphql`
  mutation ControlCreateDocumentMappingMutation(
    $input: CreateControlDocumentMappingInput!
  ) {
    createControlDocumentMapping(input: $input) {
      controlEdge {
        node {
          id
        }
      }
    }
  }
`;

// Mutation to delete a mapping between a control and a document
const deleteDocumentMappingMutation = graphql`
  mutation ControlDeleteDocumentMappingMutation(
    $input: DeleteControlDocumentMappingInput!
  ) {
    deleteControlDocumentMapping(input: $input) {
      deletedDocumentId
    }
  }
`;

const deleteControlMutation = graphql`
  mutation ControlDeleteMutation($input: DeleteControlInput!, $connections: [ID!]!) {
    deleteControl(input: $input) {
      deletedControlId @deleteEdge(connections: $connections)
    }
  }
`;

export function Control({
  controlKey,
}: {
  controlKey: ControlFragment_Control$key;
}) {
  const { organizationId, frameworkId } = useParams<{
    organizationId: string;
    frameworkId: string;
  }>();
  const control = useFragment(controlFragment, controlKey);
  const { toast } = useToast();
  const environment = useRelayEnvironment();
  const navigate = useNavigate();

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

  // Document state
  const [isDocumentMappingDialogOpen, setIsDocumentMappingDialogOpen] =
    useState(false);
  const [linkedDocumentsData, setLinkedDocumentsData] =
    useState<ControlLinkedDocumentsQuery$data | null>(null);
  const [organizationDocumentsData, setOrganizationDocumentsData] =
    useState<ControlOrganizationDocumentsQuery$data | null>(null);
  const [documentSearchQuery, setDocumentSearchQuery] = useState("");
  const [isLoadingDocuments, setIsLoadingDocuments] = useState(false);
  const [isLinkingDocument, setIsLinkingDocument] = useState(false);
  const [isUnlinkingDocument, setIsUnlinkingDocument] = useState(false);

  // Create mutation hooks
  const [commitCreateMeasureMapping] = useMutation(
    createMeasureMappingMutation
  );
  const [commitDeleteMeasureMapping] = useMutation(
    deleteMeasureMappingMutation
  );
  const [commitCreateDocumentMapping] = useMutation(createDocumentMappingMutation);
  const [commitDeleteDocumentMapping] = useMutation(deleteDocumentMappingMutation);
  const [commitDeleteControl] = useMutation(deleteControlMutation);

  // State for delete dialog
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

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
        fetchQuery<ControlLinkedMeasuresQuery>(
          environment,
          linkedMeasuresQuery,
          {
            controlId: control.id,
          }
        ).subscribe({
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

  // Load initial linked documents data
  useEffect(() => {
    if (control.id) {
      setIsLoadingDocuments(true);
      fetchQuery<ControlLinkedDocumentsQuery>(environment, linkedDocumentsQuery, {
        controlId: control.id,
      }).subscribe({
        next: (data) => {
          setLinkedDocumentsData(data);
          setIsLoadingDocuments(false);
        },
        error: (error: Error) => {
          console.error("Error loading initial documents:", error);
          setIsLoadingDocuments(false);
        },
      });
    }
  }, [control.id, environment]);

  // Load documents data
  const loadDocumentsData = useCallback(() => {
    if (!organizationId || !control.id) return;

    setIsLoadingDocuments(true);

    // Fetch all documents for the organization
    fetchQuery<ControlOrganizationDocumentsQuery>(
      environment,
      organizationDocumentsQuery,
      {
        organizationId,
      }
    ).subscribe({
      next: (data) => {
        setOrganizationDocumentsData(data);
      },
      complete: () => {
        // Fetch linked documents for this control
        fetchQuery<ControlLinkedDocumentsQuery>(
          environment,
          linkedDocumentsQuery,
          {
            controlId: control.id,
          }
        ).subscribe({
          next: (data) => {
            setLinkedDocumentsData(data);
            setIsLoadingDocuments(false);
          },
          error: (error: Error) => {
            console.error("Error fetching linked documents:", error);
            setIsLoadingDocuments(false);
            toast({
              title: "Error",
              description: "Failed to load linked documents.",
              variant: "destructive",
            });
          },
        });
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
  }, [control.id, environment, organizationId, toast]);

  // Document helper functions
  const getDocuments = useCallback(() => {
    if (!organizationDocumentsData?.organization?.documents?.edges) return [];
    return organizationDocumentsData.organization.documents.edges.map(
      (edge) => edge.node
    );
  }, [organizationDocumentsData]);

  const getLinkedDocuments = useCallback(() => {
    if (!linkedDocumentsData?.control?.documents?.edges) return [];
    return linkedDocumentsData.control.documents.edges.map((edge) => edge.node);
  }, [linkedDocumentsData]);

  const isDocumentLinked = useCallback(
    (documentId: string) => {
      const linkedDocuments = getLinkedDocuments();
      return linkedDocuments.some((document) => document.id === documentId);
    },
    [getLinkedDocuments]
  );

  const filteredDocuments = useCallback(() => {
    const documents = getDocuments();
    if (!documentSearchQuery) return documents;

    return documents.filter((document) => {
      return (
        !documentSearchQuery ||
        document.title?.toLowerCase().includes(documentSearchQuery.toLowerCase()) ||
        (document.description &&
          document.description
            .toLowerCase()
            .includes(documentSearchQuery.toLowerCase()))
      );
    });
  }, [getDocuments, documentSearchQuery]);

  // Document link/unlink handlers
  const handleLinkDocument = useCallback(
    (documentId: string) => {
      if (!control.id) return;

      setIsLinkingDocument(true);

      commitCreateDocumentMapping({
        variables: {
          input: {
            controlId: control.id,
            documentId: documentId,
          },
        },
        onCompleted: (_, errors) => {
          setIsLinkingDocument(false);

          if (errors) {
            console.error("Error linking document:", errors);
            toast({
              title: "Error",
              description: "Failed to link document. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh linked documents data
          fetchQuery<ControlLinkedDocumentsQuery>(
            environment,
            linkedDocumentsQuery,
            {
              controlId: control.id,
            }
          ).subscribe({
            next: (data) => {
              setLinkedDocumentsData(data);
            },
            error: (error: Error) => {
              console.error("Error refreshing linked documents:", error);
            },
          });

          toast({
            title: "Success",
            description: "Document successfully linked to control.",
          });
        },
        onError: (error) => {
          setIsLinkingDocument(false);
          console.error("Error linking document:", error);
          toast({
            title: "Error",
            description: "Failed to link document. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [commitCreateDocumentMapping, control.id, environment, toast]
  );

  const handleUnlinkDocument = useCallback(
    (documentId: string) => {
      if (!control.id) return;

      setIsUnlinkingDocument(true);

      commitDeleteDocumentMapping({
        variables: {
          input: {
            controlId: control.id,
            documentId: documentId,
          },
        },
        onCompleted: (_, errors) => {
          setIsUnlinkingDocument(false);

          if (errors) {
            console.error("Error unlinking document:", errors);
            toast({
              title: "Error",
              description: "Failed to unlink document. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh linked documents data
          fetchQuery<ControlLinkedDocumentsQuery>(
            environment,
            linkedDocumentsQuery,
            {
              controlId: control.id,
            }
          ).subscribe({
            next: (data) => {
              setLinkedDocumentsData(data);
            },
            error: (error: Error) => {
              console.error("Error refreshing linked documents:", error);
            },
          });

          toast({
            title: "Success",
            description: "Document successfully unlinked from control.",
          });
        },
        onError: (error) => {
          setIsUnlinkingDocument(false);
          console.error("Error unlinking document:", error);
          toast({
            title: "Error",
            description: "Failed to unlink document. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [commitDeleteDocumentMapping, control.id, environment, toast]
  );

  const handleOpenDocumentMappingDialog = useCallback(() => {
    loadDocumentsData();
    setIsDocumentMappingDialogOpen(true);
  }, [loadDocumentsData]);

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
    <>
      <div className="w-auto p-5 flex items-start gap-5">
        <div className="font-mono text-lg px-1 py-0.25 rounded-sm bg-active-bg border-mid-b border font-bold">
          {control.sectionTitle}
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

          {/* Documents Section */}
          <div className="mt-8">
            {/* Document Mapping Dialog */}
            <Dialog
              open={isDocumentMappingDialogOpen}
              onOpenChange={setIsDocumentMappingDialogOpen}
            >
              <DialogContent className="max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
                <DialogHeader>
                  <DialogTitle>Link Documents to Control</DialogTitle>
                  <DialogDescription>
                    Search and select documents to link to this control. This helps
                    track which documents address this control.
                  </DialogDescription>
                </DialogHeader>

                <div className="flex items-center space-x-4 mb-4">
                  <div className="flex-1">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-tertiary" />
                      <Input
                        placeholder="Search documents by name or content..."
                        value={documentSearchQuery}
                        onChange={(e) => setDocumentSearchQuery(e.target.value)}
                        className="w-full pl-10"
                      />
                    </div>
                  </div>
                </div>

                <div className="flex-1 overflow-hidden">
                  {isLoadingDocuments ? (
                    <div className="flex items-center justify-center h-full">
                      <Loader2 className="w-8 h-8 animate-spin text-info" />
                      <span className="ml-2">Loading documents...</span>
                    </div>
                  ) : (
                    <div className="max-h-[50vh] overflow-y-auto pr-2">
                      {filteredDocuments().length === 0 ? (
                        <div className="text-center py-8 text-secondary">
                          No documents found. Try adjusting your search.
                        </div>
                      ) : (
                        <table className="w-full bg-level-1">
                          <thead className="sticky top-0 bg-white">
                            <tr className="border-b text-left text-sm text-secondary bg-invert-bg">
                              <th className="py-3 px-4 font-medium">Name</th>
                              <th className="py-3 px-4 font-medium">
                                Review Date
                              </th>
                              <th className="py-3 px-4 font-medium text-right">
                                Actions
                              </th>
                            </tr>
                          </thead>
                          <tbody>
                            {filteredDocuments().map((document) => {
                              const isLinked = isDocumentLinked(document.id);
                              return (
                                <tr
                                  key={document.id}
                                  className="border-b hover:bg-invert-bg"
                                >
                                  <td className="py-3 px-4">
                                    <div className="font-medium">
                                      {document.title}
                                    </div>
                                    {document.description && (
                                      <div className="text-xs text-secondary line-clamp-1 mt-0.5">
                                        {document.description}
                                      </div>
                                    )}
                                  </td>
                                  <td className="py-3 px-4">
                                    {document.updatedAt
                                      ? new Date(
                                          document.updatedAt
                                        ).toLocaleDateString()
                                      : "Not set"}
                                  </td>
                                  <td className="py-3 px-4 text-right whitespace-nowrap">
                                    {isLinked ? (
                                      <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() =>
                                          handleUnlinkDocument(document.id)
                                        }
                                        disabled={isUnlinkingDocument}
                                        className="text-xs h-7 text-danger border-danger-b hover:bg-h-danger-bg"
                                      >
                                        {isUnlinkingDocument ? (
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
                                          handleLinkDocument(document.id)
                                        }
                                        disabled={isLinkingDocument}
                                        className="text-xs h-7  text-info border-info-b hover:bg-h-info-bg"
                                      >
                                        {isLinkingDocument ? (
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
                  <Button onClick={() => setIsDocumentMappingDialogOpen(false)}>
                    Close
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>

            {/* Linked Documents List */}
            <div>
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-xl font-medium text-secondary">Documents</h3>
                <Button
                  variant="outline"
                  size="sm"
                  className="flex items-center gap-1"
                  onClick={handleOpenDocumentMappingDialog}
                >
                  <LinkIcon className="w-4 h-4" />
                  <span>Link Documents</span>
                </Button>
              </div>

              {isLoadingDocuments ? (
                <div className="flex items-center justify-center h-24">
                  <Loader2 className="w-6 h-6 animate-spin text-info" />
                  <span className="ml-2">Loading documents...</span>
                </div>
              ) : linkedDocumentsData?.control?.documents?.edges &&
                linkedDocumentsData.control.documents.edges.length > 0 ? (
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
                      {getLinkedDocuments().map((document) => (
                        <tr
                          key={document.id}
                          className="border-b hover:bg-invert-bg"
                        >
                          <td className="py-3 px-4">
                            <div className="font-medium">{document.title}</div>
                            {document.description && (
                              <div className="text-xs text-secondary line-clamp-1 mt-0.5">
                                {document.description}
                              </div>
                            )}
                          </td>
                          <td className="py-3 px-4">
                            {document.updatedAt
                              ? new Date(document.updatedAt).toLocaleDateString()
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
                                  to={`/organizations/${organizationId}/documents/${document.id}`}
                                >
                                  View
                                </Link>
                              </Button>
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleUnlinkDocument(document.id)}
                                disabled={isUnlinkingDocument}
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
                  No documents linked to this control yet. Click &quot;Link
                  Documents&quot; to connect some.
                </div>
              )}
            </div>
          </div>

          {/* Delete Control Section */}
          <div className="mt-12 border-t pt-8">
            <div className="flex justify-end items-center gap-2">
              <Button
                variant="outline"
                onClick={() => navigate(`/organizations/${organizationId}/frameworks/${frameworkId}/controls/${control.id}/edit`)}
              >
                Edit Control
              </Button>
              <Button
                variant="destructive"
                onClick={() => setIsDeleteDialogOpen(true)}
              >
                Delete Control
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Control</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete the control &quot;{control.name}&quot;? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteDialogOpen(false)}
              disabled={isDeleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => {
                setIsDeleting(true);
                commitDeleteControl({
                  variables: {
                    input: {
                      controlId: control.id,
                    },
                    connections: [
                      `client:${frameworkId}:__Framework_controls_connection`,
                      `client:${frameworkId}:FrameworkLayoutView_firstControl`
                    ],
                  },
                  onCompleted: (_, errors) => {
                    setIsDeleting(false);
                    setIsDeleteDialogOpen(false);

                    if (errors) {
                      console.error("Error deleting control:", errors);
                      toast({
                        title: "Error",
                        description: "Failed to delete control. Please try again.",
                        variant: "destructive",
                      });
                      return;
                    }

                    toast({
                      title: "Success",
                      description: "Control successfully deleted.",
                    });

                    // Navigate and force reload
                    navigate(`/organizations/${organizationId}/frameworks/${frameworkId}`, { replace: true });
                    window.location.reload();
                  },
                  onError: (error) => {
                    setIsDeleting(false);
                    setIsDeleteDialogOpen(false);
                    console.error("Error deleting control:", error);
                    toast({
                      title: "Error",
                      description: "Failed to delete control. Please try again.",
                      variant: "destructive",
                    });
                  },
                });
              }}
              disabled={isDeleting}
            >
              {isDeleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
