import {
  Button,
  IconPlusLarge,
  PageHeader,
  Card,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
  Table,
  useConfirm,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { sprintf } from "@probo/helpers";
import { CreateIncidentRegistryDialog } from "./dialogs/CreateIncidentRegistryDialog";

// Mock data for now since we don't have the full GraphQL implementation
const mockIncidentRegistries = [
  {
    id: "1",
    referenceId: "INC-001",
    title: "Database Performance Issue",
    description: "Production database experiencing slow query performance",
    status: "OPEN",
    priority: "HIGH",
    severity: "MEDIUM",
    category: "Technical",
    owner: { id: "1", fullName: "John Doe" },
    incidentDate: "2024-01-15T10:30:00Z",
    resolvedDate: null,
    createdAt: "2024-01-15T10:30:00Z",
    updatedAt: "2024-01-15T10:30:00Z",
  },
  {
    id: "2",
    referenceId: "INC-002",
    title: "Security Breach Attempt",
    description: "Failed login attempts detected from suspicious IP addresses",
    status: "RESOLVED",
    priority: "CRITICAL",
    severity: "HIGH",
    category: "Security",
    owner: { id: "2", fullName: "Jane Smith" },
    incidentDate: "2024-01-14T14:20:00Z",
    resolvedDate: "2024-01-14T18:45:00Z",
    createdAt: "2024-01-14T14:20:00Z",
    updatedAt: "2024-01-14T18:45:00Z",
  },
];

export default function IncidentRegistriesPage() {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  usePageTitle(__("Incident Register"));

  const registries = mockIncidentRegistries;

  return (
    <div className="space-y-6">
      <PageHeader 
        title={__("Incident Register")} 
        description={__("Manage and track security and operational incidents")}
      >
        <CreateIncidentRegistryDialog organizationId={organizationId}>
          <Button icon={IconPlusLarge}>
            {__("Add incident")}
          </Button>
        </CreateIncidentRegistryDialog>
      </PageHeader>

      {registries.length > 0 ? (
        <Card>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Reference ID")}</Th>
                <Th>{__("Title")}</Th>
                <Th>{__("Status")}</Th>
                <Th>{__("Priority")}</Th>
                <Th>{__("Severity")}</Th>
                <Th>{__("Category")}</Th>
                <Th>{__("Owner")}</Th>
                <Th>{__("Incident Date")}</Th>
                <Th>{__("Actions")}</Th>
              </Tr>
            </Thead>
            <Tbody>
              {registries.map((registry) => (
                <RegistryRow
                  key={registry.id}
                  registry={registry}
                />
              ))}
            </Tbody>
          </Table>
        </Card>
      ) : (
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("No incident registry entries yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first incident registry entry to get started.")}
            </p>
          </div>
        </Card>
      )}
    </div>
  );
}

function RegistryRow({ registry }: { registry: any }) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const getStatusVariant = (status: string) => {
    switch (status) {
      case "OPEN":
        return "danger";
      case "IN_PROGRESS":
        return "warning";
      case "RESOLVED":
        return "success";
      case "CLOSED":
        return "neutral";
      default:
        return "neutral";
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case "OPEN":
        return __("Open");
      case "IN_PROGRESS":
        return __("In Progress");
      case "RESOLVED":
        return __("Resolved");
      case "CLOSED":
        return __("Closed");
      default:
        return status;
    }
  };

  const getPriorityVariant = (priority: string) => {
    switch (priority) {
      case "LOW":
        return "success";
      case "MEDIUM":
        return "warning";
      case "HIGH":
        return "danger";
      case "CRITICAL":
        return "danger";
      default:
        return "neutral";
    }
  };

  const getPriorityLabel = (priority: string) => {
    switch (priority) {
      case "LOW":
        return __("Low");
      case "MEDIUM":
        return __("Medium");
      case "HIGH":
        return __("High");
      case "CRITICAL":
        return __("Critical");
      default:
        return priority;
    }
  };

  const getSeverityVariant = (severity: string) => {
    switch (severity) {
      case "LOW":
        return "success";
      case "MEDIUM":
        return "warning";
      case "HIGH":
        return "danger";
      case "CRITICAL":
        return "danger";
      default:
        return "neutral";
    }
  };

  const getSeverityLabel = (severity: string) => {
    switch (severity) {
      case "LOW":
        return __("Low");
      case "MEDIUM":
        return __("Medium");
      case "HIGH":
        return __("High");
      case "CRITICAL":
        return __("Critical");
      default:
        return severity;
    }
  };

  const handleDelete = () => {
    confirm(
      () => Promise.resolve(),
      {
        message: sprintf(
          __(
            "This will permanently delete the incident registry entry %s. This action cannot be undone."
          ),
          registry.referenceId
        ),
      }
    );
  };

  return (
    <Tr to={`/organizations/${organizationId}/incident-registries/${registry.id}`}>
      <Td>
        <span className="font-mono text-sm">{registry.referenceId}</span>
      </Td>
      <Td>{registry.title}</Td>
      <Td>
        <Badge variant={getStatusVariant(registry.status)}>
          {getStatusLabel(registry.status)}
        </Badge>
      </Td>
      <Td>
        <Badge variant={getPriorityVariant(registry.priority)}>
          {getPriorityLabel(registry.priority)}
        </Badge>
      </Td>
      <Td>
        <Badge variant={getSeverityVariant(registry.severity)}>
          {getSeverityLabel(registry.severity)}
        </Badge>
      </Td>
      <Td>{registry.category || "-"}</Td>
      <Td>{registry.owner?.fullName || "-"}</Td>
      <Td>
        {registry.incidentDate ? (
          <time dateTime={registry.incidentDate}>
            {formatDate(registry.incidentDate)}
          </time>
        ) : (
          <span className="text-txt-tertiary">{__("No incident date")}</span>
        )}
      </Td>
      <Td noLink width={50} className="text-end">
        <ActionDropdown>
          <DropdownItem
            icon={IconTrashCan}
            variant="danger"
            onSelect={handleDelete}
          >
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </Td>
    </Tr>
  );
}