import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '~/lib/utils';

export function List({ children }: { children: React.ReactNode }) {
  return <div className="flex flex-col border border-ring rounded-md">{children}</div>;
}

const listItemVariants = cva('flex flex-col gap-2 border-b border-ring p-4 last:border-b-0', {
  variants: {
    variant: {
      default: 'p-4',
      slim: 'p-2',
    },
  },
  defaultVariants: {
    variant: 'default',
  },
});

export function ListItem({ children, variant }: { children: React.ReactNode } & VariantProps<typeof listItemVariants>) {
  return <div className={cn(listItemVariants({ variant }))}>{children}</div>;
}
