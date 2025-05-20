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
import type { ReactNode } from "react";
import { Button } from "../../Atoms/Button/Button";
import { useTranslate } from "@probo/i18n";

type Props = {
    onClose: () => void;
    trigger: ReactNode;
    title: ReactNode;
    children?: ReactNode;
};

export function Dialog({ onClose, trigger, title, children }: Props) {
    return (
        <Root>
            <Trigger asChild>{trigger}</Trigger>
            <Portal>
                <Overlay className="fixed inset-0 z-50 bg-dialog/40 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />
                <div className="fixed grid place-items-center inset-0 z-50">
                    <Content className="bg-level-2 rounded-2xl max-w-5xl w-[95%]">
                        <div className="flex justify-between items-center p-3 border-b border-b-border-low">
                            <Title className="text-sm font-medium">
                                {title}
                            </Title>
                            <Close asChild>
                                <Button
                                    variant="tertiary"
                                    icon={IconCrossLargeX}
                                />
                            </Close>
                        </div>
                        {children}
                    </Content>
                </div>
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

export function DialogContent({ children }: { children: ReactNode }) {
    return (
        <div
            className="overflow-y-auto"
            style={{
                maxHeight: "min(640px, calc(100vh - 140px))",
            }}
        >
            {children}
        </div>
    );
}
