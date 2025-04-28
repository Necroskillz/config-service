import { createFileRoute, useRouter } from '@tanstack/react-router';
import { z } from 'zod';
import { useAuth } from '~/auth';
import { useMutation } from '@tanstack/react-query';
import { Input } from '~/components/ui/input';
import { Button } from '~/components/ui/button';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { Alert, AlertDescription } from '~/components/ui/alert';

export const Route = createFileRoute('/login')({
  component: RouteComponent,
});

const Schema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
});

function RouteComponent() {
  const router = useRouter();
  const { login } = useAuth();
  const mutation = useMutation({
    mutationFn: async (data: { username: string; password: string }) => {
      await login(data.username, data.password);
      router.navigate({ to: '/' });
    },
  });

  const form = useAppForm({
    defaultValues: {
      username: '',
      password: '',
    },
    validators: {
      onChange: Schema,
    },
    onSubmit: async ({ value }) => {
      try {
        await mutation.mutateAsync(value);
        router.navigate({ to: '/' });
      } catch (error: any) {
        console.error(error);
      }
    },
  });

  return (
    <div className="p-8 w-96 mx-auto space-y-12">
      <form.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
        >
          {mutation.error && (
            <Alert variant="destructive">
              <AlertDescription>{mutation.error.message}</AlertDescription>
            </Alert>
          )}
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
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    type="password"
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
                  Login
                </Button>
              )}
            />
          </div>
        </form>
      </form.AppForm>
    </div>
  );
}
