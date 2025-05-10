import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import {
  DbValueValidatorType,
  dbValueValidatorType,
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions,
  getServicesServiceVersionIdQueryOptions,
  getValueTypesQueryOptions,
  getValueTypesValueTypeIdQueryOptions,
  HandlerValidatorRequest,
  HandlerVariationProperty,
  ServiceValueValidatorParameterType,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdSuspense,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspense,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense,
  useGetServicesServiceVersionIdSuspense,
  useGetServiceTypesServiceTypeIdVariationPropertiesSuspense,
  useGetValueTypesValueTypeIdSuspense,
  usePutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId,
} from '~/gen';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '~/components/ui/select';
import { useEffect, useState } from 'react';
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
import { ValidatorParameterEditor } from './-components/ValidatorParameterEditor';
import { ValueValidatorReadonlyDisplay } from './-components/ValueValidatorReadonlyDisplay';
import { createParameterValidator } from './-components/value-validator';
import { createValueValidator } from './-components/value';
import { ZodErrorMessage } from '~/components/ZodErrorMessage';
import { Breadcrumbs } from '~/components/Breadcrumbs';

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
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions(params.serviceVersionId, params.featureVersionId)
      ),
      context.queryClient
        .ensureQueryData(
          getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdQueryOptions(
            params.serviceVersionId,
            params.featureVersionId,
            params.keyId
          )
        )
        .then(async (key) => {
          await context.queryClient.ensureQueryData(getValueTypesValueTypeIdQueryOptions(key.valueTypeId));

          return key;
        }),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryOptions(
          params.serviceVersionId,
          params.featureVersionId,
          params.keyId
        )
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
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const { data: featureVersion } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense(serviceVersionId, featureVersionId);
  const { data: key } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdSuspense(serviceVersionId, featureVersionId, keyId);
  const { data: valueType } = useGetValueTypesValueTypeIdSuspense(key.valueTypeId);
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

  const propertyMap = properties.reduce((acc, property) => {
    acc[property.id] = property;
    return acc;
  }, {} as Record<string, HandlerVariationProperty>);

  const [availableValidators, setAvailableValidators] = useState(
    valueType.allowedValidators.filter((v) => !key.validators.some((kv) => kv.validatorType === v.validatorType))
  );
  const [selectedValidator, setSelectedValidator] = useState<DbValueValidatorType | null>(null);
  const validatorParameterTypes = valueType.allowedValidators.reduce((acc, validator) => {
    acc[validator.validatorType] = validator.parameterType;
    return acc;
  }, {} as Record<DbValueValidatorType, ServiceValueValidatorParameterType>);

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

  const keyValidators = key.validators.filter((validator) => !validator.isBuiltIn) as HandlerValidatorRequest[];

  const form = useAppForm({
    defaultValues: {
      description: key.description,
      validators: keyValidators,
    },
    validators: {
      onChange: z.object({
        description: z.string(),
        validators: z.array(
          z
            .object({
              validatorType: z.nativeEnum(dbValueValidatorType),
              parameter: z.string(),
              errorText: z.string(),
            })
            .superRefine((value, ctx) => {
              const parameterValidator = createParameterValidator(validatorParameterTypes[value.validatorType]);
              const result = parameterValidator.safeParse(value.parameter);
              if (!result.success) {
                result.error.errors.forEach((error) => {
                  ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: error.message,
                    path: ['parameter'],
                  });
                });

                return;
              }

              const validator = createValueValidator([value]);

              values.forEach((value) => {
                const result = validator.safeParse(value.data);
                if (!result.success) {
                  const variation = Object.entries(value.variation);
                  const variationString =
                    variation.length > 0 ? variation.map(([k, v]) => `${propertyMap[k].name}: ${v}`).join(', ') : 'Default';

                  result.error.errors.forEach((error) => {
                    ctx.addIssue({
                      code: z.ZodIssueCode.custom,
                      message: `Value "${value.data}" for variation ${variationString} is invalid: ${error.message}`,
                    });
                  });
                }
              });
            })
        ),
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

    setAvailableValidators(valueType.allowedValidators.filter((v) => !newValidators.some((v2) => v2.validatorType === v.validatorType)));
    setSelectedValidator(null);
  }

  useEffect(() => {
    form.validateField('validators', 'change');
  }, [availableValidators]);

  return (
    <SlimPage>
      <Breadcrumbs path={[serviceVersion, featureVersion, key]} />
      <PageTitle>Update Key</PageTitle>
      <div className="text-muted-foreground mb-4">
        <p className="mt-2 text-sm">
          <span className="font-bold">NOTE:</span> To edit validators, all values of the key must be valid for the new validators and their
          parameters. For existing keys, make sure to not have any changes related to this key in the current changeset.
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
          <h2 className="text-lg font-semibold">Validators</h2>
          <div className="flex flex-col gap-8 w-full">
            {valueType.validators.map((validator) => (
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
                  <form.Subscribe
                    selector={(state) => [state.errors]}
                    children={([errors]) => <ZodErrorMessage errors={errors} pathFilter={`validators.\\d+$`} />}
                  />
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
