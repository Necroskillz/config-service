"use client";

import { Input } from "~/components/ui/input";
import { SearchIcon } from "lucide-react";

export default function SearchBox({ ...props }: React.ComponentProps<"input">) {
  return (
    <div className="relative flex items-center rounded-md border focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50 pl-2 dark:bg-input/30">
      <SearchIcon className="h-5 w-5 text-muted-foreground" />
      <Input 
        type="text"
        className="border-0 focus-visible:ring-0 shadow-none dark:bg-transparent"
        {...props}
      />
    </div>
  );
}
