import { cn } from '~/lib/utils';

export function PageTitle({ children, className }: { children: React.ReactNode; className?: string }) {
  return <h1 className={cn('scroll-m-20 text-4xl mb-8', className)}>{children}</h1>;
}
