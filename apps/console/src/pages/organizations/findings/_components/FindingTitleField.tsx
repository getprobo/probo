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

import { clampFindingTitle } from "../_lib/findingTitle";
import { useSaveQueue } from "../_lib/useSaveQueue";

import { findingTitleField } from "./variants";

interface FindingTitleFieldProps {
  title: string;
  canUpdate: boolean;
  onSave: (title: string) => Promise<unknown>;
}

export function FindingTitleField({ title, canUpdate, onSave }: FindingTitleFieldProps) {
  const { t } = useTranslation("organizations/findings");
  const [draft, setDraft] = useState(title);
  const savedRef = useRef(title);
  const queuedRef = useRef(title);
  const pendingRef = useRef(0);
  const enqueue = useSaveQueue();
  const { heading, input } = findingTitleField();

  useEffect(() => {
    if (title !== savedRef.current && draft === savedRef.current) {
      setDraft(title);
    }
    savedRef.current = title;
    if (pendingRef.current === 0) {
      queuedRef.current = title;
    }
  }, [draft, title]);

  function persist(value: string) {
    const next = value.trim();
    if (!next || next === queuedRef.current) {
      setDraft(queuedRef.current);
      return;
    }
    const previous = queuedRef.current;
    queuedRef.current = next;
    pendingRef.current += 1;
    void enqueue(() => onSave(next)).then(() => {
      savedRef.current = next;
      setDraft(current => (current === value ? next : current));
    }).catch(() => {
      queuedRef.current = previous;
      setDraft(current => (current === value ? previous : current));
    }).finally(() => {
      pendingRef.current -= 1;
    });
  }

  if (!canUpdate) {
    return (
      <Heading level={1} size={6} weight="medium" highContrast className={heading()}>
        {title}
      </Heading>
    );
  }

  return (
    <input
      className={input()}
      value={draft}
      required
      aria-label={t("details.fields.title")}
      onChange={event => setDraft(clampFindingTitle(event.currentTarget.value))}
      onBlur={() => persist(draft)}
      onKeyDown={(event) => {
        if (event.key === "Enter") {
          event.preventDefault();
          event.currentTarget.blur();
        }
      }}
    />
  );
}
