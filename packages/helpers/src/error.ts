export interface GraphQLError {
  message?: string;
  source?: {
    errors?: Array<{ message: string }>;
  };
}

export function formatError(title: string, error: GraphQLError): string {
  const messages: string[] = [];

  if (error.source?.errors && Array.isArray(error.source.errors)) {
    messages.push(...error.source.errors.map((e) => e.message).filter(Boolean));
  }

  if (messages.length === 0 && error.message) {
    messages.push(error.message);
  }

  if (messages.length === 0) {
    messages.push(`${title}.`);
  }

  const errorList = messages.join(", ");

  return `${title}: ${errorList}.`;
}
