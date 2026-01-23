import { useTranslate } from "@probo/i18n";
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
import { create } from "zustand";
import { combine } from "zustand/middleware";

import { Button } from "../../Atoms/Button/Button";

import { dialog } from "./Dialog";

type State = {
  title?: string;
  message: string | null;
  variant?: ComponentProps<typeof Button>["variant"];
  label?: string;
  onConfirm: () => Promise<unknown>;
};

const useConfirmStore = create(
  combine(
    {
      message: null,
      onConfirm: () => Promise.resolve(),
    } as State,
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
  const { __ } = useTranslate();

  return useCallback(
    (cb: State["onConfirm"], props: Omit<State, "onConfirm">) => {
      open({
        onConfirm: cb,
        ...props,
        message: props.message,
        title: props.title ?? __("Are you sure ?"),
        variant: props.variant ?? "danger",
        label: props.label ?? __("Delete"),
      });
    },
    [open, __],
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

  const { __ } = useTranslate();
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
                {__("Cancel")}
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
