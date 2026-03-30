// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Field,
  IconChevronLeft,
  PageHeader,
  useToast,
} from "@probo/ui";
import type { FormEventHandler } from "react";
import { graphql, useMutation } from "react-relay";
import { Link, useLocation, useNavigate } from "react-router";

import type { NewOrganizationPageMutation } from "#/__generated__/iam/NewOrganizationPageMutation.graphql";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

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

function NewOrganizationPageInner() {
  const location = useLocation();
  const navigate = useNavigate();
  const { toast } = useToast();
  const { __ } = useTranslate();

  const [createOrganization, isCreating]
    = useMutation<NewOrganizationPageMutation>(createOrganizationMutation);

  const handleSubmit: FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const name = formData.get("name") ? (formData.get("name") as string).toString() : "";
    if (!name) {
      toast({
        title: __("Error"),
        description: __("Name is required"),
        variant: "error",
      });
      return;
    }

    void createOrganization({
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
        void navigate(`/organizations/${org!.id}`);
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
        to={(location.state as { from: string })?.from ?? "/"}
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
        <form onSubmit={e => void handleSubmit(e)} className="space-y-4">
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

export default function NewOrganizationPage() {
  return (
    <IAMRelayProvider>
      <NewOrganizationPageInner />
    </IAMRelayProvider>
  );
}
