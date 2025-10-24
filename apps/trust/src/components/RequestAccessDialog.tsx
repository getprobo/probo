import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogTitle,
  Field,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { sprintf } from "@probo/helpers";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { useTrustCenter } from "/hooks/useTrustCenter";
import { type FormEventHandler, type PropsWithChildren } from "react";
import { useIsAuthenticated } from "/hooks/useIsAuthenticated.ts";

type Props = PropsWithChildren<{
  documentId?: string;
  reportId?: string;
  trustCenterFileId?: string;
  onSuccess?: () => void;
}>;

const schema = z.object({
  name: z.string(),
  email: z.string().email(),
});

export function RequestAccessDialog({
  children,
  documentId,
  reportId,
  trustCenterFileId,
  onSuccess,
}: Props) {
  const trustCenter = useTrustCenter();
  const { toast } = useToast();
  const { __ } = useTranslate();
  const { handleSubmit, register } = useFormWithSchema(schema, {
    defaultValues: {
      name: "",
      email: "",
    },
  });
  const isAuthenticated = useIsAuthenticated();
  const dialogRef = useDialogRef();
  const [commitMutation, isMutating] = useMutation({ documentId, reportId, trustCenterFileId });

  const submitCallback = (data: z.infer<typeof schema> | null) => {
    commitMutation(data)
      .then(() => {
        onSuccess?.();
        toast({
          title: __("Success"),
          description: __("Access request submitted successfully"),
          variant: "success",
        });
        dialogRef.current?.close();
      })
      .catch((error) => {
        console.error(error);
        toast({
          title: __("Error"),
          description: __("Cannot request access"),
          variant: "error",
        });
      });
  };

  const onSubmit: FormEventHandler<HTMLFormElement> = isAuthenticated
    ? (e) => {
        e.preventDefault();
        submitCallback(null);
      }
    : handleSubmit(submitCallback);
  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      className="max-w-[500px] text-txt-primary"
    >
      <form onSubmit={onSubmit}>
        <DialogTitle className="text-2xl font-semibold mb-4 pt-4 md:pt-8 px-4 md:px-8">
          {__("Request access to documentation")}
        </DialogTitle>
        <DialogContent className="px-4 md:px-8 pb-4 md:pb-8 text-txt-primary">
          <p className="text-txt-secondary mb-4">
            {sprintf(
              __(
                "Request access to %s's Trust Center. Your request will be reviewed and you will receive an email notification with access instructions if approved.",
              ),
              trustCenter.organization.name,
            )}
          </p>
          {!isAuthenticated && (
            <div className="space-y-4">
              <Field
                label={__("Full name")}
                placeholder="John Doe"
                {...register("name")}
                type="text"
              />
              <Field
                label={__("Email")}
                placeholder="john.doe@acme.com"
                {...register("email")}
                type="email"
              />
            </div>
          )}
        </DialogContent>
        <DialogFooter>
          <Button disabled={isMutating} type="submit">
            {__("Continue")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

const requestAccessMutation = graphql`
  mutation RequestAccessDialogMutation($input: RequestAllAccessesInput!) {
    requestAllAccesses(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

const requestDocumentAccessMutation = graphql`
  mutation RequestAccessDialogDocumentMutation(
    $input: RequestDocumentAccessInput!
  ) {
    requestDocumentAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

const requestReportAccessMutation = graphql`
  mutation RequestAccessDialogReportMutation(
    $input: RequestReportAccessInput!
  ) {
    requestReportAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

const requestTrustCenterFileAccessMutation = graphql`
  mutation RequestAccessDialogTrustCenterFileMutation(
    $input: RequestTrustCenterFileAccessInput!
  ) {
    requestTrustCenterFileAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

/**
 * Use the correct mutation using the shape
 */
function useMutation({
  documentId,
  reportId,
  trustCenterFileId,
}: Pick<Props, "documentId" | "reportId" | "trustCenterFileId">): [
  (data: z.infer<typeof schema> | null) => Promise<unknown>,
  boolean,
] {
  const trustCenter = useTrustCenter();
  const [commitRequestAccess, isRequestingAccess] = useMutationWithToasts(
    requestAccessMutation,
  );
  const [commitRequestDocumentAccess, isRequestingDocumentAccess] =
    useMutationWithToasts(requestDocumentAccessMutation);
  const [commitRequestReportAccess, isRequestingReportAccess] =
    useMutationWithToasts(requestReportAccessMutation);
  const [commitRequestTrustCenterFileAccess, isRequestingTrustCenterFileAccess] =
    useMutationWithToasts(requestTrustCenterFileAccessMutation);

  if (trustCenterFileId) {
    return [
      (data) => {
        return commitRequestTrustCenterFileAccess({
          variables: {
            input: {
              trustCenterId: trustCenter.id,
              trustCenterFileId: trustCenterFileId,
              ...data,
            },
          },
        });
      },
      isRequestingTrustCenterFileAccess,
    ];
  } else if (reportId) {
    return [
      (data) => {
        return commitRequestReportAccess({
          variables: {
            input: {
              trustCenterId: trustCenter.id,
              reportId: reportId,
              ...data,
            },
          },
        });
      },
      isRequestingReportAccess,
    ];
  } else if (documentId) {
    return [
      (data) => {
        return commitRequestDocumentAccess({
          variables: {
            input: {
              trustCenterId: trustCenter.id,
              documentId: documentId,
              ...data,
            },
          },
        });
      },
      isRequestingDocumentAccess,
    ];
  } else {
    return [
      (data) => {
        return commitRequestAccess({
          variables: {
            input: {
              trustCenterId: trustCenter.id,
              ...data,
            },
          },
        });
      },
      isRequestingAccess,
    ];
  }
}
