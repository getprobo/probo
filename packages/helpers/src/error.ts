export interface GraphQLError {
  message?: string;
  extensions?: {
    code?: string;
  };
  source?: {
    errors?: Array<{ message: string; extensions?: { code?: string } }>;
  };
}

export function formatError(title: string, error: GraphQLError | GraphQLError[]): string {
  const messages: string[] = [];

  if (Array.isArray(error)) {
    messages.push(...error.map((e) => e.message).filter(Boolean) as string[]);
  } else if (error.source?.errors && Array.isArray(error.source.errors)) {
    messages.push(...error.source.errors.map((e) => e.message).filter(Boolean));
  } else if (error.message) {
    messages.push(error.message);
  }

  if (messages.length === 0) {
    return title;
  }

  const errorList = messages.join(", ");

  return `${title}: ${errorList}${errorList.endsWith('.') ? '' : '.'}`;
}
