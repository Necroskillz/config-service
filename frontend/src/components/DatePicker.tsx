import { ChevronDownIcon } from 'lucide-react';
import { useState } from 'react';
import { Button } from './ui/button';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { Calendar } from './ui/calendar';

export function DatePicker({ id, value, onChange }: { id?: string; value?: Date; onChange?: (date?: Date) => void }) {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" id={id} className="w-full justify-between font-normal">
          {value
            ? value.toLocaleDateString('en-US', {
                day: '2-digit',
                month: 'short',
                year: 'numeric',
              })
            : 'Select date'}
          <ChevronDownIcon />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto overflow-hidden p-0" align="start">
        <Calendar
          mode="single"
          selected={value}
          captionLayout="dropdown"
          onSelect={(date) => {
            onChange?.(date);
            setOpen(false);
          }}
        />
      </PopoverContent>
    </Popover>
  );
}
