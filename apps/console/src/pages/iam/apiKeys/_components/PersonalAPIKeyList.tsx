import { useState } from "react";
import { Controller } from "react-hook-form";
import {
  ConnectionHandler,
  graphql,
  useFragment,
  useMutation,
} from "react-relay";
import type { RecordSourceSelectorProxy } from "relay-runtime";
import { z } from "zod";
import { useTranslate } from "@probo/i18n";
import { formatError, type GraphQLError } from "@probo/helpers";
import {
  Breadcrumb,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Label,
  Option,
  Select,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import type { PersonalAPIKeyListFragment$key } from "./__generated__/PersonalAPIKeyListFragment.graphql";
import type { PersonalAPIKeyListCreateMutation } from "./__generated__/PersonalAPIKeyListCreateMutation.graphql";
import type { PersonalAPIKeyListRevokeMutation } from "./__generated__/PersonalAPIKeyListRevokeMutation.graphql";
import type { PersonalAPIKeyListRevealTokenMutation } from "./__generated__/PersonalAPIKeyListRevealTokenMutation.graphql";
import { PersonalAPIKeysTable } from "./PersonalAPIKeysTable";
import { PersonalAPIKeyTokenDialog } from "./PersonalAPIKeyTokenDialog";

const fragment = graphql`
  fragment PersonalAPIKeyListFragment on Identity {
    id

    personalAPIKeys(first: 1000)
      @required(action: THROW)
      @connection(key: "PersonalAPIKeyListFragment_personalAPIKeys") {
      edges @required(action: THROW) {
        node {
          id
          name
          createdAt
          expiresAt
          lastUsedAt
        }
      }
    }
  }
`;

const createMutation = graphql`
  mutation PersonalAPIKeyListCreateMutation(
    $input: CreatePersonalAPIKeyInput!
    $connections: [ID!]!
  ) {
    createPersonalAPIKey(input: $input) {
      personalAPIKeyEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          createdAt
          expiresAt
          lastUsedAt
        }
      }
      token
    }
  }
`;

const revokeMutation = graphql`
  mutation PersonalAPIKeyListRevokeMutation(
    $input: RevokePersonalAPIKeyInput!
  ) {
    revokePersonalAPIKey(input: $input) {
      success
    }
  }
`;

const revealTokenMutation = graphql`
  mutation PersonalAPIKeyListRevealTokenMutation(
    $input: RevealPersonalAPIKeyTokenInput!
  ) {
    revealPersonalAPIKeyToken(input: $input) {
      token
    }
  }
`;

const createSchema = z.object({
  name: z.string().min(1, "Name is required"),
  expiresIn: z.enum(["1month", "3months", "6months", "1year"]),
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

export function PersonalAPIKeyList(props: {
  fKey: PersonalAPIKeyListFragment$key;
}) {
  const { fKey } = props;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const createDialogRef = useDialogRef();
  const tokenDialogRef = useDialogRef();

  const [token, setToken] = useState<string>("");

  const viewer = useFragment(fragment, fKey);

  const keys = viewer.personalAPIKeys.edges.map(({ node }) => node);

  const { formState, handleSubmit, register, control, reset, watch } =
    useFormWithSchema(createSchema, {
      defaultValues: {
        name: new Date().toISOString().split("T")[0],
        expiresIn: "1month",
      },
    });

  watch();

  const [createCommit, isCreating] =
    useMutation<PersonalAPIKeyListCreateMutation>(createMutation);
  const [revokeCommit] =
    useMutation<PersonalAPIKeyListRevokeMutation>(revokeMutation);
  const [revealTokenCommit, isRevealingToken] =
    useMutation<PersonalAPIKeyListRevealTokenMutation>(revealTokenMutation);

  const handleCreate = (data: CreateFormData) => {
    const expiresAt = computeExpiresAt(data.expiresIn);
    const connectionID = ConnectionHandler.getConnectionID(
      viewer.id,
      "PersonalAPIKeyListFragment_personalAPIKeys"
    );

    createCommit({
      variables: {
        input: {
          name: data.name,
          expiresAt: expiresAt.toISOString(),
          // API keys are no longer linked to organizations; keep schema compatibility.
          organizationIds: [],
        },
        connections: [connectionID],
      },
      onCompleted: (response) => {
        toast({
          title: __("Success"),
          description: __("API key created successfully."),
          variant: "success",
        });
        const newToken = response.createPersonalAPIKey?.token;
        if (newToken) {
          setToken(newToken);
          tokenDialogRef.current?.open();
        }
        createDialogRef.current?.close();
        reset();
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to create API key."), error),
          variant: "error",
        });
      },
    });
  };

  const handleRevoke = (key: { id: string; name: string }) => {
    confirm(
      async () => {
        await new Promise<void>((resolve, reject) => {
          revokeCommit({
            variables: {
              input: { tokenId: key.id },
            },
            updater: (store: RecordSourceSelectorProxy) => {
              const viewerRecord = store.getRoot().getLinkedRecord("viewer");
              if (!viewerRecord) return;
              const connection = ConnectionHandler.getConnection(
                viewerRecord,
                "PersonalAPIKeyListFragment_personalAPIKeys"
              );
              if (connection) {
                ConnectionHandler.deleteNode(connection, key.id);
              }
            },
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                toast({
                  title: __("Error"),
                  description: formatError(
                    __("Failed to revoke API key."),
                    errors as GraphQLError[]
                  ),
                  variant: "error",
                });
                reject(errors);
                return;
              }
              toast({
                title: __("Success"),
                description: __("API key revoked successfully."),
                variant: "success",
              });
              resolve();
            },
            onError: (error) => {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to revoke API key."),
                  error
                ),
                variant: "error",
              });
              reject(error);
            },
          });
        });
      },
      {
        title: __("Revoke API Key"),
        message: __(
          `Are you sure you want to revoke the API key "${key.name}"? This action cannot be undone.`
        ),
        label: __("Revoke"),
        variant: "danger",
      }
    );
  };

  const handleShowToken = (key: { id: string; name: string }) => {
    revealTokenCommit({
      variables: {
        input: {
          tokenId: key.id,
        },
      },
      onCompleted: (response, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to reveal API key token."),
              errors as any
            ),
            variant: "error",
          });
          return;
        }

        const tokenValue = response.revealPersonalAPIKeyToken?.token;
        if (!tokenValue) {
          toast({
            title: __("Error"),
            description: __("No token returned."),
            variant: "error",
          });
          return;
        }

        setToken(tokenValue);
        tokenDialogRef.current?.open();
      },
      onError: (error: Error) => {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to reveal API key token."),
            error
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <>
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <h2 className="text-base font-medium">{__("API Keys")}</h2>
          <Button onClick={() => createDialogRef.current?.open()}>
            {__("Create API Key")}
          </Button>
        </div>

        {keys.length === 0 ? (
          <Card padded>
            <div className="text-center py-12">
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                {__("No API keys")}
              </h3>
              <p className="text-gray-600 mb-6">
                {__("Create an API key to authenticate programmatic access.")}
              </p>
            </div>
          </Card>
        ) : (
          <Card padded>
            <PersonalAPIKeysTable
              keys={keys}
              onRevoke={handleRevoke}
              onShowToken={handleShowToken}
              isShowingToken={isRevealingToken}
            />
          </Card>
        )}
      </div>

      <Dialog
        ref={createDialogRef}
        title={<Breadcrumb items={[__("API Keys"), __("Create")]} />}
        onClose={() => reset()}
      >
        <form onSubmit={handleSubmit(handleCreate)}>
          <DialogContent padded className="space-y-5">
            <Field error={formState.errors.name?.message}>
              <Label>{__("Name")}</Label>
              <Input
                {...register("name")}
                placeholder={__("e.g., Production API Key")}
              />
            </Field>

            <Field error={formState.errors.expiresIn?.message}>
              <Label>{__("Expires In")}</Label>
              <Controller
                control={control}
                name="expiresIn"
                render={({ field }) => (
                  <Select
                    {...field}
                    onValueChange={field.onChange}
                    value={field.value}
                  >
                    <Option value="1month">{__("1 Month")}</Option>
                    <Option value="3months">{__("3 Months")}</Option>
                    <Option value="6months">{__("6 Months")}</Option>
                    <Option value="1year">{__("1 Year")}</Option>
                  </Select>
                )}
              />
            </Field>
          </DialogContent>
          <DialogFooter>
            <Button type="submit" disabled={isCreating}>
              {isCreating ? __("Creating...") : __("Create")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>

      <PersonalAPIKeyTokenDialog
        dialogRef={tokenDialogRef}
        token={token}
        onDone={() => {
          tokenDialogRef.current?.close();
          setToken("");
        }}
      />
    </>
  );
}
