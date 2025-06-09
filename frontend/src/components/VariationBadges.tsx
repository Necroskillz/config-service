import { Badge } from './ui/badge';

export function VariationBadges({ variation }: { variation: Record<string, string> }) {
  return (
    <div className="inline-flex flex-wrap gap-2 mb-2">
      {Object.values(variation).map((value) => (
        <Badge key={value} variant="outline">
          {value}
        </Badge>
      ))}
    </div>
  );
}
