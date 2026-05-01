// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import {
  type CloudAccountProvider,
  formatError,
  type GraphQLError,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  type DialogRef,
  Field,
  IconShield,
  useToast,
} from "@probo/ui";
import { useRef, useState } from "react";
import { graphql, useMutation } from "react-relay";

import type { CloudAccountReconnectDialogMutation } from "#/__generated__/core/CloudAccountReconnectDialogMutation.graphql";

const rotateMutation = graphql`
  mutation CloudAccountReconnectDialogMutation(
    $input: RotateCloudAccountCredentialsInput!
  ) {
    rotateCloudAccountCredentials(input: $input) {
      cloudAccount {
        id
        status
        lastVerifiedAt
        lastProbeError
      }
      verifyStatus
      lastProbeError
    }
  }
`;

type Props = {
  ref: DialogRef;
  cloudAccountId: string;
  provider: CloudAccountProvider;
  // For Azure rotation, the existing tenant_id and client_id must be
  // resubmitted alongside the new client_secret. They are NOT secret.
  azureTenantId?: string | null;
  azureClientId?: string | null;
  // For AWS rotation, the original externalId is required so the trust
  // policy continues to match.
  awsExternalId?: string | null;
};

export function CloudAccountReconnectDialog(props: Props) {
  const {
    ref,
    cloudAccountId,
    provider,
    azureTenantId,
    azureClientId,
    awsExternalId,
  } = props;
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [rotate, isRotating]
    = useMutation<CloudAccountReconnectDialogMutation>(rotateMutation);

  // AWS: only a Role ARN; non-secret, lives in state.
  const [roleArn, setRoleArn] = useState("");

  // GCP/Azure: secret bytes live in a ref; never enter state nor a Relay
  // payload. They are POSTed via the dedicated upload endpoint.
  const credentialRef = useRef<string>("");
  const credentialNodeRef = useRef<
    HTMLTextAreaElement | HTMLInputElement | null
  >(null);
  const [hasCredential, setHasCredential] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const clearCredential = () => {
    credentialRef.current = "";
    if (credentialNodeRef.current) {
      credentialNodeRef.current.value = "";
    }
    setHasCredential(false);
  };

  const handleClose = () => {
    clearCredential();
    setRoleArn("");
    setErrorMessage(null);
    ref.current?.close();
  };

  const handleSubmitAws = () => {
    if (!roleArn.trim()) return;
    setErrorMessage(null);
    rotate({
      variables: {
        input: {
          cloudAccountId,
          provider: "AWS",
          credentialKind: "AWS_ASSUME_ROLE",
          awsRoleArn: roleArn,
          awsExternalId: awsExternalId ?? undefined,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          setErrorMessage(errors.map(e => e.message).join(", "));
          return;
        }
        toast({
          title: __("Success"),
          description: __("AWS role rotated."),
          variant: "success",
        });
        handleClose();
      },
      onError(error) {
        setErrorMessage(
          formatError(
            __("Could not rotate AWS credentials"),
            error as GraphQLError,
          ),
        );
      },
    });
  };

  const handleSubmitUploadFlow = (
    credentialKind: "GCP_SERVICE_ACCOUNT_KEY" | "AZURE_CLIENT_SECRET",
  ) => {
    const payload = credentialRef.current;
    if (!payload.trim()) return;
    setErrorMessage(null);
    if (credentialKind === "GCP_SERVICE_ACCOUNT_KEY") {
      try {
        JSON.parse(payload);
      } catch {
        setErrorMessage(__("Service-account key body is not valid JSON."));
        return;
      }
    }
    // Step 1: rotate metadata via GraphQL (no secrets travel here).
    rotate({
      variables: {
        input: {
          cloudAccountId,
          provider,
          credentialKind,
          azureTenantId:
            provider === "AZURE" ? (azureTenantId ?? undefined) : undefined,
          azureClientId:
            provider === "AZURE" ? (azureClientId ?? undefined) : undefined,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          setErrorMessage(errors.map(e => e.message).join(", "));
          return;
        }
        // Step 2: upload secret via dedicated multipart endpoint.
        const formData = new FormData();
        formData.append("cloud_account_id", cloudAccountId);
        formData.append("payload", payload);
        fetch("/api/console/v1/cloud-accounts/credentials/upload", {
          method: "POST",
          body: formData,
          credentials: "same-origin",
        })
          .then(async (res) => {
            clearCredential();
            if (!res.ok) {
              const text = await res.text().catch(() => "");
              setErrorMessage(text || __("Credential upload failed."));
              return;
            }
            toast({
              title: __("Success"),
              description: __("Credentials rotated."),
              variant: "success",
            });
            handleClose();
          })
          .catch((error) => {
            clearCredential();
            setErrorMessage(
              error instanceof Error
                ? error.message
                : __("Credential upload failed."),
            );
          });
      },
      onError(error) {
        setErrorMessage(
          formatError(
            __("Could not rotate credentials"),
            error as GraphQLError,
          ),
        );
      },
    });
  };

  return (
    <Dialog
      ref={ref}
      title={__("Reconnect cloud account")}
      className="max-w-lg"
      onClose={handleClose}
    >
      <DialogContent padded className="space-y-4">
        {errorMessage && (
          <div className="rounded-md border border-border-danger bg-danger p-3 text-sm text-txt-danger">
            {errorMessage}
          </div>
        )}
        <div className="flex items-start gap-2 rounded-md border border-border-low bg-subtle p-3 text-xs text-txt-secondary">
          <IconShield size={14} />
          <span>
            {__(
              "Rotating credentials replaces the stored secret. The previous credential is overwritten and immediately invalid.",
            )}
          </span>
        </div>

        {provider === "AWS" && (
          <Field
            label={__("New role ARN")}
            placeholder="arn:aws:iam::111122223333:role/probo-readonly"
            value={roleArn}
            onValueChange={setRoleArn}
            autoComplete="off"
            spellCheck={false}
          />
        )}

        {provider === "GCP" && (
          <label className="flex flex-col gap-1">
            <span className="text-sm font-medium">
              {__("Service-account JSON key")}
            </span>
            <textarea
              ref={(node) => {
                credentialNodeRef.current = node;
              }}
              className="font-mono text-xs rounded-md border border-border-low p-3 min-h-[160px]"
              autoComplete="off"
              spellCheck={false}
              data-1p-ignore="true"
              style={{ WebkitTextSecurity: "disc" } as React.CSSProperties}
              placeholder={__("Paste the new JSON key body here")}
              onChange={(e) => {
                credentialRef.current = e.target.value;
                setHasCredential(e.target.value.trim().length > 0);
              }}
            />
          </label>
        )}

        {provider === "AZURE" && (
          <label className="flex flex-col gap-1">
            <span className="text-sm font-medium">
              {__("New client secret")}
            </span>
            <input
              ref={(node) => {
                credentialNodeRef.current = node;
              }}
              type="text"
              autoComplete="off"
              spellCheck={false}
              data-1p-ignore="true"
              style={{ WebkitTextSecurity: "disc" } as React.CSSProperties}
              className="font-mono text-xs rounded-md border border-border-low p-3"
              placeholder={__("Paste the new client secret value")}
              onChange={(e) => {
                credentialRef.current = e.target.value;
                setHasCredential(e.target.value.trim().length > 0);
              }}
            />
          </label>
        )}
      </DialogContent>
      <DialogFooter>
        {provider === "AWS" && (
          <Button
            disabled={!roleArn.trim() || isRotating}
            onClick={handleSubmitAws}
          >
            {isRotating ? __("Rotating…") : __("Rotate role")}
          </Button>
        )}
        {provider === "GCP" && (
          <Button
            disabled={!hasCredential || isRotating}
            onClick={() => handleSubmitUploadFlow("GCP_SERVICE_ACCOUNT_KEY")}
          >
            {isRotating ? __("Rotating…") : __("Rotate key")}
          </Button>
        )}
        {provider === "AZURE" && (
          <Button
            disabled={!hasCredential || isRotating}
            onClick={() => handleSubmitUploadFlow("AZURE_CLIENT_SECRET")}
          >
            {isRotating ? __("Rotating…") : __("Rotate secret")}
          </Button>
        )}
      </DialogFooter>
    </Dialog>
  );
}
