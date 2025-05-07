import { useQueryClient } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { List, ListItem } from '~/components/List';
import { MutationErrors } from '~/components/MutationErrors';
import { PageTitle } from '~/components/PageTitle';
import { SlimPage } from '~/components/SlimPage';
import { Badge } from '~/components/ui/badge';
import { Button } from '~/components/ui/button';
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '~/components/ui/select';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import {
  useGetServicesServiceVersionIdFeaturesSuspense,
  useGetServicesServiceVersionIdFeaturesLinkableSuspense,
  useGetServicesServiceVersionIdSuspense,
  useDeleteServicesServiceVersionIdFeaturesFeatureVersionIdUnlink,
  getServicesServiceVersionIdFeaturesQueryKey,
  getServicesServiceVersionIdFeaturesLinkableQueryKey,
  usePostServicesServiceVersionIdFeaturesFeatureVersionIdLink,
  getServicesServiceVersionIdQueryOptions,
  getServicesServiceVersionIdFeaturesQueryOptions,
  getServicesServiceVersionIdFeaturesLinkableQueryOptions,
} from '~/gen';
import { useChangeset } from '~/hooks/useChangeset';
import { appTitle, versionedTitle, seo } from '~/utils/seo';

export const Route = createFileRoute('/(services)/services/$serviceVersionId/link')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(getServicesServiceVersionIdFeaturesQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(getServicesServiceVersionIdFeaturesLinkableQueryOptions(params.serviceVersionId)),
    ]);
  },
  head: ({ loaderData: [serviceVersion] }) => {
    return {
      meta: [...seo({ title: appTitle(['Link/Unlink features', versionedTitle(serviceVersion)]) })],
    };
  },
});

function RouteComponent() {
  const { serviceVersionId } = Route.useParams();
  const queryClient = useQueryClient();
  const { refresh } = useChangeset();

  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const { data: features } = useGetServicesServiceVersionIdFeaturesSuspense(serviceVersionId);
  const { data: linkableFeatures } = useGetServicesServiceVersionIdFeaturesLinkableSuspense(serviceVersionId);

  function refetchData() {
    queryClient.refetchQueries({ queryKey: getServicesServiceVersionIdFeaturesQueryKey(serviceVersionId) });
    queryClient.refetchQueries({ queryKey: getServicesServiceVersionIdFeaturesLinkableQueryKey(serviceVersionId) });
    refresh();
  }

  const unlinkMutation = useDeleteServicesServiceVersionIdFeaturesFeatureVersionIdUnlink({
    mutation: {
      onSuccess: () => {
        refetchData();
      },
    },
  });

  const linkMutation = usePostServicesServiceVersionIdFeaturesFeatureVersionIdLink({
    mutation: {
      onSuccess: () => {
        refetchData();
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      featureVersionId: 0,
    },
    validators: {
      onChange: z.object({
        featureVersionId: z.number().min(1),
      }),
    },
    onSubmit: async ({ value }) => {
      await linkMutation.mutateAsync({ service_version_id: serviceVersionId, feature_version_id: value.featureVersionId });
    },
  });

  return (
    <SlimPage>
      <PageTitle>
        Link/Unlink features of {serviceVersion.name} v{serviceVersion.version}
      </PageTitle>
      <div className="flex flex-row gap-4 w-full">
        <div className="flex flex-col gap-4 w-1/2">
          <h2 className="text-lg font-semibold">Linked features</h2>
          <MutationErrors mutations={[unlinkMutation]} />
          <List>
            {features.length ? (
              features.map((feature) => (
                <ListItem key={feature.id}>
                  <div className="flex flex-row gap-2 items-center justify-between">
                    <div className="flex flex-row gap-2 items-center">
                    <span className="text-lg font-semibold">{feature.name}</span>
                    <Badge>v{feature.version}</Badge>
                  </div>
                  {feature.canUnlink && (
                    <Button
                      variant="destructive"
                      onClick={() => unlinkMutation.mutate({ service_version_id: serviceVersionId, feature_version_id: feature.id })}
                    >
                      Unlink
                    </Button>
                  )}
                </div>
                </ListItem>
              ))
            ) : (
              <ListItem>
                No linked features
              </ListItem>
            )}
          </List>
        </div>
        <div className="flex flex-col gap-4 w-1/2">
          <h2 className="text-lg font-semibold">Link feature</h2>
          <MutationErrors mutations={[linkMutation]} />
          <form.AppForm>
            <form
              className="flex flex-col gap-2"
              onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                form.handleSubmit(e);
              }}
            >
              <form.AppField
                name="featureVersionId"
                children={(field) => (
                  <>
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
                        <SelectTrigger id={field.name} className="w-full">
                          <SelectValue placeholder="Select a feature to link" />
                        </SelectTrigger>
                        <SelectContent>
                          {linkableFeatures.length ? (
                            linkableFeatures.map((feature) => (
                              <SelectItem key={feature.id} value={feature.id.toString()}>
                                {feature.name} v{feature.version}
                              </SelectItem>
                            ))
                          ) : (
                            <SelectGroup>
                              <SelectLabel>No linkable features found</SelectLabel>
                            </SelectGroup>
                          )}
                        </SelectContent>
                      </Select>
                    </field.FormControl>
                    <field.FormMessage />
                  </>
                )}
              />
              <form.Subscribe selector={(state) => [state.canSubmit, state.isSubmitting]}>
                {([canSubmit, isSubmitting]) => (
                  <div>
                    <Button type="submit" disabled={!canSubmit || isSubmitting}>
                      Link
                    </Button>
                  </div>
                )}
              </form.Subscribe>
            </form>
          </form.AppForm>
        </div>
      </div>
    </SlimPage>
  );
}
