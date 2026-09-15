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

import { CheckIcon, CopyIcon, PlusIcon, WarningIcon } from "@phosphor-icons/react";
import { useCopy } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { FormEvent } from "react";
import { Suspense, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";

import type { ServiceAccountCredentialsDialog_serviceAccount$key } from "#/__generated__/iam/ServiceAccountCredentialsDialog_serviceAccount.graphql";
import type { ServiceAccountCredentialsDialogCreateMutation } from "#/__generated__/iam/ServiceAccountCredentialsDialogCreateMutation.graphql";
import type { ServiceAccountCredentialsDialogQuery } from "#/__generated__/iam/ServiceAccountCredentialsDialogQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { ScopePicker } from "./ScopePicker";
import { ServiceAccountCredentialList } from "./ServiceAccountCredentialList";

const serviceAccountCredentialsDialogFragment = graphql`
  fragment ServiceAccountCredentialsDialog_serviceAccount on ServiceAccount {
    id
  }
`;

export const serviceAccountCredentialsDialogQuery = graphql`
  query ServiceAccountCredentialsDialogQuery($serviceAccountId: ID!) {
    serviceAccount: node(id: $serviceAccountId) @required(action: THROW) {
      __typename
      ... on ServiceAccount {
        id
        name
        scopes
        canCreateCredential: permission(
          action: "iam:service-account-credential:create"
        )
        ...ServiceAccountCredentialList_serviceAccount
      }
    }
  }
`;

const createCredentialMutation = graphql`
  mutation ServiceAccountCredentialsDialogCreateMutation(
    $input: CreateServiceAccountCredentialInput!
    $connections: [ID!]!
  ) {
    createServiceAccountCredential(input: $input) {
      token
      serviceAccountCredentialEdge @prependEdge(connections: $connections) {
        node {
          id
          ...ServiceAccountCredentialListItem_credential
        }
      }
    }
  }
`;

interface ServiceAccountCredentialsDialogProps {
  open: boolean;
  serviceAccountKey: ServiceAccountCredentialsDialog_serviceAccount$key;
  onOpenChange: (open: boolean) => void;
}

export function ServiceAccountCredentialsDialog(
  props: ServiceAccountCredentialsDialogProps,
) {
  const { open, serviceAccountKey, onOpenChange } = props;
  const serviceAccount = useFragment(
    serviceAccountCredentialsDialogFragment,
    serviceAccountKey,
  );
  const [queryRef, loadQuery, disposeQuery]
    = useQueryLoader<ServiceAccountCredentialsDialogQuery>(
      serviceAccountCredentialsDialogQuery,
    );

  useEffect(() => {
    if (open) {
      loadQuery(
        { serviceAccountId: serviceAccount.id },
        { fetchPolicy: "store-and-network" },
      );
    } else {
      disposeQuery();
    }
  }, [disposeQuery, loadQuery, open, serviceAccount.id]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup className="max-w-4xl" lockScroll>
        <Suspense fallback={<ServiceAccountCredentialsDialogLoading />}>
          {queryRef && (
            <ServiceAccountCredentialsDialogContent
              queryRef={queryRef}
              onDone={() => onOpenChange(false)}
            />
          )}
        </Suspense>
      </DialogPopup>
    </Dialog>
  );
}

function ServiceAccountCredentialsDialogLoading() {
  const { t } = useTranslation("iam/organizations/service-accounts");
  return (
    <>
      <DialogHeader>
        <DialogTitle>{t("credentials.title")}</DialogTitle>
      </DialogHeader>
      <DialogBody className="h-56 animate-pulse rounded-3 bg-sand-3" />
    </>
  );
}

interface ServiceAccountCredentialsDialogContentProps {
  queryRef: PreloadedQuery<ServiceAccountCredentialsDialogQuery>;
  onDone: () => void;
}

function ServiceAccountCredentialsDialogContent(
  props: ServiceAccountCredentialsDialogContentProps,
) {
  const { queryRef, onDone } = props;
  const { t } = useTranslation("iam/organizations/service-accounts");
  const data = usePreloadedQuery<ServiceAccountCredentialsDialogQuery>(
    serviceAccountCredentialsDialogQuery,
    queryRef,
  );
  if (data.serviceAccount.__typename !== "ServiceAccount") {
    throw new Error("Relay node is not a service account");
  }

  const serviceAccount = data.serviceAccount;
  const [showCreate, setShowCreate] = useState(false);
  const [token, setToken] = useState<string | null>(null);

  function done() {
    setToken(null);
    onDone();
  }

  if (token != null) {
    return <CredentialToken token={token} onDone={done} />;
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {t("credentials.titleWithName", {
            name: serviceAccount.name,
          })}
        </DialogTitle>
      </DialogHeader>
      <DialogBody className="space-y-5">
        <div className="flex items-center justify-between gap-4">
          <Heading size={4}>{t("credentials.listTitle")}</Heading>
          {serviceAccount.canCreateCredential && !showCreate && (
            <Button
              size={1}
              variant="soft"
              iconStart={<PlusIcon aria-hidden />}
              onClick={() => setShowCreate(true)}
            >
              {t("credentials.create")}
            </Button>
          )}
        </div>
        {showCreate && (
          <CreateCredentialForm
            scopes={[...serviceAccount.scopes].sort()}
            serviceAccountId={serviceAccount.id}
            onCancel={() => setShowCreate(false)}
            onCreated={setToken}
          />
        )}
        <ServiceAccountCredentialList serviceAccountKey={serviceAccount} />
      </DialogBody>
      <DialogFooter>
        <Button variant="soft" onClick={done}>
          {t("actions.done")}
        </Button>
      </DialogFooter>
    </>
  );
}

interface CreateCredentialFormProps {
  scopes: readonly string[];
  serviceAccountId: string;
  onCancel: () => void;
  onCreated: (token: string) => void;
}

function CreateCredentialForm(props: CreateCredentialFormProps) {
  const { scopes: supportedScopes, serviceAccountId, onCancel, onCreated } = props;
  const { t } = useTranslation("iam/organizations/service-accounts");
  const defaultExpiry = new Date();
  defaultExpiry.setFullYear(defaultExpiry.getFullYear() + 1);
  const [name, setName] = useState("");
  const [expiresAt, setExpiresAt] = useState(
    defaultExpiry.toISOString().slice(0, 10),
  );
  const [scopes, setScopes] = useState<string[]>([]);
  const [scopeError, setScopeError] = useState(false);
  const [createCredential, isCreating]
    = useMutation<ServiceAccountCredentialsDialogCreateMutation>(
      createCredentialMutation,
      {
        successMessage: t("messages.credentialCreated"),
        errorToast: t("errors.createCredential"),
      },
    );

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (scopes.length === 0) {
      setScopeError(true);
      return;
    }

    const connectionId = ConnectionHandler.getConnectionID(
      serviceAccountId,
      "ServiceAccountCredentialList_credentials",
    );
    try {
      const response = await createCredential({
        variables: {
          input: {
            serviceAccountId,
            name: name.trim(),
            scopes,
            expiresAt: new Date(`${expiresAt}T23:59:59.999Z`).toISOString(),
          },
          connections: [connectionId],
        },
      });
      const result = response.createServiceAccountCredential;
      if (result == null) {
        throw new Error("Credential creation returned no payload");
      }
      onCreated(result.token);
    } catch {
      return;
    }
  }

  const today = new Date().toISOString().slice(0, 10);

  return (
    <form
      className="space-y-4 rounded-3 border border-sand-6 bg-sand-2 p-4"
      onSubmit={event => void submit(event)}
    >
      <Heading size={3}>{t("credentials.createTitle")}</Heading>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Field label={t("fields.credentialName")}>
          <TextField
            required
            maxLength={120}
            value={name}
            onValueChange={setName}
          />
        </Field>
        <Field label={t("fields.expiresAt")}>
          <TextField
            type="date"
            required
            min={today}
            value={expiresAt}
            onValueChange={setExpiresAt}
          />
        </Field>
      </div>
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
      <div className="flex justify-end gap-2">
        <Button variant="soft" disabled={isCreating} onClick={onCancel}>
          {t("actions.cancel")}
        </Button>
        <Button type="submit" loading={isCreating}>
          {t("credentials.create")}
        </Button>
      </div>
    </form>
  );
}

