import { useParams } from "react-router";

export function useOrganizationId(): string {
  const { organizationId } = useParams<{ organizationId?: string }>();
  if (!organizationId) {
    throw new Error("Cannot resolve organizationId in route params");
  }
  return organizationId;
}
