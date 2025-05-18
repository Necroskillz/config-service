import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { Input } from '~/components/ui/input';
import { Switch } from '~/components/ui/switch';
import { Button } from '~/components/ui/button';
import { seo, appTitle } from '~/utils/seo';
import { z } from 'zod';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { usePostUsers, usePutUsersUserId, useGetUsersUserId } from '~/gen';

interface UserFormData {
  username: string;
  password: string;
  isGlobalAdmin: boolean;
}

const Schema = z.object({
  userId: z.coerce.number().or(z.literal('create')),
});

export const Route = createFileRoute('/(admin)/admin/(users)/users/$userId')({
  component: RouteComponent,
  head: () => ({
    meta: [...seo({ title: appTitle(['User Details', 'Admin']) })],
  }),
  params: {
    parse: Schema.parse,
  },
});

function RouteComponent() {
  const { userId } = Route.useParams();
  const navigate = useNavigate();
  const isNewUser = userId === 'create';

  // queries
  const { data: userData } = useGetUsersUserId(isNewUser ? 0 : userId, {
    query: {
      enabled: !isNewUser,
    },
  });

  // mutations
  const createUser = usePostUsers({
    mutation: {
      onSuccess: async ({ newId: userId }) => {
        navigate({ to: '/admin/users' });
      },
    },
  });

  const updateUser = usePutUsersUserId({
    mutation: {
      onSuccess: async () => {
        navigate({ to: '/admin/users' });
      },
    },
  });

  const validator = z.object({
    username: z
      .string()
      .min(1, 'Username is required')
      .max(100, 'Username must have at most 100 characters')
      .regex(/^[\w\-_\.]+$/, 'Allowed characters: letters, numbers, -, _ and .'),
    password: isNewUser ? z.string().min(1, 'Password is required').max(100, 'Password must have at most 100 characters') : z.string(),
    isGlobalAdmin: z.boolean(),
  });

  const form = useAppForm({
    defaultValues: {
      username: userData?.username ?? '',
      password: '',
      isGlobalAdmin: userData?.globalAdministrator ?? false,
    } as UserFormData,
    validators: {
      onChange: validator,
      onMount: isNewUser ? validator : undefined,
    },
    onSubmit: async ({ value }) => {
      console.log('clicko');
      if (isNewUser) {
        await createUser.mutateAsync({ data: { name: value.username, password: value.password, globalAdministrator: value.isGlobalAdmin } });
      } else {
        await updateUser.mutateAsync({ user_id: userId, data: { globalAdministrator: value.isGlobalAdmin } });
      }
    },
  });

  return (
    <div className="p-4">
      <h2 className="text-2xl font-semibold mb-6">{isNewUser ? 'Create New User' : 'Edit User'}</h2>
      
      <form.AppForm>
        <form
          className="space-y-6 max-w-md"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
        >
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
                    disabled={!isNewUser}
                  />
                </field.FormControl>
                <field.FormMessage />
              </>
            )}
          />

          {isNewUser && (
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
                      disabled={!isNewUser}
                    />
                  </field.FormControl>
                  <field.FormMessage />
                </>
              )}
            />
          )}

          <form.AppField
            name="isGlobalAdmin"
            children={(field) => (
              <div className="flex items-center space-x-2">
                <field.FormControl>
                  <Switch
                    id={field.name}
                    name={field.name}
                    checked={field.state.value}
                    onCheckedChange={field.handleChange}
                  />
                </field.FormControl>
                <field.FormLabel htmlFor={field.name}>Global Administrator</field.FormLabel>
                <field.FormMessage />
              </div>
            )}
          />

          <div className="flex gap-4">
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit || isSubmitting}>
                  {isNewUser ? 'Create User' : 'Save Changes'}
                </Button>
              )}
            />
          </div>
        </form>
      </form.AppForm>
    </div>
  );
} 