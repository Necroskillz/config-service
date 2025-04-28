import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import {
  getServicesQueryKey,
  getServicesServiceVersionIdQueryKey,
  getServicesServiceVersionIdQueryOptions,
  usePutServicesServiceVersionId,
} from '~/gen';
import { z } from 'zod';
import { Label } from '@radix-ui/react-dropdown-menu';
import { MutationErrors } from '~/components/MutationErrors';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { Textarea } from '~/components/ui/textarea';
import { useQueryClient } from '@tanstack/react-query';
import { versionedTitle, seo, appTitle } from '~/utils/seo';

export const Route = createFileRoute('/(services)/services/$serviceVersionId/edit')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId));
  },
  head: (ctx) => {
    const serviceVersion = ctx.loaderData;

    return {
      meta: [
        ...seo({
          title: appTitle(['Edit', versionedTitle(serviceVersion)]),
          description: serviceVersion.description,
        }),
      ],
    };
  },
});

function RouteComponent() {
  const serviceVersion = Route.useLoaderData();
  const navigate = useNavigate();
  const mutation = usePutServicesServiceVersionId({
    mutation: {
      onSuccess: async () => {
        navigate({ to: '/services/$serviceVersionId', params: { serviceVersionId: serviceVersion.id } });
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      description: serviceVersion.description,
    },
    validators: {
      onChange: z.object({
        description: z.string().min(1, 'Description is required'),
      }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ service_version_id: serviceVersion.id, data: value });
    },
  });

  return (
    <SlimPage>
      <PageTitle>Edit Service {serviceVersion.name}</PageTitle>
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
            <Input type="text" id="name" name="name" value={serviceVersion.name} disabled />
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
          <div className="flex flex-col gap-2">
            <Label>Service type</Label>
            <Input type="text" id="service_type" name="service_type" value={serviceVersion.serviceTypeName} disabled />
          </div>
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
