import { useQueryClient } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { MutationErrors } from '~/components/MutationErrors';
import { PageTitle } from '~/components/PageTitle';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '~/components/ui/select';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { getIndent } from '~/components/VariationSelect';
import {
  getVariationPropertiesPropertyIdQueryOptions,
  getVariationPropertiesPropertyIdValueTakenValue,
  getVariationPropertiesQueryOptions,
  ServiceVariationPropertyValueDto,
  useGetVariationPropertiesPropertyIdSuspense,
  usePostVariationPropertiesPropertyIdValues,
  usePutVariationPropertiesPropertyId,
} from '~/gen';
import { cn } from '~/lib/utils';
import { appTitle } from '~/utils/seo';
import { seo } from '~/utils/seo';

export const Route = createFileRoute('/(admin)/admin/(variation-properties)/variation-properties/$propertyId')({
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

  const updateMutation = usePutVariationPropertiesPropertyId({
    mutation: {
      onSuccess: async () => {
        queryClient.refetchQueries(getVariationPropertiesQueryOptions());
      },
    },
  });

  const createValueMutation = usePostVariationPropertiesPropertyIdValues({
    mutation: {
      onSuccess: async () => {
        queryClient.refetchQueries(getVariationPropertiesPropertyIdQueryOptions(propertyId));
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
          .regex(/^[\w\-_]+$/, 'Value must not contain invalid characters'),
        parentId: z.number(),
      }),
    },
    onSubmit: async ({ value }) => {
      await createValueMutation.mutateAsync({ property_id: propertyId, data: value });
      addValueForm.reset();
    },
  });

  return (
    <div className="w-[720px] flex flex-col gap-4">
      <PageTitle>
        Variation Property <pre className="inline">{property.name}</pre>
      </PageTitle>

      <updateForm.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            updateForm.handleSubmit();
          }}
        >
          <MutationErrors mutations={[updateMutation]} />
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
      {property.values.length > 0 ? (
        <VariationPropertyValue value={{ id: 0, value: '', children: property.values }} />
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
                      {property.values.map((value) => (
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

function VariationPropertyValue({ value }: { value: ServiceVariationPropertyValueDto }) {
  return (
    <div className="mb-2">
      <div className="flex flex-row items-center gap-2 mb-2">
        <div>{value.value}</div>
      </div>
      {value.children && (
        <div className={cn('border-l border-muted-foreground pl-1', value.id !== 0 && 'ml-8')}>
          {value.children.map((child) => (
            <VariationPropertyValue key={child.id} value={child} />
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
        value.children.map((child, index) => (
          <VariationValueSelectOptions key={child.id} value={child} depth={depth + 1} isLast={index === value.children.length - 1} />
        ))}
    </>
  );
}
