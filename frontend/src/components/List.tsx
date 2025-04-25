export function List({ children }: { children: React.ReactNode }) {
  return <div className="flex flex-col border border-ring rounded-md">{children}</div>;
}

export function ListItem({ children }: { children: React.ReactNode }) {
  return <div className="flex flex-col gap-2 border-b border-ring p-4 last:border-b-0">{children}</div>;
}
