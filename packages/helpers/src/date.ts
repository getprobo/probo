export function formatDatetime(dateString?: string | null): string | undefined {
  if (!dateString) return undefined;
  return `${dateString}T00:00:00Z`;
}

export function formatDate(dateInput?: string | null): string {
  if (!dateInput) return '';

  const date = parseDate(dateInput);
  return date.toLocaleDateString();
}

function parseDate(dateString: string): Date {
  if (dateString.includes('T')) {
    return new Date(dateString);
  }
  const parts = dateString.split('-');
  return new Date(
    parseInt(parts[0], 10),
    parts[1] ? parseInt(parts[1], 10) - 1 : 0,
    parts[2] ? parseInt(parts[2], 10) : 1
  );
}
