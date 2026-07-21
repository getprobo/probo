// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { formatError } from "@probo/helpers";
import {
  Button,
  Card,
  Field,
  IconChevronLeft,
  PageHeader,
  useToast,
} from "@probo/ui";
import type { FormEventHandler } from "react";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();

  const [createOrganization, isCreating]
    = useMutation<NewOrganizationPageMutation>(createOrganizationMutation);

  const handleSubmit: FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const name = formData.get("name") ? (formData.get("name") as string).toString() : "";
    if (!name) {
      toast({
        title: t("common.error"), description: t("newOrganizationPage.errors.nameRequired"),
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
            title: t("common.error"), description: formatError(t("newOrganizationPage.errors.create"), e),
            variant: "error",
          });
          return;
        }

        const org = r.createOrganization!.organization;
        void navigate(`/organizations/${org!.id}`);
        toast({
          title: t("common.success"), description: t("newOrganizationPage.messages.created"),
          variant: "success",
        });
      },
      onError: (e) => {
        toast({
          title: t("common.error"),
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
        {t("newOrganizationPage.actions.back")}
      </Link>
      <PageHeader
        title={t("newOrganizationPage.title")}
        description={t("newOrganizationPage.description")}
      />
      <Card padded asChild>
        <form onSubmit={e => void handleSubmit(e)} className="space-y-4">
          <h2 className="text-xl font-semibold mb-1">
            {t("newOrganizationPage.details")}
          </h2>
          <p className="text-txt-tertiary text-sm mb-4">
            {t("newOrganizationPage.detailsDescription")}
          </p>
          <Field
            required
            name="name"
            type="text"
            placeholder={t("newOrganizationPage.fields.name")}
            label={t("newOrganizationPage.fields.name")}
            help={t("newOrganizationPage.fields.nameHelp")}
          />
          <Button disabled={isCreating} type="submit" className="w-full">
            {t("newOrganizationPage.actions.create")}
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
