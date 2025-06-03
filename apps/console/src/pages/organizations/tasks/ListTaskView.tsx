import {ChangeEvent, Suspense, useEffect, useState} from "react";
import {
  useQueryLoader,
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useMutation,
  useRelayEnvironment,
  fetchQuery,
} from "react-relay";
import { useSearchParams, useParams } from "react-router";
import { PageTemplate } from "@/components/PageTemplate";
import { ListTaskViewSkeleton } from "./ListTaskPage";
import { ListTaskViewQuery, TaskState } from "./__generated__/ListTaskViewQuery.graphql";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Avatar } from "@/components/ui/avatar";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogClose,
  DialogFooter,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  CircleDashed,
  Circle,
  Plus,
  Filter,
  Search,
  CornerDownRight,
  Flame,
  Building,
  X,
  ChevronDown,
  User,
  CheckCircle2,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ListTaskViewOrganizationQuery } from "./__generated__/ListTaskViewOrganizationQuery.graphql";
import {formatDate, parseISO} from "date-fns";

// Function to format ISO8601 duration to human-readable format
const formatDuration = (isoDuration: string | null | undefined): string => {
  if (!isoDuration || !isoDuration.startsWith("P")) {
    return isoDuration || "";
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

// Define organization measures query
const organizationMeasuresQuery = graphql`
  query ListTaskViewOrganizationMeasuresQuery($organizationId: ID!) {
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

// Define the create task mutation
const createTaskMutation = graphql`
  mutation ListTaskViewCreateTaskMutation(
    $input: CreateTaskInput!
    $connections: [ID!]!
  ) {
    createTask(input: $input) {
      taskEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          state
          timeEstimate
          deadline
          measure {
            id
            name
          }
          assignedTo {
            id
            fullName
          }
        }
      }
    }
  }
`;

// Define the update task mutation
const updateTaskStateMutation = graphql`
  mutation ListTaskViewUpdateTaskStateMutation(
    $input: UpdateTaskInput!
  ) {
    updateTask(input: $input) {
      task {
        id
        state
      }
    }
  }
