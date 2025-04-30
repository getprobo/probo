import { Navigate, useParams } from "react-router";

export default function HomePage() {
  const { organizationId } = useParams();
  return <Navigate to={`/organizations/${organizationId}/measures`} />;
}
