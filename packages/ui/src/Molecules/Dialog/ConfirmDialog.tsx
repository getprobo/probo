import { useTranslate } from "@probo/i18n";
import {
    Root,
    Overlay,
    Content,
    Title,
    Description,
    Cancel,
} from "@radix-ui/react-alert-dialog";
import { useCallback, useState, type ComponentProps } from "react";
import { Button } from "../../Atoms/Button/Button";
import { Root as Portal } from "@radix-ui/react-portal";
import { dialog } from "./Dialog";
import { create } from "zustand";
import { combine } from "zustand/middleware";

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
        (set) => ({
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
    const open = useConfirmStore((state) => state.open);
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
        [open],
    );
}

/**
 * Global component that displays a dialog when confirm() is called
 */
export function ConfirmDialog() {
    const { message, title, variant, label, onConfirm, close } =
        useConfirmStore();
    const { __ } = useTranslate();
    const isOpen = !!message;
    const {
        overlay,
        content,
        header,
        title: titleClassname,
        footer,
    } = dialog();
    const [loading, setLoading] = useState(false);
    const handleConfirm = async () => {
        setLoading(true);
        try {
            await onConfirm();
        } finally {
            close();
            setLoading(false);
        }
    };

    return (
        <Root open={isOpen} onOpenChange={close}>
            <Portal>
                <Overlay className={overlay()} />
                <Content className={content({ className: "max-w-[500px]" })}>
                    <header className={header()}>
                        <Title children={title} className={titleClassname()} />
                    </header>
                    <Description className="p-6" children={message} />
                    <footer className={footer()}>
                        <Cancel asChild>
                            <Button disabled={loading} variant="tertiary">
                                {__("Cancel")}
                            </Button>
                        </Cancel>
                        <Button
                            disabled={loading}
                            variant={variant}
                            onClick={handleConfirm}
                        >
                            {label}
                        </Button>
                    </footer>
                </Content>
            </Portal>
        </Root>
    );
}
