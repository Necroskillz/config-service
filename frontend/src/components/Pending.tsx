import { Spinner } from "./ui/spinner";

export function Pending() {
  return (
    <div className="flex flex-col items-center justify-center h-screen">
      <Spinner size="large" />
    </div>
  );
}

