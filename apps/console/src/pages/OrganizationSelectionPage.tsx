import { useEffect } from "react";
import { useNavigate } from "react-router";
import { graphql, useLazyLoadQuery } from "react-relay";
import { Helmet } from "react-helmet-async";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from "@/components/ui/card";
import { Plus } from "lucide-react";
import type { OrganizationSelectionPageQuery } from "./__generated__/OrganizationSelectionPageQuery.graphql";

const organizationSelectionQuery = graphql`
  query OrganizationSelectionPageQuery {
    viewer {
      organizations(first: 25) {
        edges {
          node {
            id
            name
            logoUrl
          }
        }
      }
    }
  }
`;

export default function OrganizationSelectionPage() {
  const navigate = useNavigate();
  const data = useLazyLoadQuery<OrganizationSelectionPageQuery>(
    organizationSelectionQuery,
    {},
  );

  const organizations = data.viewer.organizations.edges.map(
    (edge) => edge.node,
  );

  useEffect(() => {
    if (organizations.length === 1) {
      navigate(`/organizations/${organizations[0].id}`);
    }
  }, [organizations, navigate]);

  if (organizations.length === 0) {
    return (
      <div className="container mx-auto py-10">
        <Helmet>
          <title>Select Organization - Probo</title>
        </Helmet>
        <div className="flex flex-col items-center justify-center h-[70vh]">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>Welcome to Probo</CardTitle>
              <CardDescription>
                You don{"'"}t have any organizations yet. Create your first one
                to get started.
              </CardDescription>
            </CardHeader>
            <CardFooter>
              <Button
                className="w-full"
                onClick={() => navigate("/organizations/new")}
              >
                <Plus className="mr-2 h-4 w-4" /> Create Organization
              </Button>
            </CardFooter>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-10">
      <Helmet>
        <title>Select Organization - Probo</title>
      </Helmet>
      <div className="flex flex-col items-center">
        <h1 className="text-2xl font-bold mb-6">Select an Organization</h1>
        <div className="grid gap-4 w-full max-w-2xl">
          {organizations.map((org) => (
            <Card key={org.id} className="w-full">
              <CardHeader className="flex flex-row items-center gap-4">
                {org.logoUrl && (
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg border">
                    <img src={org.logoUrl} alt={org.name} className="h-8 w-8" />
                  </div>
                )}
                <div>
                  <CardTitle>{org.name}</CardTitle>
                </div>
              </CardHeader>
              <CardFooter>
                <Button
                  className="w-full"
                  onClick={() => navigate(`/organizations/${org.id}`)}
                >
                  Select
                </Button>
              </CardFooter>
            </Card>
          ))}
          <Card className="w-full">
            <CardHeader>
              <CardTitle>Create a New Organization</CardTitle>
              <CardDescription>
                Add a new organization to your account
              </CardDescription>
            </CardHeader>
            <CardFooter>
              <Button
                variant="outline"
                className="w-full"
                onClick={() => navigate("/organizations/new")}
              >
                <Plus className="mr-2 h-4 w-4" /> Create Organization
              </Button>
            </CardFooter>
          </Card>
        </div>
      </div>
    </div>
  );
}
