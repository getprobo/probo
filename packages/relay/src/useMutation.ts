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

import { useCallback } from "react";
import { useMutation as useRelayMutation, type UseMutationConfig } from "react-relay";
import type {
  GraphQLTaggedNode,
  MutationParameters,
  PayloadError,
} from "relay-runtime";

/**
 * App-supplied surface for rendering mutation feedback. The shared hook owns
 * *when* to notify; the host app owns *how* (toast system, i18n, error
 * formatting), keeping this package free of UI and i18n dependencies.
 *
 * `notifyError` receives an optional title override; when omitted, the
 * implementation supplies its own (localized) default.
 *
 * `handleFailure` is optional. When it returns true, the failure was consumed
 * (e.g. redirected to an auth / NDA gate) and `notifyError` is skipped.
 */
export type MutationNotifier = {
  notifySuccess: (message: string) => void;
  notifyError: (error: Error | PayloadError, title?: string) => void;
  handleFailure?: (error: Error, continueUrl: string) => boolean;
};

export type MutationFeedback = {
  // Message shown on success. Omit for no success notification.
  successMessage?: string;
  // Error notification behavior: `true` (default) notifies with the notifier's
  // default title, a string overrides that title, and `false` disables the
  // automatic notification so the caller handles the rejected promise itself.
  errorToast?: boolean | string;
  // Absolute URL returned to after an auth-gate redirect. Defaults to the
  // current page. Pass a marker-bearing URL when a deferred action must resume.
  continueUrl?: string;
};

/**
 * Builds an awaitable `useMutation` hook bound to a host-provided notifier.
 *
 * The returned hook wraps react-relay's `useMutation` so that callers can
 * `await` and continue only on success:
 *
 * - resolves with the mutation response on success;
 * - preserves every UseMutationConfig option by spreading the caller's config;
 * - on failure, optionally lets `handleFailure` consume auth gates, otherwise
 *   notifies via the injected notifier (unless disabled) AND rejects.
 *
 * Each app calls this once with its own notifier hook and re-exports the
 * result as the canonical `useMutation`.
 */
export function createUseMutation(useNotifier: () => MutationNotifier) {
  return function useMutation<T extends MutationParameters>(
    mutation: GraphQLTaggedNode,
    feedback?: MutationFeedback,
  ) {
    const [commit, isInFlight] = useRelayMutation<T>(mutation);
    const notifier = useNotifier();

    const {
      successMessage: baseSuccess,
      errorToast: baseErrorToast = true,
      continueUrl: baseContinueUrl,
    } = feedback ?? {};

    const mutate = useCallback(
      (config: UseMutationConfig<T>, overrides?: MutationFeedback): Promise<T["response"]> => {
        const successMessage = overrides?.successMessage ?? baseSuccess;
        const errorToast = overrides?.errorToast ?? baseErrorToast;
        const continueUrl
          = overrides?.continueUrl
            ?? baseContinueUrl
            ?? (typeof window !== "undefined" ? window.location.href : "");

        function notifyError(error: Error | PayloadError) {
          if (errorToast === false) {
            return;
          }
          notifier.notifyError(
            error,
            typeof errorToast === "string" ? errorToast : undefined,
          );
        }

        function toError(value: unknown): Error {
          return value instanceof Error ? value : new Error(String(value));
        }

        function consumeFailure(error: Error): boolean {
          return notifier.handleFailure?.(error, continueUrl) === true;
        }

        return new Promise<T["response"]>((resolve, reject) => {
          commit({
            ...config,
            onCompleted: (response, errors) => {
              // A throwing caller callback must still settle the wrapper promise,
              // otherwise `await mutate()` would hang forever.
              try {
                config.onCompleted?.(response, errors);
              } catch (callbackError) {
                const error = toError(callbackError);
                if (!consumeFailure(error)) {
                  notifyError(error);
                }
                reject(error);
                return;
              }
              if (errors && errors.length > 0) {
                const [payloadError] = errors;
                const error
                  = payloadError instanceof Error
                    ? payloadError
                    : new Error(payloadError.message);
                if (!consumeFailure(error)) {
                  notifyError(payloadError);
                }
                reject(error);
                return;
              }
              if (successMessage) {
                notifier.notifySuccess(successMessage);
              }
              resolve(response);
            },
            onError: (error) => {
              // Auth / NDA gates are consumed before the caller's onError so
              // every mutation redirects consistently without per-call boilerplate.
              if (consumeFailure(error)) {
                reject(error);
                return;
              }
              // Swallow a throwing caller callback so the original mutation error
              // still flows through to the notifier and the rejection.
              try {
                config.onError?.(error);
              } catch {
                // Intentionally ignored: the mutation error below is authoritative.
              }
              notifyError(error);
              reject(error);
            },
          });
        });
      },
      [commit, notifier, baseSuccess, baseErrorToast, baseContinueUrl],
    );

    return [mutate, isInFlight] as const;
  };
}
