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

import { PaperPlaneTiltIcon, PlusIcon, UploadSimpleIcon, XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment } from "react-relay";

import type { DocumentPublishForm_document$key } from "#/__generated__/core/DocumentPublishForm_document.graphql";
import type { DocumentPublishReviewCalloutQuery } from "#/__generated__/core/DocumentPublishReviewCalloutQuery.graphql";
import { PeopleMenu } from "#/components/people/PeopleMenu";

import { usePublishDocument } from "../../_lib/usePublishDocument";
import { documentPublishForm } from "../../variants";

import { DocumentPublishReviewCallout } from "./DocumentPublishReviewCallout";

const documentPublishFormFragment = graphql`
  fragment DocumentPublishForm_document on Document {
    id
    defaultApprovers {
      id
      fullName
    }
  }
`;

type Approver = { id: string; fullName: string };

interface DocumentPublishFormProps {
  documentKey: DocumentPublishForm_document$key;
  reviewQueryRef: PreloadedQuery<DocumentPublishReviewCalloutQuery> | null;
  onDone: () => void;
}

// A minor version publishes right away. A major version asks the chosen
// approvers first, or publishes right away when there are none.
export function DocumentPublishForm({ documentKey, reviewQueryRef, onDone }: DocumentPublishFormProps) {
  const document = useFragment(documentPublishFormFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const [publishDocument, isPublishing] = usePublishDocument();
  const [approvers, setApprovers] = useState<Approver[]>(
    () => document.defaultApprovers.map(({ id, fullName }) => ({ id, fullName })),
  );
  const {
    form, body, approversSection, approverRows, approverRow, approverName, addApprover,
  } = documentPublishForm();
  const requestsApproval = approvers.length > 0;

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const submitter = (event.nativeEvent as SubmitEvent).submitter;
    const minor = submitter instanceof HTMLButtonElement && submitter.value === "minor";
    const value = new FormData(event.currentTarget).get("changelog");
    const changelog = typeof value === "string" ? value.trim() : "";
    if (changelog === "") {
      return;
    }

    void publishDocument(
      {
        variables: {
          input: {
            documentId: document.id,
            minor,
            approverIds: minor ? null : approvers.map(approver => approver.id),
            changelog,
          },
        },
      },
      {
        successMessage: !minor && requestsApproval
          ? t("details.actions.messages.approvalRequested")
          : t("details.actions.messages.published"),
      },
    ).then(onDone).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <form className={form()} onSubmit={handleSubmit}>
      <DialogHeader>
        <DialogTitle>{t("details.actions.publish.title")}</DialogTitle>
        <DialogDescription>{t("details.actions.publish.description")}</DialogDescription>
      </DialogHeader>
      <DialogBody>
        <div className={body()}>
          <DocumentPublishReviewCallout queryRef={reviewQueryRef} />
          <Field label={t("details.actions.publish.changelog")} required>
            <Textarea
              name="changelog"
              placeholder={t("details.actions.publish.changelogPlaceholder")}
              disabled={isPublishing}
            />
          </Field>
          <div className={approversSection()}>
            <Text size={2} weight="medium">{t("details.actions.publish.approvers")}</Text>
            <Text size={1} color="faint">{t("details.actions.publish.approversDescription")}</Text>
            {approvers.length > 0 && (
              <div className={approverRows()}>
                {approvers.map(approver => (
                  <div key={approver.id} className={approverRow()}>
                    <Text size={2} className={approverName()} title={approver.fullName}>
                      {approver.fullName}
                    </Text>
                    <IconButton
                      size={1}
                      variant="ghost"
                      color="neutral"
                      disabled={isPublishing}
                      aria-label={t("details.actions.publish.removeApprover", { name: approver.fullName })}
                      onClick={() => {
                        setApprovers(current => current.filter(item => item.id !== approver.id));
                      }}
                    >
                      <XIcon />
                    </IconButton>
                  </div>
                ))}
              </div>
            )}
            <Dropdown>
              <DropdownTrigger
                disabled={isPublishing}
                render={(
                  <Button
                    size={1}
                    variant="soft"
                    color="neutral"
                    className={addApprover()}
                    iconStart={<PlusIcon />}
                  />
                )}
              >
                {t("details.actions.publish.addApprover")}
              </DropdownTrigger>
              <DropdownPopup>
                <PeopleMenu
                  value={null}
                  emptyValue={null}
                  disabled={isPublishing}
                  onValueChange={() => {}}
                  onSelect={(person) => {
                    if (person == null) {
                      return;
                    }
                    setApprovers(current => current.some(item => item.id === person.id)
                      ? current
                      : [...current, { id: person.id, fullName: person.fullName }]);
                  }}
                />
              </DropdownPopup>
            </Dropdown>
          </div>
        </div>
      </DialogBody>
      <DialogFooter>
        <DialogClose
          render={(
            <Button variant="soft" color="neutral" disabled={isPublishing}>
              {t("details.actions.cancel")}
            </Button>
          )}
        />
        <Button
          type="submit"
          value="minor"
          variant="soft"
          color="neutral"
          disabled={isPublishing}
          iconStart={<UploadSimpleIcon />}
        >
          {t("details.actions.publish.minor")}
        </Button>
        <Button
          type="submit"
          value="major"
          variant="solid"
          color="neutral"
          highContrast
          loading={isPublishing}
          iconStart={requestsApproval ? <PaperPlaneTiltIcon /> : <UploadSimpleIcon />}
        >
          {requestsApproval
            ? t("details.actions.publish.requestApproval")
            : t("details.actions.publish.major")}
        </Button>
      </DialogFooter>
    </form>
  );
}
