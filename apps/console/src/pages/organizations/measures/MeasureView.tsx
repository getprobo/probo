import {
  Suspense,
  useEffect,
  useState,
  useRef,
  DragEvent,
  useCallback,
} from "react";
import { useParams, useNavigate, useSearchParams } from "react-router";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
  fetchQuery,
  useRelayEnvironment,
  ConnectionHandler,
} from "react-relay";
import {
  AlertTriangle,
  Check,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  Download,
  Eye,
  File as FileGeneric,
  FileIcon,
  FileText,
  Image,
  Link2,
  Loader2,
  Plus,
  Search,
  ShieldCheck,
  Trash2,
  Upload,
  User,
  UserMinus,
  UserPlus,
  X,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import ReactMarkdown from "react-markdown";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetClose,
} from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";

import { PageTemplate } from "@/components/PageTemplate";
import { MeasureViewSkeleton } from "./MeasurePage";
import { MeasureViewUpdateTaskStateMutation as MeasureViewUpdateTaskStateMutationType } from "./__generated__/MeasureViewUpdateTaskStateMutation.graphql";
import { MeasureViewCreateTaskMutation as MeasureViewCreateTaskMutationType } from "./__generated__/MeasureViewCreateTaskMutation.graphql";
import { MeasureViewDeleteTaskMutation as MeasureViewDeleteTaskMutationType } from "./__generated__/MeasureViewDeleteTaskMutation.graphql";
import { MeasureViewDeleteEvidenceMutation as MeasureViewDeleteEvidenceMutationType } from "./__generated__/MeasureViewDeleteEvidenceMutation.graphql";
import { MeasureViewAssignTaskMutation as MeasureViewAssignTaskMutationType } from "./__generated__/MeasureViewAssignTaskMutation.graphql";
import { MeasureViewUnassignTaskMutation as MeasureViewUnassignTaskMutationType } from "./__generated__/MeasureViewUnassignTaskMutation.graphql";
import { MeasureViewUpdateMeasureStateMutation as MeasureViewUpdateMeasureStateMutationType } from "./__generated__/MeasureViewUpdateMeasureStateMutation.graphql";
import { MeasureViewQuery as MeasureViewQueryType } from "./__generated__/MeasureViewQuery.graphql";
import {
  MeasureViewOrganizationQuery,
  MeasureViewOrganizationQuery$data,
} from "./__generated__/MeasureViewOrganizationQuery.graphql";
import {
  MeasureViewFrameworksQuery,
  MeasureViewFrameworksQuery$data,
} from "./__generated__/MeasureViewFrameworksQuery.graphql";
import {
  MeasureViewLinkedControlsQuery,
  MeasureViewLinkedControlsQuery$data,
} from "./__generated__/MeasureViewLinkedControlsQuery.graphql";
import { MeasureViewFulfillEvidenceMutation as MeasureViewFulfillEvidenceMutationType } from "./__generated__/MeasureViewFulfillEvidenceMutation.graphql";
import { MeasureViewCreateControlMappingMutation } from "./__generated__/MeasureViewCreateControlMappingMutation.graphql";
import { MeasureViewDeleteControlMappingMutation } from "./__generated__/MeasureViewDeleteControlMappingMutation.graphql";
import { MeasureViewRisksQuery$data } from "./__generated__/MeasureViewRisksQuery.graphql";
import { MeasureViewRisksQuery } from "./__generated__/MeasureViewRisksQuery.graphql";
import { MeasureViewDeleteMeasureMutation } from "./__generated__/MeasureViewDeleteMeasureMutation.graphql";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import { format, formatISO, parseISO } from "date-fns";

// Function to format ISO8601 duration to human-readable format
const formatDuration = (isoDuration: string): string => {
  if (!isoDuration || !isoDuration.startsWith("P")) {
    return isoDuration;
  }

  try {
    const durationRegex =
      /P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?/;
    const matches = isoDuration.match(durationRegex);

    if (!matches) return isoDuration;

    const years = matches[1] ? parseInt(matches[1]) : 0;
    const months = matches[2] ? parseInt(matches[2]) : 0;
    const days = matches[3] ? parseInt(matches[3]) : 0;
    const hours = matches[4] ? parseInt(matches[4]) : 0;
    const minutes = matches[5] ? parseInt(matches[5]) : 0;
    const seconds = matches[6] ? parseInt(matches[6]) : 0;

    const parts = [];
    if (years) parts.push(`${years} ${years === 1 ? "year" : "years"}`);
    if (months) parts.push(`${months} ${months === 1 ? "month" : "months"}`);
    if (days) parts.push(`${days} ${days === 1 ? "day" : "days"}`);
    if (hours) parts.push(`${hours} ${hours === 1 ? "hour" : "hours"}`);
    if (minutes)
      parts.push(`${minutes} ${minutes === 1 ? "minute" : "minutes"}`);
    if (seconds)
      parts.push(`${seconds} ${seconds === 1 ? "second" : "seconds"}`);

    return parts.length > 0 ? parts.join(", ") : "No duration";
  } catch (error) {
    console.error("Error parsing duration:", error);
    return isoDuration;
  }
};

const measureViewQuery = graphql`
  query MeasureViewQuery($measureId: ID!) {
    measure: node(id: $measureId) {
      id
      ... on Measure {
        name
        description
        state
        category
        evidences(first: 100) @connection(key: "MeasureView_evidences") {
          __id
          edges {
            node {
              id
              mimeType
              filename
              size
              state
              type
              url
              createdAt
              description
            }
          }
        }
        tasks(first: 100) @connection(key: "MeasureView_tasks") {
          __id
          edges {
            node {
              id
              name
              description
              state
              timeEstimate
              deadline
              assignedTo {
                id
                fullName
                primaryEmailAddress
              }
              evidences(first: 50)
                @connection(key: "MeasureView_task_evidences") {
                __id
                edges {
                  node {
                    id
                    mimeType
                    filename
                    size
                    state
                    type
                    url
                    createdAt
                    description
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;

const updateTaskStateMutation = graphql`
  mutation MeasureViewUpdateTaskStateMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        id
        state
        timeEstimate
        deadline
      }
    }
  }
`;

const createTaskMutation = graphql`
  mutation MeasureViewCreateTaskMutation(
    $input: CreateTaskInput!
    $connections: [ID!]!
  ) {
    createTask(input: $input) {
      taskEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          timeEstimate
          deadline
          state
          assignedTo {
            id
            fullName
            primaryEmailAddress
          }
        }
      }
    }
  }
`;

const deleteTaskMutation = graphql`
  mutation MeasureViewDeleteTaskMutation(
    $input: DeleteTaskInput!
    $connections: [ID!]!
  ) {
    deleteTask(input: $input) {
      deletedTaskId @deleteEdge(connections: $connections)
    }
  }
`;

const requestEvidenceMutation = graphql`
  mutation MeasureViewRequestEvidenceMutation(
    $input: RequestEvidenceInput!
    $connections: [ID!]!
  ) {
    requestEvidence(input: $input) {
      evidenceEdge @appendEdge(connections: $connections) {
        node {
          id
          filename
          fileUrl
          mimeType
          type
          url
          size
          state
          createdAt
          description
        }
      }
    }
  }
`;

const deleteEvidenceMutation = graphql`
  mutation MeasureViewDeleteEvidenceMutation(
    $input: DeleteEvidenceInput!
    $connections: [ID!]!
  ) {
    deleteEvidence(input: $input) {
      deletedEvidenceId @deleteEdge(connections: $connections)
    }
  }
`;

// Add a GraphQL query to fetch the fileUrl for an evidence item
const getEvidenceFileUrlQuery = graphql`
  query MeasureViewGetEvidenceFileUrlQuery($evidenceId: ID!) {
    node(id: $evidenceId) {
      ... on Evidence {
        id
        fileUrl
      }
    }
  }
`;

const assignTaskMutation = graphql`
  mutation MeasureViewAssignTaskMutation($input: AssignTaskInput!) {
    assignTask(input: $input) {
      task {
        id
        assignedTo {
          id
          fullName
          primaryEmailAddress
        }
      }
    }
  }
`;

const unassignTaskMutation = graphql`
  mutation MeasureViewUnassignTaskMutation($input: UnassignTaskInput!) {
    unassignTask(input: $input) {
      task {
        id
        assignedTo {
          id
          fullName
          primaryEmailAddress
        }
      }
    }
  }
`;

const updateMeasureStateMutation = graphql`
  mutation MeasureViewUpdateMeasureStateMutation($input: UpdateMeasureInput!) {
    updateMeasure(input: $input) {
      measure {
        id
        state
      }
    }
  }
