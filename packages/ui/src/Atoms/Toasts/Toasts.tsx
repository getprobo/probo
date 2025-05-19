import { create } from "zustand";
import { combine } from "zustand/middleware";
import { tv } from "tailwind-variants";

type Toast = {
    title: string;
    description: string;
    variant?: "success" | "error" | "warning" | "info";
    id: string;
};

const useToasts = create(
    combine({ toasts: [] as Toast[] }, (set) => ({
        add: (toast: Pick<Toast, "title" | "description" | "variant">) =>
            set((state) => ({
                toasts: [
                    { ...toast, id: Date.now().toString() },
                    ...state.toasts,
                ],
            })),
        remove: (id: string) =>
            set((state) => ({
                toasts: state.toasts.filter((toast) => toast.id !== id),
            })),
    })),
);

/**
 * Hook used to trigger toasts
 */
export function useToast() {
    return {
        toast: useToasts().add,
    };
}

/**
 * Toasts displayed at the bottom right of the screen.
 */
export function Toasts() {
    const { toasts, remove } = useToasts();
    return (
        <div className="fixed bottom-4 right-4 space-y-2 w-85">
            {toasts.map((toast) => (
                <div key={toast.id}>
                    <Toast {...toast} onClose={() => remove(toast.id)} />
                </div>
            ))}
        </div>
    );
}

const toast = tv({
    base: "p-2 border rounded relative animate-in slide-in-from-right w-full cursor-pointer",
    variants: {
        variant: {
            success:
                "bg-success text-invert text-txt-success border-border-success",
            error: "bg-danger text-invert text-txt-danger border-border-danger",
            warning:
                "bg-warning text-invert text-txt-warning border-border-warning",
            info: "bg-info text-invert text-txt-info border-border-info",
        },
    },
    defaultVariants: {
        variant: "success",
    },
});

export function Toast({ onClose, ...props }: Toast & { onClose: () => void }) {
    return (
        <div className={toast(props)} onClick={onClose}>
            <div className="font-semibold">{props.title}</div>
            <div className="text-sm opacity-80">{props.description}</div>
        </div>
    );
}
