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

import {
  documentClassifications,
  documentTypes,
  getDocumentClassificationLabel,
  getDocumentTypeLabel,
} from "@probo/helpers";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
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
  const { root } = documentProperties();
  const pending = document.pendingApproval.edges[0]?.node != null;
  const locked = !document.canUpdate || document.status === "ARCHIVED" || pending;
  const canEditFields = !locked && document.writeMode !== "GENERATED";

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
      </div>
    </Card>
  );
}
