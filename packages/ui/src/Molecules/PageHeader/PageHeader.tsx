import type { PropsWithChildren } from "react";

type Props = PropsWithChildren<{
    title: string;
    description?: string;
}>;

export function PageHeader({ title, description, children }: Props) {
    return (
        <div className="flex justify-between items-start w-full">
            <div className=" space-y-1">
                <h1 className="text-2xl ">{title}</h1>
                {description && (
                    <p className="text-sm text-txt-secondary">{description}</p>
                )}
            </div>
            <div>{children}</div>
        </div>
    );
}
