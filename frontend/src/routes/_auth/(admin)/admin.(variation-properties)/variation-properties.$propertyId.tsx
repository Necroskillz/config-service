import { useQueryClient } from '@tanstack/react-query';
import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { ChevronUp, ChevronDown, EllipsisIcon } from 'lucide-react';
import { z } from 'zod';
import { MutationErrors } from '~/components/MutationErrors';
import { PageTitle } from '~/components/PageTitle';
import { Button } from '~/components/ui/button';
import { DropdownMenuItem } from '~/components/ui/dropdown-menu';
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from '~/components/ui/dropdown-menu';
import { Input } from '~/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '~/components/ui/select';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { getIndent } from '~/components/VariationSelect';
import {
  getVariationPropertiesPropertyIdQueryOptions,
  getVariationPropertiesPropertyIdValueTakenValue,
  getVariationPropertiesQueryOptions,
  ServiceVariationPropertyValueDto,
  useDeleteVariationPropertiesPropertyId,
  useDeleteVariationPropertiesPropertyIdValuesValueId,
  useGetVariationPropertiesPropertyIdSuspense,
  usePostVariationPropertiesPropertyIdValues,
  usePutVariationPropertiesPropertyId,
  usePutVariationPropertiesPropertyIdValuesValueIdArchive,
  usePutVariationPropertiesPropertyIdValuesValueIdOrder,
  usePutVariationPropertiesPropertyIdValuesValueIdUnarchive,
} from '~/gen';
import { cn } from '~/lib/utils';
import { appTitle } from '~/utils/seo';
import { seo } from '~/utils/seo';

export const Route = createFileRoute('/_auth/(admin)/admin/(variation-properties)/variation-properties/$propertyId')({
  component: RouteComponent,
  params: {
    parse: z.object({
      propertyId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return context.queryClient.ensureQueryData(getVariationPropertiesPropertyIdQueryOptions(params.propertyId));
  },
  head: ({ loaderData: property }) => ({
    meta: [...seo({ title: appTitle([property.name, 'Variation Properties', 'Admin']) })],
  }),
});

function RouteComponent() {
  const { propertyId } = Route.useParams();
  const { data: property } = useGetVariationPropertiesPropertyIdSuspense(propertyId);
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  function refetchProperty() {
    queryClient.invalidateQueries(getVariationPropertiesPropertyIdQueryOptions(propertyId));
  }

  function refetchPropertyList() {
    queryClient.invalidateQueries(getVariationPropertiesQueryOptions());
  }

  const updateMutation = usePutVariationPropertiesPropertyId({
    mutation: {
      onSuccess: async () => {
        refetchPropertyList();
      },
    },
  });

  const deleteMutation = useDeleteVariationPropertiesPropertyId({
    mutation: {
      onSuccess: async () => {
        refetchPropertyList();
        navigate({ to: '/admin/variation-properties' });
      },
    },
  });

  const createValueMutation = usePostVariationPropertiesPropertyIdValues({
    mutation: {
      onSuccess: async () => {
        refetchProperty();
      },
    },
  });

  const updateOrderMutation = usePutVariationPropertiesPropertyIdValuesValueIdOrder({
    mutation: {
      onSuccess: async () => {
        refetchProperty();
      },
    },
  });

  const deleteValueMutation = useDeleteVariationPropertiesPropertyIdValuesValueId({
    mutation: {
      onSuccess: async () => {
        refetchProperty();
      },
    },
  });

  const archiveValueMutation = usePutVariationPropertiesPropertyIdValuesValueIdArchive({
    mutation: {
      onSuccess: async () => {
        refetchProperty();
      },
    },
  });

  const unarchiveValueMutation = usePutVariationPropertiesPropertyIdValuesValueIdUnarchive({
    mutation: {
      onSuccess: async () => {
        refetchProperty();
      },
    },
  });

  const updateForm = useAppForm({
    defaultValues: {
      displayName: property.displayName,
    },
    validators: {
      onChange: z.object({
        displayName: z
          .string()
          .min(1, 'Display name is required')
          .max(20, 'Display name must have at most 20 characters')
          .regex(/^[a-zA-Z\- ]+$/, 'Display name must not contain invalid characters'),
      }),
    },
    onSubmit: async ({ value }) => {
      await updateMutation.mutateAsync({ property_id: propertyId, data: value });
    },
  });

  const addValueForm = useAppForm({
    defaultValues: {
      value: '',
      parentId: 0,
    },
    validators: {
      onChange: z.object({
        value: z
          .string()
          .min(1, 'Value is required')
          .max(20, 'Value must have at most 20 characters')
          .regex(/^[\w\-_\.]+$/, 'Value must not contain invalid characters'),
        parentId: z.number(),
      }),
    },
    onSubmit: async ({ value }) => {
      await createValueMutation.mutateAsync({ property_id: propertyId, data: value });
      addValueForm.reset();
    },
  });

  return (
    <div className="w-[720px] pl-4 flex flex-col gap-4">
      <div className="flex items-center justify-between mb-8">
        <PageTitle className="mb-0">
          Variation Property <pre className="inline">{property.name}</pre>
        </PageTitle>
        <div className="flex items-center">
          {property.usageCount === 0 && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon">
                  <EllipsisIcon className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem variant="destructive" onClick={() => deleteMutation.mutate({ property_id: propertyId })}>
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </div>
      <updateForm.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            updateForm.handleSubmit();
          }}
        >
          <MutationErrors mutations={[updateMutation, deleteMutation]} />
          <updateForm.AppField
            name="displayName"
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Display name</field.FormLabel>
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
            <updateForm.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit || isSubmitting}>
                  Update
                </Button>
              )}
            />
          </div>
        </form>
      </updateForm.AppForm>
      <h2 className="text-lg font-medium">Values</h2>
      <MutationErrors mutations={[updateOrderMutation, deleteValueMutation, archiveValueMutation, unarchiveValueMutation]} />
      {property.values.length > 0 ? (
        <VariationPropertyValue
          value={{ id: 0, value: '', children: property.values, usageCount: 0, archived: false }}
          onOrderChange={(id, order) => updateOrderMutation.mutate({ property_id: propertyId, value_id: id, data: { order } })}
          onDelete={(id) => deleteValueMutation.mutate({ property_id: propertyId, value_id: id })}
          onArchive={(id) => archiveValueMutation.mutate({ property_id: propertyId, value_id: id })}
          onUnarchive={(id) => unarchiveValueMutation.mutate({ property_id: propertyId, value_id: id })}
          disabled={
            updateOrderMutation.isPending ||
            deleteValueMutation.isPending ||
            archiveValueMutation.isPending ||
            unarchiveValueMutation.isPending
          }
        />
      ) : (
        <div className="text-sm text-muted-foreground">No values</div>
      )}
      <addValueForm.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            addValueForm.handleSubmit();
          }}
        >
          <MutationErrors mutations={[createValueMutation]} />
          <addValueForm.AppField
            name="value"
            validators={{
              onChangeAsync: async ({ value }) => {
                if (!value) {
                  return;
                }

                const response = await getVariationPropertiesPropertyIdValueTakenValue(propertyId, value);
                const isTaken = response.value;
                if (isTaken) {
                  return 'Variation property value is already taken';
                }
              },
            }}
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Value</field.FormLabel>
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
          <addValueForm.AppField
            name="parentId"
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Parent</field.FormLabel>
                <field.FormControl>
                  <Select value={field.state.value.toString()} onValueChange={(value) => field.handleChange(parseInt(value))}>
                    <SelectTrigger className="min-w-[180px]">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="0">(root)</SelectItem>
                      {property.values.filter((value) => !value.archived).map((value) => (
                        <VariationValueSelectOptions key={value.id} value={value} depth={0} isLast={false} />
                      ))}
                    </SelectContent>
                  </Select>
                </field.FormControl>
                <field.FormMessage />
              </>
            )}
          />
          <div>
            <addValueForm.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit || isSubmitting}>
                  Add Value
                </Button>
              )}
            />
          </div>
        </form>
      </addValueForm.AppForm>
    </div>
  );
}

