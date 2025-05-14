import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import { z } from 'zod';
import {
  getServicesServiceVersionIdFeaturesNameTakenName,
  getServicesServiceVersionIdQueryOptions,
  useGetServicesServiceVersionIdSuspense,
  usePostServicesServiceVersionIdFeatures,
} from '~/gen';
import { seo, appTitle, versionedTitle } from '~/utils/seo';
import { MutationErrors } from '~/components/MutationErrors';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { Textarea } from '~/components/ui/textarea';
import { useChangeset } from '~/hooks/useChangeset';
import { Breadcrumbs } from '~/components/Breadcrumbs';

export const Route = createFileRoute('/(features)/services/$serviceVersionId/features/create')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId));
  },
  head: ({ loaderData: serviceVersion }) => {
    return {
      meta: [...seo({ title: appTitle(['Create Feature', versionedTitle(serviceVersion)]) })],
    };
  },
});

function RouteComponent() {
  const { serviceVersionId } = Route.useParams();
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);

  const navigate = useNavigate();
  const { refresh } = useChangeset();
  const mutation = usePostServicesServiceVersionIdFeatures({
    mutation: {
      onSuccess: ({ newId }) => {
        navigate({
          to: '/services/$serviceVersionId/features/$featureVersionId',
          params: { featureVersionId: newId, serviceVersionId },
        });
        refresh();
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      name: '',
      description: '',
    },
    validators: {
      onChange: z.object({
        name: z
          .string()
          .min(1, 'Name is required')
          .max(100, 'Name must have at most 100 characters')
          .regex(/^[\w\-_\.]+$/, 'Allowed characters: letters, numbers, -, _ and .'),
        description: z.string().min(1, 'Description is required').max(1000, 'Description must have at most 1000 characters'),
      }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ service_version_id: serviceVersionId, data: value });
    },
  });

  return (
    <SlimPage>
      <Breadcrumbs path={[serviceVersion]} />
      <PageTitle>Create Feature</PageTitle>

      <form.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
        >
          <MutationErrors mutations={[mutation]} />
          <form.AppField
            name="name"
            validators={{
              onChangeAsync: async ({ value }) => {
                if (!value) {
                  return;
                }

                const response = await getServicesServiceVersionIdFeaturesNameTakenName(serviceVersionId, value);
                const isTaken = response.value;
                if (isTaken) {
                  return 'Feature name is already taken';
                }
              },
            }}
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Name</field.FormLabel>
                <field.FormControl>
                  <Input
                    type="text"
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                  />
                </field.FormControl>
                <field.FormMessage />
              </>
            )}
          />
          <form.AppField
            name="description"
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Description</field.FormLabel>
                <field.FormControl>
                  <Textarea
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                  />
                </field.FormControl>
                <field.FormMessage />
              </>
            )}
          />
          <div>
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit || isSubmitting}>
                  Create
                </Button>
              )}
            />
          </div>
        </form>
      </form.AppForm>
    </SlimPage>
  );
}
