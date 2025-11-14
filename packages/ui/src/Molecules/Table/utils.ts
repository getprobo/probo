export function getKey<T>(item: T): string {
    if (
        item &&
        typeof item === "object" &&
        "id" in item &&
        typeof item.id === "string"
    ) {
        return item.id.toString();
    }
    if (typeof item === "string" || typeof item === "number") {
        return item.toString();
    }
    if (item === undefined) {
        return "";
    }
    console.error("Cannot compute a key from item", item);
    return "";
}
