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

import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { ConnectorNameHeadingUpdateNameMutation } from "#/__generated__/core/ConnectorNameHeadingUpdateNameMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const updateConnectorNameMutation = graphql`
  mutation ConnectorNameHeadingUpdateNameMutation($input: UpdateConnectorInput!) {
    updateConnector(input: $input) {
      connector {
        id
        name
      }
    }
  }
`;

interface ConnectorNameHeadingProps {
  connectorId: string;
  name: string;
  canUpdate: boolean;
}

export function ConnectorNameHeading({
  connectorId,
  name,
  canUpdate,
}: ConnectorNameHeadingProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(name);
  const [error, setError] = useState<string | undefined>();
  const canceling = useRef(false);
  const saving = useRef(false);
  const [updateName, isUpdating] = useMutation<ConnectorNameHeadingUpdateNameMutation>(
    updateConnectorNameMutation,
  );

  async function save(value: string) {
    if (saving.current) {
      return;
    }
    const trimmed = value.trim();
    if (trimmed === "") {
      setError(t("connectForm.name.required"));
      return;
    }
    if (trimmed === name) {
      setError(undefined);
      setEditing(false);
      return;
    }
    saving.current = true;
    try {
      await updateName({
        variables: { input: { connectorId, name: trimmed } },
      }, { errorToast: t("connectForm.name.saveFailed") });
      setError(undefined);
      setEditing(false);
    } catch {
      setDraft(trimmed);
    } finally {
      saving.current = false;
    }
  }

  if (!canUpdate) {
    return (
      <Heading level={2} size={3} weight="medium" highContrast className="truncate">
        {name}
      </Heading>
    );
  }

  if (!editing) {
    return (
      <button
        type="button"
        className="pointer-events-auto relative z-10 min-w-0 cursor-text text-left"
        onClick={(event) => {
          event.stopPropagation();
          setDraft(name);
          setError(undefined);
          setEditing(true);
        }}
      >
        <Heading level={2} size={3} weight="medium" highContrast className="truncate">
          {name}
        </Heading>
      </button>
    );
  }

  return (
    <div
      className="pointer-events-auto relative z-10 min-w-0"
      onClick={event => event.stopPropagation()}
      onKeyDown={event => event.stopPropagation()}
    >
      <Field error={error}>
        <TextField
          value={draft}
          autoFocus
          disabled={isUpdating}
          aria-label={t("connectForm.name.label")}
          onChange={event => setDraft(event.target.value)}
          onBlur={() => {
            if (canceling.current) {
              canceling.current = false;
              return;
            }
            void save(draft);
          }}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              event.preventDefault();
              void save(draft);
            }
            if (event.key === "Escape") {
              event.preventDefault();
              canceling.current = true;
              setDraft(name);
              setError(undefined);
              setEditing(false);
            }
          }}
        />
      </Field>
    </div>
  );
}
