export function formatDatetime(dateString?: string | null): string | undefined {
  if (!dateString) return undefined;
  return `${dateString}T00:00:00Z`;
}

export function toDateInput(dateString?: string | null): string {
  if (!dateString) return '';
  return dateString.split('T')[0];
}

export function formatDate(dateInput?: string | null): string {
  if (!dateInput) return "";

  const date = parseDate(dateInput);
  return date.toLocaleDateString();
}

export function parseDate(dateString: string): Date {
  if (dateString.includes("T")) {
    return new Date(dateString);
  }
  const parts = dateString.split("-");
  return new Date(
    parseInt(parts[0], 10),
    parts[1] ? parseInt(parts[1], 10) - 1 : 0,
    parts[2] ? parseInt(parts[2], 10) : 1,
  );
}

export function formatDuration(duration?: string | null, __?: (s: string) => string): string | null {
  if (!duration || !__) return null;

  const timeMatch = duration.match(/PT(\d+)([MH])/);
  if (timeMatch) {
    const amount = parseInt(timeMatch[1], 10) || 0;
    const unit = timeMatch[2];
    if (unit === "M") {
      return `${amount} ${amount === 1 ? __("Minute") : __("Minutes")}`;
    } else if (unit === "H") {
      return `${amount} ${amount === 1 ? __("Hour") : __("Hours")}`;
    }
  }

  const dateMatch = duration.match(/P(\d+)([DW])/);
  if (dateMatch) {
    const amount = parseInt(dateMatch[1], 10) || 0;
    const unit = dateMatch[2];
    if (unit === "W") {
      return `${amount} ${amount === 1 ? __("Week") : __("Weeks")}`;
    } else if (unit === "D") {
      const days = amount;
      if (days % 7 === 0 && days > 0) {
        const weeks = days / 7;
        return `${weeks} ${weeks === 1 ? __("Week") : __("Weeks")}`;
      }
      return `${days} ${days === 1 ? __("Day") : __("Days")}`;
    }
  }

  return null;
}
