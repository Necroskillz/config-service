import { createFileRoute } from '@tanstack/react-router';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import { seo, appTitle } from '~/utils/seo';

export const Route = createFileRoute('/(changesets)/changesets/empty')({
  component: RouteComponent,
  head: () => {
    return {
      meta: [...seo({ title: appTitle([`Empty Changeset`]) })],
    };
  },
});

function RouteComponent() {
  return (
    <SlimPage>
      <PageTitle>No Open Changeset</PageTitle>
      <div className="flex flex-col gap-4">
        <p>You haven't made any changes</p>
      </div>
    </SlimPage>
  );
}
