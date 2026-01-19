import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Field,
  IconChevronLeft,
  PageHeader,
  useToast,
} from "@probo/ui";
import { graphql, useMutation } from "react-relay";
import type { FormEventHandler } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { formatError } from "@probo/helpers";
import type { NewOrganizationPageMutation } from "/__generated__/iam/NewOrganizationPageMutation.graphql";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";

const createOrganizationMutation = graphql`
  mutation NewOrganizationPageMutation($input: CreateOrganizationInput!) {
    createOrganization(input: $input) {
      organization {
        id
        name
      }
    }
  }
`;

function NewOrganizationPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { __ } = useTranslate();

  const [createOrganization, isCreating] =
    useMutation<NewOrganizationPageMutation>(createOrganizationMutation);

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

    createOrganization({
      variables: {
        input: {
          name,
        },
      },
      onCompleted: (r, e) => {
        if (e) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to create organization"), e),
            variant: "error",
          });
          return;
        }

        const org = r.createOrganization!.organization;
        navigate(`/organizations/${org!.id}`);
        toast({
          title: __("Success"),
          description: __("Organization has been created successfully"),
          variant: "success",
        });
      },
      onError: (e) => {
        toast({
          title: __("Error"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-6">
      <Link
        to={location.state?.from ?? "/"}
        className="mb-4 inline-flex gap-2 items-center"
      >
        <IconChevronLeft size={16} />
        {__("Back")}
      </Link>
      <PageHeader
        title={__("Create Organization")}
        description={__(
          "Create a new organization to manage your compliance and security needs.",
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
              "The name of your organization as it will appear throughout the platform.",
            )}
          />
          <Button disabled={isCreating} type="submit" className="w-full">
            {__("Create Organization")}
          </Button>
        </form>
      </Card>
    </div>
  );
}

export default function () {
  return (
    <IAMRelayProvider>
      <NewOrganizationPage />
    </IAMRelayProvider>
  );
}
