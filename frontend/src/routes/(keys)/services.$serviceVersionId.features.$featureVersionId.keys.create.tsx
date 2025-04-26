import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import {
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysNameTakenName,
  getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions,
  getServicesServiceVersionIdQueryOptions,
  getValueTypesQueryOptions,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense,
  useGetValueTypes,
  usePostServicesServiceVersionIdFeaturesFeatureVersionIdKeys,
} from '~/gen';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectGroup, SelectLabel, SelectItem } from '~/components/ui/select';
import { useEffect } from 'react';
import { MutationErrors } from '~/components/MutationErrors';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form';
import { Textarea } from '~/components/ui/textarea';
import { versionedTitle } from '~/utils/seo';
import { appTitle } from '~/utils/seo';
import { seo } from '~/utils/seo';

export const Route = createFileRoute('/(keys)/services/$serviceVersionId/features/$featureVersionId/keys/create')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
      featureVersionId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    context.queryClient.ensureQueryData(getValueTypesQueryOptions());

    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions(params.serviceVersionId, params.featureVersionId)
      ),
    ]);
  },
  head: ({ loaderData: [serviceVersion, featureVersion] }) => {
    return {
      meta: [...seo({ title: appTitle(['Create Key', versionedTitle(featureVersion), versionedTitle(serviceVersion)]) })],
    };
  },
});

function RouteComponent() {
  const { serviceVersionId, featureVersionId } = Route.useParams();
  const navigate = useNavigate();
  const { data: featureVersion } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense(serviceVersionId, featureVersionId);
  const { data: valueTypes, isLoading } = useGetValueTypes({
    query: {
      staleTime: Infinity,
    },
  });
  
  const mutation = usePostServicesServiceVersionIdFeaturesFeatureVersionIdKeys({
    mutation: {
      onSuccess: async ({ newId }) => {
        navigate({
          to: '/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/values',
          params: { serviceVersionId, featureVersionId, keyId: newId },
        });
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      name: '',
      description: '',
      valueTypeId: 0,
      defaultValue: '',
    },
    validators: {
      onChange: z.object({
        name: z.string().min(1, 'Name is required'),
        description: z.string(),
        valueTypeId: z.number().min(1, 'Value type is required'),
        defaultValue: z.string(),
      }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ service_version_id: serviceVersionId, feature_version_id: featureVersionId, data: value });
    },
  });

  useEffect(() => {
    if (!isLoading && !form.state.values.valueTypeId && valueTypes?.[0].value) {
      form.setFieldValue('valueTypeId', parseInt(valueTypes?.[0].value), {
        dontUpdateMeta: true,
      });
    }
  }, [isLoading]);

  return (
    <SlimPage>
      <PageTitle>Create Key</PageTitle>
      <div className="text-muted-foreground mb-4">
        <p>
          Create a new key for {featureVersion?.name} v{featureVersion?.version}
        </p>
      </div>
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

                const response = await getServicesServiceVersionIdFeaturesFeatureVersionIdKeysNameTakenName(
                  serviceVersionId,
                  featureVersionId,
                  value
                );
                const isTaken = response.value;
                if (isTaken) {
                  return 'Key name is already taken';
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
          <form.AppField
            name="valueTypeId"
            children={(field) => (
              <field.FormItem>
                <field.FormLabel htmlFor={field.name}>Value Type</field.FormLabel>
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
                      <SelectValue placeholder="Select a value type" />
                    </SelectTrigger>
                    <SelectContent>
                      {isLoading && (
                        <SelectGroup>
                          <SelectLabel>Loading...</SelectLabel>
                        </SelectGroup>
                      )}
                      {valueTypes?.map((valueType) => (
                        <SelectItem key={valueType.value} value={valueType.value}>
                          {valueType.text}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </field.FormControl>
                <field.FormMessage />
              </field.FormItem>
            )}
          />
          <form.AppField
            name="defaultValue"
            children={(field) => (
              <field.FormItem>
                <field.FormLabel htmlFor={field.name}>Default Value</field.FormLabel>
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
