import { Button } from '~/components/ui/button';
import { useAppForm } from '~/components/ui/tanstack-form';
import { z } from 'zod';
import { Input } from '~/components/ui/input';
import {
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysNameTakenName,
  useGetValueTypes,
  usePostServicesServiceVersionIdFeaturesFeatureVersionIdKeys,
} from '~/gen';
import { Textarea } from '~/components/ui/textarea';
import { SelectGroup, SelectLabel, SelectTrigger, SelectValue, SelectItem } from '~/components/ui/select';
import { SelectContent } from '~/components/ui/select';
import { Select } from '~/components/ui/select';
import { useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { MutationErrors } from '~/components/MutationErrors';

const Schema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string(),
  valueTypeId: z.number().min(1, 'Value type is required'),
  defaultValue: z.string(),
});

export function KeyForm({ serviceVersionId, featureVersionId }: { serviceVersionId: number; featureVersionId: number }) {
  const navigate = useNavigate();
  const { data: valueTypes, isLoading } = useGetValueTypes();
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
      onChange: Schema,
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
                value,
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
  );
}
