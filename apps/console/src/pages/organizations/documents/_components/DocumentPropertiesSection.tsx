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

import { UserPlusIcon, XIcon } from "@phosphor-icons/react";
import {
  documentClassifications,
  documentTypes,
  getDocumentClassificationLabel,
  getDocumentTypeLabel,
} from "@probo/helpers";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Field } from "@probo/ui/src/v2/form/Field";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentPropertiesSection_document$key } from "#/__generated__/core/DocumentPropertiesSection_document.graphql";
import type { useUpdateDocumentMutation$variables } from "#/__generated__/core/useUpdateDocumentMutation.graphql";
import { PeopleMenu } from "#/components/people/PeopleMenu";

import { useUpdateDocument } from "../_lib/useUpdateDocument";
import { documentProperties } from "../variants";

const documentPropertiesSectionFragment = graphql`
  fragment DocumentPropertiesSection_document on Document {
    id
    documentType
    classification
    status
    writeMode
    canUpdate: permission(action: "core:document:update")
    pendingApproval: versions(first: 1, filter: { statuses: [PENDING_APPROVAL] }) {
      edges {
        node {
          id
        }
      }
    }
    defaultApprovers {
      id
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
    }
  }
`;

interface DocumentPropertiesSectionProps {
  documentKey: DocumentPropertiesSection_document$key;
}

export function DocumentPropertiesSection({ documentKey }: DocumentPropertiesSectionProps) {
  const document = useFragment(documentPropertiesSectionFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const { t: tApp } = useTranslation();
  const [updateDocument, isUpdating] = useUpdateDocument();
  const { root, approvers, approver, identity, avatar, name } = documentProperties();
  const pending = document.pendingApproval.edges[0]?.node != null;
  const locked = !document.canUpdate || document.status === "ARCHIVED" || pending;
  const canEditFields = !locked && document.writeMode !== "GENERATED";
  const canEditApprovers = !locked;
  const approverIds = document.defaultApprovers.map(person => person.id);

  function save(input: Omit<useUpdateDocumentMutation$variables["input"], "id">) {
    void updateDocument({
      variables: {
        input: {
          id: document.id,
          ...input,
        },
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <Heading level={2} size={2} weight="medium">
          {t("details.properties.title")}
        </Heading>
        <Field label={t("details.properties.type")}>
          {canEditFields
            ? (
                <Select
                  value={document.documentType}
                  disabled={isUpdating}
                  onValueChange={(next: string | null) => {
                    if (next == null || next === document.documentType) {
                      return;
                    }
                    save({ documentType: next as typeof document.documentType });
                  }}
                >
                  <SelectTrigger size={1} variant="ghost" aria-label={t("details.properties.type")}>
                    {(selected: string | null) => selected == null
                      ? null
                      : getDocumentTypeLabel(tApp, selected) ?? selected}
                  </SelectTrigger>
                  <SelectPopup>
                    {documentTypes.map(type => (
                      <SelectItem key={type} value={type}>
                        {getDocumentTypeLabel(tApp, type) ?? type}
                      </SelectItem>
                    ))}
                  </SelectPopup>
                </Select>
              )
            : (
                <Text size={2}>
                  {getDocumentTypeLabel(tApp, document.documentType) ?? document.documentType}
                </Text>
              )}
        </Field>
        <Field label={t("details.properties.classification")}>
          {canEditFields
            ? (
                <Select
                  value={document.classification}
                  disabled={isUpdating}
                  onValueChange={(next: string | null) => {
                    if (next == null || next === document.classification) {
                      return;
                    }
                    save({ classification: next as typeof document.classification });
                  }}
                >
                  <SelectTrigger size={1} variant="ghost" aria-label={t("details.properties.classification")}>
                    {(selected: string | null) => selected == null
                      ? null
                      : getDocumentClassificationLabel(tApp, selected) ?? selected}
                  </SelectTrigger>
                  <SelectPopup>
                    {documentClassifications.map(classification => (
                      <SelectItem key={classification} value={classification}>
                        {getDocumentClassificationLabel(tApp, classification) ?? classification}
                      </SelectItem>
                    ))}
                  </SelectPopup>
                </Select>
              )
            : (
                <Text size={2}>
                  {getDocumentClassificationLabel(tApp, document.classification) ?? document.classification}
                </Text>
              )}
        </Field>
        <Field label={t("details.properties.approvers")}>
          <div className={approvers()}>
            {document.defaultApprovers.length > 0
              ? document.defaultApprovers.map(person => (
                  <div key={person.id} className={approver()}>
                    <div className={identity()}>
                      <Avatar
                        size={1}
                        className={avatar()}
                        name={person.fullName}
                        email={person.emailAddress}
                        src={person.avatar?.downloadUrl}
                      />
                      <Text size={1} weight="medium" color="neutral" className={name()} title={person.fullName}>
                        {person.fullName}
                      </Text>
                    </div>
                    {canEditApprovers && (
                      <IconButton
                        size={1}
                        variant="ghost"
                        color="neutral"
                        disabled={isUpdating}
                        className="shrink-0"
                        aria-label={t("details.properties.removeApprover", { name: person.fullName })}
                        onClick={() => {
                          save({
                            defaultApproverIds: approverIds.filter(id => id !== person.id),
                          });
                        }}
                      >
                        <XIcon />
                      </IconButton>
                    )}
                  </div>
                ))
              : (
                  <Text size={2}>{t("details.properties.none")}</Text>
                )}
            {canEditApprovers && (
              <Dropdown>
                <DropdownTrigger
                  disabled={isUpdating}
                  render={(
                    <Button
                      variant="ghost"
                      color="neutral"
                      size={1}
                      className="self-start"
                      iconStart={<UserPlusIcon />}
                    >
                      {t("details.properties.addApprover")}
                    </Button>
                  )}
                />
              <DropdownPopup size={1}>
                <PeopleMenu
                  value={null}
                  emptyValue={null}
                  disabled={isUpdating}
                  onValueChange={(approverId) => {
                    if (approverId == null || approverIds.includes(approverId)) {
                      return;
                    }
                    save({ defaultApproverIds: [...approverIds, approverId] });
                  }}
                />
              </DropdownPopup>
              </Dropdown>
            )}
          </div>
        </Field>
      </div>
    </Card>
  );
}
