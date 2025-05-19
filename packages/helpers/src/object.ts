/**
 * A type safe version of Object.keys
 */
export function objectKeys<T>(object: T) {
    return Object.keys(object) as (keyof T)[];
}
