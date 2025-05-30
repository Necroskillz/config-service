import { VariantProps } from 'class-variance-authority';
import { Badge, badgeVariants } from '~/components/ui/badge';
import { DbChangesetStateEnum } from '~/gen';

function getStateBadgeVariant(state: DbChangesetStateEnum): VariantProps<typeof badgeVariants>['variant'] {
  switch (state) {
    case 'open':
      return 'default';
    case 'discarded':
      return 'destructive';
    case 'applied':
      return 'secondary';
    case 'stashed':
      return 'outline';
    case 'committed':
      return 'outline';
    default:
      return 'default';
  }
}

export function ChangesetState({ state }: { state: DbChangesetStateEnum }) {
  return <Badge variant={getStateBadgeVariant(state)}>{state}</Badge>;
}
