import { useTranslate } from "@probo/i18n";
import {
  Close,
  Content,
  Overlay,
  Portal,
  Root,
  Title,
  Trigger,
} from "@radix-ui/react-dialog";
import { clsx } from "clsx";
import {
  Children,
  cloneElement,
  type ComponentProps,
  type CSSProperties,
  type HTMLAttributes,
  isValidElement,
  type ReactNode,
  type RefObject,
  useEffect,
  useRef,
  useState,
} from "react";
import { tv } from "tailwind-variants";

import { Button } from "../../Atoms/Button/Button";
import { IconCrossLargeX } from "../../Atoms/Icons";

export const dialog = tv({
  slots: {
    overlay:
            "fixed inset-0 z-50 bg-dialog/40 duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
    content:
            "text-txt-secondary text-sm fixed inset-0 m-auto z-50 w-full h-max bg-level-2 rounded-2xl max-w-5xl w-[95%] duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-top-5 data-[state=open]:slide-in-from-top-5",
    header: "flex justify-between items-center p-3 border-b border-b-border-low",
    title: "text-sm font-medium text-txt-primary",
    footer: "flex justify-end items-center p-3 border-t border-t-border-low gap-2",
  },
});

export type DialogRef = RefObject<{
  open: () => void;
  close: () => void;
} | null>;

type Props = {
  trigger?: ReactNode;
  title?: ReactNode;
  children?: ReactNode;
  defaultOpen?: boolean;
  className?: string;
  ref?: DialogRef;
  onClose?: () => void;
  closable?: boolean;
};

export const useDialogRef = (): DialogRef => {
  return useRef(null);
};

export function Dialog({
  trigger,
  title,
  children,
  className,
  ref,
  defaultOpen,
  onClose,
  closable = true,
}: Props) {
  const { overlay, content, header, title: titleClassname } = dialog();
  const [open, setOpen] = useState(!!defaultOpen);

  useEffect(() => {
    if (ref) {
      ref.current = {
        open() {
          setOpen(true);
        },
        close() {
          setOpen(false);
        },
      };
    }
  });

  const onOpenChange = (open: boolean) => {
    if (!open && !closable) {
      return;
    }
    setOpen(open);
    if (!open) {
      onClose?.();
    }
  };

  const contentProps = closable
    ? {}
    : {
        onEscapeKeyDown: (e: Event) => e.preventDefault(),
        onPointerDownOutside: (e: Event) => e.preventDefault(),
        onInteractOutside: (e: Event) => e.preventDefault(),
      };

  return (
    <Root open={open} onOpenChange={closable ? onOpenChange : undefined}>
      {trigger && <Trigger asChild>{trigger}</Trigger>}
      <Portal>
        <Overlay className={overlay()} />
        <Content
          aria-describedby={undefined}
          className={content({ className })}
          {...contentProps}
        >
          {title
            ? (
                <div className={header()}>
                  <Title className={titleClassname()}>
                    {" "}
                    {title}
                  </Title>
                  {closable && (
                    <Close asChild>
                      <Button
                        tabIndex={-1}
                        variant="tertiary"
                        icon={IconCrossLargeX}
                      />
                    </Close>
                  )}
                </div>
              )
            : (
                closable && (
                  <Close asChild>
                    <Button
                      tabIndex={-1}
                      variant="tertiary"
                      className="absolute top-4 right-4"
                      icon={IconCrossLargeX}
                    />
                  </Close>
                )
              )}
          {children}
        </Content>
      </Portal>
    </Root>
  );
}

export function DialogFooter({
  children,
  exitLabel,
  className,
}: {
  children?: ReactNode;
  exitLabel?: string;
  className?: string;
}) {
  const { __ } = useTranslate();
  const { footer } = dialog();
  return (
    <footer className={footer({ className })}>
      <Close asChild>
        <Button variant="secondary">{exitLabel ?? __("Cancel")}</Button>
      </Close>
      {children}
    </footer>
  );
}

export function DialogContent({
  padded,
  scrollableChildren,
  ...props
}: HTMLAttributes<HTMLDivElement> & {
  padded?: boolean;
  scrollableChildren?: boolean;
}) {
  let children = props.children;
  if (scrollableChildren) {
    children = Children.map(props.children, (c) => {
      if (
        isValidElement<{
          className?: string;
          style?: CSSProperties;
        }>(c)
      ) {
        return cloneElement(c, {
          ...c.props,
          className: clsx(c.props.className, "overflow-y-auto"),
          style: {
            maxHeight: "var(--maxHeight)",
            ...c.props.style,
          },
        });
      }
      return c;
    });
  }
  return (
    <div
      {...props}
      className={clsx(
        "overflow-y-auto",
        props.className,
        padded && "p-6",
      )}
      style={
        {
          "--maxHeight": "min(640px, calc(100vh - 140px))",
          "maxHeight": "var(--maxHeight)",
        } as CSSProperties
      }
    >
      {children}
    </div>
  );
}

export function DialogTitle(props: ComponentProps<typeof Title>) {
  return <Title {...props} />;
}
