import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function variationToQueryParams(variation: Record<string | number, string> | undefined): string[] | undefined {
  if (!variation || Object.keys(variation).length === 0) {
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

export function queryParamsToVariation(queryParams: string[]): Record<string, string> {
  const variation: Record<string, string> = {};

  for (const param of queryParams) {
    const [key, value] = param.split(':');
    variation[key] = value;
  }

  return variation;
}
