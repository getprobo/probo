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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { Button, useToast } from "@probo/ui";
import {
  type ClipboardEvent,
  type KeyboardEvent,
  useCallback,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { DeviceActivationPageMutation } from "#/__generated__/iam/DeviceActivationPageMutation.graphql";
import type { DeviceActivationPageQuery } from "#/__generated__/iam/DeviceActivationPageQuery.graphql";

export const deviceActivationPageQuery = graphql`
  query DeviceActivationPageQuery {
    viewer {
      __typename
    }
  }
`;

const authorizeDeviceMutation = graphql`
  mutation DeviceActivationPageMutation($input: AuthorizeDeviceInput!) {
    authorizeDevice(input: $input) {
      success
      consentId
    }
  }
`;

export default function DeviceActivationPage(props: {
  queryRef: PreloadedQuery<DeviceActivationPageQuery>;
}) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  usePreloadedQuery<DeviceActivationPageQuery>(deviceActivationPageQuery, props.queryRef);
  usePageTitle(t("deviceActivationPage.pageTitle"));

  const preset = (searchParams.get("user_code") ?? "").replace(/-/g, "");
  const [values, setValues] = useState<string[]>(() => {
    const chars = preset.split("").slice(0, 8);
    return Array.from({ length: 8 }, (_, i) => chars[i] ?? "");
  });
  const [status, setStatus] = useState<"idle" | "success">("idle");
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const [authorizeDevice, isInFlight]
    = useMutation<DeviceActivationPageMutation>(authorizeDeviceMutation);

  const syncAndFocus = useCallback(
    (next: string[], focusIdx?: number) => {
      setValues(next);
      if (focusIdx !== undefined && inputRefs.current[focusIdx]) {
        inputRefs.current[focusIdx].focus();
      }
    },
    [],
  );

  const handleInput = useCallback(
    (idx: number, char: string) => {
      const cleaned = char.replace(/[^a-zA-Z0-9]/g, "").slice(0, 1);
      const next = [...values];
      next[idx] = cleaned;
      syncAndFocus(next, cleaned ? Math.min(idx + 1, 7) : undefined);
    },
    [values, syncAndFocus],
  );

  const handleKeyDown = useCallback(
    (idx: number, e: KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Backspace" && !values[idx] && idx > 0) {
        const next = [...values];
        next[idx - 1] = "";
        syncAndFocus(next, idx - 1);
      }
    },
    [values, syncAndFocus],
  );

  const handlePaste = useCallback(
    (idx: number, e: ClipboardEvent<HTMLInputElement>) => {
      e.preventDefault();
      const text = e.clipboardData.getData("text").replace(/[^a-zA-Z0-9]/g, "");
      const next = [...values];
      for (let j = 0; j < text.length && idx + j < 8; j++) {
        next[idx + j] = text[j];
      }
      syncAndFocus(next, Math.min(idx + text.length, 7));
    },
    [values, syncAndFocus],
  );

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      const code = values.join("").toUpperCase();
      if (code.length !== 8) return;

      const userCode = code.slice(0, 4) + "-" + code.slice(4);

      authorizeDevice({
        variables: { input: { userCode } },
        onCompleted: (response, errors) => {
          if (errors) {
            toast({
              title: t("deviceActivationPage.errors.authorizationFailed"),
              description: formatError(
                t("deviceActivationPage.errors.invalidCode"),
                errors,
              ),
              variant: "error",
            });
            return;
          }

          const result = response.authorizeDevice;
          if (!result) return;

          if (result.success) {
            setStatus("success");
          } else if (result.consentId) {
            void navigate(`/auth/consent?consent_id=${result.consentId}`);
          }
        },
        onError: (err) => {
          toast({
            title: t("common.error"),
            description: err.message || t("deviceActivationPage.errors.generic"),
            variant: "error",
          });
        },
      });
    },
    [values, authorizeDevice, t, toast, navigate],
  );

  const isFilled = values.every(v => v.length === 1);

  if (status === "success") {
    return (
      <div className="w-full max-w-md mx-auto pt-8 space-y-6 text-center">
        <h1 className="text-2xl font-bold">{t("deviceActivationPage.authorized.title")}</h1>
        <p className="text-txt-tertiary">
          {t("deviceActivationPage.authorized.description")}
        </p>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md mx-auto pt-8 space-y-6">
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold">{t("deviceActivationPage.title")}</h1>
        <p className="text-txt-tertiary">
          {t("deviceActivationPage.description")}
        </p>
      </div>

      <form onSubmit={e => void handleSubmit(e)} className="space-y-6">
        <div className="flex items-center justify-center gap-2">
          {values.map((val, idx) => (
            <div key={idx} className="contents">
              {idx === 4 && (
                <span className="text-xl text-txt-tertiary select-none px-0.5">&ndash;</span>
              )}
              <input
                ref={(el) => { inputRefs.current[idx] = el; }}
                type="text"
                inputMode="text"
                maxLength={1}
                value={val}
                onChange={e => handleInput(idx, e.target.value)}
                onKeyDown={e => handleKeyDown(idx, e)}
                onPaste={e => handlePaste(idx, e)}
                autoComplete="off"
                autoCorrect="off"
                autoCapitalize="characters"
                spellCheck={false}
                autoFocus={idx === 0}
                aria-label={t("deviceActivationPage.codeCharacter", { count: idx + 1 })}
                className="w-11 h-13 text-center text-lg font-mono font-medium uppercase rounded-lg border border-border-mid bg-level-1 text-txt-primary outline-none transition-colors focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
              />
            </div>
          ))}
        </div>

        <Button
          type="submit"
          className="w-full h-10"
          disabled={!isFilled || isInFlight}
        >
          {isInFlight ? t("deviceActivationPage.actions.authorizing") : t("deviceActivationPage.actions.continue")}
        </Button>
      </form>

      <p className="text-center text-sm text-txt-tertiary">
        {t("deviceActivationPage.notice")}
      </p>
    </div>
  );
}