function VariationPropertyValue({
  value,
  onOrderChange,
  onDelete,
  onArchive,
  onUnarchive,
  disabled,
  order = 0,
  isLast = false,
}: {
  value: ServiceVariationPropertyValueDto;
  onOrderChange: (id: number, order: number) => void;
  onDelete: (id: number) => void;
  onArchive: (id: number) => void;
  onUnarchive: (id: number) => void;
  disabled: boolean;
  order?: number;
  isLast?: boolean;
}) {
  const canChangeOrderUp = !disabled && order > 1;
  const canChangeOrderDown = !disabled && !isLast;
  const isDeletable = value.usageCount === 0 && value.children.length === 0;
  const isArchivable = !value.archived && value.children.every((child) => child.archived);
  const isUnarchivable = value.archived;

  return (
    <div className="mb-2">
      {value.id !== 0 && (
        <div className="flex flex-row items-center gap-2 mb-2">
          <div className="flex flex-col gap-1">
            <ChevronUp
              className={cn('w-4 h-4', canChangeOrderUp ? 'cursor-pointer' : 'opacity-50')}
              onClick={() => canChangeOrderUp && onOrderChange(value.id, order - 1)}
            />
            <ChevronDown
              className={cn('w-4 h-4', canChangeOrderDown ? 'cursor-pointer' : 'opacity-50')}
              onClick={() => canChangeOrderDown && onOrderChange(value.id, order + 1)}
            />
          </div>
          <pre className={cn(isUnarchivable && 'line-through text-muted-foreground')}>{value.value}</pre>
          {(isDeletable || isArchivable || isUnarchivable) && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon">
                  <EllipsisIcon className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                {isUnarchivable && (
                  <DropdownMenuItem disabled={disabled} onClick={() => onUnarchive(value.id)}>
                    Unarchive
                  </DropdownMenuItem>
                )}
                {isDeletable && (
                  <DropdownMenuItem disabled={disabled} variant="destructive" onClick={() => onDelete(value.id)}>
                    Delete
                  </DropdownMenuItem>
                )}
                {!isDeletable && isArchivable && (
                  <DropdownMenuItem disabled={disabled} variant="destructive" onClick={() => onArchive(value.id)}>
                    Archive
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      )}
      {value.children.length > 0 && (
        <div className={cn('border-l border-muted-foreground pl-1', value.id !== 0 && 'ml-8')}>
          {value.children.map((child, index) => (
            <VariationPropertyValue
              key={child.id}
              value={child}
              order={index + 1}
              onOrderChange={onOrderChange}
              onDelete={onDelete}
              onArchive={onArchive}
              onUnarchive={onUnarchive}
              disabled={disabled}
              isLast={index === value.children.length - 1}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function VariationValueSelectOptions({
  value,
  depth,
  isLast,
}: {
  value: ServiceVariationPropertyValueDto;
  depth: number;
  isLast: boolean;
}) {
  return (
    <>
      <SelectItem value={value.id.toString()} prefix={getIndent(depth, isLast)}>
        {value.value}
      </SelectItem>
      {value.children &&
        value.children.filter((child) => !child.archived).map((child, index) => (
          <VariationValueSelectOptions key={child.id} value={child} depth={depth + 1} isLast={index === value.children.length - 1} />
        ))}
    </>
  );
}
