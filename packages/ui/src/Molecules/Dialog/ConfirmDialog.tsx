import { useTranslate } from "@probo/i18n";
import { Dialog, DialogContent, DialogFooter } from "./Dialog";
import {
    Children,
    cloneElement,
    isValidElement,
    useState,
    type ComponentProps,
    type MouseEvent,
    type ReactNode,
} from "react";
import { Button } from "../../Atoms/Button/Button";
import { Root as Portal } from "@radix-ui/react-portal";

type Props = {
    children?: ReactNode;
    title?: ReactNode;
    message: string;
    variant?: ComponentProps<typeof Button>["variant"];
    label?: string;
    onConfirm?: () => void;
};

export function ConfirmDialog(props: Props) {
    const { __ } = useTranslate();
    const title = props.title ?? __("Are you sure ?");
    const variant = props.variant ?? "danger";
    const label = variant === "danger" ? __("Delete") : props.label;
    const [open, setOpen] = useState(false);
    const onConfirm = () => {
        setOpen(false);
        props.onConfirm?.();
        document.body.click();
    };
    const children = Children.map(props.children, (child) => {
        if (!isValidElement(child)) {
            throw new Error("Cannot use non element children");
        }
        return cloneElement(child, {
            // @ts-expect-error We inject a handler on an unknown element
            onClick: (e: MouseEvent) => {
                e.preventDefault();
                setOpen(true);
            },
        });
    });
    return (
        <>
            {children}
            <Portal>
                <Dialog
                    open={open}
                    onOpenChange={setOpen}
                    title={title}
                    className="max-w-[500px]"
                >
                    <DialogContent>
                        <p className="text-sm text-txt-secondary p-4">
                            {props.message}
                        </p>
                    </DialogContent>
                    <DialogFooter>
                        <Button variant={variant} onClick={onConfirm}>
                            {label}
                        </Button>
                    </DialogFooter>
                </Dialog>
            </Portal>
        </>
    );
}
