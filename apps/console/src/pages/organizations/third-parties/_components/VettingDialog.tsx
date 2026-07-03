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
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { VettingDialogFragment$key } from "#/__generated__/core/VettingDialogFragment.graphql";
import type { VettingDialogMutation } from "#/__generated__/core/VettingDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const vettingDialogFragment = graphql`
  fragment VettingDialogFragment on ThirdParty {
    id
    name
    legalName
    websiteUrl
  }
`;

const schema = z.object({
  name: z.string().min(1),
  legalName: z.string().optional(),
  url: z.string().url(),
});

const vetMutation = graphql`
  mutation VettingDialogMutation($input: VetThirdPartyInput!) {
    vetThirdParty(input: $input) {
      thirdParty {
        id
        name
        legalName
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
  thirdParty: VettingDialogFragment$key;
  children: ReactNode;
}

export function VettingDialog({ thirdParty: thirdPartyKey, children }: VettingDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const thirdParty = useFragment(vettingDialogFragment, thirdPartyKey);
  const { register, handleSubmit, reset, formState } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: thirdParty.name ?? "",
        legalName: thirdParty.legalName ?? "",
        url: thirdParty.websiteUrl ?? "",
      },
    },
  );
  const [vet, isVetting] = useMutation<VettingDialogMutation>(vetMutation);

  const onSubmit = (data: z.infer<typeof schema>) => {
    const name = data.name.trim();
    const legalName = data.legalName?.trim() ?? "";

    // Only send identity when changed: writing it needs update permission,
    // so an unchanged confirm stays vet-only. Empty legal name clears it.
    const nameChanged = name !== (thirdParty.name ?? "");
    const legalNameChanged = legalName !== (thirdParty.legalName ?? "");

    vet({
      variables: {
        input: {
          id: thirdParty.id,
          websiteUrl: data.url,
          ...(nameChanged ? { name } : {}),
          ...(legalNameChanged ? { legalName } : {}),
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
        <DialogContent padded className="space-y-4">
          <Field
            required
            label={__("Name")}
            type="text"
            {...register("name")}
            error={formState.errors.name?.message}
          />
          <Field
            label={__("Legal name")}
            type="text"
            {...register("legalName")}
            error={formState.errors.legalName?.message}
          />
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
