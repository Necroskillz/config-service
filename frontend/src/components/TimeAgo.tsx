import { useEffect } from 'react';

import { useMemo, useState } from 'react';

const SECOND = 1;
const MINUTE = 60 * SECOND;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;
const WEEK = 7 * DAY;
const MONTH = 30 * DAY;
const YEAR = 365 * DAY;

const intervals = [
  { seconds: YEAR, fn: (years: number) => (years === 1 ? 'a year ago' : `${years} years ago`) },
  { seconds: MONTH, fn: (months: number) => (months === 1 ? 'a month ago' : `${months} months ago`) },
  { seconds: WEEK, fn: (weeks: number) => (weeks === 1 ? 'a week ago' : `${weeks} weeks ago`) },
  { seconds: DAY, fn: (days: number) => (days === 1 ? 'yesterday' : `${days} days ago`) },
  { seconds: HOUR, fn: (hours: number) => (hours === 1 ? '1 hour ago' : `${hours} hours ago`) },
  { seconds: MINUTE, fn: (minutes: number) => (minutes === 1 ? '1 minute ago' : `${minutes} minutes ago`) },
];

function getRelativeTime(date: Date) {
  const now = new Date();
  const diff = (now.getTime() - date.getTime()) / 1000;

  for (const interval of intervals) {
    if (diff >= interval.seconds) {
      const value = Math.floor(diff / interval.seconds);
      return interval.fn(value);
    }
  }

  return 'just now';
}

export function TimeAgo({ datetime }: { datetime: string }) {
  const date = useMemo(() => new Date(datetime), [datetime]);
  const [relativeTime, setRelativeTime] = useState(getRelativeTime(date));
  const formattedDate = useMemo(() => date.toLocaleString(), [date]);
  
  useEffect(() => {
    const timer = setInterval(() => {
      setRelativeTime(getRelativeTime(date));
    }, 1000 * 60);
    return () => clearInterval(timer);
  }, [date]);

  return <time title={formattedDate}>{relativeTime}</time>;
}
