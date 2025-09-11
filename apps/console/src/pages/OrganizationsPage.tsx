import { useTranslate } from "@probo/i18n";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";
import type { OrganizationsPageQuery as OrganizationsPageQueryType } from "./__generated__/OrganizationsPageQuery.graphql";
import { useEffect } from "react";
import { Link, useNavigate } from "react-router";
import {
  Avatar,
  Button,
  Card,
  IconPlusLarge,
} from "@probo/ui";
import { usePageTitle } from "@probo/hooks";

const OrganizationsPageQuery = graphql`
  query OrganizationsPageQuery {
    viewer {
      organizations(first: 1000, orderBy: {field: NAME, direction: ASC}) @connection(key: "OrganizationsPage_organizations") {
        __id
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

export default function OrganizationsPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const data = useLazyLoadQuery<OrganizationsPageQueryType>(
    OrganizationsPageQuery,
    {}
  );

  const organizations = data.viewer.organizations.edges.map(
    (edge) => edge.node
  );

  usePageTitle(__("Select an organization"));

  useEffect(() => {
    if (organizations.length === 1) {
      navigate(`/organizations/${organizations[0].id}`);
    } else if (organizations.length === 0) {
      navigate("/organizations/new");
    }
  }, [organizations]);

  return (
    <>
      <div className="space-y-6 w-full py-6">
        <h1 className="text-3xl font-bold text-center">
          {__("Select an organization")}
        </h1>
        <div className="space-y-4 w-full">
          {organizations.map((organization) => (
            <OrganizationCard
              key={organization.id}
              organization={organization}
            />
          ))}
          <Card padded>
            <h2 className="text-xl font-semibold mb-1">
              {__("Create an organization")}
            </h2>
            <p className="text-txt-tertiary mb-4">
              {__("Add a new organization to your account")}
            </p>
            <Button
              to="/organizations/new"
              variant="quaternary"
              icon={IconPlusLarge}
              className="w-full"
            >
              {__("Create organization")}
            </Button>
          </Card>
        </div>
      </div>
    </>
  );
}

type OrganizationCardProps = {
  organization: {
    id: string;
    name: string;
    logoUrl: string | null | undefined;
  };
};

function OrganizationCard({ organization }: OrganizationCardProps) {
  const { __ } = useTranslate();

  return (
    <Card padded className="w-full">
      <div className="flex items-center justify-between">
        <Link
          to={`/organizations/${organization.id}`}
          className="flex items-center gap-4 hover:text-primary flex-1"
        >
          <Avatar
            src={organization.logoUrl}
            name={organization.name}
            size="l"
          />
          <h2 className="font-semibold text-xl">{organization.name}</h2>
        </Link>
        <div className="flex items-center gap-3">
          <Button asChild>
            <Link to={`/organizations/${organization.id}`}>
              {__("Select")}
            </Link>
          </Button>
        </div>
      </div>
    </Card>
  );
}
