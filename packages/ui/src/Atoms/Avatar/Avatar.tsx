import clsx from "clsx";

type Props = {
    name: string;
    src?: string | null;
    size?: "s" | "m" | "l" | "xl";
};

export function Avatar({ name, src, size = "m" }: Props) {
    const className = clsx(
        "bg-txt-success text-txt-invert rounded-full font-semibold flex items-center justify-center flex-none",
        size === "s" && "size-5 text-xss",
        size === "m" && "size-6 text-xxs",
        size === "l" && "size-8 text-sm",
        size === "xl" && "size-16 text-3xl",
    );
    if (src) {
        return <img className={className} src={src} alt={name} />;
    }
    return <div className={className}>{extractInitials(name ?? "")}</div>;
}

function extractInitials(name: string) {
    const words = name.split(" ");
    if (words.length === 2) {
        return words
            .map((word) => word[0])
            .join("")
            .toUpperCase();
    }
    return name.substring(0, 2).toUpperCase();
}
