import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import {
  DbValueValidatorType,
  dbValueValidatorType,
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysNameTakenName,
  getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions,
  getServicesServiceVersionIdQueryOptions,
  getValueTypesQueryOptions,
  HandlerValidatorRequest,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense,
  useGetValueTypesSuspense,
  usePostServicesServiceVersionIdFeaturesFeatureVersionIdKeys,
  useGetServicesServiceVersionIdSuspense,
  ValuetypeValueTypeDto,
  ValidationValueValidatorParameterType,
} from '~/gen';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '~/components/ui/select';
import { useEffect, useMemo, useState } from 'react';
import { MutationErrors } from '~/components/MutationErrors';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { Textarea } from '~/components/ui/textarea';
import { versionedTitle } from '~/utils/seo';
import { appTitle } from '~/utils/seo';
import { seo } from '~/utils/seo';
import { useChangeset } from '~/hooks/use-changeset';
import { ValueEditor } from './-components/ValueEditor';
import { createDefaultValue, createValueValidator, ValueValidatorDef } from './-components/value';
import { ValidatorParameterEditor } from './-components/ValidatorParameterEditor';
import { Label } from '~/components/ui/label';
import { ValueValidatorReadonlyDisplay } from './-components/ValueValidatorReadonlyDisplay';
import { createParameterValidator } from './-components/value-validator';
import { Breadcrumbs } from '~/components/Breadcrumbs';
export const Route = createFileRoute('/_auth/(keys)/services/$serviceVersionId/features/$featureVersionId/keys/create')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
      featureVersionId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions(params.serviceVersionId, params.featureVersionId)
      ),
      context.queryClient.ensureQueryData(getValueTypesQueryOptions()),
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
  const { refresh } = useChangeset();
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const { data: featureVersion } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense(serviceVersionId, featureVersionId);
  const { data: valueTypes } = useGetValueTypesSuspense({
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
        refresh();
      },
    },
  });

  const [selectedValueType, setSelectedValueType] = useState<ValuetypeValueTypeDto>(valueTypes[0]);
  const [availableValidators, setAvailableValidators] = useState(selectedValueType.allowedValidators);
  const [selectedValidator, setSelectedValidator] = useState<DbValueValidatorType | null>(null);
  const validatorParameterTypes = useMemo(() => {
    return selectedValueType.allowedValidators.reduce((acc, validator) => {
      acc[validator.validatorType] = validator.parameterType;
      return acc;
    }, {} as Record<DbValueValidatorType, ValidationValueValidatorParameterType>);
  }, [selectedValueType]);

  function addValidator(validatorType: DbValueValidatorType) {
    form.setFieldValue('validators', [
      ...form.state.values.validators,
      {
        validatorType,
        parameter: '',
        errorText: '',
      },
    ]);

    setAvailableValidators((old) => old.filter((validator) => validator.validatorType !== validatorType));
    setSelectedValidator(null);
  }

  function removeValidator(index: number) {
    const newValidators = form.state.values.validators.filter((_, i) => i !== index);
    form.setFieldValue('validators', newValidators);

    setAvailableValidators(
      selectedValueType.allowedValidators.filter((v) => !newValidators.some((v2) => v2.validatorType === v.validatorType))
    );
    setSelectedValidator(null);
  }

  function setValueType(id: number) {
    const valueType = valueTypes.find((valueType) => valueType.id === id);
    if (!valueType) {
      return;
    }
    setSelectedValueType(valueType);
    setAvailableValidators(valueType.allowedValidators);
    setSelectedValidator(null);
    form.setFieldValue('valueTypeId', valueType.id);
    form.setFieldValue('defaultValue', createDefaultValue(valueType.kind));
    form.setFieldValue('validators', []);
  }

  const form = useAppForm({
    defaultValues: {
      name: '',
      description: '',
      valueTypeId: selectedValueType.id,
      defaultValue: createDefaultValue(selectedValueType.kind),
      validators: [] as HandlerValidatorRequest[],
    },
    validators: {
      onChange: z
        .object({
          name: z
            .string()
            .min(1, 'Name is required')
            .max(100, 'Name must have at most 100 characters')
            .regex(/^[a-zA-Z_][\w_]*$/, 'Allowed characters: letters, numbers, _. Must start with a letter or underscore.'),
          description: z.string().max(1000, 'Description must have at most 1000 characters'),
          valueTypeId: z.number().min(1, 'Value type is required'),
          defaultValue: z.string(),
          validators: z.array(
            z.object({
              validatorType: z.nativeEnum(dbValueValidatorType),
              parameter: z.string(),
              errorText: z.string(),
            })
          ),
        })
        .superRefine((value, ctx) => {
          const validValidators = [] as HandlerValidatorRequest[];

          value.validators.forEach((validator, i) => {
            const parameterValidator = createParameterValidator(validatorParameterTypes[validator.validatorType]);
            const result = parameterValidator.safeParse(validator.parameter);

            if (!result.success) {
              result.error.errors.forEach((error) => {
                ctx.addIssue({
                  code: z.ZodIssueCode.custom,
                  message: error.message,
                  path: ['validators', i, 'parameter'],
                });
              });
            } else {
              validValidators.push(validator);
            }
          });

          const validator = createValueValidator((selectedValueType.validators as ValueValidatorDef[]).concat(validValidators));
          const result = validator.safeParse(value.defaultValue);
          if (!result.success) {
            result.error.errors.forEach((error) => {
              ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: error.message,
                path: ['defaultValue'],
              });
            });
          }
        }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ service_version_id: serviceVersionId, feature_version_id: featureVersionId, data: value });
    },
  });

  useEffect(() => {
    form.validateSync('change');
  }, [selectedValueType, availableValidators]);

  return (
    <SlimPage>
      <Breadcrumbs path={[serviceVersion, featureVersion]} />
      <PageTitle>Create Key</PageTitle>
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
            name="valueTypeId"
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Value Type</field.FormLabel>
                <field.FormControl>
                  <Select
                    name={field.name}
                    value={field.state.value.toString()}
                    onValueChange={(value) => {
                      if (!value) {
                        return;
                      }
                      setValueType(parseInt(value));
                    }}
                  >
                    <SelectTrigger id={field.name} className="w-[180px]">
                      <SelectValue placeholder="Select a value type" />
                    </SelectTrigger>
                    <SelectContent>
                      {valueTypes.map((valueType) => (
                        <SelectItem key={valueType.id} value={valueType.id.toString()}>
                          {valueType.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </field.FormControl>
                <field.FormMessage />
              </>
            )}
          />
          <form.AppField
            name="defaultValue"
            children={(field) => (
              <>
                <field.FormLabel htmlFor={field.name}>Default Value</field.FormLabel>
                <field.FormControl>
                  <ValueEditor
                    valueType={selectedValueType.kind}
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onChange={(value) => field.handleChange(value)}
                    onBlur={field.handleBlur}
                  />
                </field.FormControl>
                <field.FormMessage />
              </>
            )}
          />
          <h2 className="text-lg font-semibold">Validators</h2>
          <div className="flex flex-col gap-8 w-full">
            {selectedValueType.validators.map((validator) => (
              <ValueValidatorReadonlyDisplay key={validator.validatorType} validator={validator} />
            ))}
            <form.AppField name="validators" mode="array">
              {(field) => (
                <div className="flex flex-col gap-8 w-full">
                  {field.state.value.map((_, i) => (
                    <div key={i} className="flex flex-row gap-4 w-full">
                      <div className="flex flex-3/12 flex-col gap-2">
                        <Label>Validator</Label>
                        <span className="text-lg">{field.state.value[i].validatorType}</span>
                      </div>
                      <div className="flex flex-1/3 flex-col gap-2">
                        <form.AppField
                          name={`validators[${i}].parameter`}
                          children={(subField) => (
                            <>
                              <subField.FormLabel htmlFor={subField.name}>Parameter</subField.FormLabel>
                              <subField.FormControl>
                                <ValidatorParameterEditor
                                  parameterType={validatorParameterTypes[field.state.value[i].validatorType]}
                                  parameter={subField.state.value}
                                  onChange={(value) => subField.handleChange(value)}
                                  onBlur={subField.handleBlur}
                                />
                              </subField.FormControl>
                              <subField.FormMessage />
                            </>
                          )}
                        />
                      </div>
                      <div className="flex flex-1/3 flex-col gap-2">
                        <form.AppField
                          name={`validators[${i}].errorText`}
                          children={(subField) => (
                            <>
                              <subField.FormLabel htmlFor={subField.name}>Error Text</subField.FormLabel>
                              <subField.FormControl>
                                <Input
                                  type="text"
                                  id={subField.name}
                                  name={subField.name}
                                  value={subField.state.value}
                                  onChange={(e) => subField.handleChange(e.target.value)}
                                  onBlur={subField.handleBlur}
                                />
                              </subField.FormControl>
                              <subField.FormMessage />
                            </>
                          )}
                        />
                      </div>
                      <div className="flex flex-1/12 flex-col gap-2">
                        <Label>&nbsp;</Label>
                        <Button type="button" variant="destructive" onClick={() => removeValidator(i)}>
                          Remove
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </form.AppField>
            {availableValidators.length > 0 && (
              <div className="flex flex-row gap-2">
                <Label>Add Validator</Label>
                <Select value={selectedValidator ?? ''} onValueChange={(value) => setSelectedValidator(value as DbValueValidatorType)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a validator" />
                  </SelectTrigger>
                  <SelectContent>
                    {availableValidators.map((validator) => (
                      <SelectItem key={validator.validatorType} value={validator.validatorType}>
                        {validator.validatorType}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button type="button" variant="outline" onClick={() => addValidator(selectedValidator!)} disabled={!selectedValidator}>
                  Add
                </Button>
              </div>
            )}
          </div>
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
