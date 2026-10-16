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

import { TASKS_NAMESPACE } from "./i18n";
import { taskNameField } from "./variants";

const taskNameMaxLength = 1000;
const taskNameSaveDelayMs = 1000;

export interface TaskNameFieldProps {
  name: string;
  canUpdate: boolean;
  onSave: (name: string) => Promise<void>;
}

export function TaskNameField({ name, canUpdate, onSave }: TaskNameFieldProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const [draft, setDraft] = useState(name);
  const savedRef = useRef(name);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { root, input } = taskNameField();

  useEffect(() => {
    if (name !== savedRef.current && draft === savedRef.current) {
      setDraft(name);
    }
    savedRef.current = name;
  }, [draft, name]);

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

    void onSave(next).then(() => {
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
    }, taskNameSaveDelayMs);
  }

  function flush() {
    if (timeoutRef.current != null) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    persist(draft);
  }

  if (!canUpdate) {
    return (
      <Heading level={1} size={6} weight="medium" highContrast className="min-w-0 truncate">
        {name}
      </Heading>
    );
  }

  return (
    <div className={root()}>
      <input
        className={input()}
        value={draft}
        required
        maxLength={taskNameMaxLength}
        aria-label={t("fields.name")}
        onChange={(event) => {
          const next = event.currentTarget.value;
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
