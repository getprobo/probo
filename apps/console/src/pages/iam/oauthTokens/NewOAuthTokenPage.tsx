// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Card,
  Checkbox,
  Field,
  Input,
  Label,
  Option,
  PageHeader,
  Select,
  useDialogRef,
} from "@probo/ui";
import { useMemo, useState } from "react";
import { Controller } from "react-hook-form";
import { ConnectionHandler, useLazyLoadQuery } from "react-relay";
import { Link, useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { NewOAuthTokenPageCreateMutation } from "#/__generated__/iam/NewOAuthTokenPageCreateMutation.graphql";
import type { NewOAuthTokenPageQuery } from "#/__generated__/iam/NewOAuthTokenPageQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { OAuthTokenCredentialsDialog } from "./_components/OAuthTokenCredentialsDialog";
import { formatApiScopeLabel } from "./_components/scopeLabels";

const pageQuery = graphql`
  query NewOAuthTokenPageQuery {
    oauth2ScopesSupported
    viewer @required(action: THROW) {
      id
      canCreateOAuth2AccessToken: permission(
        action: "iam:oauth2-access-token:create"
      )
    }
  }
`;

const createMutation = graphql`
  mutation NewOAuthTokenPageCreateMutation(
    $input: CreateOAuth2AccessTokenInput!
    $connections: [ID!]!
  ) {
    createOAuth2AccessToken(input: $input) {
      token
      oauth2AccessTokenEdge @prependEdge(connections: $connections) {
        node {
          id
          ...OAuthTokenRowFragment
        }
      }
    }
  }
`;

const createSchema = z.object({
  name: z.string().min(1, "Name is required"),
  expiresIn: z.enum(["1month", "3months", "6months", "1year"]),
  scopes: z.array(z.string()).min(1, "Select at least one scope"),
});

type CreateFormData = z.infer<typeof createSchema>;

function computeExpiresAt(expiresIn: CreateFormData["expiresIn"]) {
  const now = new Date();
  const expiresAt = new Date(now);
  switch (expiresIn) {
    case "1month":
      expiresAt.setMonth(now.getMonth() + 1);
      break;
    case "3months":
      expiresAt.setMonth(now.getMonth() + 3);
      break;
    case "6months":
      expiresAt.setMonth(now.getMonth() + 6);
      break;
    case "1year":
      expiresAt.setFullYear(now.getFullYear() + 1);
      break;
  }
  return expiresAt;
}

export function NewOAuthTokenPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const tokenDialogRef = useDialogRef();
  const [token, setToken] = useState("");

  usePageTitle(__("New OAuth token"));

  const data = useLazyLoadQuery<NewOAuthTokenPageQuery>(pageQuery, {});

  const viewer = data.viewer;

  const supportedScopes = useMemo(
    () => [...data.oauth2ScopesSupported].sort(),
    [data.oauth2ScopesSupported],
  );

  const { formState, handleSubmit, register, control, watch, setValue }
    = useFormWithSchema(createSchema, {
      defaultValues: {
        name: new Date().toISOString().split("T")[0],
        expiresIn: "1year",
        scopes: [] as string[],
      },
    });

  const selectedScopes = watch("scopes");

  const [create, isCreating] = useMutationWithToasts<NewOAuthTokenPageCreateMutation>(
    createMutation,
    {
      successMessage: "OAuth token created successfully.",
      errorMessage: "Failed to create OAuth token",
    },
  );

  if (!viewer.canCreateOAuth2AccessToken) {
    throw new Error("forbidden");
  }

  const toggleScope = (scope: string, checked: boolean) => {
    const current = new Set(selectedScopes);
    if (checked) {
      current.add(scope);
    } else {
      current.delete(scope);
    }
    setValue("scopes", [...current], { shouldValidate: true });
  };

  const allScopesSelected
    = supportedScopes.length > 0
      && selectedScopes.length === supportedScopes.length;

  const toggleAllScopes = () => {
    setValue(
      "scopes",
      allScopesSelected ? [] : [...supportedScopes],
      { shouldValidate: true },
    );
  };

  const handleCreate = (formData: CreateFormData) => {
    const connectionID = ConnectionHandler.getConnectionID(
      viewer.id,
      "OAuthTokensPage_oauth2AccessTokens",
    );

    void create({
      variables: {
        input: {
          name: formData.name,
          expiresAt: computeExpiresAt(formData.expiresIn).toISOString(),
          scopes: formData.scopes,
        },
        connections: [connectionID],
      },
      onCompleted: (response) => {
        const newToken = response.createOAuth2AccessToken?.token;
        if (newToken) {
          setToken(newToken);
          tokenDialogRef.current?.open();
        }
      },
    });
  };

  const handleDone = () => {
    tokenDialogRef.current?.close();
    void navigate("/me/oauth-tokens");
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("OAuth tokens"),
            to: "/me/oauth-tokens",
          },
          { label: __("New token") },
        ]}
      />

      <PageHeader
        title={__("New OAuth token")}
        description={__(
          "Create a bearer token with scoped access to the Probo API.",
        )}
      />

      <Card padded>
        <form className="space-y-6" onSubmit={e => void handleSubmit(handleCreate)(e)}>
          <div className="max-w-xl space-y-6">
            <Field>
              <Label htmlFor="name">{__("Name")}</Label>
              <Input id="name" {...register("name")} />
              {formState.errors.name && (
                <p className="text-sm text-danger mt-1">
                  {formState.errors.name.message}
                </p>
              )}
            </Field>

            <Field>
              <Label htmlFor="expiresIn">{__("Expiration")}</Label>
              <Controller
                name="expiresIn"
                control={control}
                render={({ field }) => (
                  <Select
                    id="expiresIn"
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    <Option value="1month">{__("1 month")}</Option>
                    <Option value="3months">{__("3 months")}</Option>
                    <Option value="6months">{__("6 months")}</Option>
                    <Option value="1year">{__("1 year")}</Option>
                  </Select>
                )}
              />
            </Field>
          </div>

          <Field>
            <Label>{__("Scopes")}</Label>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 mt-2">
              {supportedScopes.map(scope => (
                <label key={scope} className="flex items-start gap-2">
                  <Checkbox
                    checked={selectedScopes.includes(scope)}
                    onChange={checked => toggleScope(scope, checked)}
                  />
                  <span className="min-w-0">
                    <span className="block font-medium">
                      {formatApiScopeLabel(scope)}
                    </span>
                    <span className="block text-sm text-txt-secondary break-all">
                      {scope}
                    </span>
                  </span>
                </label>
              ))}
            </div>
            {formState.errors.scopes && (
              <p className="text-sm text-danger mt-1">
                {formState.errors.scopes.message}
              </p>
            )}
          </Field>

          <div className="flex gap-3 max-w-xl">
            <Button type="submit" disabled={isCreating}>
              {__("Create token")}
            </Button>
            <Button
              type="button"
              variant="secondary"
              onClick={toggleAllScopes}
            >
              {allScopesSelected ? __("Deselect all") : __("Select all")}
            </Button>
            <Button variant="secondary" asChild>
              <Link to="/me/oauth-tokens">
                {__("Cancel")}
              </Link>
            </Button>
          </div>
        </form>
      </Card>

      <OAuthTokenCredentialsDialog
        dialogRef={tokenDialogRef}
        token={token}
        onDone={handleDone}
      />
    </div>
  );
}
