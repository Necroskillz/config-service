import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import {
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysNameTakenName,
  getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions,
  getServicesServiceVersionIdQueryOptions,
  getValueTypesQueryOptions,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdSuspense,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense,
  useGetValueTypes,
  usePostServicesServiceVersionIdFeaturesFeatureVersionIdKeys,
  usePutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId,
} from '~/gen';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectGroup, SelectLabel, SelectItem } from '~/components/ui/select';
import { useEffect } from 'react';
import { MutationErrors } from '~/components/MutationErrors';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { Textarea } from '~/components/ui/textarea';
import { versionedTitle } from '~/utils/seo';
import { appTitle } from '~/utils/seo';
import { seo } from '~/utils/seo';
import { useChangeset } from '~/hooks/useChangeset';
import { Label } from '~/components/ui/label';

export const Route = createFileRoute('/(keys)/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/edit')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
      featureVersionId: z.coerce.number(),
      keyId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    context.queryClient.ensureQueryData(getValueTypesQueryOptions());

    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions(params.serviceVersionId, params.featureVersionId)
      ),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdQueryOptions(
          params.serviceVersionId,
          params.featureVersionId,
          params.keyId
        )
      ),
    ]);
  },
  head: ({ loaderData: [serviceVersion, featureVersion, key] }) => {
    return {
      meta: [...seo({ title: appTitle([key.name, versionedTitle(featureVersion), versionedTitle(serviceVersion)]) })],
      description: key.description,
    };
  },
});

function RouteComponent() {
  const { serviceVersionId, featureVersionId, keyId } = Route.useParams();
  const navigate = useNavigate();
  const { refresh } = useChangeset();
  const { data: featureVersion } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense(serviceVersionId, featureVersionId);
  const { data: key } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdSuspense(serviceVersionId, featureVersionId, keyId);

  const mutation = usePutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId({
    mutation: {
      onSuccess: async () => {
        navigate({
          to: '/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/values',
          params: { serviceVersionId, featureVersionId, keyId },
        });
        refresh();
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      description: '',
    },
    validators: {
      onChange: z.object({
        description: z.string(),
      }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({
        service_version_id: serviceVersionId,
        feature_version_id: featureVersionId,
        key_id: keyId,
        data: value,
      });
    },
  });

  return (
    <SlimPage>
      <PageTitle>Update Key</PageTitle>
      <div className="text-muted-foreground mb-4">
        <p>
          Update key {key.name} of {featureVersion?.name} v{featureVersion?.version}
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
          <div className="flex flex-col gap-2">
            <Label>Name</Label>
            <Input type="text" id="name" name="name" value={key.name} disabled />
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
            <Label>Value Type</Label>
            <Input type="text" id="valueType" name="valueType" value={key.valueType} disabled />
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
