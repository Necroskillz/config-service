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
  const mutation = usePostServicesServiceVersionIdFeatures({
    mutation: {
      onSuccess: ({ newId }) => {
        navigate({
          to: '/services/$serviceVersionId/features/$featureVersionId',
          params: { featureVersionId: newId, serviceVersionId },
        });
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
        name: z.string().min(1, 'Name is required'),
        description: z.string().min(1, 'Description is required'),
      }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ service_version_id: serviceVersionId, data: value });
    },
  });

  return (
    <SlimPage>
      <PageTitle>Create Feature</PageTitle>

      <div className="text-muted-foreground mb-4">
        <p>
          Created feature will be linked to{' '}
          <Link className="text-accent-foreground" to="/services/$serviceVersionId" params={{ serviceVersionId }}>
            {serviceVersion?.name} v{serviceVersion?.version}
          </Link>
        </p>
      </div>

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
