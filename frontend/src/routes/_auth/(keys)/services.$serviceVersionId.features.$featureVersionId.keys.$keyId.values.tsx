import { createFileRoute, Link } from '@tanstack/react-router';
import { z } from 'zod';
import { Fragment, useEffect, useMemo, useState } from 'react';
import {
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesCanAdd,
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesValueIdCanEdit,
  getServicesServiceVersionIdQueryOptions,
  getServiceTypesServiceTypeIdVariationPropertiesQueryOptions,
  HandlerValueRequest,
  useDeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesValueId,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdSuspense,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspense,
  useGetServicesServiceVersionIdSuspense,
  useGetServiceTypesServiceTypeIdVariationPropertiesSuspense,
  usePostServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues,
  usePutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesValueId,
  getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions,
  useDeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense,
  ValueVariationValueDto,
  ValueNewValueInfo,
  VariationpropertyFlatVariationPropertyValueDto,
} from '~/gen';
import { ColumnDef, useReactTable, getCoreRowModel, flexRender } from '@tanstack/react-table';
import { HttpError } from '~/axios';
import { MutationErrors } from '~/components/MutationErrors';
import { PageTitle } from '~/components/PageTitle';
import { Button } from '~/components/ui/button';
import { TableHeader, TableRow, TableHead, TableBody, TableCell, TableFooter, Table } from '~/components/ui/table';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { VariationSelect } from '~/components/VariationSelect';
import { cn, variationToQueryParams, variationToRequestParams } from '~/lib/utils';
import { useChangeset } from '~/hooks/use-changeset';
import { versionedTitle } from '~/utils/seo';
import { appTitle } from '~/utils/seo';
import { seo } from '~/utils/seo';
import { DropdownMenuItem } from '~/components/ui/dropdown-menu';
import { CircleAlert } from 'lucide-react';
import { createDefaultValue, createValueValidator } from './-components/value';
import { ValueEditor } from './-components/ValueEditor';
import { ValueViewer } from './-components/ValueViewer';
import { useNavigate } from '@tanstack/react-router';
import { ZodErrorMessage } from '~/components/ZodErrorMessage';
import { Breadcrumbs } from '~/components/Breadcrumbs';
import { useQueryClient } from '@tanstack/react-query';
import { Tooltip, TooltipContent, TooltipTrigger } from '~/components/ui/tooltip';
import { DotDotDot } from '~/components/DotDotDot';

export const Route = createFileRoute('/_auth/(keys)/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/values')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
      featureVersionId: z.coerce.number(),
      keyId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)).then(async (serviceVersion) => {
        await context.queryClient.ensureQueryData(
          getServiceTypesServiceTypeIdVariationPropertiesQueryOptions(serviceVersion.serviceTypeId)
        );

        return serviceVersion;
      }),
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
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryOptions(
          params.serviceVersionId,
          params.featureVersionId,
          params.keyId
        )
      ),
    ]);
  },
  head: ({ loaderData: [serviceVersion, featureVersion, key] }) => {
    return {
      meta: [
        ...seo({
          title: appTitle(['Values', key.name, versionedTitle(featureVersion), versionedTitle(serviceVersion)]),
        }),
      ],
    };
  },
});

export type ValueFormData = {
  data: string;
  variation: Record<number, string>;
};

