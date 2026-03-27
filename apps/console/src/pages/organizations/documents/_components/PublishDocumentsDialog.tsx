import { formatError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconWarning,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { PublishDocumentsDialogMutation } from "#/__generated__/core/PublishDocumentsDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

type Props = {
  documentIds: string[];
  children: ReactNode;
  onSave: () => void;
};

const documentsPublishMutation = graphql`
  mutation PublishDocumentsDialogMutation(
    $input: BulkPublishDocumentVersionsInput!
  ) {
    bulkPublishDocumentVersions(input: $input) {
      documentVersions {
        id
      }
      documents {
        id
        ...DocumentListItemFragment
      }
    }
  }
`;

export function PublishDocumentsDialog({
  documentIds,
  children,
  onSave,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const schema = z.object({
    changelog: z.string().min(1, __("Changelog is required")),
  });

  const [publishMutation, isPublishing] = useMutation<PublishDocumentsDialogMutation>(documentsPublishMutation);

  const {
    handleSubmit,
    register,
    formState: { errors },
  } = useFormWithSchema(schema, {
    defaultValues: {
      changelog: "",
    },
  });

  const onSubmit = (data: z.infer<typeof schema>) => {
    publishMutation({
      variables: {
        input: {
          documentIds,
          changelog: data.changelog,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to publish documents"), errors),
            variant: "error",
          });
        } else {
          toast({
            title: __("Success"),
            description: sprintf(__("%s documents published"), documentIds.length),
            variant: "success",
          });
          dialogRef.current?.close();
          onSave();
        }
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      className="max-w-xl"
      ref={dialogRef}
      trigger={children}
      title={__("Publish documents")}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded>
          <div className="space-y-4">
            <div className="flex items-start gap-2 rounded-lg bg-bg-warning/10 border border-border-warning p-3">
              <IconWarning size={16} className="text-txt-warning shrink-0 mt-0.5" />
              <p className="text-sm text-txt-warning">
                {__("This will publish the selected documents directly without requiring approval.")}
              </p>
            </div>
            <div>
              <label htmlFor="changelog" className="text-sm font-medium text-txt-primary mb-1 block">
                {__("Changelog")}
              </label>
              <Textarea
                id="changelog"
                aria-label={__("Changelog")}
                required
                autogrow
                placeholder={__("Describe what changed in this version...")}
                {...register("changelog")}
              />
              {errors.changelog?.message && (
                <p className="text-xs text-txt-danger mt-1">{errors.changelog.message}</p>
              )}
            </div>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isPublishing}>
            {sprintf(__("Publish %s documents"), documentIds.length)}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
