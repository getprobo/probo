import {
    Root,
    Trigger,
    Portal,
    Overlay,
    Content,
    Title,
    Close,
} from "@radix-ui/react-dialog";
import { IconCrossLargeX } from "../../Atoms/Icons";
import { type HTMLAttributes, type ReactNode } from "react";
import { Button } from "../../Atoms/Button/Button";
import { useTranslate } from "@probo/i18n";
import clsx from "clsx";

type Props = {
    trigger?: ReactNode;
    title: ReactNode;
    children?: ReactNode;
    onOpenChange?: (open: boolean) => void;
    open?: boolean;
};

export function Dialog({
    trigger,
    title,
    children,
    onOpenChange,
    open,
}: Props) {
    return (
        <Root open={open} onOpenChange={onOpenChange}>
            {trigger && <Trigger asChild>{trigger}</Trigger>}
            <Portal>
                <Overlay
                    className={clsx(
                        "fixed inset-0 z-50 bg-dialog/40",
                        `duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0`,
                    )}
                />
                <Content
                    aria-describedby={undefined}
                    className={clsx(
                        "fixed inset-0 m-auto z-50 w-full h-max bg-level-2 rounded-2xl max-w-5xl w-[95%]",
                        `duration-200
                            data-[state=open]:animate-in data-[state=closed]:animate-out
                            data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0
                            data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95
                            data-[state=closed]:slide-out-to-top-5
                            data-[state=open]:slide-in-from-top-5`,
                    )}
                >
                    <div className="flex justify-between items-center p-3 border-b border-b-border-low">
                        <Title className="text-sm font-medium">{title}</Title>
                        <Close asChild>
                            <Button variant="tertiary" icon={IconCrossLargeX} />
                        </Close>
                    </div>
                    {children}
                </Content>
            </Portal>
        </Root>
    );
}

export function DialogFooter({ children }: { children?: ReactNode }) {
    const { __ } = useTranslate();
    return (
        <footer className="flex justify-end items-center p-3 border-t border-t-border-low gap-2">
            <Close asChild>
                <Button variant="secondary">{__("Cancel")}</Button>
            </Close>
            {children}
        </footer>
    );
}

export function DialogContent(props: HTMLAttributes<HTMLDivElement>) {
    return (
        <div
            {...props}
            className={clsx("overflow-y-auto", props.className)}
            style={{
                maxHeight: "min(640px, calc(100vh - 140px))",
            }}
        />
    );
}
