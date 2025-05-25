import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function variationToParams(variation: Record<string | number, string>): Record<string, string[]> {
  const params: Record<string, string[]> = {};
  
  for (const [key, value] of Object.entries(variation)) {
    if (value && value !== 'any') {
      if (!params.variation) {
        params.variation = [];
      }
      params.variation.push(`${key}:${value}`);
    }
  }
  
  return params;
}
