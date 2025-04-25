import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import { ChangesetDetail } from './-components/ChangesetDetail';
import { PageTitle } from '~/components/PageTitle';
import { Suspense } from 'react';
const Schema = z.object({
  changesetId: z.coerce.number(),
});

export const Route = createFileRoute('/(changesets)/changesets/$changesetId')({
  component: RouteComponent,
  params: {
    parse: Schema.parse,
  },
});

function RouteComponent() {
  const { changesetId } = Route.useParams();

  return (
    <SlimPage>
      <PageTitle>Changeset #{changesetId}</PageTitle>
      <Suspense fallback={<div>Loading...</div>}>
        <ChangesetDetail changesetId={changesetId} />
      </Suspense>
    </SlimPage>
  );
}
