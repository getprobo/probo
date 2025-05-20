import { useEffect } from "react";

export function usePageTitle(title: string) {
    useEffect(() => {
        document.title = title + " - Probo";
        return () => {
            document.title = "Probo";
        };
    }, [title]);
}
