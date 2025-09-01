import {
  PageHeader,
  Card,
  PropertyRow,
  Badge,
  Button,
  IconPencil,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";

// Mock data for now since we don't have the full GraphQL implementation
const mockIncidentRegistry = {
  id: "1",
  referenceId: "INC-001",
  title: "Database Performance Issue",
  description: "Production database experiencing slow query performance affecting multiple services. Users are reporting slow load times and timeouts.",
  status: "OPEN",
  priority: "HIGH",
  severity: "MEDIUM",
  category: "Technical",
  source: "Monitoring System Alert",
  owner: { id: "1", fullName: "John Doe" },
  incidentDate: "2024-01-15T10:30:00Z",
  resolvedDate: null,
  createdAt: "2024-01-15T10:30:00Z",
  updatedAt: "2024-01-15T10:30:00Z",
};

export default function IncidentRegistryDetailsPage() {
  const { __ } = useTranslate();

  const registry = mockIncidentRegistry;

  usePageTitle(`${__("Incident")} ${registry.referenceId}`);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
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

  return (
    <div className="space-y-6">
      <PageHeader
        title={`${__("Incident")} ${registry.referenceId}`}
        description={registry.title}
      >
        <Button icon={IconPencil} variant="secondary">
          {__("Edit incident")}
        </Button>
      </PageHeader>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <div className="p-6">
            <h3 className="text-lg font-semibold mb-4">{__("Incident Details")}</h3>
            <div className="space-y-4">
              <PropertyRow label={__("Reference ID")}>
                <span className="font-mono">{registry.referenceId}</span>
              </PropertyRow>
              
              <PropertyRow label={__("Title")}>
                {registry.title}
              </PropertyRow>

              <PropertyRow label={__("Description")}>
                {registry.description || __("No description provided")}
              </PropertyRow>

              <PropertyRow label={__("Category")}>
                {registry.category || __("No category specified")}
              </PropertyRow>

              <PropertyRow label={__("Source")}>
                {registry.source || __("No source specified")}
              </PropertyRow>
            </div>
          </div>
        </Card>

        <Card>
          <div className="p-6">
            <h3 className="text-lg font-semibold mb-4">{__("Status & Classification")}</h3>
            <div className="space-y-4">
              <PropertyRow label={__("Status")}>
                <Badge variant={getStatusVariant(registry.status)}>
                  {getStatusLabel(registry.status)}
                </Badge>
              </PropertyRow>

              <PropertyRow label={__("Priority")}>
                <Badge variant={getPriorityVariant(registry.priority)}>
                  {getPriorityLabel(registry.priority)}
                </Badge>
              </PropertyRow>

              <PropertyRow label={__("Severity")}>
                <Badge variant={getSeverityVariant(registry.severity)}>
                  {getSeverityLabel(registry.severity)}
                </Badge>
              </PropertyRow>

              <PropertyRow label={__("Owner")}>
                {registry.owner?.fullName || __("Unassigned")}
              </PropertyRow>
            </div>
          </div>
        </Card>

        <Card>
          <div className="p-6">
            <h3 className="text-lg font-semibold mb-4">{__("Timeline")}</h3>
            <div className="space-y-4">
              <PropertyRow label={__("Incident Date")}>
                {registry.incidentDate ? (
                  <time dateTime={registry.incidentDate}>
                    {formatDate(registry.incidentDate)}
                  </time>
                ) : (
                  __("Not specified")
                )}
              </PropertyRow>

              <PropertyRow label={__("Resolved Date")}>
                {registry.resolvedDate ? (
                  <time dateTime={registry.resolvedDate}>
                    {formatDate(registry.resolvedDate)}
                  </time>
                ) : (
                  __("Not resolved")
                )}
              </PropertyRow>

              <PropertyRow label={__("Created")}>
                <time dateTime={registry.createdAt}>
                  {formatDate(registry.createdAt)}
                </time>
              </PropertyRow>

              <PropertyRow label={__("Last Updated")}>
                <time dateTime={registry.updatedAt}>
                  {formatDate(registry.updatedAt)}
                </time>
              </PropertyRow>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}