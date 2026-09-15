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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import type { FormEvent } from "react";
import { Suspense, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";

import type { CreateServiceAccountDialogCreateMutation } from "#/__generated__/iam/CreateServiceAccountDialogCreateMutation.graphql";
import type { CreateServiceAccountDialogQuery } from "#/__generated__/iam/CreateServiceAccountDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { ScopePicker } from "./ScopePicker";

export const createServiceAccountDialogQuery = graphql`
  query CreateServiceAccountDialogQuery {
    oauth2ScopesSupported
  }
`;

const createServiceAccountMutation = graphql`
  mutation CreateServiceAccountDialogCreateMutation(
    $input: CreateServiceAccountInput!
    $connections: [ID!]!
  ) {
    createServiceAccount(input: $input) {
      serviceAccountEdge @appendEdge(connections: $connections) {
        node {
          id
          ...ServiceAccountListItem_serviceAccount
        }
      }
    }
  }
`;

interface CreateServiceAccountDialogProps {
  connectionId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateServiceAccountDialog(
  props: CreateServiceAccountDialogProps,
) {
  const { connectionId, open, onOpenChange } = props;
  const [queryRef, loadQuery, disposeQuery]
    = useQueryLoader<CreateServiceAccountDialogQuery>(
      createServiceAccountDialogQuery,
    );

  useEffect(() => {
    if (open) {
      loadQuery({}, { fetchPolicy: "store-or-network" });
    } else {
      disposeQuery();
    }
  }, [disposeQuery, loadQuery, open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup className="max-w-3xl">
        <Suspense fallback={<CreateServiceAccountDialogLoading />}>
          {queryRef && (
            <CreateServiceAccountDialogForm
              connectionId={connectionId}
              queryRef={queryRef}
              onCreated={() => onOpenChange(false)}
            />
          )}
        </Suspense>
      </DialogPopup>
    </Dialog>
  );
}

function CreateServiceAccountDialogLoading() {
  const { t } = useTranslation("iam/organizations/service-accounts");
  return (
    <>
      <DialogHeader>
        <DialogTitle>{t("create.title")}</DialogTitle>
      </DialogHeader>
      <DialogBody className="h-48 animate-pulse rounded-3 bg-sand-3" />
    </>
  );
}

interface CreateServiceAccountDialogFormProps {
  connectionId: string;
  queryRef: PreloadedQuery<CreateServiceAccountDialogQuery>;
  onCreated: () => void;
}

function CreateServiceAccountDialogForm(
  props: CreateServiceAccountDialogFormProps,
) {
  const { connectionId, queryRef, onCreated } = props;
  const { t } = useTranslation("iam/organizations/service-accounts");
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<CreateServiceAccountDialogQuery>(
    createServiceAccountDialogQuery,
    queryRef,
  );
  const supportedScopes = [...data.oauth2ScopesSupported].sort();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [scopes, setScopes] = useState<string[]>([]);
  const [scopeError, setScopeError] = useState(false);
  const [createServiceAccount, isCreating]
    = useMutation<CreateServiceAccountDialogCreateMutation>(
      createServiceAccountMutation,
      {
        successMessage: t("messages.created"),
        errorToast: t("errors.create"),
      },
    );

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (scopes.length === 0) {
      setScopeError(true);
      return;
    }

    try {
      await createServiceAccount({
        variables: {
          input: {
            organizationId,
            name: name.trim(),
            description: description.trim() || null,
            scopes,
          },
          connections: [connectionId],
        },
      });
      onCreated();
    } catch {
      return;
    }
  }

  return (
    <form onSubmit={event => void submit(event)}>
      <DialogHeader>
        <DialogTitle>{t("create.title")}</DialogTitle>
      </DialogHeader>
      <DialogBody className="space-y-5">
        <Field label={t("fields.name")}>
          <TextField
            name="name"
            required
            maxLength={120}
            value={name}
            onValueChange={setName}
          />
        </Field>
        <Field label={t("fields.description")}>
          <Textarea
            name="description"
            maxLength={500}
            value={description}
            onChange={event => setDescription(event.target.value)}
          />
        </Field>
        <div>
          <ScopePicker
            disabled={isCreating}
            scopes={supportedScopes}
            selectedScopes={scopes}
            onChange={(nextScopes) => {
              setScopes(nextScopes);
              setScopeError(nextScopes.length === 0);
            }}
          />
          {scopeError && (
            <p className="mt-2 text-1 text-red-11">
              {t("validation.scopeRequired")}
            </p>
          )}
        </div>
      </DialogBody>
      <DialogFooter>
        <DialogClose render={<Button variant="soft" />}>
          {t("actions.cancel")}
        </DialogClose>
        <Button type="submit" loading={isCreating}>
          {t("actions.create")}
        </Button>
      </DialogFooter>
    </form>
  );
}
