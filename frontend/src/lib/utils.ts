import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function variationToQueryParams(variation: Record<string | number, string> | undefined): string[] | undefined {
  if (!variation) {
    return undefined;
  }

  const params: string[] = [];

  for (const [key, value] of Object.entries(variation)) {
    if (value && value !== 'any') {
      params.push(`${key}:${value}`);
    }
  }

  return params.length > 0 ? params : undefined;
}

export function variationToRequestParams(variation: Record<string | number, string>): Record<string, string> {
  const params: Record<string | number, string> = {};

  for (const [key, value] of Object.entries(variation)) {
    if (value && value !== 'any') {
      params[key] = value;
    }
  }

  return params;
}
