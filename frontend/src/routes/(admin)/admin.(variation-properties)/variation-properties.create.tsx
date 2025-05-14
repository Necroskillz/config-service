import { useQueryClient } from '@tanstack/react-query';
import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import { MutationErrors } from '~/components/MutationErrors';
import { PageTitle } from '~/components/PageTitle';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { getVariationPropertiesNameTakenName, getVariationPropertiesQueryOptions, usePostVariationProperties } from '~/gen';
import { appTitle } from '~/utils/seo';
import { seo } from '~/utils/seo';

export const Route = createFileRoute('/(admin)/admin/(variation-properties)/variation-properties/create')({
  component: RouteComponent,
  head: () => ({
    meta: [...seo({ title: appTitle(['Create Property', 'Variation Properties', 'Admin']) })],
  }),
});

function RouteComponent() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const mutation = usePostVariationProperties({
    mutation: {
      onSuccess: async ({ newId }) => {
        navigate({ to: '/admin/variation-properties/$propertyId', params: { propertyId: newId } });
        queryClient.refetchQueries(getVariationPropertiesQueryOptions());
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      name: '',
      displayName: '',
    },
    validators: {
      onChange: z.object({
        name: z
          .string()
          .min(1, 'Name is required')
          .max(20, 'Name must have at most 20 characters')
          .regex(/^[a-z_\-]+$/, 'Name must not contain invalid characters'),
        displayName: z
          .string()
          .max(20, 'Display name must have at most 20 characters')
          .regex(/^[a-zA-Z\- ]*$/, 'Display name must not contain invalid characters'),
      }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ data: value });
    },
  });

  return (
    <div className="w-[720px]">
      <PageTitle>Create Variation Property</PageTitle>

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

                const response = await getVariationPropertiesNameTakenName(value);
                const isTaken = response.value;
                if (isTaken) {
                  return 'Variation property name is already taken';
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
            name="displayName"
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Display name</field.FormLabel>
                <field.FormControl>
                  <Input
                    placeholder="Defaults to name"
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
    </div>
  );
}
