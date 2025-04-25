export function SlimPage({ children }: { children: React.ReactNode }) {
  return (
    <div className="max-w-[1280px] mx-auto p-4">
      {children}
    </div>
  );
}