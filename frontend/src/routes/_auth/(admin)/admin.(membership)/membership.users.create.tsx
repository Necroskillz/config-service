import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { Input } from '~/components/ui/input';
import { Switch } from '~/components/ui/switch';
import { Button } from '~/components/ui/button';
import { seo, appTitle } from '~/utils/seo';
import { z } from 'zod';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { usePostMembershipUsers } from '~/gen';
import { MutationErrors } from '~/components/MutationErrors';

export const Route = createFileRoute('/_auth/(admin)/admin/(membership)/membership/users/create')({
  component: RouteComponent,
  head: () => ({
    meta: [...seo({ title: appTitle(['Create User', 'Admin']) })],
  }),
});

function RouteComponent() {
  const navigate = useNavigate();

  const createUser = usePostMembershipUsers({
    mutation: {
      onSuccess: async () => {
        navigate({ to: '/admin/membership' });
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      username: '',
      password: '',
      globalAdministrator: false,
    },
    validators: {
      onChange: z.object({
        username: z
          .string()
          .min(1, 'Username is required')
          .max(100, 'Username must have at most 100 characters')
          .regex(/^[\w\-_\.]+$/, 'Allowed characters: letters, numbers, -, _ and .'),
        password: z.string().min(8, 'Password must be at least 8 characters'),
        globalAdministrator: z.boolean(),
      }),
    },
    onSubmit: async ({ value }) => {
      await createUser.mutateAsync({ data: value });
    },
  });

  return (
    <div className="p-4 w-[720px]">
      <h2 className="text-2xl font-semibold mb-6">Create New User</h2>
      <form.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
        >
          <MutationErrors mutations={[createUser]} />
          <form.AppField
            name="username"
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Username</field.FormLabel>
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
            name="password"
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Password</field.FormLabel>
                <field.FormControl>
                  <Input
                    type="password"
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
            name="globalAdministrator"
            children={(field) => (
              <div className="flex items-center space-x-2">
                <field.FormControl>
                  <Switch id={field.name} name={field.name} checked={field.state.value} onCheckedChange={field.handleChange} />
                </field.FormControl>
                <field.FormLabel htmlFor={field.name}>Global Administrator</field.FormLabel>
                <field.FormMessage />
              </div>
            )}
          />

          <div>
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit || isSubmitting}>
                  Create User
                </Button>
              )}
            />
          </div>
        </form>
      </form.AppForm>
    </div>
  );
}
