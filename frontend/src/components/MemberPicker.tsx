import { useEffect, useRef, useState } from 'react';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { Button } from './ui/button';
import { Check, ChevronsUpDown } from 'lucide-react';
import { Command, CommandInput, CommandList, CommandEmpty, CommandGroup, CommandItem, CommandLoading } from './ui/command';
import { useGetMembership } from '~/gen/hooks/useGetMembership';
import { MembershipMembershipObjectDto } from '~/gen';
import { cn } from '~/lib/utils';
import { z } from 'zod';

export function requiredMember(message: string) {
  return z.custom<MembershipMembershipObjectDto | undefined>().refine((val) => val != null, message);
}

export function MemberPicker({
  value,
  type,
  onValueChange,
  onBlur,
}: {
  value?: MembershipMembershipObjectDto;
  type?: 'user' | 'group';
  onValueChange?: (value: MembershipMembershipObjectDto | undefined) => void;
  onBlur?: () => void;
}) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState<string | undefined>(undefined);
  const [selectedMembershipObject, setSelectedMembershipObject] = useState<MembershipMembershipObjectDto | undefined>(value);
  const interacted = useRef(value != null);

  function fetchMembershipObjects() {
    interacted.current = true;
  }

  useEffect(() => {
    setSelectedMembershipObject(value);
  }, [value]);

  const { data: membershipObjects, isLoading: membershipObjectsLoading } = useGetMembership(
    { page: 1, pageSize: 50, name: search, type },
    { query: { enabled: interacted.current } }
  );

  return (
    <Popover
      open={open}
      onOpenChange={(open) => {
        setOpen(open);
        if (!open) {
          onBlur?.();
        }
      }}
    >
      <PopoverTrigger asChild onMouseEnter={() => fetchMembershipObjects()} onTouchStart={() => fetchMembershipObjects()}>
        <Button variant="outline" role="combobox" aria-expanded={open} className="w-72 justify-between">
          {selectedMembershipObject ? selectedMembershipObject.name : 'Select a user or group'}
          <ChevronsUpDown className="opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="p-0">
        <Command shouldFilter={false}>
          <CommandInput value={search} onValueChange={(value) => setSearch(value || undefined)} placeholder="Search..." />
          <CommandList>
            <CommandEmpty>No results found</CommandEmpty>
            {membershipObjectsLoading && <CommandLoading>Loading...</CommandLoading>}
            <CommandGroup>
              {membershipObjects?.items.map((membershipObject) => (
                <CommandItem
                  key={membershipObject.id}
                  onSelect={() => {
                    setSearch(undefined);
                    setOpen(false);
                    setSelectedMembershipObject(membershipObject);
                    onValueChange?.(membershipObject);
                  }}
                >
                  {membershipObject.name}
                  <Check className={cn('ml-auto', membershipObject.id === selectedMembershipObject?.id ? 'opacity-100' : 'opacity-0')} />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