function RouteComponent() {
  const { serviceVersionId, featureVersionId, keyId } = Route.useParams();
  const { refresh } = useChangeset();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: key } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdSuspense(serviceVersionId, featureVersionId, keyId);
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const { data: featureVersion } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense(serviceVersionId, featureVersionId);
  const { data: values } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspense(
    serviceVersionId,
    featureVersionId,
    keyId
  );
  const { data: properties } = useGetServiceTypesServiceTypeIdVariationPropertiesSuspense(serviceVersion.serviceTypeId, {
    query: {
      staleTime: Infinity,
    },
  });

  function convertVariation(variation: Record<number, string>): string[] | undefined {
    const converted: Record<string, string> = {};
    for (const [propertyId, value] of Object.entries(variation)) {
      const property = properties.find((p) => p.id === parseInt(propertyId));
      if (property) {
        converted[property.name] = value;
      }
    }

    return variationToQueryParams(converted);
  }

  const [data, setData] = useState<ValueVariationValueDto[]>(values);

  useEffect(() => {
    setData(values);
  }, [values]);

  function compareOrder(a: number[], b: number[]): number {
    const len = a.length;
    for (let i = 0; i < len; i++) {
      const av = a[i];
      const bv = b[i];

      if (av !== bv) {
        return av - bv;
      }
    }
    return 0;
  }

  function updateData(newItem: ValueVariationValueDto, removeId?: number) {
    const newData: ValueVariationValueDto[] = [];
    let foundInsert = false;

    for (let i = 0; i < data.length; i++) {
      const item = data[i];
      if (item.id === removeId) continue;

      if (!foundInsert && compareOrder(newItem.order, item.order) < 0) {
        newData.push(newItem);
        foundInsert = true;
      }

      newData.push(item);
    }

    // If newOrder is greater than all, insert at the end
    if (!foundInsert) {
      newData.push(newItem);
    }

    setData(newData);
    queryClient.invalidateQueries(
      getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryOptions(serviceVersionId, featureVersionId, keyId)
    );
  }

  function createNewItem(info: ValueNewValueInfo, data: HandlerValueRequest) {
    return {
      id: info.id,
      data: data.data,
      variation: data.variation,
      order: info.order,
      canEdit: true,
      rank: 0,
    };
  }

  const updateMutation = usePutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesValueId({
    mutation: {
      onSuccess: (info, variables) => {
        const newItem: ValueVariationValueDto = createNewItem(info, variables.data);

        updateData(newItem, editingValueId!);
        refresh();

        setEditingValueId(null);
      },
    },
  });

  const createMutation = usePostServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues({
    mutation: {
      onSuccess: (info, variables) => {
        const newItem: ValueVariationValueDto = createNewItem(info, variables.data);

        updateData(newItem);
        refresh();
        setShowAddForm(false);
        addForm.reset();
      },
    },
  });

  const deleteMutation = useDeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesValueId({
    mutation: {
      onSuccess: (_, variables) => {
        setData((old) => old.filter((item) => item.id !== variables.value_id));
        refresh();
      },
    },
  });

  const deleteKeyMutation = useDeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId({
    mutation: {
      onSuccess: () => {
        refresh();
        navigate({ to: '/services/$serviceVersionId/features/$featureVersionId', params: { serviceVersionId, featureVersionId } });
      },
    },
  });

  const variationPropertyValues = useMemo(() => {
    return properties.reduce((acc, property) => {
      acc[property.id] = property.values;
      return acc;
    }, {} as Record<number, VariationpropertyFlatVariationPropertyValueDto[]>);
  }, [properties]);

  const [editingValueId, setEditingValueId] = useState<number | null>(null);
  const [showAddForm, setShowAddForm] = useState(false);

  function asyncValidatorErrorMessage(err: any) {
    const httpError = err as HttpError;

    if (httpError.status === 403) {
      return 'You do not have permission to save value with this variation';
    } else if (httpError.status === 409) {
      return 'Value with this variation already exists';
    }

    return httpError.message;
  }

  function cachePerVariation(fn: (variation: Record<number, string>) => Promise<string | undefined>) {
    const cache = new Map<string, string | undefined>();

    return async (variation: Record<number, string>): Promise<string | undefined> => {
      const key = JSON.stringify(variation);
      if (cache.has(key)) {
        return cache.get(key);
      }

      const result = await fn(variation);
      cache.set(key, result);
      return result;
    };
  }

  const editValidationFn = cachePerVariation(async (variation) => {
    try {
      await getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesValueIdCanEdit(
        serviceVersionId,
        featureVersionId,
        keyId,
        editingValueId!,
        {
          'variation[]': variationToQueryParams(variation),
        }
      );
    } catch (err: any) {
      return asyncValidatorErrorMessage(err);
    }
  });

  const editForm = useAppForm({
    defaultValues: {
      data: createDefaultValue(key.valueType),
      variation: {},
    } as ValueFormData,
    validators: {
      // triggers twice the first time, https://github.com/TanStack/form/issues/1431
      onChangeAsync: z.object({
        data: createValueValidator(key.validators),
        variation: z.record(z.string(), z.string()).superRefine(async (value, ctx) => {
          const error = await editValidationFn(value);

          if (error) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: error,
            });
          }
        }),
      }),
    },
    onSubmit: async ({ value }) => {
      await updateMutation.mutateAsync({
        service_version_id: serviceVersionId,
        feature_version_id: featureVersionId,
        key_id: keyId,
        value_id: editingValueId!,
        data: {
          data: value.data,
          variation: variationToRequestParams(value.variation),
        },
      });
    },
  });

  const addValidationFn = cachePerVariation(async (variation) => {
    try {
      await getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesCanAdd(serviceVersionId, featureVersionId, keyId, {
        'variation[]': variationToQueryParams(variation),
      });
    } catch (err: any) {
      return asyncValidatorErrorMessage(err);
    }
  });

  const addForm = useAppForm({
    defaultValues: {
      data: createDefaultValue(key.valueType),
      variation: {},
    } as ValueFormData,
    validators: {
      onChangeAsync: z.object({
        data: createValueValidator(key.validators),
        variation: z.record(z.string(), z.string()).superRefine(async (value, ctx) => {
          const error = await addValidationFn(value);

          if (error) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: error,
            });
          }
        }),
      }),
    },
    onSubmit: async ({ value }) => {
      await createMutation.mutateAsync({
        service_version_id: serviceVersionId,
        feature_version_id: featureVersionId,
        key_id: keyId,
        data: {
          data: value.data,
          variation: variationToRequestParams(value.variation),
        },
      });
    },
  });

  const columns = useMemo<ColumnDef<ValueVariationValueDto, any>[]>(
    () => [
      {
        header: 'Value',
        accessorKey: 'data',
        meta: {
          sizeClass: 'max-w-[400px] min-w-[200px]',
          editingValueId,
        },
        cell: (info) => {
          const value = info.getValue();
          const id = info.row.original.id;

          return (
            <>
              {editingValueId === id ? (
                <editForm.AppField
                  name="data"
                  children={(field) => (
                    <>
                      <field.FormControl>
                        <ValueEditor
                          valueType={key.valueType}
                          id={`edit-${field.name}`}
                          name={field.name}
                          value={field.state.value}
                          onChange={(value) => field.handleChange(value)}
                          onBlur={field.handleBlur}
                          disabled={editForm.state.isSubmitting}
                        />
                      </field.FormControl>
                    </>
                  )}
                />
              ) : (
                <ValueViewer valueType={key.valueType} value={value} />
              )}
            </>
          );
        },
        footer: () => {
          return (
            <addForm.AppField
              name="data"
              children={(field) => (
                <>
                  <field.FormControl>
                    <ValueEditor
                      valueType={key.valueType}
                      id={`add-${field.name}`}
                      name={field.name}
                      value={field.state.value}
                      onChange={(value) => field.handleChange(value)}
                      onBlur={field.handleBlur}
                      disabled={addForm.state.isSubmitting}
                    />
                  </field.FormControl>
                </>
              )}
            />
          );
        },
      },
      ...properties.map<ColumnDef<ValueVariationValueDto, any>>((property) => ({
        header: property.displayName,
        id: property.name,
        accessorFn: (row) => row.variation[property.id] ?? 'any',
        meta: {
          sizeClass: 'min-w-[120px]',
        },
        cell: (info) => {
          const value = info.getValue();
          const id = info.row.original.id;
          const propertyValues = variationPropertyValues[property.id];

          return (
            <>
              {editingValueId === id ? (
                <editForm.AppField
                  name={`variation.${property.id}`}
                  validators={{
                    onMount: ({ value }) => {
                      if (value && !propertyValues.some((v) => v.value === value)) {
                        return 'This value was probably archived and cannot be used anymore. Please change it or delete this row.';
                      }

                      return undefined;
                    },
                  }}
                  children={(field) => (
                    <div className="flex flex-row gap-2 items-center">
                      <VariationSelect
                        values={propertyValues}
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onValueChange={(value) => field.handleChange(value)}
                        disabled={editForm.state.isSubmitting}
                      ></VariationSelect>
                      {field.state.meta.errors.length > 0 && (
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <CircleAlert className="size-4 text-destructive" />
                          </TooltipTrigger>
                          <TooltipContent>
                            <field.FormMessage className="text-primary-foreground" immediate />
                          </TooltipContent>
                        </Tooltip>
                      )}
                    </div>
                  )}
                />
              ) : (
                <>{value}</>
              )}
            </>
          );
        },
        footer: () => {
          const propertyValues = variationPropertyValues[property.id];

          return (
            <addForm.AppField
              name={`variation.${property.id}`}
              children={(field) => (
                <VariationSelect
                  values={propertyValues}
                  id={field.name}
                  name={field.name}
                  value={field.state.value}
                  onValueChange={(value) => field.handleChange(value)}
                  disabled={addForm.state.isSubmitting}
                ></VariationSelect>
              )}
            />
          );
        },
      })),
      {
        header: 'Actions',
        id: 'actions',
        meta: {
          hide: data.every((v) => !v.canEdit),
        },
        cell: (info) => {
          const id = info.row.original.id;
          const isDefaultValue = Object.keys(info.row.original.variation).length === 0;

          function setEditing() {
            editForm.setFieldValue('data', info.row.original.data);
            editForm.setFieldValue('variation', info.row.original.variation);
            setEditingValueId(id);
          }

          return (
            <div className="flex gap-2">
              {editingValueId === id ? (
                <editForm.Subscribe
                  selector={(state) => [state.canSubmit, state.isSubmitting]}
                  children={([canSubmit, isSubmitting]) => (
                    <>
                      <Button variant="outline" onClick={() => editForm.handleSubmit()} disabled={!canSubmit || isSubmitting}>
                        Save
                      </Button>
                      <Button variant="outline" onClick={() => setEditingValueId(null)} disabled={isSubmitting}>
                        Cancel
                      </Button>
                    </>
                  )}
                />
              ) : (
                <>
                  {info.row.original.canEdit && (
                    <>
                      <Button variant="outline" onClick={setEditing}>
                        Edit
                      </Button>
                      {!isDefaultValue && (
                        <Button
                          variant="destructive"
                          onClick={() =>
                            deleteMutation.mutate({
                              service_version_id: serviceVersionId,
                              feature_version_id: featureVersionId,
                              key_id: keyId,
                              value_id: id,
                            })
                          }
                        >
                          Delete
                        </Button>
                      )}
                      <DotDotDot>
                        <Link
                          to="/change-history"
                          search={{
                            serviceId: serviceVersion.serviceId,
                            serviceVersionId,
                            featureId: featureVersion.featureId,
                            featureVersionId,
                            keyName: key.name,
                            applyVariation: true,
                            'variation[]': convertVariation(info.row.original.variation),
                          }}
                        >
                          <DropdownMenuItem>History</DropdownMenuItem>
                        </Link>
                      </DotDotDot>
                    </>
                  )}
                </>
              )}
            </div>
          );
        },
        footer: () => {
          return (
            <addForm.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <div className="flex gap-2 items-center">
                  <Button onClick={() => addForm.handleSubmit()} disabled={!canSubmit || isSubmitting}>
                    Add
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setShowAddForm(false);
                      addForm.reset();
                    }}
                    disabled={isSubmitting}
                  >
                    Cancel
                  </Button>
                </div>
              )}
            />
          );
        },
      },
    ],
    [editingValueId, data]
  );

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    //getFilteredRowModel: getFilteredRowModel(),
  });

  return (
    <div className="p-4">
      <Breadcrumbs path={[serviceVersion, featureVersion]} />
      <div className="flex items-center justify-between mb-8">
        <PageTitle className="mb-0">Key {key.name}</PageTitle>
        <div className="flex items-center">
          {key.canEdit && (
            <DotDotDot>
              <Link
                to="/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/edit"
                params={{ serviceVersionId, featureVersionId, keyId }}
              >
                <DropdownMenuItem>Edit</DropdownMenuItem>
              </Link>
              <Link
                to="/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/permissions"
                params={{ serviceVersionId, featureVersionId, keyId }}
              >
                <DropdownMenuItem>Permissions</DropdownMenuItem>
              </Link>
              <Link
                to="/change-history"
                search={{
                  serviceId: serviceVersion.serviceId,
                  serviceVersionId,
                  featureId: featureVersion.featureId,
                  featureVersionId,
                  keyName: key.name,
                }}
              >
                <DropdownMenuItem>History</DropdownMenuItem>
              </Link>
              <DropdownMenuItem
                variant="destructive"
                onClick={() =>
                  deleteKeyMutation.mutate({ service_version_id: serviceVersionId, feature_version_id: featureVersionId, key_id: keyId })
                }
              >
                Delete
              </DropdownMenuItem>
            </DotDotDot>
          )}
        </div>
      </div>
      <div className="flex flex-col gap-4">
        {key.description && <p className="text-muted-foreground">{key.description}</p>}
        <MutationErrors mutations={[createMutation, updateMutation, deleteMutation, deleteKeyMutation]} />
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    className={(header.column.columnDef.meta as any)?.sizeClass}
                    hidden={(header.column.columnDef.meta as any)?.hide}
                  >
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.map((row) => (
              <Fragment key={row.id}>
                <TableRow
                  className={cn({
                    'border-b-0': editingValueId == row.original.id,
                  })}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                  ))}
                </TableRow>
                {editingValueId == row.original.id && (
                  <editForm.Subscribe
                    selector={(state) => [state.errors]}
                    children={([errors]) =>
                      errors.length > 0 ? (
                        <TableRow className="border-t-0">
                          <TableCell className="pt-0" colSpan={row.getVisibleCells().length}>
                            <ZodErrorMessage errors={errors} />
                          </TableCell>
                        </TableRow>
                      ) : null
                    }
                  />
                )}
              </Fragment>
            ))}
          </TableBody>
          <TableFooter className="font-normal">
            {showAddForm &&
              table.getFooterGroups().map((footerGroup) => (
                <Fragment key={footerGroup.id}>
                  <TableRow className="border-b-0">
                    {footerGroup.headers.map((header) => (
                      <TableCell key={header.id}>{flexRender(header.column.columnDef.footer, header.getContext())}</TableCell>
                    ))}
                  </TableRow>
                  {showAddForm && (
                    <addForm.Subscribe
                      selector={(state) => [state.errors]}
                      children={([errors]) =>
                        errors.length > 0 ? (
                          <TableRow className="border-t-0">
                            <TableCell className="pt-0" colSpan={footerGroup.headers.length}>
                              <ZodErrorMessage errors={errors} />
                            </TableCell>
                          </TableRow>
                        ) : null
                      }
                    />
                  )}
                </Fragment>
              ))}
          </TableFooter>
        </Table>
        {!showAddForm && (
          <div>
            <Button onClick={() => setShowAddForm(true)}>Add Value</Button>
          </div>
        )}
      </div>
    </div>
  );
}
