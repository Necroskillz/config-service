import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';

const Schema = z.object({
  userId: z.coerce.number(),
});

export const Route = createFileRoute('/(users)/users/$userId')({
  component: RouteComponent,
  params: {
    parse: Schema.parse,
  },
});

function RouteComponent() {
  return <div>Hello "/(users)/user/$userId"!</div>;
}
