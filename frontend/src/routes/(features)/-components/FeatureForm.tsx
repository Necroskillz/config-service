import { Button } from '~/components/ui/button';
import { useAppForm } from '~/components/ui/tanstack-form';
import { z } from 'zod';
import { Input } from '~/components/ui/input';
import { Textarea } from '~/components/ui/textarea';
import { useNavigate } from '@tanstack/react-router';
import { getServicesServiceVersionIdFeaturesNameTakenName, usePostServicesServiceVersionIdFeatures } from '~/gen';
import { MutationErrors } from '~/components/MutationErrors';

const Schema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().min(1, 'Description is required'),
});

export function FeatureForm({ serviceVersionId }: { serviceVersionId: number }) {
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
      onChange: Schema,
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ service_version_id: serviceVersionId, data: value });
    },
  });

  return (
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
            <field.FormItem>
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
            </field.FormItem>
          )}
        />
        <form.AppField
          name="description"
          children={(field) => (
            <field.FormItem>
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
            </field.FormItem>
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
  );
}
