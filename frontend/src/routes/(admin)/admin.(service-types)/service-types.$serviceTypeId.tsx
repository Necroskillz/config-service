import { SelectGroup, SelectLabel } from '@radix-ui/react-select';
import { useQueryClient } from '@tanstack/react-query';
import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { ChevronDown, ChevronUp, EllipsisIcon } from 'lucide-react';
import { useEffect, useState } from 'react';
import { z } from 'zod';
import { MutationErrors } from '~/components/MutationErrors';
import { PageTitle } from '~/components/PageTitle';
import { Button } from '~/components/ui/button';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '~/components/ui/dropdown-menu';
import { Select, SelectItem, SelectContent, SelectTrigger, SelectValue } from '~/components/ui/select';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import {
  getServiceTypesQueryOptions,
  getServiceTypesServiceTypeIdQueryOptions,
  ServicetypeServiceTypeVariationPropertyLinkDto,
  useDeleteServiceTypesServiceTypeId,
  useDeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId,
  useGetServiceTypesServiceTypeIdSuspense,
  useGetVariationProperties,
  usePostServiceTypesServiceTypeIdVariationProperties,
  usePutServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdPriority,
  VariationpropertyVariationPropertyItemDto,
} from '~/gen';
import { cn } from '~/lib/utils';
import { appTitle } from '~/utils/seo';
import { seo } from '~/utils/seo';

export const Route = createFileRoute('/(admin)/admin/(service-types)/service-types/$serviceTypeId')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceTypeId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return context.queryClient.ensureQueryData(getServiceTypesServiceTypeIdQueryOptions(params.serviceTypeId));
  },
  head: ({ loaderData: serviceType }) => ({
    meta: [...seo({ title: appTitle([serviceType.name, 'Service Types', 'Admin']) })],
  }),
});

function RouteComponent() {
  const { serviceTypeId } = Route.useParams();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const { data: serviceType } = useGetServiceTypesServiceTypeIdSuspense(serviceTypeId);
  const { data: variationProperties, isLoading: isVariationPropertiesLoading } = useGetVariationProperties();

  const [availableProperties, setAvailableProperties] = useState<VariationpropertyVariationPropertyItemDto[]>([]);

  useEffect(() => {
    if (variationProperties) {
      setAvailableProperties(variationProperties.filter((p) => !serviceType.properties.some((link) => link.propertyId === p.id)));
      form.setFieldValue('propertyId', 0);
    }
  }, [variationProperties, serviceType.properties]);

  const linkMutation = usePostServiceTypesServiceTypeIdVariationProperties({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getServiceTypesServiceTypeIdQueryOptions(serviceTypeId));
      },
    },
  });

  const unlinkMutation = useDeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getServiceTypesServiceTypeIdQueryOptions(serviceTypeId));
      },
    },
  });

  const updatePriorityMutation = usePutServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdPriority({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getServiceTypesServiceTypeIdQueryOptions(serviceTypeId));
      },
    },
  });

  const deleteMutation = useDeleteServiceTypesServiceTypeId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getServiceTypesQueryOptions());
        navigate({ to: '/admin/service-types' });
      },
    },
  });
  const form = useAppForm({
    defaultValues: {
      propertyId: 0,
    },
    validators: {
      onChange: z.object({
        propertyId: z.number().min(1, 'Property is required'),
      }),
    },
    onSubmit: async ({ value }) => {
      await linkMutation.mutateAsync({
        service_type_id: serviceTypeId,
        data: {
          variation_property_id: value.propertyId,
        },
      });
    },
  });

  return (
    <div className="w-[720px] flex flex-col gap-4">
      <div className="flex items-center justify-between mb-8">
        <PageTitle className="mb-0">
          Service type <pre className="inline">{serviceType.name}</pre>
        </PageTitle>
        <div className="flex items-center">
          {serviceType.usageCount === 0 && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon">
                  <EllipsisIcon className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem variant="destructive" onClick={() => deleteMutation.mutate({ service_type_id: serviceTypeId })}>
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </div>
      <h2 className="text-lg font-semibold">Linked Properties</h2>
      <MutationErrors mutations={[updatePriorityMutation, unlinkMutation, deleteMutation]} />
      <div className="flex flex-col gap-4">
        {serviceType.properties.length > 0 ? (
          serviceType.properties.map((link, index) => (
            <ServiceTypeVariationPropertyLink
              key={link.id}
              link={link}
              isLast={index === serviceType.properties.length - 1}
              onOrderChange={(priority) =>
                updatePriorityMutation.mutate({
                  service_type_id: serviceTypeId,
                  variation_property_id: link.propertyId,
                  data: { priority: priority },
                })
              }
              onDelete={() => unlinkMutation.mutate({ service_type_id: serviceTypeId, variation_property_id: link.propertyId })}
              disabled={unlinkMutation.isPending}
            />
          ))
        ) : (
          <p className="text-sm text-muted-foreground">No properties linked</p>
        )}
      </div>
      <h2 className="text-lg font-semibold">Link Property</h2>
      <form.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
        >
          <MutationErrors mutations={[linkMutation]} />
          <form.AppField name="propertyId">
            {(field) => (
              <Select
                value={field.state.value === 0 ? '' : field.state.value.toString()}
                onValueChange={(value) => field.handleChange(parseInt(value))}
              >
                <SelectTrigger className="min-w-[180px]">
                  <SelectValue placeholder="Select a property to link" />
                </SelectTrigger>
                <SelectContent>
                  {isVariationPropertiesLoading ? (
                    <SelectGroup>
                      <SelectLabel>Loading...</SelectLabel>
                    </SelectGroup>
                  ) : availableProperties.length > 0 ? (
                    availableProperties?.map((p) => (
                      <SelectItem key={p.id} value={p.id.toString()}>
                        {p.name}
                      </SelectItem>
                    ))
                  ) : (
                    <SelectGroup>
                      <SelectLabel>No properties available</SelectLabel>
                    </SelectGroup>
                  )}
                </SelectContent>
              </Select>
            )}
          </form.AppField>
          <div>
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit || isSubmitting}>
                  Link Property
                </Button>
              )}
            />
          </div>
        </form>
      </form.AppForm>
    </div>
  );
}

function ServiceTypeVariationPropertyLink({
  link,
  isLast,
  onOrderChange,
  onDelete,
  disabled,
}: {
  link: ServicetypeServiceTypeVariationPropertyLinkDto;
  isLast: boolean;
  disabled: boolean;
  onOrderChange: (priority: number) => void;
  onDelete: () => void;
}) {
  const canChangeOrderUp = link.priority > 1;
  const canChangeOrderDown = !isLast;
  const isDeletable = link.usageCount === 0;

  return (
    <div className="flex flex-row items-center gap-2 mb-2">
      <div className="flex flex-col gap-1">
        <ChevronUp
          className={cn('w-4 h-4', canChangeOrderUp ? 'cursor-pointer' : 'opacity-50')}
          onClick={() => canChangeOrderUp && onOrderChange(link.priority - 1)}
        />
        <ChevronDown
          className={cn('w-4 h-4', canChangeOrderDown ? 'cursor-pointer' : 'opacity-50')}
          onClick={() => canChangeOrderDown && onOrderChange(link.priority + 1)}
        />
      </div>
      <div className="flex flex-row items-center gap-2">
        <pre>{link.name}</pre>
        <span>({link.displayName})</span>
      </div>
      {isDeletable && (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <EllipsisIcon className="size-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            {isDeletable && (
              <DropdownMenuItem disabled={disabled} variant="destructive" onClick={() => onDelete()}>
                Unlink
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      )}
    </div>
  );
}
