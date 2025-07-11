export function withViewTransition(fn: () => void) {
    if (!document.startViewTransition) {
        fn();
        return;
    }
    document.startViewTransition(fn);
}

export function downloadFile(url: string | undefined | null, filename: string) {
    if (!url) {
        alert("Cannot download this file, fileUrl is not provided");
        return;
    }
    const link = document.createElement("a");
    link.setAttribute("href", url);
    link.setAttribute("hidden", "hidden");
    link.setAttribute("download", filename);
    link.style.display = "none";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}