interface CredentialTokenProps {
  token: string;
  onDone: () => void;
}

function CredentialToken({ token, onDone }: CredentialTokenProps) {
  const { t } = useTranslation("iam/organizations/service-accounts");
  const [isCopied, copy] = useCopy();

  return (
    <>
      <DialogHeader>
        <DialogTitle>{t("token.title")}</DialogTitle>
      </DialogHeader>
      <DialogBody className="space-y-4">
        <Callout color="amber" icon={<WarningIcon weight="fill" />}>
          {t("token.warning")}
        </Callout>
        <div className="flex items-start gap-3 rounded-3 border border-sand-6 bg-sand-2 p-4">
          <Code className="min-w-0 flex-1 break-all" size={2}>
            {token}
          </Code>
          <Button
            size={1}
            variant="soft"
            iconStart={isCopied
              ? <CheckIcon aria-hidden />
              : <CopyIcon aria-hidden />}
            onClick={() => copy(token)}
          >
            {isCopied ? t("actions.copied") : t("actions.copy")}
          </Button>
        </div>
        <Text size={2} color="faint">
          {t("token.description")}
        </Text>
      </DialogBody>
      <DialogFooter>
        <Button onClick={onDone}>{t("actions.done")}</Button>
      </DialogFooter>
    </>
  );
}
