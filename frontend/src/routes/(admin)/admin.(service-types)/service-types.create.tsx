import { useQueryClient } from '@tanstack/react-query';
import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import { MutationErrors } from '~/components/MutationErrors';
import { PageTitle } from '~/components/PageTitle';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { getServiceTypesQueryOptions, getVariationPropertiesNameTakenName, usePostServiceTypes } from '~/gen';

export const Route = createFileRoute('/(admin)/admin/(service-types)/service-types/create')({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const mutation = usePostServiceTypes({
    mutation: {
      onSuccess: async ({ newId }) => {
        navigate({ to: '/admin/service-types/$serviceTypeId', params: { serviceTypeId: newId } });
        queryClient.refetchQueries(getServiceTypesQueryOptions());
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      name: '',
    },
    validators: {
      onChange: z.object({
        name: z
          .string()
          .min(1, 'Name is required')
          .max(50, 'Name must have at most 50 characters')
          .regex(/^[\w\-_\. ]+$/, 'Name must not contain invalid characters'),
      }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ data: value });
    },
  });

  return (
    <div className="w-[720px]">
      <PageTitle>Create Service Type</PageTitle>

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
