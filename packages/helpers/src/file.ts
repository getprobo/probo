/**
 * Return the file size in a human readable format
 */
export function fileSize(__: (s: string) => string, size: number): string {
    if (size < 0) return "";
    if (size === 0) return `0 ${__("B")}`;

    const units = [__("B"), __("KB"), __("MB"), __("GB"), __("TB")];
    const i = Math.floor(Math.log(size) / Math.log(1024));

    // Don't go beyond available units
    const unitIndex = Math.min(i, units.length - 1);

    // Convert to the appropriate unit with 2 decimal places
    const convertedSize = size / Math.pow(1024, unitIndex);
    const formattedSize = Math.round(convertedSize * 100) / 100;

    return `${formattedSize} ${units[unitIndex]}`;
}

type FileInfo = {
    type: string;
    mimeType: string;
};

export function fileType(__: (s: string) => string, info: FileInfo): string {
    if (
        info.type !== "FILE" ||
        (info.mimeType !== "text/uri-list" && info.mimeType !== "text/uri")
    ) {
        return __("Document");
    }
    return __("Link");
}
