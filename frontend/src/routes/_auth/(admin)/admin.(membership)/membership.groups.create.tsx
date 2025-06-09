import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import { MutationErrors } from '~/components/MutationErrors';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { usePostMembershipGroups } from '~/gen';
import { appTitle } from '~/utils/seo';
import { seo } from '~/utils/seo';

export const Route = createFileRoute('/_auth/(admin)/admin/(membership)/membership/groups/create')({
  component: RouteComponent,
  head: () => ({
    meta: [...seo({ title: appTitle(['Create Group', 'Admin']) })],
  }),
});

function RouteComponent() {
  const navigate = useNavigate();
  const createGroup = usePostMembershipGroups({
    mutation: {
      onSuccess: async ({ newId }) => {
        navigate({ to: '/admin/membership/groups/$groupId', params: { groupId: newId } });
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
          .max(100, 'Name must have at most 100 characters')
          .regex(/^[\w\-_\.]+$/, 'Allowed characters: letters, numbers, - and .'),
      }),
    },
    onSubmit: async ({ value }) => {
      await createGroup.mutateAsync({
        data: value,
      });
    },
  });

  return (
    <div className="p-4 w-[720px]">
      <h2 className="text-2xl font-semibold mb-6">Create New Group</h2>
      <form.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
        >
          <MutationErrors mutations={[createGroup]} />
          <form.AppField
            name="name"
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
                  Create Group
                </Button>
              )}
            />
          </div>
        </form>
      </form.AppForm>
    </div>
  );
}
