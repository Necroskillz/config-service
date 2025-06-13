const sizes = {
  xs: 'max-w-[720px]',
  sm: 'max-w-[1080px]',
  md: 'max-w-[1280px]',
  lg: 'max-w-[1440px]',
}

export function SlimPage({ children, size = 'md' }: { children: React.ReactNode, size?: keyof typeof sizes }) {
  return (
    <div className={`mx-auto p-4 ${sizes[size]}`}>
      {children}
    </div>
  );
}