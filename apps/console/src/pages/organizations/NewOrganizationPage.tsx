import { Button, Card, Field, PageHeader, useToast } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { ConnectionHandler, graphql } from "relay-runtime";
import { useLazyLoadQuery, useMutation } from "react-relay";
import type { NewOrganizationPageQuery as NewOrganizationPageQueryType } from "./__generated__/NewOrganizationPageQuery.graphql";
import type { NewOrganizationPageMutation as NewOrganizationPageMutationType } from "./__generated__/NewOrganizationPageMutation.graphql";
import { useState, type FormEventHandler } from "react";
import { useNavigate } from "react-router";
import { formatError, type GraphQLError } from "@probo/helpers";

const createOrganizationMutation = graphql`
  mutation NewOrganizationPageMutation(
    $input: CreateOrganizationInput!
    $connections: [ID!]!
  ) {
    createOrganization(input: $input) {
      organizationEdge @appendEdge(connections: $connections) {
        node {
          id
          name
          logoUrl
        }
      }
    }
  }
`;

const viewerQuery = graphql`
  query NewOrganizationPageQuery {
    viewer {
      id
    }
  }
`;

export default function NewOrganizationPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [isFetching, setIsFetching] = useState(false);

  const data = useLazyLoadQuery<NewOrganizationPageQueryType>(viewerQuery, {});
  const [createOrganization] = useMutation<NewOrganizationPageMutationType>(
    createOrganizationMutation
  );

  const handleSubmit: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const name = formData.get("name")?.toString();
    if (!name) {
      toast({
        title: __("Error"),
        description: __("Name is required"),
        variant: "error",
      });
      return;
    }

    setIsFetching(true);

    createOrganization({
      variables: {
        input: {
          name,
        },
        connections: [
          ConnectionHandler.getConnectionID(
            data.viewer.id,
            "OrganizationsPage_organizations"
          ),
          ConnectionHandler.getConnectionID(
            data.viewer.id,
            "MainLayout_OrganizationSelector_organizations"
          ),
        ],
      },
      onCompleted: (r) => {
        setIsFetching(false);
        const org = r.createOrganization.organizationEdge.node;
        navigate(`/organizations/${org.id}`);
        toast({
          title: __("Success"),
          description: __("Organization has been created successfully"),
          variant: "success",
        });
      },
      onError: (e: GraphQLError) => {
        setIsFetching(false);

        toast({
          title: __("Error"),
          description: formatError(__("Failed to create organization"), e),
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Create Organization")}
        description={__(
          "Create a new organization to manage your compliance and security needs."
        )}
      />
      <Card padded asChild>
        <form onSubmit={handleSubmit} className="space-y-4">
          <h2 className="text-xl font-semibold mb-1">
            {__("Organization Details")}
          </h2>
          <p className="text-txt-tertiary text-sm mb-4">
            {__("Enter the basic information about your organization.")}
          </p>
          <Field
            required
            name="name"
            type="text"
            placeholder={__("Organization name")}
            label={__("Organization name")}
            help={__(
              "The name of your organization as it will appear throughout the platform."
            )}
          />
          <Button disabled={isFetching} type="submit" className="w-full">
            {__("Create Organization")}
          </Button>
        </form>
      </Card>
    </div>
  );
}
