// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { formatError } from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { VettingDialogMutation } from "#/__generated__/core/VettingDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

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
  const { t } = useTranslation();
  const schema = z.object({
    url: z.string().url(t("vettingDialog.validation.url")),
  });
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
            title: t("vettingDialog.messages.error"),
            description: formatError(
              t("vettingDialog.errors.start"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("vettingDialog.messages.success"),
          description: t("vettingDialog.messages.started"),
          variant: "success",
        });
        dialogRef.current?.close();
        reset();
      },
      onError(error) {
        toast({
          title: t("vettingDialog.messages.error"),
          description: formatError(
            t("vettingDialog.errors.start"),
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
      title={t("vettingDialog.title")}
      className="max-w-lg"
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded>
          <Field
            required
            label={t("vettingDialog.fields.websiteUrl")}
            type="text"
            {...register("url")}
            error={formState.errors.url?.message}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isVetting}>
            {t("vettingDialog.actions.start")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