`;

const deleteMeasureMutation = graphql`
  mutation MeasureViewDeleteMeasureMutation(
    $input: DeleteMeasureInput!
    $connections: [ID!]!
  ) {
    deleteMeasure(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

const organizationQuery = graphql`
  query MeasureViewOrganizationQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        peoples(first: 100, orderBy: { direction: ASC, field: FULL_NAME })
          @connection(key: "MeasureView_peoples") {
          edges {
            node {
              id
              fullName
              primaryEmailAddress
            }
          }
        }
      }
    }
  }
`;

// New queries and mutations for Control Mapping
const frameworksQuery = graphql`
  query MeasureViewFrameworksQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        frameworks(first: 100) @connection(key: "Organization__frameworks") {
          edges {
            node {
              id
              name
              controls(first: 100) @connection(key: "Framework__controls") {
                edges {
                  node {
                    id
                    referenceId
                    name
                    description
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;

const linkedControlsQuery = graphql`
  query MeasureViewLinkedControlsQuery($measureId: ID!) {
    measure: node(id: $measureId) {
      id
      ... on Measure {
        controls(first: 100) @connection(key: "MeasureView_controls") {
          edges {
            node {
              id
              referenceId
              name
              description
            }
          }
        }
      }
    }
  }
`;

const createControlMappingMutation = graphql`
  mutation MeasureViewCreateControlMappingMutation(
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

const deleteControlMappingMutation = graphql`
  mutation MeasureViewDeleteControlMappingMutation(
    $input: DeleteControlMeasureMappingInput!
  ) {
    deleteControlMeasureMapping(input: $input) {
      deletedMeasureId
    }
  }
`;

// Add a mutation to fulfill evidence
const fulfillEvidenceMutation = graphql`
  mutation MeasureViewFulfillEvidenceMutation(
    $input: FulfillEvidenceInput!
    $connections: [ID!]!
  ) {
    fulfillEvidence(input: $input) {
      evidenceEdge @appendEdge(connections: $connections) {
        node {
          id
          filename
          fileUrl
          mimeType
          type
          url
          size
          state
          createdAt
          description
        }
      }
    }
  }
`;

// Add query to fetch risks linked to a measure
const measureRisksQuery = graphql`
  query MeasureViewRisksQuery($measureId: ID!) {
    measure: node(id: $measureId) {
      id
      ... on Measure {
        risks(first: 100) @connection(key: "MeasureView_risks") {
          edges {
            node {
              id
              name
              description
              inherentLikelihood
              inherentImpact
              inherentRiskScore
              residualLikelihood
              residualImpact
              residualRiskScore
              createdAt
              updatedAt
            }
          }
        }
      }
    }
  }
`;

// Add a mutation for uploadMeasureEvidence
const uploadMeasureEvidenceMutation = graphql`
  mutation MeasureViewUploadMeasureEvidenceMutation(
    $input: UploadMeasureEvidenceInput!
    $connections: [ID!]!
  ) {
    uploadMeasureEvidence(input: $input) {
      evidenceEdge @appendEdge(connections: $connections) {
        node {
          id
          filename
          fileUrl
          mimeType
          type
          url
          size
          state
          createdAt
          description
        }
      }
    }
  }
`;

// Add the uploadTaskEvidence mutation
const uploadTaskEvidenceMutation = graphql`
  mutation MeasureViewUploadTaskEvidenceMutation(
    $input: UploadTaskEvidenceInput!
    $connections: [ID!]!
  ) {
    uploadTaskEvidence(input: $input) {
      evidenceEdge @appendEdge(connections: $connections) {
        node {
          id
          filename
          fileUrl
          mimeType
          type
          url
          size
          state
          createdAt
          description
        }
      }
    }
  }
`;

function MeasureViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<MeasureViewQueryType>;
}) {
  const data = usePreloadedQuery<MeasureViewQueryType>(
    measureViewQuery,
    queryRef
  );

  // Define a type for the measure with evidences field
  type MeasureWithEvidences = typeof data.measure & {
    evidences?: {
      __id: string;
      edges: Array<{
        node: {
          id: string;
          mimeType: string;
          filename: string;
          size: number;
          state: string;
          type: string;
          url?: string;
          createdAt: string;
          description: string;
        };
      } | null> | null;
    };
  };

  // Cast data.measure to include evidences field
  const measureWithEvidences = data.measure as MeasureWithEvidences;

  const navigate = useNavigate();
  const { toast } = useToast();
  const { organizationId, measureId } = useParams<{
    organizationId: string;
    measureId: string;
  }>();
  const environment = useRelayEnvironment();

  const [commitDeleteMeasure, isDeletingMeasure] = useMutation<any>(
    deleteMeasureMutation
  );
  const [isDeleteMeasureOpen, setIsDeleteMeasureOpen] = useState(false);

  // Add state for main content tabs
  const [mainContentTab, setMainContentTab] = useState<string>("evidence");

  // Add URLSearchParams handling for task persistence
  const [searchParams, setSearchParams] = useSearchParams();
  const taskIdFromUrl = searchParams.get("taskId");

  // Load organization data for people selector
  const [organizationData, setOrganizationData] =
    useState<MeasureViewOrganizationQuery$data | null>(null);

  // Control mapping state
  const [isControlMappingDialogOpen, setIsControlMappingDialogOpen] =
    useState(false);
  const [frameworksData, setFrameworksData] =
    useState<MeasureViewFrameworksQuery$data | null>(null);
  const [linkedControlsData, setLinkedControlsData] =
    useState<MeasureViewLinkedControlsQuery$data | null>(null);
  const [controlSearchQuery, setControlSearchQuery] = useState("");
  const [selectedFrameworkId, setSelectedFrameworkId] = useState<string | null>(
    null
  );
  const [isLoadingControls, setIsLoadingControls] = useState(false);
  const [isLinkingControl, setIsLinkingControl] = useState(false);
  const [isUnlinkingControl, setIsUnlinkingControl] = useState(false);

  // Create mutation hooks for control mapping
  const [commitCreateControlMapping] =
    useMutation<MeasureViewCreateControlMappingMutation>(
      createControlMappingMutation
    );
  const [commitDeleteControlMapping] =
    useMutation<MeasureViewDeleteControlMappingMutation>(
      deleteControlMappingMutation
    );

  useEffect(() => {
    if (organizationId) {
      fetchQuery<MeasureViewOrganizationQuery>(environment, organizationQuery, {
        organizationId,
      }).subscribe({
        next: (data) => {
          setOrganizationData(data);
        },
        error: (error: Error) => {
          console.error("Error fetching organization:", error);
        },
      });
    }
  }, [environment, organizationId]);

  // Load linked controls when component mounts
  useEffect(() => {
    if (measureId) {
      fetchQuery<MeasureViewLinkedControlsQuery>(
        environment,
        linkedControlsQuery,
        { measureId }
      ).subscribe({
        next: (data) => {
          setLinkedControlsData(data);
        },
        error: (error: Error) => {
          console.error("Error fetching linked controls:", error);
        },
      });
    }
  }, [environment, measureId]);

  // Add state for risks data
  const [risksData, setRisksData] = useState<MeasureViewRisksQuery$data | null>(
    null
  );

  // Load risks data when component mounts
  useEffect(() => {
    if (measureId) {
      fetchQuery<MeasureViewRisksQuery>(environment, measureRisksQuery, {
        measureId,
      }).subscribe({
        next: (data) => {
          setRisksData(data);
        },
        error: (error: Error) => {
          console.error("Error fetching risks:", error);
        },
      });
    }
  }, [environment, measureId]);

  const formatState = (state: string | undefined): string => {
    if (!state) return "";

    const upperState = state.toUpperCase();

    if (upperState === "NOT_STARTED") return "Not Started";
    if (upperState === "IN_PROGRESS") return "In Progress";
    if (upperState === "NOT_APPLICABLE") return "Not Applicable";
    if (upperState === "IMPLEMENTED") return "Implemented";

    const formatted = state.toLowerCase();
    return formatted.charAt(0).toUpperCase() + formatted.slice(1);
  };

  const getStateColor = (state: string | undefined): string => {
    if (!state) return "bg-secondary-bg text-primary";

    const upperState = state.toUpperCase();

    if (upperState === "NOT_STARTED") return "bg-secondary-bg text-primary";
    if (upperState === "IN_PROGRESS") return "bg-blue-100 text-blue-800";
    if (upperState === "NOT_APPLICABLE") return "bg-purple-100 text-purple-800";
    if (upperState === "IMPLEMENTED") return "bg-green-100 text-green-800";

    return "bg-secondary-bg text-primary";
  };

  const [updateTask] = useMutation<MeasureViewUpdateTaskStateMutationType>(
    updateTaskStateMutation
  );
  const [createTask] =
    useMutation<MeasureViewCreateTaskMutationType>(createTaskMutation);
  const [deleteTask] =
    useMutation<MeasureViewDeleteTaskMutationType>(deleteTaskMutation);
  const [requestEvidence] = useMutation<any>(requestEvidenceMutation);
  const [deleteEvidence] = useMutation<MeasureViewDeleteEvidenceMutationType>(
    deleteEvidenceMutation
  );
  const [assignTask] =
    useMutation<MeasureViewAssignTaskMutationType>(assignTaskMutation);
  const [unassignTask] =
    useMutation<MeasureViewUnassignTaskMutationType>(unassignTaskMutation);
  const [fulfillEvidence] = useMutation<MeasureViewFulfillEvidenceMutationType>(
    fulfillEvidenceMutation
  );

  const [updateMeasureState] =
    useMutation<MeasureViewUpdateMeasureStateMutationType>(
      updateMeasureStateMutation
    );

  const [isCreateTaskOpen, setIsCreateTaskOpen] = useState(false);
  const [newTaskName, setNewTaskName] = useState("");
  const [newTaskDescription, setNewTaskDescription] = useState("");
  const [newDeadline, setNewDeadline] = useState("");
  const [timeEstimateDays, setTimeEstimateDays] = useState("");
  const [timeEstimateHours, setTimeEstimateHours] = useState("");
  const [timeEstimateMinutes, setTimeEstimateMinutes] = useState("");

  const [isDeleteTaskOpen, setIsDeleteTaskOpen] = useState(false);
  const [taskToDelete, setTaskToDelete] = useState<{
    id: string;
    name: string;
  } | null>(null);

  const [taskForEvidence, setTaskForEvidence] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const hiddenFileInputRef = useRef<HTMLInputElement>(null);

  const [draggedOverTaskId, setDraggedOverTaskId] = useState<string | null>(
    null
  );
  const [uploadingTaskId, setUploadingTaskId] = useState<string | null>(null);
  const [isDraggingFile, setIsDraggingFile] = useState(false);

  // State to track which task's evidence list is expanded
  const [expandedEvidenceTaskId, setExpandedEvidenceTaskId] = useState<
    string | null
  >(null);

  // Add state for the preview modal
  const [isPreviewModalOpen, setIsPreviewModalOpen] = useState(false);
  const [previewEvidence, setPreviewEvidence] = useState<{
    id: string;
    filename: string;
    mimeType: string;
    fileUrl?: string;
  } | null>(null);
  const [isLoadingFileUrl, setIsLoadingFileUrl] = useState(false);

  const [isDeleteEvidenceOpen, setIsDeleteEvidenceOpen] = useState(false);
  const [evidenceToDelete, setEvidenceToDelete] = useState<{
    id: string;
    filename: string;
    taskId: string;
  } | null>(null);

  // Add state for people selector
  const [peoplePopoverOpen, setPeoplePopoverOpen] = useState<{
    [key: string]: boolean;
  }>({});

  // Add state for people search
  const [peopleSearch, setPeopleSearch] = useState<{
    [key: string]: string;
  }>({});

  // Add state variables for the evidence dialog and link evidence
  const [evidenceDialogOpen, setEvidenceDialogOpen] = useState(false);
  const [linkEvidenceName, setLinkEvidenceName] = useState("");
  const [linkEvidenceUrl, setLinkEvidenceUrl] = useState("");
  const [linkEvidenceDescription, setLinkEvidenceDescription] = useState("");
  const [activeTab, setActiveTab] = useState<"file" | "link">("file");

  // State for fulfill evidence dialog
  const [fulfillEvidenceDialogOpen, setFulfillEvidenceDialogOpen] =
    useState(false);
  const [evidenceToFulfill, setEvidenceToFulfill] = useState<{
    id: string;
    filename: string;
  } | null>(null);

  // Add state for selected task panel
  const [selectedTask, setSelectedTask] = useState<(typeof tasks)[0] | null>(
    null
  );

  // Track if task panel is open
  const [isTaskPanelOpen, setIsTaskPanelOpen] = useState(false);

  // Add state for tracking when files are dragged over the evidence zone
  const [isDraggingOverEvidenceZone, setIsDraggingOverEvidenceZone] =
    useState(false);

  const tasks = data.measure.tasks?.edges.map((edge) => edge.node) || [];

  // Add useEffect to handle URL parameters for task selection
  useEffect(() => {
    // If there's a taskId in the URL, find that task and select it
    if (taskIdFromUrl && tasks.length > 0) {
      const taskFromUrl = tasks.find((task) => task.id === taskIdFromUrl);
      if (taskFromUrl) {
        setSelectedTask(taskFromUrl);
        setIsTaskPanelOpen(true);
      } else {
        // If task ID is invalid, remove it from URL
        searchParams.delete("taskId");
        setSearchParams(searchParams);
      }
    }
  }, [taskIdFromUrl, tasks, searchParams, setSearchParams]);

  const getEvidenceConnectionId = useCallback(
    (taskId: string) => {
      const task = tasks.find((t) => t.id === taskId);
      if (task?.evidences?.__id) {
        return task.evidences.__id;
      }
      return null;
    },
    [tasks]
  );

  // Add a function to get the measure evidence connection ID
  const getMeasureEvidenceConnectionId = useCallback(() => {
    return measureWithEvidences.evidences?.__id || null;
  }, [measureWithEvidences.evidences]);

  // Function to convert days, hours, and minutes to ISO 8601 duration format
  const convertToISODuration = useCallback(() => {
    let duration = "P";

    if (timeEstimateDays && parseInt(timeEstimateDays) > 0) {
      duration += `${parseInt(timeEstimateDays)}D`;
    }

    if (
      (timeEstimateHours && parseInt(timeEstimateHours) > 0) ||
      (timeEstimateMinutes && parseInt(timeEstimateMinutes) > 0)
    ) {
      duration += "T";

      if (timeEstimateHours && parseInt(timeEstimateHours) > 0) {
        duration += `${parseInt(timeEstimateHours)}H`;
      }

      if (timeEstimateMinutes && parseInt(timeEstimateMinutes) > 0) {
        duration += `${parseInt(timeEstimateMinutes)}M`;
      }
    }

    // Return empty string if no time components were provided
    return duration === "P" ? "" : duration;
  }, [timeEstimateDays, timeEstimateHours, timeEstimateMinutes]);

  useEffect(() => {
    const handleDragEnter = (e: globalThis.DragEvent) => {
      e.preventDefault();
      if (e.dataTransfer?.types.includes("Files")) {
        setIsDraggingFile(true);
      }
    };

    const handleDragLeave = (e: globalThis.DragEvent) => {
      e.preventDefault();
      // Only set to false if we're leaving the window
      if (!e.relatedTarget || (e.relatedTarget as Node).nodeName === "HTML") {
        setIsDraggingFile(false);
      }
    };

    const handleDragOver = (e: globalThis.DragEvent) => {
      e.preventDefault();
    };

    const handleDrop = (e: globalThis.DragEvent) => {
      e.preventDefault();
      setIsDraggingFile(false);
    };

    document.addEventListener("dragenter", handleDragEnter);
    document.addEventListener("dragleave", handleDragLeave);
    document.addEventListener("dragover", handleDragOver);
    document.addEventListener("drop", handleDrop);

    return () => {
      document.removeEventListener("dragenter", handleDragEnter);
      document.removeEventListener("dragleave", handleDragLeave);
      document.removeEventListener("dragover", handleDragOver);
      document.removeEventListener("drop", handleDrop);
    };
  }, []);

  const handleTaskClick = (task: (typeof tasks)[0]) => {
    if (!task) return;
    setSelectedTask(task);
    setIsTaskPanelOpen(true);

    // Add the task ID to URL parameters
    searchParams.set("taskId", task.id);
    setSearchParams(searchParams);
  };

  const handleToggleTaskState = (taskId: string, currentState: string) => {
    const newState = currentState === "DONE" ? "TODO" : "DONE";

    updateTask({
      variables: {
        input: {
          taskId,
          state: newState,
        },
      },
      onCompleted: () => {
        if (selectedTask && selectedTask.id === taskId) {
          setSelectedTask({
            ...selectedTask,
            state: newState,
          });
        }
      },
      onError: (error) => {
        toast({
          title: "Error updating task",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  const handleCreateTask = () => {
    if (!newTaskName.trim()) {
      toast({
        title: "Error creating task",
        description: "Task name is required",
        variant: "destructive",
      });
      return;
    }

    if (!data.measure.id) {
      toast({
        title: "Error creating task",
        description: "Measure ID is missing",
        variant: "destructive",
      });
      return;
    }
    // Convert the time estimate components to ISO 8601 format
    const isoTimeEstimate = convertToISODuration();

    createTask({
      variables: {
        connections: [`${data.measure.tasks?.__id}`],
        input: {
          organizationId: organizationId!,
          measureId: data.measure.id,
          name: newTaskName,
          description: newTaskDescription,
          timeEstimate: isoTimeEstimate === "" ? null : isoTimeEstimate,
          deadline: newDeadline === "" ? null : formatISO(newDeadline),
        },
      },
      onCompleted: () => {
        toast({
          title: "Task created",
          description: "New task has been created successfully.",
        });
        setNewTaskName("");
        setNewTaskDescription("");
        setTimeEstimateDays("");
        setTimeEstimateHours("");
        setTimeEstimateMinutes("");
        setIsCreateTaskOpen(false);
        setNewDeadline("");
      },
      onError: (error) => {
        toast({
          title: "Error creating task",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  const handleDeleteTask = (taskId: string, taskName: string) => {
    setTaskToDelete({ id: taskId, name: taskName });
    setIsDeleteTaskOpen(true);
  };

  const confirmDeleteTask = () => {
    if (!taskToDelete) return;

    deleteTask({
      variables: {
        connections: [`${data.measure.tasks?.__id}`],
        input: {
          taskId: taskToDelete.id,
        },
      },
      onCompleted: () => {
        toast({
          title: "Task deleted",
          description: "Task has been deleted successfully.",
        });
        setIsDeleteTaskOpen(false);
        setTaskToDelete(null);
      },
      onError: (error) => {
        toast({
          title: "Error deleting task",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  const handleEditMeasure = () => {
    navigate(`/organizations/${organizationId}/measures/${measureId}/edit`);
  };

  const handleCreateEvidence = (taskId: string, taskName: string) => {
    setTaskForEvidence({ id: taskId, name: taskName });
    setEvidenceDialogOpen(true);
    // Reset form fields
    setLinkEvidenceName("");
    setLinkEvidenceUrl("");
    setLinkEvidenceDescription("");
  };

  const handleFileSelected = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length === 0 || !taskForEvidence)
      return;

    const file = e.target.files[0];

    // Show toast for add started
    toast({
      title: "Adding document",
      description: `Adding ${file.name}...`,
      variant: "default",
    });

    // Get the evidence connection ID for this task
    const evidenceConnectionId = getEvidenceConnectionId(taskForEvidence.id);

    uploadTaskEvidence({
      variables: {
        input: {
          taskId: taskForEvidence.id,
          file: null,
        },
        connections: evidenceConnectionId ? [evidenceConnectionId] : [],
      },
      uploadables: {
        "input.file": file,
      },
      onCompleted: () => {
        setTaskForEvidence(null);
        // Reset the file input
        if (hiddenFileInputRef.current) {
          hiddenFileInputRef.current.value = "";
        }
      },
      onError: (error) => {
        toast({
          title: "Error adding document",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  const handleLinkEvidenceSubmit = () => {
    // Validate form
    if (!linkEvidenceName.trim()) {
      toast({
        title: "Error",
        description: "Please provide a name for the evidence",
        variant: "destructive",
      });
      return;
    }

    if (!linkEvidenceUrl.trim()) {
      toast({
        title: "Error",
        description: "Please provide a URL for the evidence",
        variant: "destructive",
      });
      return;
    }

    try {
      console.log("Creating URI file with URL:", linkEvidenceUrl.trim());

      // Create a URI file with the link
      const linkContent = linkEvidenceUrl.trim();
      const fileName = `${linkEvidenceName.trim()}.uri`;
      const file = new File([linkContent], fileName, { type: "text/uri-list" });

      console.log("File created:", file.name, file.type, file.size);

      // Show toast for add started
      toast({
        title: "Adding link evidence",
        description: `Adding ${fileName}...`,
        variant: "default",
      });

      if (taskForEvidence) {
        // This is a task-specific evidence
        const evidenceConnectionId = getEvidenceConnectionId(
          taskForEvidence.id
        );

        // Upload the URI file as task evidence
        uploadTaskEvidence({
          variables: {
            input: {
              taskId: taskForEvidence.id,
              file: null,
            },
            connections: evidenceConnectionId ? [evidenceConnectionId] : [],
          },
          uploadables: {
            "input.file": file,
          },
          onCompleted: () => {
            setTaskForEvidence(null);
            setEvidenceDialogOpen(false);
            // Reset form fields
            setLinkEvidenceName("");
            setLinkEvidenceUrl("");
            setLinkEvidenceDescription("");

            toast({
              title: "Link evidence added",
              description: "Link evidence has been added successfully.",
              variant: "default",
            });
          },
          onError: (error) => {
            console.error("Error uploading URI file:", error);
            toast({
              title: "Error adding link evidence",
              description: error.message || "Unknown error occurred",
              variant: "destructive",
            });
          },
        });
      } else {
        // This is a measure-level evidence
        const evidenceConnectionId = getMeasureEvidenceConnectionId();

        // Upload the URI file as measure evidence
        uploadMeasureEvidence({
          variables: {
            input: {
              measureId: measureId,
              file: null,
            },
            connections: evidenceConnectionId ? [evidenceConnectionId] : [],
          },
          uploadables: {
            "input.file": file,
          },
          onCompleted: () => {
            setEvidenceDialogOpen(false);
            // Reset form fields
            setLinkEvidenceName("");
            setLinkEvidenceUrl("");
            setLinkEvidenceDescription("");

            toast({
              title: "Link evidence added",
              description: "Link evidence has been added successfully.",
              variant: "default",
            });
          },
          onError: (error) => {
            console.error("Error uploading URI file:", error);
            toast({
              title: "Error adding link evidence",
              description: error.message || "Unknown error occurred",
              variant: "destructive",
            });
          },
        });
      }
    } catch (error) {
      console.error("Error in handleLinkEvidenceSubmit:", error);
      toast({
        title: "Error creating URI file",
        description:
          error instanceof Error ? error.message : "Unknown error occurred",
        variant: "destructive",
      });
    }
  };

  const handleDragOver = (e: DragEvent<HTMLDivElement>, taskId: string) => {
    e.preventDefault();
    e.stopPropagation();
    if (draggedOverTaskId !== taskId) {
      setDraggedOverTaskId(taskId);
    }
  };

  const handleDragLeave = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setDraggedOverTaskId(null);
  };

  const handleDrop = async (e: DragEvent<HTMLDivElement>, taskId: string) => {
    e.preventDefault();
    e.stopPropagation();
    setDraggedOverTaskId(null);
    setIsDraggingFile(false);

    const files = e.dataTransfer.files;
    if (files.length === 0) return;

    const file = files[0];
    setUploadingTaskId(taskId);

    // Show toast for add started
    toast({
      title: "Adding document",
      description: `Adding ${file.name}...`,
      variant: "default",
    });

    // Get the evidence connection ID for this task
    const evidenceConnectionId = getEvidenceConnectionId(taskId);

    uploadTaskEvidence({
      variables: {
        input: {
          taskId: taskId,
          file: null,
        },
        connections: evidenceConnectionId ? [evidenceConnectionId] : [],
      },
      uploadables: {
        "input.file": file,
      },
      onCompleted: () => {
        setUploadingTaskId(null);
        toast({
          title: "Document added",
          description: "Document evidence has been added successfully.",
          variant: "default",
        });
      },
      onError: (error) => {
        setUploadingTaskId(null);
        toast({
          title: "Error adding document",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  // Function to toggle evidence list expansion
  const toggleEvidenceList = (taskId: string) => {
    if (expandedEvidenceTaskId === taskId) {
      setExpandedEvidenceTaskId(null);
    } else {
      setExpandedEvidenceTaskId(taskId);
    }
  };

  // Function to get file icon based on mime type and evidence type
  const getFileIcon = (mimeType: string, evidenceType: string) => {
    if (
      evidenceType === "LINK" ||
      mimeType === "text/uri-list" ||
      mimeType === "text/uri"
    ) {
      return <Link2 className="w-4 h-4 text-blue-600" />;
    } else if (mimeType.startsWith("image/")) {
      return <Image className="w-4 h-4 text-blue-500" />;
    } else if (mimeType.includes("pdf")) {
      return <FileText className="w-4 h-4 text-red-500" />;
    } else if (mimeType.includes("word") || mimeType.includes("document")) {
      return <FileText className="w-4 h-4 text-blue-600" />;
    } else if (mimeType.includes("excel") || mimeType.includes("spreadsheet")) {
      return <FileText className="w-4 h-4 text-green-600" />;
    } else {
      return <FileGeneric className="w-4 h-4 text-secondary" />;
    }
  };

  // Function to get URL from URI file
  const getUrlFromUriFile = async (fileUrl: string): Promise<string | null> => {
    try {
      const response = await fetch(fileUrl);
      const text = await response.text();
      // URI files typically have the URL on the first line
      const firstLine = text.trim().split("\n")[0];
      return firstLine || null;
    } catch (error) {
      console.error("Error extracting URL from URI file:", error);
      return null;
    }
  };

  // Simplified preview handler that opens the modal and sets the evidence
  const handlePreviewEvidence = (evidence: {
    id: string;
    filename: string;
    mimeType: string;
  }) => {
    setPreviewEvidence({
      id: evidence.id,
      filename: evidence.filename,
      mimeType: evidence.mimeType,
    });
    setIsPreviewModalOpen(true);
    setIsLoadingFileUrl(true);
    fetchQuery(environment, getEvidenceFileUrlQuery, {
      evidenceId: evidence.id,
    })
      .toPromise()
      .then((response) => {
        const data = response as { node?: { id: string; fileUrl?: string } };

        if (data?.node?.fileUrl) {
          setPreviewEvidence((prev) => {
            if (!prev) return null;
            return { ...prev, fileUrl: data.node!.fileUrl };
          });

          // If it's a URI file, fetch and extract the URL
          if (
            evidence.mimeType === "text/uri-list" ||
            evidence.mimeType === "text/uri"
          ) {
            getUrlFromUriFile(data.node!.fileUrl!)
              .then((url) => {
                if (url) {
                  window.open(url, "_blank");
                  setIsPreviewModalOpen(false);
                }
              })
              .catch((error) => {
                console.error("Error handling URI file:", error);
              });
          }
        } else {
          throw new Error("File URL not available in response");
        }
      })
      .catch((error) => {
        console.error("Error fetching file URL:", error);
        toast({
          title: "Error fetching file URL",
          description: error.message || "Could not load the file preview",
          variant: "destructive",
        });
      })
      .finally(() => {
        setIsLoadingFileUrl(false);
      });
  };

  const handleDeleteEvidence = (
    evidenceId: string,
    filename: string,
    taskId: string
  ) => {
    setEvidenceToDelete({ id: evidenceId, filename, taskId });
    setIsDeleteEvidenceOpen(true);
  };

  const handleFulfillEvidence = (evidenceId: string, filename: string) => {
    setEvidenceToFulfill({ id: evidenceId, filename });
    setFulfillEvidenceDialogOpen(true);
    setActiveTab("file");
    setLinkEvidenceUrl("");
    setLinkEvidenceDescription("");
  };

  const handleFulfillEvidenceWithFile = (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    if (!e.target.files || e.target.files.length === 0 || !evidenceToFulfill)
      return;

    const file = e.target.files[0];

    // Show toast for upload started
    toast({
      title: "Uploading evidence",
      description: `Uploading ${file.name}...`,
      variant: "default",
    });

    const evidenceId = evidenceToFulfill.id;
    // Get connection ID for the parent task (assuming it's available in the view)
    const task = tasks.find((task) =>
      task.evidences?.edges.some((edge) => edge?.node?.id === evidenceId)
    );

    const evidenceConnectionId = task?.id
      ? getEvidenceConnectionId(task.id)
      : null;

    fulfillEvidence({
      variables: {
        input: {
          evidenceId: evidenceId,
          file: null,
        },
        connections: evidenceConnectionId ? [evidenceConnectionId] : [],
      },
      uploadables: {
        "input.file": file,
      },
      onCompleted: () => {
        setFulfillEvidenceDialogOpen(false);
        setEvidenceToFulfill(null);
        toast({
          title: "Evidence uploaded",
          description: "Evidence has been fulfilled successfully.",
          variant: "default",
        });
        // Reset the file input
        if (hiddenFileInputRef.current) {
          hiddenFileInputRef.current.value = "";
        }
      },
      onError: (error) => {
        toast({
          title: "Error uploading evidence",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  const handleFulfillEvidenceWithLink = () => {
    if (!evidenceToFulfill) {
      toast({
        title: "Error",
        description: "No evidence to fulfill",
        variant: "destructive",
      });
      return;
    }

    // Validate URL
    if (!linkEvidenceUrl.trim()) {
      toast({
        title: "Error",
        description: "Please provide a URL for the evidence",
        variant: "destructive",
      });
      return;
    }

    try {
      console.log(
        "Creating URI file to fulfill evidence:",
        evidenceToFulfill.id
      );

      // Create a URI file with the link
      const linkContent = linkEvidenceUrl.trim();
      // Use provided name or fallback to the original filename
      const baseName = linkEvidenceName.trim()
        ? linkEvidenceName.trim()
        : evidenceToFulfill.filename.replace(/\.[^/.]+$/, "");
      const fileName = `${baseName}.uri`;
      const file = new File([linkContent], fileName, { type: "text/uri-list" });

      console.log("File created:", file.name, file.type, file.size);

      // Get connection ID for the parent task
      const task = tasks.find((task) =>
        task.evidences?.edges.some(
          (edge) => edge?.node?.id === evidenceToFulfill.id
        )
      );

      const evidenceConnectionId = task?.id
        ? getEvidenceConnectionId(task.id)
        : getMeasureEvidenceConnectionId();

      console.log("Using connection ID:", evidenceConnectionId);
      console.log("Task found:", task?.id);

      // Show toast for upload started
      toast({
        title: "Fulfilling evidence with link",
        description: `Creating link file for ${fileName}...`,
        variant: "default",
      });

      // Upload the URI file to fulfill the evidence
      fulfillEvidence({
        variables: {
          input: {
            evidenceId: evidenceToFulfill.id,
            file: null,
          },
          connections: evidenceConnectionId ? [evidenceConnectionId] : [],
        },
        uploadables: {
          "input.file": file,
        },
        onCompleted: (response) => {
          console.log("Evidence fulfilled successfully:", response);
          setFulfillEvidenceDialogOpen(false);
          setEvidenceToFulfill(null);
          setLinkEvidenceName("");
          setLinkEvidenceUrl("");
          setLinkEvidenceDescription("");
          toast({
            title: "Evidence fulfilled",
            description: "Evidence has been fulfilled with link successfully.",
            variant: "default",
          });
        },
        onError: (error) => {
          console.error("Error fulfilling evidence:", error);
          toast({
            title: "Error fulfilling evidence",
            description: error.message || "Unknown error occurred",
            variant: "destructive",
          });
        },
      });
    } catch (error) {
      console.error("Error in handleFulfillEvidenceWithLink:", error);
      toast({
        title: "Error creating URI file",
        description:
          error instanceof Error ? error.message : "Unknown error occurred",
        variant: "destructive",
      });
    }
  };

  const confirmDeleteEvidence = () => {
    if (!evidenceToDelete) return;

    let evidenceConnectionId;

    // If taskId is empty, it's a measure-level evidence
    if (!evidenceToDelete.taskId) {
      evidenceConnectionId = getMeasureEvidenceConnectionId();
    } else {
      // Otherwise it's a task-level evidence
      evidenceConnectionId = getEvidenceConnectionId(evidenceToDelete.taskId);
    }

    deleteEvidence({
      variables: {
        input: {
          evidenceId: evidenceToDelete.id,
        },
        connections: evidenceConnectionId ? [evidenceConnectionId] : [],
      },
      onCompleted: () => {
        setIsDeleteEvidenceOpen(false);
        setEvidenceToDelete(null);
        toast({
          title: "Evidence deleted",
          description: "Evidence has been deleted successfully.",
        });
      },
      onError: (error) => {
        toast({
          title: "Error deleting evidence",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  // Function to handle assigning a person to a task
  const handleAssignPerson = (taskId: string, personId: string) => {
    assignTask({
      variables: {
        input: {
          taskId,
          assignedToId: personId,
        },
      },
      onCompleted: () => {
        setPeoplePopoverOpen((prev) => ({ ...prev, [taskId]: false }));
      },
      onError: (error) => {
        toast({
          title: "Error assigning task",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  // Function to handle unassigning a person from a task
  const handleUnassignPerson = (taskId: string) => {
    unassignTask({
      variables: {
        input: {
          taskId,
        },
      },
      onError: (error) => {
        toast({
          title: "Error unassigning task",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  // Function to handle measure state change
  const handleMeasureStateChange = (newState: string) => {
    updateMeasureState({
      variables: {
        input: {
          id: data.measure.id,
          state: newState as
            | "NOT_STARTED"
            | "IN_PROGRESS"
            | "IMPLEMENTED"
            | "NOT_APPLICABLE",
        },
      },
      onCompleted: () => {
        toast({
          title: "Measure state updated",
          description: `Measure state has been updated to ${formatState(
            newState
          )}.`,
        });
      },
      onError: (error) => {
        toast({
          title: "Error updating measure state",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  // Update SheetContent to handle closing
  const handleCloseTaskPanel = () => {
    setSelectedTask(null);
    setIsTaskPanelOpen(false);
    // Remove taskId from URL when panel is closed
    searchParams.delete("taskId");
    setSearchParams(searchParams);
  };

  // Add state variables for tracking edit mode and duration components
  const [isEditingDuration, setIsEditingDuration] = useState(false);
  const [isEditingDeadline, setIsEditingDeadline] = useState(false);
  const [editTimeEstimateDays, setEditTimeEstimateDays] = useState("");
  const [editTimeEstimateHours, setEditTimeEstimateHours] = useState("");
  const [editTimeEstimateMinutes, setEditTimeEstimateMinutes] = useState("");
  const [editDeadline, setEditDeadline] = useState("");

  // Function to parse ISO duration string into components for editing
  const parseISODuration = useCallback(
    (duration: string | null | undefined) => {
      if (!duration || !duration.startsWith("P")) {
        return { days: "", hours: "", minutes: "" };
      }

      try {
        const durationRegex =
          /P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?/;
        const matches = duration.match(durationRegex);

        if (!matches) return { days: "", hours: "", minutes: "" };

        // We only care about days, hours, and minutes
        const days = matches[3] ? matches[3] : "";
        const hours = matches[4] ? matches[4] : "";
        const minutes = matches[5] ? matches[5] : "";

        return { days, hours, minutes };
      } catch (error) {
        console.error("Error parsing duration:", error);
        return { days: "", hours: "", minutes: "" };
      }
    },
    []
  );

  // Function to handle saving the updated deadline
  const handleSaveDeadline = useCallback(
    (taskId: string) => {
      const newDeadline = editDeadline === "" ? null : formatISO(editDeadline);
      updateTask({
        variables: {
          input: {
            taskId,
            deadline: newDeadline,
          },
        },
        onCompleted: () => {
          setIsEditingDeadline(false);

          // Update the selected task state if it's the current task
          if (selectedTask && selectedTask.id === taskId) {
            setSelectedTask({
              ...selectedTask,
              deadline: newDeadline,
            });
          }
        },
        onError: (error) => {
          toast({
            title: "Error updating task",
            description: error.message,
            variant: "destructive",
          });
        },
      });
    },
    [editDeadline, updateTask, toast, selectedTask, setSelectedTask]
  );

  // Function to handle saving the updated duration
  const handleSaveDuration = useCallback(
    (taskId: string) => {
      // Convert to ISO duration format
      let duration = "P";

      if (editTimeEstimateDays && parseInt(editTimeEstimateDays) > 0) {
        duration += `${parseInt(editTimeEstimateDays)}D`;
      }

      if (
        (editTimeEstimateHours && parseInt(editTimeEstimateHours) > 0) ||
        (editTimeEstimateMinutes && parseInt(editTimeEstimateMinutes) > 0)
      ) {
        duration += "T";

        if (editTimeEstimateHours && parseInt(editTimeEstimateHours) > 0) {
          duration += `${parseInt(editTimeEstimateHours)}H`;
        }

        if (editTimeEstimateMinutes && parseInt(editTimeEstimateMinutes) > 0) {
          duration += `${parseInt(editTimeEstimateMinutes)}M`;
        }
      }

      // If no valid time components were provided, use null (remove the time estimate)
      const timeEstimate = duration === "P" ? null : duration;

      updateTask({
        variables: {
          input: {
            taskId,
            timeEstimate,
          },
        },
        onCompleted: () => {
          setIsEditingDuration(false);

          // Update the selected task state if it's the current task
          if (selectedTask && selectedTask.id === taskId) {
            setSelectedTask({
              ...selectedTask,
              timeEstimate,
            });
          }
        },
        onError: (error) => {
          toast({
            title: "Error updating task",
            description: error.message,
            variant: "destructive",
          });
        },
      });
    },
    [
      editTimeEstimateDays,
      editTimeEstimateHours,
      editTimeEstimateMinutes,
      updateTask,
      toast,
      selectedTask,
      setSelectedTask,
    ]
  );

  // Control mapping functions
  const loadFrameworksAndControls = useCallback(() => {
    if (!organizationId || !measureId) return;

    setIsLoadingControls(true);

    // Fetch all frameworks and their controls
    fetchQuery<MeasureViewFrameworksQuery>(environment, frameworksQuery, {
      organizationId,
    }).subscribe({
      next: (data) => {
        setFrameworksData(data);
        if (
          data?.organization?.frameworks?.edges &&
          data.organization.frameworks.edges.length > 0 &&
          !selectedFrameworkId
        ) {
          // Select the first framework by default if none is selected
          const frameworks = data.organization.frameworks.edges;
          if (frameworks[0]?.node?.id) {
            setSelectedFrameworkId(frameworks[0].node.id);
          }
        }
      },
      complete: () => {
        // Fetch already linked controls for this measure
        fetchQuery<MeasureViewLinkedControlsQuery>(
          environment,
          linkedControlsQuery,
          {
            measureId,
          }
        ).subscribe({
          next: (data) => {
            setLinkedControlsData(data);
            setIsLoadingControls(false);
          },
          error: (error: Error) => {
            console.error("Error fetching linked controls:", error);
            setIsLoadingControls(false);
            toast({
              title: "Error",
              description: "Failed to load linked controls.",
              variant: "destructive",
            });
          },
        });
      },
      error: (error: Error) => {
        console.error("Error fetching frameworks:", error);
        setIsLoadingControls(false);
        toast({
          title: "Error",
          description: "Failed to load frameworks and controls.",
          variant: "destructive",
        });
      },
    });
  }, [environment, measureId, organizationId, selectedFrameworkId, toast]);

  const getControls = useCallback(() => {
    if (!frameworksData?.organization?.frameworks?.edges) return [];

    // Get controls from the selected framework
    const frameworks = frameworksData.organization.frameworks.edges;
    if (selectedFrameworkId) {
      const selectedFramework = frameworks.find(
        (edge) => edge.node.id === selectedFrameworkId
      );

      if (selectedFramework?.node?.controls?.edges) {
        return selectedFramework.node.controls.edges.map((edge) => edge.node);
      }
    }

    // If no framework is selected or it doesn't have controls, return controls from all frameworks
    return frameworks.flatMap((framework) =>
      framework.node.controls.edges.map((edge) => edge.node)
    );
  }, [frameworksData, selectedFrameworkId]);

  const getLinkedControls = useCallback(() => {
    if (!linkedControlsData?.measure?.controls?.edges) return [];
    return (linkedControlsData.measure.controls.edges || []).map(
      (edge) => edge.node
    );
  }, [linkedControlsData]);

  const isControlLinked = useCallback(
    (controlId: string) => {
      const linkedControls = getLinkedControls();
      return linkedControls.some((control) => control.id === controlId);
    },
    [getLinkedControls]
  );

  const handleLinkControl = useCallback(
    (controlId: string) => {
      if (!measureId) return;

      setIsLinkingControl(true);

      commitCreateControlMapping({
        variables: {
          input: {
            controlId,
            measureId,
          },
        },
        onCompleted: (_, errors) => {
          setIsLinkingControl(false);

          if (errors) {
            console.error("Error linking control:", errors);
            toast({
              title: "Error",
              description: "Failed to link control. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh linked controls data
          fetchQuery<MeasureViewLinkedControlsQuery>(
            environment,
            linkedControlsQuery,
            {
              measureId,
            }
          ).subscribe({
            next: (data) => {
              setLinkedControlsData(data);
            },
            error: (error: Error) => {
              console.error("Error refreshing linked controls:", error);
            },
          });

          toast({
            title: "Success",
            description: "Control successfully linked to measure.",
          });
        },
        onError: (error) => {
          setIsLinkingControl(false);
          console.error("Error linking control:", error);
          toast({
            title: "Error",
            description: "Failed to link control. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [commitCreateControlMapping, environment, measureId, toast]
  );

  const handleUnlinkControl = useCallback(
    (controlId: string) => {
      if (!measureId) return;

      setIsUnlinkingControl(true);

      commitDeleteControlMapping({
        variables: {
          input: {
            controlId,
            measureId,
          },
        },
        onCompleted: (_, errors) => {
          setIsUnlinkingControl(false);

          if (errors) {
            console.error("Error unlinking control:", errors);
            toast({
              title: "Error",
              description: "Failed to unlink control. Please try again.",
              variant: "destructive",
            });
            return;
          }

          // Refresh linked controls data
          fetchQuery(environment, linkedControlsQuery, {
            measureId,
          }).subscribe({
            next: (data: unknown) => {
              setLinkedControlsData(
                data as MeasureViewLinkedControlsQuery$data
              );
            },
            error: (error: Error) => {
              console.error("Error refreshing linked controls:", error);
            },
          });

          toast({
            title: "Success",
            description: "Control successfully unlinked from measure.",
          });
        },
        onError: (error) => {
          setIsUnlinkingControl(false);
          console.error("Error unlinking control:", error);
          toast({
            title: "Error",
            description: "Failed to unlink control. Please try again.",
            variant: "destructive",
          });
        },
      });
    },
    [commitDeleteControlMapping, environment, measureId, toast]
  );

  const handleOpenControlMappingDialog = useCallback(() => {
    loadFrameworksAndControls();
    setIsControlMappingDialogOpen(true);
  }, [loadFrameworksAndControls]);

  const filteredControls = useCallback(() => {
    const controls = getControls();
    if (!controlSearchQuery) return controls;

    const lowerQuery = controlSearchQuery.toLowerCase();
    return controls.filter(
      (control) =>
        control.referenceId.toLowerCase().includes(lowerQuery) ||
        control.name.toLowerCase().includes(lowerQuery) ||
        (control.description &&
          control.description.toLowerCase().includes(lowerQuery))
    );
  }, [controlSearchQuery, getControls]);

  // Helper function to get risk severity color
  const getRiskSeverityColor = (severity: number): string => {
    if (severity >= 0.75) return "bg-red-100 text-red-800";
    if (severity >= 0.5) return "bg-orange-100 text-orange-800";
    if (severity >= 0.25) return "bg-yellow-100 text-yellow-800";
    return "bg-green-100 text-green-800";
  };

  // Helper function to get risk severity text
  const getRiskSeverityText = (severity: number): string => {
    if (severity >= 0.75) return "Critical";
    if (severity >= 0.5) return "High";
    if (severity >= 0.25) return "Medium";
    return "Low";
  };

  // Format likelihood as text
  const formatLikelihood = (value: number): string => {
    if (value <= 0.1) return "Very Low";
    if (value <= 0.3) return "Low";
    if (value <= 0.5) return "Medium";
    if (value <= 0.7) return "High";
    return "Very High";
  };

  // Format impact as text
  const formatImpact = (value: number): string => {
    if (value <= 0.1) return "Very Low";
    if (value <= 0.3) return "Low";
    if (value <= 0.5) return "Medium";
    if (value <= 0.7) return "High";
    return "Very High";
  };

  const handleDeleteMeasure = () => {
    setIsDeleteMeasureOpen(true);
  };

  const confirmDeleteMeasure = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId!,
      "MeasureListView_measures"
    );

    commitDeleteMeasure({
      variables: {
        connections: [connectionId],
        input: {
          measureId: measureId!,
        },
      },
      onCompleted: () => {
        toast({
          title: "Measure deleted",
          description: "Measure has been deleted successfully.",
        });
        navigate(`/organizations/${organizationId}/measures`);
      },
      onError: (error) => {
        toast({
          title: "Error deleting measure",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  // Add the mutation hook for uploadMeasureEvidence
  const [uploadMeasureEvidence] = useMutation<any>(
    uploadMeasureEvidenceMutation
  );

  // Add a function to handle uploading evidence directly to the measure
  const handleUploadMeasureEvidence = (file: File) => {
    // Show toast for add started
    toast({
      title: "Adding document",
      description: `Adding ${file.name}...`,
      variant: "default",
    });

    const evidenceConnectionId = getMeasureEvidenceConnectionId();

    uploadMeasureEvidence({
      variables: {
        input: {
          measureId: data.measure.id,
          file: null,
        },
        connections: evidenceConnectionId ? [evidenceConnectionId] : [],
      },
      uploadables: {
        "input.file": file,
      },
      onCompleted: () => {
        toast({
          title: "Document added",
          description: "Document evidence has been added successfully.",
          variant: "default",
        });
      },
      onError: (error) => {
        toast({
          title: "Error adding document",
          description: error.message,
          variant: "destructive",
        });
      },
    });
  };

  // Add a function to create evidence at the measure level
  const handleCreateMeasureEvidence = () => {
    setEvidenceDialogOpen(true);
    setActiveTab("file");
    // Reset form fields
    setLinkEvidenceName("");
    setLinkEvidenceUrl("");
    setLinkEvidenceDescription("");
  };

  // Add uploadTaskEvidence mutation hook
  const [uploadTaskEvidence] = useMutation<any>(uploadTaskEvidenceMutation);

  return (
    <PageTemplate
      title={data.measure?.name || "Measure"}
      actions={
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleEditMeasure}>
            Edit
          </Button>
          <select
            value={data.measure?.state || ""}
            onChange={(e) => handleMeasureStateChange(e.target.value)}
            className="rounded-full cursor-pointer inline-flex items-center justify-center gap-2 whitespace-nowrap text-sm font-medium transition-colors focus-visible:outline-hidden focus-visible:ring-1 focus-visible:ring-active-b disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 border border-low-b hover:bg-h-tertiary-bg active:bg-p-tertiary-bg focus:bg-tertiary-bg shadow-xs px-2"
          >
            <option value="">Select state</option>
            <option value="NOT_STARTED">Not Started</option>
            <option value="IN_PROGRESS">In Progress</option>
            <option value="NOT_APPLICABLE">Not Applicable</option>
            <option value="IMPLEMENTED">Implemented</option>
          </select>
          <Button
            variant="destructive"
            onClick={() => handleDeleteMeasure()}
            disabled={isDeletingMeasure}
          >
            {isDeletingMeasure ? "Deleting..." : "Delete"}
          </Button>
        </div>
      }
    >
      <div className="space-y-4 mb-8">
        <Card className="mt-4">
          <CardContent className="pt-6">
            <div className="prose prose-gray prose-sm md:prose-base text-secondary max-w-3xl">
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                rehypePlugins={[rehypeRaw]}
              >
                {data.measure.description}
              </ReactMarkdown>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content Tabs */}
      <Tabs
        defaultValue="evidence"
        value={mainContentTab}
        onValueChange={setMainContentTab}
        className="mb-4"
      >
        <TabsList className="border-b w-full p-0 h-auto">
          <TabsTrigger value="evidence" className="flex items-center gap-2">
            <FileText className="w-4 h-4" />
            Evidence
            {((measureWithEvidences.evidences?.edges &&
              measureWithEvidences.evidences.edges.length > 0) ||
              tasks.some(
                (task) =>
                  task.evidences?.edges && task.evidences.edges.length > 0
              )) && (
              <span className="ml-1.5 bg-blue-100 text-blue-800 rounded-full text-xs px-2 py-0.5">
                {(measureWithEvidences.evidences?.edges?.length || 0) +
                  tasks.reduce(
                    (count, task) =>
                      count + (task.evidences?.edges?.length || 0),
                    0
                  )}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="tasks" className="flex items-center gap-2">
            <Check className="w-4 h-4" />
            Tasks
            {tasks.length > 0 && (
              <span className="ml-1.5 bg-blue-100 text-blue-800 rounded-full text-xs px-2 py-0.5">
                {tasks.length}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="controls" className="flex items-center gap-2">
            <ShieldCheck className="w-4 h-4" />
            Controls
            {linkedControlsData?.measure?.controls?.edges &&
              linkedControlsData.measure.controls.edges.length > 0 && (
                <span className="ml-1.5 bg-blue-100 text-blue-800 rounded-full text-xs px-2 py-0.5">
                  {linkedControlsData.measure.controls.edges.length}
                </span>
              )}
          </TabsTrigger>
          <TabsTrigger value="risks" className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4" />
            Risks
            {risksData?.measure?.risks?.edges &&
              risksData.measure.risks.edges.length > 0 && (
                <span className="ml-1.5 bg-blue-100 text-blue-800 rounded-full text-xs px-2 py-0.5">
                  {risksData.measure.risks.edges.length}
                </span>
              )}
          </TabsTrigger>
        </TabsList>

        {/* Evidence Tab Content */}
        <TabsContent value="evidence">
          <div>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold">Evidence</h2>
              <div className="flex items-center gap-3">
                <div className="text-sm text-secondary flex items-center bg-invert-bg px-3 py-1.5 rounded-md border border-mid-b">
                  <FileIcon className="w-4 h-4 mr-2 text-blue-500" />
                  <span>Drag & drop files here to add evidence</span>
                </div>
                <Button
                  size="sm"
                  className="flex items-center gap-1"
                  onClick={handleCreateMeasureEvidence}
                >
                  <Plus className="w-4 h-4" />
                  <span>Add Evidence</span>
                </Button>
              </div>
            </div>

            <div
              className={`space-y-4 p-6 border-2 rounded-md transition-all duration-200 relative ${
                isDraggingOverEvidenceZone
                  ? "border-dashed border-blue-400 bg-blue-50 shadow-md"
                  : "border-dashed border-mid-b hover:border-blue-200"
              }`}
              onDragOver={(e) => {
                e.preventDefault();
                e.stopPropagation();
                setIsDraggingOverEvidenceZone(true);
              }}
              onDragLeave={(e) => {
                e.preventDefault();
                e.stopPropagation();
                const rect = e.currentTarget.getBoundingClientRect();
                const x = e.clientX;
                const y = e.clientY;

                // Only set to false if cursor is outside the drop zone
                if (
                  x < rect.left ||
                  x > rect.right ||
                  y < rect.top ||
                  y > rect.bottom
                ) {
                  setIsDraggingOverEvidenceZone(false);
                }
              }}
              onDrop={(e) => {
                e.preventDefault();
                e.stopPropagation();
                setIsDraggingOverEvidenceZone(false);
                setIsDraggingFile(false);

                const files = e.dataTransfer.files;
                if (files.length === 0) return;

                const file = files[0];
                handleUploadMeasureEvidence(file);
              }}
            >
              {isDraggingOverEvidenceZone && (
                <div className="absolute inset-0 flex items-center justify-center bg-blue-50 bg-opacity-80 rounded-md z-10">
                  <div className="bg-white p-6 rounded-lg shadow-lg text-center">
                    <FileText className="w-12 h-12 text-blue-500 mx-auto mb-2" />
                    <p className="text-lg font-medium">
                      Drop file to add evidence
                    </p>
                  </div>
                </div>
              )}

              {measureWithEvidences.evidences?.edges &&
              measureWithEvidences.evidences.edges.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full border-collapse">
                    <thead className="bg-invert-bg">
                      <tr className="border-b">
                        <th className="text-left font-medium text-sm py-3 px-4">
                          File
                        </th>
                        <th className="text-left font-medium text-sm py-3 px-4">
                          Type
                        </th>
                        <th className="text-left font-medium text-sm py-3 px-4">
                          Size
                        </th>
                        <th className="text-left font-medium text-sm py-3 px-4">
                          Created
                        </th>
                        <th className="text-left font-medium text-sm py-3 px-4">
                          Description
                        </th>
                        <th className="text-right font-medium text-sm py-3 px-4">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {measureWithEvidences.evidences.edges.map(
                        (evidenceEdge) => {
                          const evidence = evidenceEdge?.node;
                          if (!evidence) return null;

                          return (
                            <tr
                              key={evidence.id}
                              className="border-b hover:bg-invert-bg transition-colors"
                            >
                              <td className="py-3 px-4">
                                <div className="flex items-center gap-2">
                                  <div className="bg-white p-1.5 rounded-md border border-mid-b">
                                    {getFileIcon(
                                      evidence.mimeType,
                                      evidence.type
                                    )}
                                  </div>
                                  <div className="flex-1 overflow-hidden">
                                    <div className="text-sm font-medium text-primary flex items-center gap-2 truncate">
                                      {evidence.filename}
                                      {evidence.state === "REQUESTED" && (
                                        <Badge
                                          variant="outline"
                                          className="text-xs bg-yellow-50 text-yellow-700 border-yellow-200 shrink-0"
                                        >
                                          Requested
                                        </Badge>
                                      )}
                                    </div>
                                  </div>
                                </div>
                              </td>
                              <td className="py-3 px-4 text-sm">
                                {evidence.type === "FILE"
                                  ? evidence.mimeType === "text/uri-list" ||
                                    evidence.mimeType === "text/uri"
                                    ? "Link"
                                    : "Document"
                                  : "Link"}
                              </td>
                              <td className="py-3 px-4 text-sm">
                                {evidence.type === "FILE" &&
                                !(
                                  evidence.mimeType === "text/uri-list" ||
                                  evidence.mimeType === "text/uri"
                                )
                                  ? formatFileSize(evidence.size)
                                  : "-"}
                              </td>
                              <td className="py-3 px-4 text-sm">
                                {formatDate(evidence.createdAt)}
                              </td>
                              <td className="py-3 px-4 text-sm max-w-xs">
                                <div className="truncate">
                                  {evidence.description ||
                                    (evidence.url ? evidence.url : "-")}
                                </div>
                              </td>
                              <td className="py-3 px-4 text-right">
                                <div className="flex items-center justify-end gap-1">
                                  {evidence.state === "REQUESTED" ? (
                                    <button
                                      onClick={() =>
                                        handleFulfillEvidence(
                                          evidence.id,
                                          evidence.filename
                                        )
                                      }
                                      className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                      title="Fulfill Evidence"
                                    >
                                      <Upload className="w-4 h-4 text-green-600" />
                                    </button>
                                  ) : evidence.type === "FILE" ? (
                                    <>
                                      {evidence.mimeType === "text/uri-list" ||
                                      evidence.mimeType === "text/uri" ? (
                                        <button
                                          onClick={() =>
                                            handlePreviewEvidence(evidence)
                                          }
                                          className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                          title="Open Link"
                                        >
                                          <Link2 className="w-4 h-4 text-blue-600" />
                                        </button>
                                      ) : evidence.mimeType.startsWith(
                                          "image/"
                                        ) ? (
                                        <button
                                          onClick={() =>
                                            handlePreviewEvidence(evidence)
                                          }
                                          className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                          title="Preview Image"
                                        >
                                          <Eye className="w-4 h-4 text-blue-600" />
                                        </button>
                                      ) : (
                                        <button
                                          onClick={() =>
                                            handlePreviewEvidence(evidence)
                                          }
                                          className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                          title="Download"
                                        >
                                          <Download className="w-4 h-4 text-blue-600" />
                                        </button>
                                      )}
                                    </>
                                  ) : evidence.url ? (
                                    <button
                                      onClick={() => {
                                        if (evidence.url) {
                                          window.open(evidence.url, "_blank");
                                        }
                                      }}
                                      className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                      title="Open Link"
                                    >
                                      <Link2 className="w-4 h-4 text-blue-600" />
                                    </button>
                                  ) : null}
                                  <button
                                    onClick={() =>
                                      handleDeleteEvidence(
                                        evidence.id,
                                        evidence.filename,
                                        ""
                                      )
                                    }
                                    className="p-1.5 rounded-full hover:bg-red-50 hover:shadow-sm transition-all"
                                    title="Delete"
                                  >
                                    <Trash2 className="w-4 h-4 text-red-500" />
                                  </button>
                                </div>
                              </td>
                            </tr>
                          );
                        }
                      )}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center py-12">
                  <FileIcon className="w-12 h-12 text-tertiary mx-auto mb-3" />
                  <p className="text-lg font-medium">No evidence added yet</p>
                  <p className="text-secondary mt-1">
                    Drag and drop files here or use the "Add Evidence" button
                  </p>
                </div>
              )}
            </div>
          </div>
        </TabsContent>

        {/* Tasks Tab Content */}
        <TabsContent value="tasks">
          <div>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold">Tasks</h2>
              <Button
                size="sm"
                className="flex items-center gap-1"
                onClick={() => setIsCreateTaskOpen(true)}
              >
                <Plus className="w-4 h-4" />
                <span>Add Task</span>
              </Button>
            </div>

            <div className="space-y-3">
              {tasks.length > 0 ? (
                tasks.map((task) => (
                  <div
                    key={task.id}
                    className="p-4 border rounded-md hover:shadow-sm transition-shadow cursor-pointer"
                    onClick={() => handleTaskClick(task)}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div
                          className="rounded-full w-5 h-5 border flex items-center justify-center cursor-pointer"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleToggleTaskState(
                              task.id,
                              task.state || "TODO"
                            );
                          }}
                        >
                          {task.state === "DONE" && (
                            <Check className="w-3 h-3 text-primary" />
                          )}
                        </div>
                        <h3
                          className={`font-medium ${
                            task.state === "DONE"
                              ? "line-through text-secondary"
                              : ""
                          }`}
                        >
                          {task.name}
                        </h3>
                        <Badge
                          variant="outline"
                          className={`text-xs ${getStateColor(task.state)}`}
                        >
                          {formatState(task.state)}
                        </Badge>
                      </div>

                      <div className="flex items-center gap-2">
                        {task.assignedTo && (
                          <div className="flex items-center gap-1 text-sm text-secondary bg-invert-bg px-2 py-1 rounded">
                            <User className="w-3 h-3" />
                            <span>{task.assignedTo.fullName}</span>
                          </div>
                        )}
                        {task.timeEstimate && (
                          <div className="text-sm text-secondary">
                            {formatDuration(task.timeEstimate)}
                          </div>
                        )}
                        {task.deadline && (
                          <div className="text-sm text-secondary">
                            {formatDate(task.deadline)} 
                          </div>
                        )}
                      </div>
                    </div>

                    {task.description && (
                      <div className="mt-2 text-sm text-secondary">
                        {task.description.length > 150
                          ? `${task.description.substring(0, 150)}...`
                          : task.description}
                      </div>
                    )}

                    {task.evidences?.edges &&
                      task.evidences.edges.length > 0 && (
                        <div className="mt-3 flex items-center gap-1">
                          <FileText className="w-4 h-4 text-blue-500" />
                          <span className="text-sm text-secondary">
                            {task.evidences.edges.length} evidence file(s)
                          </span>
                        </div>
                      )}
                  </div>
                ))
              ) : (
                <div className="text-center py-12 border-2 border-dashed rounded-md">
                  <Check className="w-12 h-12 text-tertiary mx-auto mb-3" />
                  <p className="text-lg font-medium">No tasks added yet</p>
                  <p className="text-secondary mt-1">
                    Add tasks to track the implementation of this measure
                  </p>
                </div>
              )}
            </div>
          </div>
        </TabsContent>

        {/* Controls Tab Content */}
        <TabsContent value="controls">
          <div>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold">Controls</h2>
              <Button
                size="sm"
                className="flex items-center gap-1"
                onClick={handleOpenControlMappingDialog}
              >
                <Plus className="w-4 h-4" />
                <span>Link Control</span>
              </Button>
            </div>

            <div className="space-y-3">
              {linkedControlsData?.measure?.controls?.edges &&
              linkedControlsData.measure.controls.edges.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full border-collapse">
                    <thead className="bg-invert-bg">
                      <tr className="border-b">
                        <th className="text-left font-medium text-sm py-3 px-4">
                          ID
                        </th>
                        <th className="text-left font-medium text-sm py-3 px-4">
                          Name
                        </th>
                        <th className="text-right font-medium text-sm py-3 px-4">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {linkedControlsData.measure.controls.edges.map((edge) => {
                        const control = edge.node;
                        return (
                          <tr
                            key={control.id}
                            className="border-b hover:bg-invert-bg transition-colors"
                          >
                            <td className="py-3 px-4">
                              <div className="flex items-center gap-2">
                                <ShieldCheck className="w-4 h-4 text-green-600" />
                                <span className="font-medium">
                                  {control.referenceId}
                                </span>
                              </div>
                            </td>
                            <td className="py-3 px-4">{control.name}</td>
                            <td className="py-3 px-4 text-right">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleUnlinkControl(control.id)}
                                className="text-red-500 hover:text-red-700 hover:bg-red-50 border-red-200"
                              >
                                <X className="w-4 h-4 mr-1" />
                                Unlink
                              </Button>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center py-12 border-2 border-dashed rounded-md">
                  <ShieldCheck className="w-12 h-12 text-tertiary mx-auto mb-3" />
                  <p className="text-lg font-medium">No controls linked yet</p>
                  <p className="text-secondary mt-1">
                    Link controls from your frameworks to this measure
                  </p>
                </div>
              )}
            </div>
          </div>
        </TabsContent>

        {/* Risks Tab Content */}
        <TabsContent value="risks">
          <div>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold">Linked Risks</h2>
            </div>

            <div className="space-y-3">
              {risksData?.measure?.risks?.edges &&
              risksData.measure.risks.edges.length > 0 ? (
                risksData.measure.risks.edges.map((edge) => {
                  const risk = edge.node;
                  return (
                    <div
                      key={risk.id}
                      className="p-4 border rounded-md hover:shadow-sm transition-shadow"
                    >
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-2">
                          <AlertTriangle className="w-4 h-4 text-amber-500" />
                          <h3 className="font-medium">{risk.name}</h3>
                        </div>
                        <div className="flex items-center gap-2">
                          <Badge
                            variant="outline"
                            className={getRiskSeverityColor(
                              risk.inherentRiskScore
                            )}
                          >
                            Inherent:{" "}
                            {getRiskSeverityText(risk.inherentRiskScore)}
                          </Badge>
                          <Badge
                            variant="outline"
                            className={getRiskSeverityColor(
                              risk.residualRiskScore
                            )}
                          >
                            Residual:{" "}
                            {getRiskSeverityText(risk.residualRiskScore)}
                          </Badge>
                        </div>
                      </div>

                      {risk.description && (
                        <p className="text-sm text-secondary mt-2">
                          {risk.description.length > 200
                            ? `${risk.description.substring(0, 200)}...`
                            : risk.description}
                        </p>
                      )}

                      <div className="grid grid-cols-2 gap-4 mt-3 text-sm">
                        <div>
                          <p className="text-secondary">Inherent risk:</p>
                          <p>
                            Likelihood:{" "}
                            {formatLikelihood(risk.inherentLikelihood)}
                          </p>
                          <p>Impact: {formatImpact(risk.inherentImpact)}</p>
                        </div>
                        <div>
                          <p className="text-secondary">Residual risk:</p>
                          <p>
                            Likelihood:{" "}
                            {formatLikelihood(risk.residualLikelihood)}
                          </p>
                          <p>Impact: {formatImpact(risk.residualImpact)}</p>
                        </div>
                      </div>
                    </div>
                  );
                })
              ) : (
                <div className="text-center py-12 border-2 border-dashed rounded-md">
                  <AlertTriangle className="w-12 h-12 text-tertiary mx-auto mb-3" />
                  <p className="text-lg font-medium">
                    No risks linked to this measure
                  </p>
                  <p className="text-secondary mt-1">
                    This measure is not currently linked to any risks
                  </p>
                </div>
              )}
            </div>
          </div>
        </TabsContent>
      </Tabs>

      {/* Right task panel */}
      <Sheet open={isTaskPanelOpen} onOpenChange={handleCloseTaskPanel}>
        <SheetContent
          side="right"
          className="!max-w-[50vw] !w-[50vw] p-0 overflow-y-auto"
        >
          {selectedTask && (
            <div className="flex flex-col h-full">
              <SheetHeader className="px-6 py-4 border-b sticky top-0 bg-white z-10">
                <div className="flex items-center justify-between">
                  <SheetTitle className="text-xl font-semibold">
                    {selectedTask.name}
                  </SheetTitle>
                  <SheetClose className="rounded-full p-1 hover:bg-secondary-bg">
                    <X className="h-5 w-5" />
                  </SheetClose>
                </div>
              </SheetHeader>

              <div className="flex-1 overflow-auto p-6 space-y-6">
                {/* Task content here */}
                <div className="space-y-6">
                  {/* Task status */}
                  <div>
                    <h3 className="text-sm font-medium text-secondary mb-2">
                      Status
                    </h3>
                    <Badge
                      variant="outline"
                      className={`${getStateColor(selectedTask.state)}`}
                    >
                      {formatState(selectedTask.state)}
                    </Badge>
                  </div>

                  {/* Task description */}
                  {selectedTask.description && (
                    <div>
                      <h3 className="text-sm font-medium text-secondary mb-2">
                        Description
                      </h3>
                      <div className="prose prose-gray prose-sm max-w-none text-primary bg-invert-bg p-4 rounded-md border border-mid-b">
                        <ReactMarkdown
                          remarkPlugins={[remarkGfm]}
                          rehypePlugins={[rehypeRaw]}
                        >
                          {selectedTask.description}
                        </ReactMarkdown>
                      </div>
                    </div>
                  )}

                  {/* Assigned person */}
                  <div>
                    <h3 className="text-sm font-medium text-secondary mb-2">
                      Assigned To
                    </h3>
                    {selectedTask.assignedTo ? (
                      <div className="flex items-center gap-2">
                        <div className="bg-blue-100 text-blue-600 rounded-full p-1">
                          <User className="w-4 h-4" />
                        </div>
                        <span>{selectedTask.assignedTo.fullName}</span>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-red-500 p-1 h-auto"
                          onClick={() => handleUnassignPerson(selectedTask.id)}
                        >
                          <UserMinus className="w-4 h-4" />
                        </Button>
                      </div>
                    ) : (
                      <Popover
                        open={peoplePopoverOpen[selectedTask.id] || false}
                        onOpenChange={(open) =>
                          setPeoplePopoverOpen({
                            ...peoplePopoverOpen,
                            [selectedTask.id]: open,
                          })
                        }
                      >
                        <PopoverTrigger asChild>
                          <Button
                            variant="outline"
                            size="sm"
                            className="flex items-center gap-1"
                          >
                            <UserPlus className="w-4 h-4" />
                            <span>Assign Person</span>
                          </Button>
                        </PopoverTrigger>
                        <PopoverContent className="w-80">
                          <div className="space-y-2">
                            <div className="relative">
                              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-tertiary" />
                              <Input
                                placeholder="Search people..."
                                value={peopleSearch[selectedTask.id] || ""}
                                onChange={(e) =>
                                  setPeopleSearch({
                                    ...peopleSearch,
                                    [selectedTask.id]: e.target.value,
                                  })
                                }
                                className="pl-9"
                              />
                            </div>
                            <div className="max-h-52 overflow-y-auto">
                              {organizationData?.organization?.peoples?.edges
                                ?.filter((edge) => {
                                  const search = (
                                    peopleSearch[selectedTask.id] || ""
                                  ).toLowerCase();
                                  if (!search) return true;
                                  const person = edge?.node;
                                  return (
                                    person?.fullName
                                      ?.toLowerCase()
                                      .includes(search) ||
                                    person?.primaryEmailAddress
                                      ?.toLowerCase()
                                      .includes(search)
                                  );
                                })
                                .map((edge) => {
                                  const person = edge?.node;
                                  if (!person) return null;
                                  return (
                                    <div
                                      key={person.id}
                                      className="flex items-center justify-between p-2 hover:bg-invert-bg cursor-pointer rounded-md"
                                      onClick={() =>
                                        handleAssignPerson(
                                          selectedTask.id,
                                          person.id
                                        )
                                      }
                                    >
                                      <div className="flex items-center gap-2">
                                        <User className="w-4 h-4 text-tertiary" />
                                        <span>{person.fullName}</span>
                                      </div>
                                      <Button
                                        variant="ghost"
                                        size="sm"
                                        className="h-7 px-2"
                                      >
                                        <UserPlus className="w-3.5 h-3.5" />
                                      </Button>
                                    </div>
                                  );
                                })}
                            </div>
                          </div>
                        </PopoverContent>
                      </Popover>
                    )}
                  </div>

                  {/* Time estimate */}
                  <div>
                    <h3 className="text-sm font-medium text-secondary mb-2">
                      Time Estimate
                    </h3>
                    {isEditingDuration ? (
                      <div className="space-y-2">
                        <div className="flex gap-2">
                          <div className="flex-1">
                            <Input
                              type="number"
                              min="0"
                              placeholder="Days"
                              value={editTimeEstimateDays}
                              onChange={(e) =>
                                setEditTimeEstimateDays(e.target.value)
                              }
                            />
                          </div>
                          <div className="flex-1">
                            <Input
                              type="number"
                              min="0"
                              max="23"
                              placeholder="Hours"
                              value={editTimeEstimateHours}
                              onChange={(e) =>
                                setEditTimeEstimateHours(e.target.value)
                              }
                            />
                          </div>
                          <div className="flex-1">
                            <Input
                              type="number"
                              min="0"
                              max="59"
                              placeholder="Minutes"
                              value={editTimeEstimateMinutes}
                              onChange={(e) =>
                                setEditTimeEstimateMinutes(e.target.value)
                              }
                            />
                          </div>
                        </div>
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setIsEditingDuration(false)}
                          >
                            Cancel
                          </Button>
                          <Button
                            size="sm"
                            onClick={() => handleSaveDuration(selectedTask.id)}
                          >
                            Save
                          </Button>
                        </div>
                      </div>
                    ) : (
                      <div className="flex items-center justify-between">
                        <div>
                          {selectedTask.timeEstimate ? (
                            <span>
                              {formatDuration(selectedTask.timeEstimate)}
                            </span>
                          ) : (
                            <span className="text-secondary">
                              No time estimate
                            </span>
                          )}
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="p-1 h-auto"
                          onClick={() => {
                            // Parse current duration into components
                            const { days, hours, minutes } = parseISODuration(
                              selectedTask.timeEstimate
                            );
                            setEditTimeEstimateDays(days);
                            setEditTimeEstimateHours(hours);
                            setEditTimeEstimateMinutes(minutes);
                            setIsEditingDuration(true);
                          }}
                        >
                          Edit
                        </Button>
                      </div>
                    )}
                  </div>

                  {/* Deadline */}
                  <div>
                    <h3 className="text-sm font-medium text-secondary mb-2">
                      Deadline
                    </h3>
                    {isEditingDeadline ? (
                      <div className="space-y-2">
                        <div className="flex gap-2">
                          <div className="flex-1">
                            <Input
                              type="date"
                              value={editDeadline}
                              onChange={(e) =>
                                setEditDeadline(e.target.value)
                              }
                            />
                          </div>                        
                        </div>
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setIsEditingDeadline(false)}
                          >
                            Cancel
                          </Button>
                          <Button
                            size="sm"
                            onClick={() => handleSaveDeadline(selectedTask.id)}
                          >
                            Save
                          </Button>
                        </div>
                      </div>
                    ) : (
                      <div className="flex items-center justify-between">
                        <div>
                          {selectedTask.deadline ? (
                            <span>
                              {formatDate(selectedTask.deadline)}
                            </span>
                          ) : (
                            <span className="text-secondary">
                              No deadline
                            </span>
                          )}
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="p-1 h-auto"
                          onClick={() => {
                            const selectedDeadline = selectedTask.deadline ? format(selectedTask.deadline, 'yyyy-MM-dd') : "";
                            setEditDeadline(selectedDeadline);
                            setIsEditingDeadline(true);
                          }}
                        >
                          Edit
                        </Button>
                      </div>
                    )}
                  </div>

                  {/* Evidence section */}
                  <div>
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="text-sm font-medium text-secondary">
                        Evidence
                      </h3>
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex items-center gap-1"
                        onClick={() =>
                          handleCreateEvidence(
                            selectedTask.id,
                            selectedTask.name
                          )
                        }
                      >
                        <Plus className="w-4 h-4" />
                        <span>Add Evidence</span>
                      </Button>
                    </div>

                    {selectedTask.evidences?.edges &&
                    selectedTask.evidences.edges.length > 0 ? (
                      <div className="space-y-2">
                        {selectedTask.evidences.edges.map((edge) => {
                          const evidence = edge?.node;
                          if (!evidence) return null;

                          return (
                            <div
                              key={evidence.id}
                              className="flex items-center justify-between p-2 rounded-md border border-mid-b hover:bg-invert-bg"
                            >
                              <div className="flex items-center gap-2">
                                <div className="bg-white p-1.5 rounded-md border border-mid-b">
                                  {getFileIcon(
                                    evidence.mimeType,
                                    evidence.type
                                  )}
                                </div>
                                <div>
                                  <div className="text-sm font-medium flex items-center gap-2">
                                    {evidence.filename}
                                    {evidence.state === "REQUESTED" && (
                                      <Badge
                                        variant="outline"
                                        className="text-xs bg-yellow-50 text-yellow-700 border-yellow-200"
                                      >
                                        Requested
                                      </Badge>
                                    )}
                                  </div>
                                  <div className="text-xs text-secondary">
                                    {formatDate(evidence.createdAt)} •{" "}
                                    {formatFileSize(evidence.size)}
                                  </div>
                                </div>
                              </div>

                              <div className="flex items-center gap-1">
                                {evidence.state === "REQUESTED" ? (
                                  <button
                                    onClick={() =>
                                      handleFulfillEvidence(
                                        evidence.id,
                                        evidence.filename
                                      )
                                    }
                                    className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                    title="Fulfill Evidence"
                                  >
                                    <Upload className="w-4 h-4 text-green-600" />
                                  </button>
                                ) : evidence.type === "FILE" ? (
                                  <>
                                    {evidence.mimeType === "text/uri-list" ||
                                    evidence.mimeType === "text/uri" ? (
                                      <button
                                        onClick={() =>
                                          handlePreviewEvidence(evidence)
                                        }
                                        className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                        title="Open Link"
                                      >
                                        <Link2 className="w-4 h-4 text-blue-600" />
                                      </button>
                                    ) : evidence.mimeType.startsWith(
                                        "image/"
                                      ) ? (
                                      <button
                                        onClick={() =>
                                          handlePreviewEvidence(evidence)
                                        }
                                        className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                        title="Preview Image"
                                      >
                                        <Eye className="w-4 h-4 text-blue-600" />
                                      </button>
                                    ) : (
                                      <button
                                        onClick={() =>
                                          handlePreviewEvidence(evidence)
                                        }
                                        className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                        title="Download"
                                      >
                                        <Download className="w-4 h-4 text-blue-600" />
                                      </button>
                                    )}
                                  </>
                                ) : evidence.url ? (
                                  <button
                                    onClick={() => {
                                      if (evidence.url) {
                                        window.open(evidence.url, "_blank");
                                      }
                                    }}
                                    className="p-1.5 rounded-full hover:bg-white hover:shadow-sm transition-all"
                                    title="Open Link"
                                  >
                                    <Link2 className="w-4 h-4 text-blue-600" />
                                  </button>
                                ) : null}
                                <button
                                  onClick={() =>
                                    handleDeleteEvidence(
                                      evidence.id,
                                      evidence.filename,
                                      selectedTask.id
                                    )
                                  }
                                  className="p-1.5 rounded-full hover:bg-red-50 hover:shadow-sm transition-all"
                                  title="Delete"
                                >
                                  <Trash2 className="w-4 h-4 text-red-500" />
                                </button>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    ) : (
                      <div className="text-center py-8 border-2 border-dashed rounded-md">
                        <FileIcon className="w-8 h-8 text-tertiary mx-auto mb-2" />
                        <p className="text-sm text-secondary">
                          No evidence added to this task yet
                        </p>
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* Task actions */}
              <div className="border-t p-4 sticky bottom-0 bg-white">
                <div className="flex gap-3 justify-end">
                  {selectedTask && selectedTask.state !== "DONE" ? (
                    <Button
                      variant="default"
                      className="flex items-center gap-1"
                      onClick={() =>
                        handleToggleTaskState(
                          selectedTask.id,
                          selectedTask.state || "TODO"
                        )
                      }
                    >
                      <CheckCircle2 className="w-4 h-4" />
                      <span>Mark Complete</span>
                    </Button>
                  ) : (
                    <Button
                      variant="outline"
                      className="flex items-center gap-1"
                      onClick={() => {
                        if (selectedTask) {
                          handleToggleTaskState(
                            selectedTask.id,
                            selectedTask.state || "DONE"
                          );
                        }
                      }}
                    >
                      <X className="w-4 h-4" />
                      <span>Reopen Task</span>
                    </Button>
                  )}
                  <Button
                    variant="destructive"
                    className="flex items-center gap-1"
                    onClick={() => {
                      if (selectedTask) {
                        handleDeleteTask(selectedTask.id, selectedTask.name);
                        handleCloseTaskPanel();
                      }
                    }}
                  >
                    <Trash2 className="w-4 h-4" />
                    <span>Delete Task</span>
                  </Button>
                </div>
              </div>
            </div>
          )}
        </SheetContent>
      </Sheet>

      {/* Evidence Add Dialog */}
      <Dialog open={evidenceDialogOpen} onOpenChange={setEvidenceDialogOpen}>
        <DialogContent className="sm:max-w-[525px]">
          <DialogHeader>
            <DialogTitle>Add Evidence</DialogTitle>
            <DialogDescription>
              Choose the type of evidence you want to add to this task.
            </DialogDescription>
          </DialogHeader>

          <Tabs
            value={activeTab}
            onValueChange={(value: string) =>
              setActiveTab(value as "file" | "link")
            }
          >
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="file">Document</TabsTrigger>
              <TabsTrigger value="link">Link</TabsTrigger>
            </TabsList>

            <TabsContent value="file" className="space-y-4">
              <div className="space-y-4 pt-4">
                <p>Select a document to add as evidence.</p>
                <Button
                  onClick={() => {
                    if (hiddenFileInputRef.current) {
                      hiddenFileInputRef.current.click();
                      setEvidenceDialogOpen(false);
                    }
                  }}
                  className="w-full"
                >
                  Select Document
                </Button>
              </div>
            </TabsContent>

            <TabsContent value="link" className="space-y-4 pt-4">
              <div className="space-y-4">
                <div className="grid w-full items-center gap-1.5">
                  <Label htmlFor="evidence-name">Name</Label>
                  <Input
                    id="evidence-name"
                    value={linkEvidenceName}
                    onChange={(e) => setLinkEvidenceName(e.target.value)}
                    placeholder="Link name"
                    type="text"
                  />
                </div>

                <div className="grid w-full items-center gap-1.5">
                  <Label htmlFor="evidence-url">URL</Label>
                  <Input
                    id="evidence-url"
                    value={linkEvidenceUrl}
                    onChange={(e) => setLinkEvidenceUrl(e.target.value)}
                    placeholder="https://example.com"
                    type="url"
                  />
                  <p className="text-xs text-secondary">
                    This will create a .uri file with the URL inside
                  </p>
                </div>

                <div className="grid w-full items-center gap-1.5">
                  <Label htmlFor="evidence-description">
                    Description (optional)
                  </Label>
                  <Textarea
                    id="evidence-description"
                    value={linkEvidenceDescription}
                    onChange={(e) => setLinkEvidenceDescription(e.target.value)}
                    placeholder="Describe this evidence (optional)"
                    className="min-h-[100px]"
                  />
                </div>
              </div>
            </TabsContent>
          </Tabs>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setEvidenceDialogOpen(false)}
            >
              Cancel
            </Button>
            {activeTab === "link" && (
              <Button onClick={handleLinkEvidenceSubmit}>
                Create URI File
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Hidden file input for direct uploads */}
      <input
        type="file"
        ref={hiddenFileInputRef}
        onChange={handleFileSelected}
        style={{ display: "none" }}
        accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.txt"
      />

      {/* Delete Task Confirmation Dialog */}
      <Dialog open={isDeleteTaskOpen} onOpenChange={setIsDeleteTaskOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Task</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete the task &quot;
              {taskToDelete?.name}&quot;? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteTaskOpen(false)}
            >
              Cancel
            </Button>
            <Button variant="destructive" onClick={confirmDeleteTask}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add the Preview Modal */}
      <Dialog open={isPreviewModalOpen} onOpenChange={setIsPreviewModalOpen}>
        <DialogContent className="sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle className="flex items-center justify-between">
              <span>{previewEvidence?.filename}</span>
            </DialogTitle>
            <DialogDescription>
              Preview of the evidence file. You can view or download the file
              from here.
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col items-center justify-center min-h-[300px] bg-invert-bg rounded-md p-4">
            {isLoadingFileUrl ? (
              <div className="flex flex-col items-center gap-2">
                <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
                <p className="text-secondary">Loading preview...</p>
              </div>
            ) : previewEvidence?.fileUrl ? (
              previewEvidence.mimeType.startsWith("image/") ? (
                <img
                  src={previewEvidence.fileUrl}
                  alt={previewEvidence.filename}
                  className="max-h-[70vh] object-contain"
                />
              ) : previewEvidence.mimeType.includes("pdf") ? (
                <iframe
                  src={previewEvidence.fileUrl}
                  className="w-full h-[70vh]"
                  title={previewEvidence.filename}
                />
              ) : (
                <div className="flex flex-col items-center gap-4">
                  <FileGeneric className="w-16 h-16 text-tertiary" />
                  <p className="text-secondary">
                    Preview not available for this file type
                  </p>
                  <Button
                    onClick={() => {
                      if (previewEvidence?.fileUrl) {
                        window.open(previewEvidence.fileUrl, "_blank");
                      }
                    }}
                  >
                    Download File
                  </Button>
                </div>
              )
            ) : (
              <div className="flex flex-col items-center gap-2">
                <FileGeneric className="w-12 h-12 text-tertiary" />
                <p className="text-secondary">Failed to load preview</p>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsPreviewModalOpen(false)}
            >
              Close
            </Button>
            {previewEvidence?.fileUrl && (
              <Button
                onClick={() => {
                  if (previewEvidence?.fileUrl) {
                    window.open(previewEvidence.fileUrl, "_blank");
                  }
                }}
              >
                Download
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Delete Evidence Confirmation Dialog */}
      <Dialog
        open={isDeleteEvidenceOpen}
        onOpenChange={setIsDeleteEvidenceOpen}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Evidence</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete the evidence &quot;
              {evidenceToDelete?.filename}&quot;? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteEvidenceOpen(false)}
            >
              Cancel
            </Button>
            <Button variant="destructive" onClick={confirmDeleteEvidence}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Fulfill Evidence Dialog */}
      <Dialog
        open={fulfillEvidenceDialogOpen}
        onOpenChange={setFulfillEvidenceDialogOpen}
      >
        <DialogContent className="sm:max-w-[525px]">
          <DialogHeader>
            <DialogTitle>Fulfill Evidence</DialogTitle>
            <DialogDescription>
              Provide the requested evidence: {evidenceToFulfill?.filename}
            </DialogDescription>
          </DialogHeader>

          <Tabs
            value={activeTab}
            onValueChange={(value: string) =>
              setActiveTab(value as "file" | "link")
            }
          >
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="file">Document</TabsTrigger>
              <TabsTrigger value="link">Link</TabsTrigger>
            </TabsList>

            <TabsContent value="file" className="space-y-4">
              <div className="space-y-4 pt-4">
                <p>Upload a document to fulfill this evidence request.</p>
                <Button
                  onClick={() => {
                    if (hiddenFileInputRef.current) {
                      // Set a temp onchange handler for the file input
                      const originalOnChange =
                        hiddenFileInputRef.current.onchange;
                      hiddenFileInputRef.current.onchange = (e) => {
                        hiddenFileInputRef.current!.onchange = originalOnChange;
                        handleFulfillEvidenceWithFile({
                          target: {
                            files: (e.target as HTMLInputElement)?.files,
                          },
                        } as React.ChangeEvent<HTMLInputElement>);
                      };
                      hiddenFileInputRef.current.click();
                      setFulfillEvidenceDialogOpen(false);
                    }
                  }}
                  className="w-full"
                >
                  Select Document
                </Button>
              </div>
            </TabsContent>

            <TabsContent value="link" className="space-y-4 pt-4">
              <div className="space-y-4">
                <div className="grid w-full items-center gap-1.5">
                  <Label htmlFor="fulfill-evidence-name">Name</Label>
                  <Input
                    id="fulfill-evidence-name"
                    value={linkEvidenceName}
                    onChange={(e) => setLinkEvidenceName(e.target.value)}
                    placeholder="Link name (optional)"
                    type="text"
                  />
                  <p className="text-xs text-secondary">
                    Optional - will use evidence name if not provided
                  </p>
                </div>

                <div className="grid w-full items-center gap-1.5">
                  <Label htmlFor="fulfill-evidence-url">URL</Label>
                  <Input
                    id="fulfill-evidence-url"
                    value={linkEvidenceUrl}
                    onChange={(e) => setLinkEvidenceUrl(e.target.value)}
                    placeholder="https://example.com"
                    type="url"
                  />
                  <p className="text-xs text-secondary">
                    This will create a .uri file with the URL inside
                  </p>
                </div>
              </div>
            </TabsContent>
          </Tabs>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setFulfillEvidenceDialogOpen(false)}
            >
              Cancel
            </Button>
            {activeTab === "link" && (
              <Button onClick={handleFulfillEvidenceWithLink}>
                Create URI File
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Measure Confirmation Dialog */}
      <Dialog open={isDeleteMeasureOpen} onOpenChange={setIsDeleteMeasureOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Measure</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this measure? This action cannot
              be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteMeasureOpen(false)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDeleteMeasure}
              disabled={isDeletingMeasure}
            >
              {isDeletingMeasure ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Control Mapping Dialog */}
      <Dialog
        open={isControlMappingDialogOpen}
        onOpenChange={setIsControlMappingDialogOpen}
      >
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle>Link Controls to Measure</DialogTitle>
            <DialogDescription>
              Search and select controls to link to this measure.
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col flex-1 overflow-hidden mt-4">
            {/* Framework Selector and Search */}
            <div className="flex gap-4 mb-4">
              <div className="flex-1">
                <Select
                  value={selectedFrameworkId || "all"}
                  onValueChange={(value) =>
                    setSelectedFrameworkId(value !== "all" ? value : null)
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a framework" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Frameworks</SelectItem>
                    {frameworksData?.organization?.frameworks?.edges.map(
                      (edge) => (
                        <SelectItem key={edge.node.id} value={edge.node.id}>
                          {edge.node.name}
                        </SelectItem>
                      )
                    )}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex-1">
                <div className="relative">
                  <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-tertiary" />
                  <Input
                    placeholder="Search controls..."
                    value={controlSearchQuery}
                    onChange={(e) => setControlSearchQuery(e.target.value)}
                    className="pl-9"
                  />
                </div>
              </div>
            </div>

            {/* Controls Table */}
            <div className="flex-1 overflow-y-auto border rounded-md">
              {isLoadingControls ? (
                <div className="flex items-center justify-center h-40">
                  <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
                </div>
              ) : filteredControls().length > 0 ? (
                <table className="w-full border-collapse">
                  <thead className="bg-invert-bg">
                    <tr className="border-b">
                      <th className="text-left font-medium text-sm py-3 px-4">
                        ID
                      </th>
                      <th className="text-left font-medium text-sm py-3 px-4">
                        Name
                      </th>
                      <th className="text-right font-medium text-sm py-3 px-4">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredControls().map((control) => {
                      const isLinked = isControlLinked(control.id);
                      return (
                        <tr
                          key={control.id}
                          className="border-b hover:bg-invert-bg transition-colors"
                        >
                          <td className="py-3 px-4">
                            <div className="flex items-center gap-2">
                              <ShieldCheck className="w-4 h-4 text-green-600" />
                              <span className="font-medium">
                                {control.referenceId}
                              </span>
                            </div>
                          </td>
                          <td className="py-3 px-4">{control.name}</td>
                          <td className="py-3 px-4 text-right">
                            {isLinked ? (
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleUnlinkControl(control.id)}
                                disabled={isUnlinkingControl}
                                className="text-red-500 hover:text-red-700 hover:bg-red-50 border-red-200"
                              >
                                {isUnlinkingControl ? (
                                  <Loader2 className="w-4 h-4 animate-spin" />
                                ) : (
                                  <X className="w-4 h-4 mr-1" />
                                )}
                                <span>Unlink</span>
                              </Button>
                            ) : (
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleLinkControl(control.id)}
                                disabled={isLinkingControl}
                                className="text-blue-500 hover:text-blue-700 hover:bg-blue-50 border-blue-200"
                              >
                                {isLinkingControl ? (
                                  <Loader2 className="w-4 h-4 animate-spin" />
                                ) : (
                                  <Link2 className="w-4 h-4 mr-1" />
                                )}
                                <span>Link</span>
                              </Button>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              ) : (
                <div className="text-center py-12">
                  <ShieldCheck className="w-12 h-12 text-tertiary mx-auto mb-3" />
                  <p className="text-lg font-medium">No controls found</p>
                  <p className="text-secondary mt-1">
                    Try adjusting your search or framework filter
                  </p>
                </div>
              )}
            </div>
          </div>

          <DialogFooter className="mt-4">
            <Button onClick={() => setIsControlMappingDialogOpen(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create Task Dialog */}
      <Dialog open={isCreateTaskOpen} onOpenChange={setIsCreateTaskOpen}>
        <DialogContent className="sm:max-w-[525px]">
          <DialogHeader>
            <DialogTitle>Create New Task</DialogTitle>
            <DialogDescription>
              Add a new task to this measure.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="grid w-full items-center gap-1.5">
              <Label htmlFor="task-name">Task Name</Label>
              <Input
                id="task-name"
                value={newTaskName}
                onChange={(e) => setNewTaskName(e.target.value)}
                placeholder="Enter task name"
              />
            </div>

            <div className="grid w-full items-center gap-1.5">
              <Label htmlFor="task-description">Description (optional)</Label>
              <Textarea
                id="task-description"
                value={newTaskDescription}
                onChange={(e) => setNewTaskDescription(e.target.value)}
                placeholder="Describe what needs to be done"
                className="min-h-[100px]"
              />
            </div>

            <div>
              <Label>Time Estimate (optional)</Label>
              <div className="flex gap-2 mt-1.5">
                <div className="flex-1">
                  <Input
                    type="number"
                    min="0"
                    placeholder="Days"
                    value={timeEstimateDays}
                    onChange={(e) => setTimeEstimateDays(e.target.value)}
                  />
                </div>
                <div className="flex-1">
                  <Input
                    type="number"
                    min="0"
                    max="23"
                    placeholder="Hours"
                    value={timeEstimateHours}
                    onChange={(e) => setTimeEstimateHours(e.target.value)}
                  />
                </div>
                <div className="flex-1">
                  <Input
                    type="number"
                    min="0"
                    max="59"
                    placeholder="Minutes"
                    value={timeEstimateMinutes}
                    onChange={(e) => setTimeEstimateMinutes(e.target.value)}
                  />
                </div>
              </div>
            </div>
            <div className="grid w-full items-center gap-1.5">
              <Label htmlFor="task-deadline">Deadline (optional)</Label>
              <Input
                id="task-deadline"
                type="date"
                value={newDeadline}
                onChange={(e) => setNewDeadline(e.target.value)}
              />
            </div>
          </div>

          <DialogFooter className="mt-4">
            <Button
              variant="outline"
              onClick={() => setIsCreateTaskOpen(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleCreateTask}>Create Task</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageTemplate>
  );
}

export default function MeasureView() {
  const { measureId } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<MeasureViewQueryType>(measureViewQuery);

  useEffect(() => {
    if (measureId) {
      loadQuery({ measureId });
    }
  }, [measureId, loadQuery]);

  if (!queryRef) {
    return <MeasureViewSkeleton />;
  }

  return (
    <Suspense fallback={<MeasureViewSkeleton />}>
      <MeasureViewContent queryRef={queryRef} />
    </Suspense>
  );
}
