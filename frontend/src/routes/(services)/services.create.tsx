import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import { getServicesNameTakenName, getServiceTypesQueryOptions, useGetServiceTypes, usePostServices } from '~/gen';
import { useEffect } from 'react';
import { Select, SelectItem, SelectGroup, SelectContent, SelectTrigger, SelectValue, SelectLabel } from '~/components/ui/select';
import { z } from 'zod';
import { MutationErrors } from '~/components/MutationErrors';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { Textarea } from '~/components/ui/textarea';
import { seo, appTitle } from '~/utils/seo';
import { useChangeset } from '~/hooks/useChangeset';
export const Route = createFileRoute('/(services)/services/create')({
  component: RouteComponent,
  loader: async ({ context }) => {
    context.queryClient.ensureQueryData(getServiceTypesQueryOptions());
  },
  head: () => {
    return {
      meta: [...seo({ title: appTitle(['Create Service']) })],
    };
  },
});

function RouteComponent() {
  const navigate = useNavigate();
  const { refresh } = useChangeset();
  const { data: serviceTypes, isLoading } = useGetServiceTypes({
    query: {
      staleTime: Infinity,
    },
  });
  const mutation = usePostServices({
    mutation: {
      onSuccess: async ({ newId }) => {
        navigate({ to: '/services/$serviceVersionId', params: { serviceVersionId: newId } });
        refresh();
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      name: '',
      description: '',
      serviceTypeId: 0,
    },
    validators: {
      onChange: z.object({
        name: z.string().min(1, 'Name is required'),
        description: z.string().min(1, 'Description is required'),
        serviceTypeId: z.number().min(1, 'Service type is required'),
      }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ data: value });
    },
  });

  useEffect(() => {
    if (!isLoading && !form.state.values.serviceTypeId && serviceTypes?.[0].value) {
      form.setFieldValue('serviceTypeId', parseInt(serviceTypes?.[0].value), {
        dontUpdateMeta: true,
      });
    }
  }, [isLoading]);

  return (
    <SlimPage>
      <PageTitle>Create Service</PageTitle>

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

                const response = await getServicesNameTakenName(value);
                const isTaken = response.value;
                if (isTaken) {
                  return 'Service name is already taken';
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
          <form.AppField
            name="serviceTypeId"
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Service Type</field.FormLabel>
                <field.FormControl>
                  <Select
                    name={field.name}
                    value={field.state.value.toString()}
                    onValueChange={(value) => {
                      if (!value) {
                        return;
                      }
                      field.handleChange(parseInt(value));
                    }}
                  >
                    <SelectTrigger id={field.name} className="w-[180px]">
                      <SelectValue placeholder="Select a service type" />
                    </SelectTrigger>
                    <SelectContent>
                      {isLoading && (
                        <SelectGroup>
                          <SelectLabel>Loading...</SelectLabel>
                        </SelectGroup>
                      )}
                      {serviceTypes?.map((serviceType) => (
                        <SelectItem key={serviceType.value} value={serviceType.value}>
                          {serviceType.text}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
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
