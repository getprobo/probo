import { useQuery, type UseQueryOptions } from "@tanstack/react-query";

/**
 * Wrapper around useQuery to fetch data from an API endpoint
 */
export function useFetchQuery<T>(
  url: string,
  options?: Omit<UseQueryOptions<T>, "queryKey" | "queryFn">,
) {
  return useQuery<T>({
    ...options,
    queryKey: [url],
    queryFn: () => fetch(url).then(res => res.json()),
  });
}
