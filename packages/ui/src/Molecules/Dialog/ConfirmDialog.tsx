import { useTranslate } from "@probo/i18n";
import {
    Root,
    Trigger,
    Overlay,
    Content,
    Title,
    Description,
    Cancel,
} from "@radix-ui/react-alert-dialog";
import {
    useRef,
    useState,
    type ComponentProps,
    type ReactNode,
    type RefObject,
} from "react";
import { Button } from "../../Atoms/Button/Button";
import { Root as Portal } from "@radix-ui/react-portal";
import { dialog } from "./Dialog";

type Props = {
    children?: ReactNode;
    title?: ReactNode;
    message: string;
    variant?: ComponentProps<typeof Button>["variant"];
    label?: string;
    onConfirm?: () => Promise<void>;
    ref?: RefObject<{ open: () => void } | null>;
};

export const useConfirmDialogRef = () => {
    return useRef<{ open: () => void } | null>(null);
};

export function ConfirmDialog(props: Props) {
    const { __ } = useTranslate();
    const title = props.title ?? __("Are you sure ?");
    const variant = props.variant ?? "danger";
    const label = variant === "danger" ? __("Delete") : props.label;
    const [open, setOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    const {
        overlay,
        content,
        header,
        footer,
        title: titleClassname,
    } = dialog();
    const onConfirm = () => {
        setLoading(true);
        props.onConfirm?.().then(() => {
            setLoading(false);
            setOpen(false);
        });
    };
    if (props.ref) {
        props.ref.current = {
            open: () => setOpen(true),
        };
    }
    return (
        <Root open={open} onOpenChange={setOpen}>
            {props.children && <Trigger asChild children={props.children} />}
            <Portal>
                <Overlay className={overlay()} />
                <Content className={content({ className: "max-w-[500px]" })}>
                    <header className={header()}>
                        <Title children={title} className={titleClassname()} />
                    </header>
                    <Description className="p-6" children={props.message} />
                    <footer className={footer()}>
                        <Cancel asChild>
                            <Button disabled={loading} variant="tertiary">
                                {__("Cancel")}
                            </Button>
                        </Cancel>
                        <Button
                            disabled={loading}
                            variant={variant}
                            onClick={onConfirm}
                        >
                            {label}
                        </Button>
                    </footer>
                </Content>
            </Portal>
        </Root>
    );
}
