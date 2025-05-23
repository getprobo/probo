export function times<T>(n: number, cb: (i: number) => T): T[] {
    return Array.from({ length: n }, (_, i) => cb(i));
}

export function groupBy<T>(
    arr: T[],
    key: (item: T) => string,
): Record<string, T[]> {
    return arr.reduce(
        (acc, item) => {
            const k = key(item);
            if (!acc[k]) {
                acc[k] = [];
            }
            acc[k].push(item);
            return acc;
        },
        {} as Record<string, T[]>,
    );
}

/**
 * Check that a value is empty (null, undefined, empty string, empty array, empty object)
 */
export function isEmpty(v: unknown): boolean {
    if (Array.isArray(v)) {
        return v.find((v) => !isEmpty(v)) === undefined;
    }
    return !v;
}
