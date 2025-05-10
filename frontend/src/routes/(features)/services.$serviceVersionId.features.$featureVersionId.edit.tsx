import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import { getServicesServiceVersionIdQueryOptions, getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions, usePutServicesServiceVersionIdFeaturesFeatureVersionId, useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense, useGetServicesServiceVersionIdSuspense } from '~/gen';
import { z } from 'zod';
import { Label } from '@radix-ui/react-dropdown-menu';
import { MutationErrors } from '~/components/MutationErrors';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { Textarea } from '~/components/ui/textarea';
import { versionedTitle, seo, appTitle } from '~/utils/seo';
import { Breadcrumbs } from '~/components/Breadcrumbs';

export const Route = createFileRoute('/(features)/services/$serviceVersionId/features/$featureVersionId/edit')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
      featureVersionId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions(params.serviceVersionId, params.featureVersionId)),
    ]);
  },
  head: (ctx) => {
    const [serviceVersion, featureVersion] = ctx.loaderData;

    return {
      meta: [
        ...seo({
          title: appTitle(['Edit', versionedTitle(featureVersion), versionedTitle(serviceVersion)]),
          description: featureVersion.description,
        }),
      ],
    };
  },
});

function RouteComponent() {
  const { serviceVersionId, featureVersionId } = Route.useParams();
  const { data: featureVersion } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense(serviceVersionId, featureVersionId);
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const navigate = useNavigate();
  const mutation = usePutServicesServiceVersionIdFeaturesFeatureVersionId({
    mutation: {
      onSuccess: async () => {
        navigate({ to: '/services/$serviceVersionId/features/$featureVersionId', params: { serviceVersionId, featureVersionId } });
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      description: featureVersion.description,
    },
    validators: {
      onChange: z.object({
        description: z.string().min(1, 'Description is required'),
      }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ service_version_id: serviceVersionId, feature_version_id: featureVersionId, data: value });
    },
  });

  return (
    <SlimPage>
      <Breadcrumbs path={[serviceVersion, featureVersion]} />
      <PageTitle>Edit feature</PageTitle>

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
          <div className="flex flex-col gap-2">
            <Label>Name</Label>
            <Input type="text" id="name" name="name" value={featureVersion.name} disabled />
          </div>
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
                  Save
                </Button>
              )}
            />
          </div>
        </form>
      </form.AppForm>
    </SlimPage>
  );
}
