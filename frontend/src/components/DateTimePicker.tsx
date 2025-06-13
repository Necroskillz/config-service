import { DatePicker } from './DatePicker';
import { Input } from './ui/input';
import { cn } from '~/lib/utils';

function formatTime(date: Date | undefined) {
  if (!date) {
    return '00:00:00';
  }
  const [hours, minutes, seconds] = [date.getHours(), date.getMinutes(), date.getSeconds()];
  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}

function combineDateAndTime(date: Date | undefined, time: string) {
  if (!date) return undefined;
  const [h, m, s] = time.split(':').map(Number);
  const result = new Date(date);
  result.setHours(h, m, s, 0);
  return result;
}

export function DateTimePicker({
  id,
  value,
  onChange,
  className,
}: {
  id?: string;
  value?: Date;
  onChange?: (date?: Date) => void;
  className?: string;
}) {
  const handleDateChange = (newDate: Date | undefined) => {
    if (newDate && value) {
      onChange?.(combineDateAndTime(newDate, formatTime(value)));
    } else {
      onChange?.(newDate);
    }
  };

  const handleTimeChange = (newTime: string) => {
    if (value && newTime) {
      onChange?.(combineDateAndTime(value, newTime));
    }
  };

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div className="flex-1/2">
        <DatePicker id={id ? `${id}-date` : undefined} value={value} onChange={handleDateChange} />
      </div>
      <Input
        type="time"
        id={id ? `${id}-time` : undefined}
        step="1"
        value={formatTime(value)}
        onChange={(e) => handleTimeChange(e.target.value)}
        className="flex-1/2 bg-background appearance-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
      />
    </div>
  );
}
