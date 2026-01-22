export function getKey<T>(item: T): string | undefined {
  if (
    item
    && typeof item === "object"
    && "id" in item
    && (typeof item.id === "string" || typeof item.id === "number")
  ) {
    return item.id.toString();
  }
  if (typeof item === "string" || typeof item === "number") {
    return item.toString();
  }
  if (item === undefined) {
    return undefined;
  }
  console.error("Cannot compute a key from item", item);
  return "";
}
