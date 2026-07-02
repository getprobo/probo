// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  useToast,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { VettingDialogMutation } from "#/__generated__/core/VettingDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const schema = z.object({
  url: z.string().url(),
});

const vetMutation = graphql`
  mutation VettingDialogMutation($input: VetThirdPartyInput!) {
    vetThirdParty(input: $input) {
      thirdParty {
        id
        name
        websiteUrl
        vettingStatus
        ...useThirdPartyFormFragment
        ...ThirdPartyCompliancePageFragment
        ...ThirdPartyRiskAssessmentPageFragment
      }
    }
  }
`;

interface VettingDialogProps {
  thirdPartyId: string;
  websiteUrl?: string | null;
  children: ReactNode;
}

export function VettingDialog({ thirdPartyId, websiteUrl, children }: VettingDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const { register, handleSubmit, reset, formState } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        url: websiteUrl ?? "",
      },
    },
  );
  const [vet, isVetting] = useMutation<VettingDialogMutation>(vetMutation);

  const onSubmit = (data: z.infer<typeof schema>) => {
    vet({
      variables: {
        input: {
          id: thirdPartyId,
          websiteUrl: data.url,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to start vetting."),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("The third party is being vetted in the background."),
          variant: "success",
        });
        dialogRef.current?.close();
        reset();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to start vetting."),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={__("Start Vetting")}
      className="max-w-lg"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded>
          <Field
            required
            label={__("Website URL")}
            type="text"
            {...register("url")}
            error={formState.errors.url?.message}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isVetting}>
            {__("Start Vetting")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
