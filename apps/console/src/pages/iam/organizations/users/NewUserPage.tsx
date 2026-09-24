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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { toFieldErrors } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { NewUserPageMutation } from "#/__generated__/iam/NewUserPageMutation.graphql";
import type { NewUserPageQuery } from "#/__generated__/iam/NewUserPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";

import { UserForm, type UserFormValues } from "./_components/UserForm";
import { newUserPage } from "./variants";

export const newUserPageQuery = graphql`
  query NewUserPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateUser: permission(
          action: "iam:membership-profile:create"
          attributes: { target_role: "VIEWER" }
        )
      }
    }
  }
`;

const createUserMutation = graphql`
  mutation NewUserPageMutation($input: CreateUserInput!) {
    createUser(input: $input) {
      profileEdge {
        node {
          id
        }
      }
    }
  }
`;

const userFormErrorKeys = {
  contract_start_date: "contractStart",
  contract_end_date: "contractEnd",
} as const;

function mapUserFormErrors(errors: Record<string, string>) {
  const next: Record<string, string> = {};
  for (const [key, message] of Object.entries(errors)) {
    const mapped = userFormErrorKeys[key as keyof typeof userFormErrorKeys];
    next[mapped ?? key] = message;
  }
  return next;
}

interface NewUserPageProps {
  queryRef: PreloadedQuery<NewUserPageQuery>;
}

export function NewUserPage({ queryRef }: NewUserPageProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const title = t("newUserPage.title");
  usePageTitle(title);

  const { organization } = usePreloadedQuery<NewUserPageQuery>(
    newUserPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization" || !organization.canCreateUser) {
    throw new NotFoundError(t("newUserPage.notFound"));
  }

  const { root, back, intro } = newUserPage();
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [createUser, isCreating] = useMutation<NewUserPageMutation>(
    createUserMutation,
    {
      successMessage: t("userForm.messages.created"),
      errorToast: t("userForm.errors.create"),
    },
  );

  function handleSubmit(values: UserFormValues) {
    void createUser({
      variables: {
        input: {
          organizationId,
          fullName: values.fullName,
          emailAddress: values.emailAddress,
          role: values.role,
          kind: values.kind,
          position: values.position,
          additionalEmailAddresses: values.additionalEmailAddresses,
          contract: values.contract,
        },
      },
      onCompleted(response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(mapUserFormErrors(fieldErrors));
          return;
        }
        const profileId = response.createUser?.profileEdge.node.id;
        if (profileId != null && profileId !== "") {
          void navigate(`/organizations/${organizationId}/settings/users/${profileId}`);
        }
      },
    }).catch(() => {
      // Field errors are mapped in onCompleted; other failures toast.
    });
  }

  return (
    <div className={root()}>
      <Link
        to=".."
        size={2}
        color="neutral"
        underline={false}
        iconStart={<CaretLeftIcon />}
        className={back()}
      >
        {t("userPage.back")}
      </Link>
      <div className={intro()}>
        <Heading level={1} size={6} weight="medium" highContrast>
          {title}
        </Heading>
        <Text size={2} color="faint">
          {t("newUserPage.description")}
        </Text>
      </div>
      <UserForm
        disabled={isCreating}
        errors={errors}
        onErrorsChange={setErrors}
        onSubmit={handleSubmit}
      />
    </div>
  );
}
