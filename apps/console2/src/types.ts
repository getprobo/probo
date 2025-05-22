export type NodeOf<T> = T extends
  | { readonly edges: ReadonlyArray<{ readonly node: infer U }> }
  | undefined
  ? U
  : never;

export type ItemOf<T> = T extends (infer U)[] ? U : never;
