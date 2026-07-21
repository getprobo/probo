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

import {
  Cancel,
  Content,
  Description,
  Overlay,
  Root,
  Title,
} from "@radix-ui/react-alert-dialog";
import { Root as Portal } from "@radix-ui/react-portal";
import { type ComponentProps, useCallback, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { create } from "zustand";
import { combine } from "zustand/middleware";

import { Button } from "../../Atoms/Button/Button";

import { dialog } from "./Dialog";

type State = {
  title?: string;
  message: string | null;
  variant?: ComponentProps<typeof Button>["variant"];
  label?: string;
  onConfirm: () => void | Promise<unknown>;
};

const initialConfirmState: State = {
  message: null,
  onConfirm: () => Promise.resolve(),
};

const useConfirmStore = create(
  combine(
    initialConfirmState,
    set => ({
      open: (props: State) => {
        set(props);
      },
      close: () => {
        set({
          message: null,
        });
      },
    }),
  ),
);

/**
 * Hook used to open a confirm dialog
 */
export function useConfirm() {
  const open = useConfirmStore(state => state.open);
  const { t } = useTranslation();

  return useCallback(
    (cb: State["onConfirm"], props: Omit<State, "onConfirm">) => {
      open({
        onConfirm: cb,
        ...props,
        message: props.message,
        title: props.title ?? t("ui.confirmDialog.defaultTitle"),
        variant: props.variant ?? "danger",
        label: props.label ?? t("ui.actions.delete"),
      });
    },
    [open, t],
  );
}

/**
 * Global component that displays a dialog when confirm() is called
 */
export function ConfirmDialog() {
  const message = useConfirmStore(state => state.message);
  const isOpen = !!message;

  if (!isOpen) {
    return null;
  }

  return <ConfirmDialogContent />;
}

function ConfirmDialogContent() {
  const message = useConfirmStore(state => state.message);
  const title = useConfirmStore(state => state.title);
  const variant = useConfirmStore(state => state.variant);
  const label = useConfirmStore(state => state.label);
  const onConfirm = useConfirmStore(state => state.onConfirm);
  const close = useConfirmStore(state => state.close);

  const { t } = useTranslation();
  const isOpen = !!message;

  const [loading, setLoading] = useState(false);

  const dialogStyles = useMemo(() => {
    const styles = dialog();
    return {
      overlay: styles.overlay(),
      content: styles.content({ className: "max-w-[500px]" }),
      header: styles.header(),
      title: styles.title(),
      footer: styles.footer(),
    };
  }, []);

  const handleConfirm = async () => {
    setLoading(true);
    try {
      await onConfirm();
    } catch (error) {
      console.error("Confirm action failed:", error);
    } finally {
      close();
      setLoading(false);
    }
  };

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setLoading(false);
      close();
    }
  };

  return (
    <Root open={isOpen} onOpenChange={handleOpenChange}>
      <Portal>
        <Overlay className={dialogStyles.overlay} />
        <Content className={dialogStyles.content}>
          <header className={dialogStyles.header}>
            <Title className={dialogStyles.title}>{title}</Title>
          </header>
          <Description className="p-6">{message}</Description>
          <footer className={dialogStyles.footer}>
            <Cancel asChild>
              <Button disabled={loading} variant="tertiary">
                {t("ui.actions.cancel")}
              </Button>
            </Cancel>
            <Button
              disabled={loading}
              variant={variant}
              onClick={() => void handleConfirm()}
            >
              {label}
            </Button>
          </footer>
        </Content>
      </Portal>
    </Root>
  );
}