`;

const listTaskViewQuery = graphql`
  query ListTaskViewQuery($organizationId: ID!) {
    node(id: $organizationId) {
      id
      ... on Organization {
        name
        tasks(first: 100) @connection(key: "ListTaskView_tasks") {
          __id
          edges {
            node {
              id
              name
              description
              state
              timeEstimate
              deadline
              measure {
                id
                name
              }
              assignedTo {
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

// Define organization members query
const organizationQuery = graphql`
  query ListTaskViewOrganizationQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        peoples(first: 100, orderBy: { direction: ASC, field: FULL_NAME })
          @connection(key: "ListTaskView_peoples") {
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

interface Badge {
  id: string;
  name: string;
  type: 'FIRE' | 'BANK' | string;
}

interface Task {
  id: string;
  name: string;
  description?: string;
  state: string;
  timeEstimate?: number | null;
  deadline?: string | null;
  assignedTo?: {
    id: string;
    fullName: string;
  } | null;
  parent?: {
    id: string;
    name: string;
  } | null;
  badges?: Badge[];
}

function TaskCard({ task, onClick, onToggleState }: {
  task: Task;
  onClick?: (task: Task) => void;
  onToggleState?: (taskId: string, newState: TaskState) => void;
}) {
  return (
    <div
      className="flex justify-between items-center px-6 py-3 border-b border-gray-100 hover:bg-gray-50 cursor-pointer"
      onClick={() => onClick && onClick(task)}
    >
      <div className="flex gap-4 items-center">
        <div
          className="flex items-center cursor-pointer"
          onClick={(e) => {
            e.stopPropagation();
            if (onToggleState) {
              const newState = task.state === "TODO" ? "DONE" : "TODO";
              onToggleState(task.id, newState as TaskState);
            }
          }}
        >
          {task.state === "TODO" ? (
            <CircleDashed className="w-5 h-5 text-green-500" />
          ) : task.state === "DONE" ? (
            <CheckCircle2 className="w-5 h-5 text-gray-500" />
          ) : (
            <Circle className="w-5 h-5 text-gray-300" />
          )}
        </div>
        <div className="flex flex-col">
          <div className="flex items-center gap-2">
            <span className={`text-sm font-medium ${task.state === "DONE" ? "text-gray-500 line-through" : "text-gray-900"}`}>
              {task.name}
            </span>
            <Badge
              variant="outline"
              className={`text-xs px-1.5 py-0.5 rounded-lg flex items-center gap-1
                ${task.state === "TODO"
                  ? "bg-green-50 text-green-700 border-green-200"
                  : "bg-gray-50 text-gray-700 border-gray-200"}
              `}
            >
              {task.state === "TODO" ? "To Do" : "Done"}
            </Badge>
            {task.timeEstimate && (
              <span className="text-xs text-gray-500">
                {formatDuration(task.timeEstimate.toString())}
              </span>
            )}
            {task.deadline && (
              <span className="text-xs text-gray-500">
                {formatDate(task.deadline, "dd MMM yyyy")}
              </span>
            )}
          </div>
          {task.parent && (
            <div className="flex items-center gap-1 text-sm text-gray-500">
              <CornerDownRight className="w-3 h-3 text-gray-500" />
              <span>{task.parent.name}</span>
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center gap-4">
        <div className="flex gap-2">
          {task.badges && task.badges.map((badge) => (
            <Badge
              key={badge.id}
              variant="outline"
              className="text-xs px-1.5 py-0.5 rounded-lg flex items-center gap-1 text-gray-600 border-gray-200/50"
            >
              {badge.type === "FIRE" ? (
                <Flame className="w-3 h-3 text-gray-600" />
              ) : badge.type === "BANK" ? (
                <Building className="w-3 h-3 text-gray-600" />
              ) : null}
              {badge.name}
            </Badge>
          ))}
        </div>
        {task.assignedTo && (
          <Avatar className="h-6 w-6 border border-white shadow-sm">
            <div className="bg-green-100 h-full w-full rounded-full flex items-center justify-center text-[10px] font-medium text-green-800">
              {task.assignedTo.fullName.split(' ').map((n: string) => n[0]).join('')}
            </div>
          </Avatar>
        )}
      </div>
    </div>
  );
}

function ListTaskContent({
  queryRef,
  organizationQueryRef,
}: {
  queryRef: PreloadedQuery<ListTaskViewQuery>;
  organizationQueryRef: PreloadedQuery<ListTaskViewOrganizationQuery>;
}) {
  const data = usePreloadedQuery<ListTaskViewQuery>(listTaskViewQuery, queryRef);
  const organizationData = usePreloadedQuery<ListTaskViewOrganizationQuery>(
    organizationQuery,
    organizationQueryRef
  );

  const organization = data.node;
  const [isNewTaskOpen, setIsNewTaskOpen] = useState(false);

  // Get organization members
  const members = organizationData.organization?.peoples?.edges.map(edge => edge.node) || [];

  // Fetch measures
  const environment = useRelayEnvironment();
  const [measures, setMeasures] = useState<Array<{id: string, name: string, category: string, description?: string, state?: string}>>([]);
  const [measuresLoading, setMeasuresLoading] = useState(false);
  const [measureSearchQuery, setMeasureSearchQuery] = useState("");

  // Task form state
  const [taskTitle, setTaskTitle] = useState("");
  const [taskDescription, setTaskDescription] = useState("");
  const [taskStatus, setTaskStatus] = useState<"TODO" | "DONE">("TODO");
  const [assignee, setAssignee] = useState<{id: string, fullName: string} | null>(null);
  const [measureId, setMeasureId] = useState<string | null>(null);
  const [selectedMeasure, setSelectedMeasure] = useState<{id: string, name: string} | null>(null);
  const [timeEstimate, setTimeEstimate] = useState<number | null>(null);
  const [timeUnit, setTimeUnit] = useState<"minutes" | "hours" | "days" | "weeks">("hours");
  const [deadlineString, setDeadlineString] = useState<string>('');
  const [deadline, setDeadline] = useState<Date | null>(null);

  // Create task mutation
  const [createTask, isCreatingTask] = useMutation(createTaskMutation);

  // Update task state mutation
  const [updateTaskState] = useMutation(updateTaskStateMutation);

  // Fetch measures from the organization
  useEffect(() => {
    if (organization?.id) {
      setMeasuresLoading(true);
      fetchQuery(
        environment,
        organizationMeasuresQuery,
        { organizationId: organization.id }
      ).subscribe({
        next: (data: any) => {
          if (data.organization?.measures?.edges) {
            const measuresList = data.organization.measures.edges.map((edge: any) => ({
              id: edge.node.id,
              name: edge.node.name,
              category: edge.node.category,
              description: edge.node.description,
              state: edge.node.state
            }));
            setMeasures(measuresList);
          }
          setMeasuresLoading(false);
        },
        error: () => {
          setMeasuresLoading(false);
        }
      });
    }
  }, [environment, organization?.id]);

  // Parse deadline string into Date-object
  const handleDeadlineChange = (event: ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    setDeadlineString(value);

    if (value) {
      const date = parseISO(value);
      setDeadline(date);
    } else {
      setDeadline(null)
    }
  };

  // Filter measures based on search query
  const filteredMeasures = measures.filter(measure => 
    measure.name.toLowerCase().includes(measureSearchQuery.toLowerCase()) || 
    (measure.description && measure.description.toLowerCase().includes(measureSearchQuery.toLowerCase()))
  );

  const allTasks = data.node?.tasks?.edges.map((edge) => ({
    ...edge.node,
    badges: [
      { id: "1", name: "Policy violation", type: "FIRE" },
      { id: "2", name: "SOC2", type: "BANK" },
    ],
    parent: edge.node.measure ? { id: edge.node.measure.id, name: edge.node.measure.name } : null,
  })) || [];

  const todoTasks = allTasks.filter(task => task.state === "TODO");
  const doneTasks = allTasks.filter(task => task.state === "DONE");

  const handleTaskClick = (task: Task) => {
    console.log("Task clicked:", task);
    // Navigate to task detail page
    // window.location.href = `/organizations/${organizationId}/tasks/${task.id}`;
  };

  // Function to convert timeEstimate to ISO8601 duration
  const convertToISODuration = (value: number | null, unit: string): string | null => {
    if (value === null || value <= 0) return null;

    switch (unit) {
      case "minutes":
        return `PT${value}M`;
      case "hours":
        return `PT${value}H`;
      case "days":
        return `P${value}D`;
      case "weeks":
        return `P${value * 7}D`;
      default:
        return null;
    }
  };

  const handleCreateTask = () => {
    if (!taskTitle.trim()) return;

    // Convert timeEstimate to ISO8601 duration
    const formattedTimeEstimate = convertToISODuration(timeEstimate, timeUnit);

    createTask({
      variables: {
        input: {
          name: taskTitle,
          description: taskDescription,
          organizationId: organization?.id || "",
          measureId: selectedMeasure?.id || null,
          assignedToId: assignee?.id,
          timeEstimate: formattedTimeEstimate,
          deadline: deadline
        },
        connections: [data.node?.tasks?.__id || ""],
      },
      onCompleted: () => {
        setIsNewTaskOpen(false);
        resetTaskForm();
      },
      onError: error => {
        console.error("Error creating task:", error);
      }
    });
  };

  const resetTaskForm = () => {
    setTaskTitle("");
    setTaskDescription("");
    setTaskStatus("TODO");
    setAssignee(null);
    setSelectedMeasure(null);
    setMeasureId(null);
    setTimeEstimate(null);
    setTimeUnit("hours");
    setMeasureSearchQuery("");
    setDeadlineString('');
    setDeadline(null);
  };

  // Get initials from name
  const getInitials = (name: string) => {
    return name.split(' ').map(n => n[0]).join('');
  };

  // Function to handle toggling task state
  const handleToggleTaskState = (taskId: string, newState: TaskState) => {
    updateTaskState({
      variables: {
        input: {
          taskId: taskId,
          state: newState
        }
      },
      optimisticResponse: {
        updateTask: {
          task: {
            id: taskId,
            state: newState,
          }
        }
      }
    });
  };

  return (
    <PageTemplate
      title="Tasks"
      description="Track your assigned compliance tasks and keep progress on track."
      actions={
        <>
          <Button
            className="bg-gray-900 text-white hover:bg-gray-800 gap-1.5"
            onClick={() => setIsNewTaskOpen(true)}
          >
            <Plus className="w-4 h-4" />
            New task
          </Button>
        </>
      }
    >
      <div className="space-y-6">
        <Card className="border border-gray-100 rounded-xl p-0 overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <Tabs defaultValue="todo" className="w-full">
              <div className="flex justify-between items-center w-full mb-6">
                <TabsList className="w-full">
                  <TabsTrigger value="todo" className="border-b-2 border-transparent data-[state=active]:border-gray-900">
                    <div className="flex items-center gap-2">
                      <CircleDashed className="w-4 h-4 text-green-500" />
                      <span>To Do</span>
                      <div className="flex items-center justify-center bg-green-100 rounded-full h-6 w-6 text-xs text-green-800">
                        {todoTasks.length}
                      </div>
                    </div>
                  </TabsTrigger>
                  <TabsTrigger value="done" className="border-b-2 border-transparent data-[state=active]:border-gray-900">
                    <div className="flex items-center gap-2">
                      <Circle className="w-4 h-4 text-gray-500" />
                      <span>Done</span>
                      <div className="flex items-center justify-center bg-gray-100 rounded-full h-6 w-6 text-xs text-gray-800">
                        {doneTasks.length}
                      </div>
                    </div>
                  </TabsTrigger>
                  <TabsTrigger value="all" className="border-b-2 border-transparent data-[state=active]:border-gray-900">
                    <div className="flex items-center gap-2">
                      <span>All</span>
                      <div className="flex items-center justify-center bg-gray-100 rounded-full h-6 w-6 text-xs">
                        {allTasks.length}
                      </div>
                    </div>
                  </TabsTrigger>
                </TabsList>
              </div>

              <TabsContent value="todo">
                {todoTasks.length > 0 ? (
                  todoTasks.map((task) => (
                    <TaskCard
                      key={task.id}
                      task={task}
                      onClick={handleTaskClick}
                      onToggleState={handleToggleTaskState}
                    />
                  ))
                ) : (
                  <div className="p-4 text-center text-gray-500">
                    No todo tasks found. Create a new task to get started.
                  </div>
                )}
              </TabsContent>

              <TabsContent value="done">
                {doneTasks.length > 0 ? (
                  doneTasks.map((task) => (
                    <TaskCard
                      key={task.id}
                      task={task}
                      onClick={handleTaskClick}
                      onToggleState={handleToggleTaskState}
                    />
                  ))
                ) : (
                  <div className="p-4 text-center text-gray-500">
                    No completed tasks found.
                  </div>
                )}
              </TabsContent>

              <TabsContent value="all">
                {allTasks.length > 0 ? (
                  allTasks.map((task) => (
                    <TaskCard
                      key={task.id}
                      task={task}
                      onClick={handleTaskClick}
                      onToggleState={handleToggleTaskState}
                    />
                  ))
                ) : (
                  <div className="p-4 text-center text-gray-500">
                    No tasks found. Create a new task to get started.
                  </div>
                )}
              </TabsContent>
            </Tabs>
          </div>
        </Card>
      </div>

      {/* New Task Modal */}
      <Dialog open={isNewTaskOpen} onOpenChange={setIsNewTaskOpen}>
        <DialogContent className="sm:max-w-[1080px] p-0 rounded-lg overflow-hidden border border-gray-200">
          {/* Modal Header */}
          <div className="flex justify-between items-center px-6 py-4 border-b border-[rgba(2,42,2,0.08)]">
            <div className="flex items-center gap-2">
              <div className="text-sm text-gray-500">List of your assets</div>
              <div className="text-gray-500">
                <CornerDownRight className="w-3 h-3" />
              </div>
              <div className="text-sm font-medium">New task</div>
            </div>
            <DialogClose className="rounded-full w-8 h-8 flex items-center justify-center hover:bg-gray-100"/>
          </div>

          {/* Task Form - Two Column Layout */}
          <div className="flex">
            {/* Left Column - Task Details */}
            <div className="flex-1 border-r border-[rgba(2,42,2,0.08)] p-6">
              <Input
                placeholder="Task title"
                className="text-2xl font-semibold border-none px-0 mb-4 w-full"
                value={taskTitle}
                onChange={(e) => setTaskTitle(e.target.value)}
              />
              <Textarea
                placeholder="Add description..."
                className="resize-none border-none px-0 text-base w-full min-h-[400px]"
                value={taskDescription}
                onChange={(e) => setTaskDescription(e.target.value)}
              />
            </div>

            {/* Right Column - Properties */}
            <div className="w-[400px] p-6">
              <h3 className="text-base font-medium text-gray-900 mb-4">Properties</h3>

              <div className="space-y-4">
                {/* Status */}
                <div className="flex justify-between items-center border-b border-[rgba(2,42,2,0.08)] py-3">
                  <Label className="text-sm font-medium text-gray-500">Status</Label>
                  <div className="bg-[rgba(0,39,0,0.05)] px-[10px] py-[6px] rounded-lg flex items-center gap-1.5">
                    <img
                      src="/images/radio-unchecked.svg"
                      alt="Radio"
                      className="w-4 h-4"
                    />
                    <span className="text-sm font-medium">To do</span>
                  </div>
                </div>

                {/* Assignee */}
                <div className="flex justify-between items-center border-b border-[rgba(2,42,2,0.08)] py-3">
                  <Label className="text-sm font-medium text-gray-500">Assignee</Label>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <div className="bg-[rgba(0,39,0,0.05)] px-[10px] py-[6px] rounded-lg flex items-center gap-1.5 cursor-pointer">
                        {assignee ? (
                          <>
                            <Avatar className="h-4 w-4">
                              <div className="bg-blue-100 h-full w-full rounded-full flex items-center justify-center text-[7px] font-medium text-blue-800">
                                {getInitials(assignee.fullName)}
                              </div>
                            </Avatar>
                            <span className="text-sm font-medium">{assignee.fullName}</span>
                          </>
                        ) : (
                          <>
                            <User className="h-4 w-4 text-gray-500" />
                            <span className="text-sm font-medium">Unassigned</span>
                          </>
                        )}
                        <ChevronDown className="h-3 w-3 text-gray-500 ml-1" />
                      </div>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-[220px] max-h-[300px] overflow-y-auto">
                      <DropdownMenuItem
                        className="flex items-center gap-2 cursor-pointer"
                        onClick={() => setAssignee(null)}
                      >
                        <User className="h-4 w-4 text-gray-500" />
                        <span>Unassigned</span>
                      </DropdownMenuItem>

                      {members.map(member => (
                        <DropdownMenuItem
                          key={member.id}
                          className="flex items-center gap-2 cursor-pointer"
                          onClick={() => setAssignee({id: member.id, fullName: member.fullName})}
                        >
                          <Avatar className="h-4 w-4">
                            <div className="bg-blue-100 h-full w-full rounded-full flex items-center justify-center text-[7px] font-medium text-blue-800">
                              {getInitials(member.fullName)}
                            </div>
                          </Avatar>
                          <span>{member.fullName}</span>
                        </DropdownMenuItem>
                      ))}
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>

                {/* Measure */}
                <div className="flex justify-between items-center border-b border-[rgba(2,42,2,0.08)] py-3">
                  <Label className="text-sm font-medium text-gray-500">Measure</Label>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <div className="bg-[rgba(0,39,0,0.05)] px-[10px] py-[6px] rounded-lg flex items-center gap-1.5 cursor-pointer">
                        {selectedMeasure ? (
                          <span className="text-sm font-medium">{selectedMeasure.name}</span>
                        ) : (
                          <span className="text-sm font-medium">No measure</span>
                        )}
                        <ChevronDown className="h-3 w-3 text-gray-500 ml-1" />
                      </div>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-[320px] max-h-[300px] overflow-y-auto">
                      <div className="p-2">
                        <div className="relative mb-2">
                          <Search className="absolute left-2 top-2.5 h-4 w-4 text-gray-500" />
                          <Input
                            placeholder="Search measures..."
                            value={measureSearchQuery}
                            onChange={(e) => setMeasureSearchQuery(e.target.value)}
                            className="pl-8"
                          />
                        </div>
                      </div>

                      <DropdownMenuItem
                        className="flex items-center gap-2 cursor-pointer"
                        onClick={() => {
                          setSelectedMeasure(null);
                          setMeasureId(null);
                        }}
                      >
                        <span>No measure</span>
                      </DropdownMenuItem>

                      {measuresLoading ? (
                        <div className="p-3 text-center text-sm text-gray-500">
                          Loading measures...
                        </div>
                      ) : filteredMeasures.length > 0 ? (
                        filteredMeasures.map(measure => (
                          <DropdownMenuItem
                            key={measure.id}
                            className="flex items-start gap-2 cursor-pointer py-2"
                            onClick={() => {
                              setSelectedMeasure({id: measure.id, name: measure.name});
                              setMeasureId(measure.id);
                            }}
                          >
                            <div className="flex flex-col">
                              <span className="font-medium">{measure.name}</span>
                              {measure.category && (
                                <span className="text-xs text-gray-500">{measure.category}</span>
                              )}
                            </div>
                          </DropdownMenuItem>
                        ))
                      ) : (
                        <div className="p-3 text-center text-sm text-gray-500">
                          No measures found
                        </div>
                      )}
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>

                {/* Time estimate */}
                <div className="flex justify-between items-center border-b border-[rgba(2,42,2,0.08)] py-3">
                  <Label className="text-sm font-medium text-gray-500">Time estimate</Label>
                  <div className="flex items-center gap-2">
                    {timeEstimate === null ? (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 px-2"
                        onClick={() => setTimeEstimate(0)}
                      >
                        <Plus className="w-4 h-4" />
                      </Button>
                    ) : (
                      <>
                        <Input
                          type="number"
                          min="0"
                          value={timeEstimate}
                          onChange={(e) => setTimeEstimate(parseInt(e.target.value) || 0)}
                          className="w-16 h-8 text-sm"
                        />
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-8 flex items-center gap-1"
                            >
                              <span>{timeUnit}</span>
                              <ChevronDown className="h-3 w-3" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => setTimeUnit("minutes")}>
                              minutes
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => setTimeUnit("hours")}>
                              hours
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => setTimeUnit("days")}>
                              days
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => setTimeUnit("weeks")}>
                              weeks
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-8 p-0"
                          onClick={() => setTimeEstimate(null)}
                        >
                          <X className="w-4 h-4" />
                        </Button>
                      </>
                    )}
                  </div>
                </div>

                {/* Deadline */}
                <div className="flex justify-between items-center broder-b border-[rgba(2,42,2,0.08)] py-3">
                  <Label className="text-sm font-medium text-gray-500">Deadline</Label>
                  <div className="flex items-center gap-1.5">
                    <Input type="date" value={deadlineString} onChange={handleDeadlineChange}
                           className="text-sm font-medium bg-[rgba(0,39,0,0.05)] px-[10px] py-[6px] rounded-lg"/>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Modal Footer */}
          <div className="flex justify-between items-center px-6 py-4 border-t border-[rgba(2,42,2,0.08)]">
            <Button
              variant="outline"
              className="h-8 flex items-center gap-1.5 border-[rgba(2,42,2,0.08)]"
            >
              <img src="/images/attachment.svg" alt="Attachment" className="w-4 h-4" />
              Upload evidence
            </Button>

            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                className="h-8 border-[rgba(2,42,2,0.08)]"
                onClick={() => setIsNewTaskOpen(false)}
              >
                Cancel
              </Button>
              <Button
                className="h-8 bg-gray-900 text-white"
                onClick={handleCreateTask}
                disabled={!taskTitle.trim()}
              >
                Create task
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </PageTemplate>
  );
}

export default function ListTaskView() {
  const [searchParams] = useSearchParams();
  const [queryRef, loadQuery] =
    useQueryLoader<ListTaskViewQuery>(listTaskViewQuery);
  const [organizationQueryRef, loadOrganizationQuery] =
    useQueryLoader<ListTaskViewOrganizationQuery>(organizationQuery);

  const { organizationId } = useParams();

  useEffect(() => {
    loadQuery({ organizationId: organizationId! });
    loadOrganizationQuery({ organizationId: organizationId! });
  }, [loadQuery, loadOrganizationQuery, organizationId]);

  if (!queryRef || !organizationQueryRef) {
    return <ListTaskViewSkeleton />;
  }

  return (
    <Suspense fallback={<ListTaskViewSkeleton />}>
      <ListTaskContent queryRef={queryRef} organizationQueryRef={organizationQueryRef} />
    </Suspense>
  );
}
