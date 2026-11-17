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

import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentTitleField_document$key } from "#/__generated__/core/DocumentTitleField_document.graphql";

import { clampDocumentTitle } from "../_lib/documentTitle";
import { useUpdateDocument } from "../_lib/useUpdateDocument";
import { documentTitleField } from "../variants";

const documentTitleSaveDelayMs = 1000;

const documentTitleFieldFragment = graphql`
  fragment DocumentTitleField_document on Document {
    id
    title
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

interface DocumentTitleFieldProps {
  documentKey: DocumentTitleField_document$key;
}

export function DocumentTitleField({ documentKey }: DocumentTitleFieldProps) {
  const document = useFragment(documentTitleFieldFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const [updateDocument] = useUpdateDocument();
  const [draft, setDraft] = useState(document.title);
  const savedRef = useRef(document.title);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { root, input, heading } = documentTitleField();
  const canEdit = document.canUpdate
    && document.status !== "ARCHIVED"
    && document.writeMode !== "GENERATED"
    && document.pendingApproval.edges[0]?.node == null;

  useEffect(() => {
    if (document.title !== savedRef.current && draft === savedRef.current) {
      setDraft(document.title);
    }
    savedRef.current = document.title;
  }, [draft, document.title]);

  useEffect(() => {
    return () => {
      if (timeoutRef.current != null) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  function persist(value: string) {
    const next = value.trim();
    if (!next || next === savedRef.current) {
      setDraft(current => (current === value ? savedRef.current : current));
      return;
    }

    void updateDocument({
      variables: {
        input: {
          id: document.id,
          title: next,
        },
      },
    }).then(() => {
      setDraft(current => (current.trim() === next ? next : current));
    }).catch(() => {
      setDraft(savedRef.current);
    });
  }

  function schedule(value: string) {
    if (timeoutRef.current != null) {
      clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = setTimeout(() => {
      persist(value);
    }, documentTitleSaveDelayMs);
  }

  function flush() {
    if (timeoutRef.current != null) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    persist(draft);
  }

  if (!canEdit) {
    return (
      <Heading level={1} size={6} weight="medium" highContrast className={heading()}>
        {document.title}
      </Heading>
    );
  }

  return (
    <div className={root()}>
      <input
        className={input()}
        value={draft}
        required
        aria-label={t("details.fields.title")}
        onChange={(event) => {
          const next = clampDocumentTitle(event.currentTarget.value);
          setDraft(next);
          schedule(next);
        }}
        onBlur={flush}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            event.preventDefault();
            event.currentTarget.blur();
          }
        }}
      />
    </div>
  );
}
