import {
  DbPermissionLevelEnum,
  getMembershipPermissionsQueryOptions,
  MembershipMembershipObjectDto,
  useDeleteMembershipPermissionsPermissionId,
  useGetMembershipPermissions,
  useGetServiceTypesServiceTypeIdVariationProperties,
  usePostMembershipPermissions,
} from '~/gen';
import { List, ListItem } from './List';
import { MutationErrors } from './MutationErrors';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from './ui/select';
import { useEffect, useState } from 'react';
import { z } from 'zod';
import { variationToQueryParams, variationToRequestParams } from '~/lib/utils';
import { Button } from './ui/button';
import { useAppForm } from './ui/tanstack-form-hook';
import { ZodErrorMessage } from './ZodErrorMessage';
import { useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { VariationSelect } from './VariationSelect';
import { RenderQuery } from './RenderQuery';
import { MemberPicker, requiredMember } from './MemberPicker';

type FormValues = {
  member?: MembershipMembershipObjectDto;
  permission: string;
};

export function PermissionEditor({
  serviceVersionId,
  featureVersionId,
  keyId,
  serviceTypeId,
}: {
  serviceVersionId: number;
  featureVersionId?: number;
  keyId?: number;
  serviceTypeId?: number;
}) {
  const queryClient = useQueryClient();
  const [variation, setVariation] = useState<Record<number, string>>({});

  const membershipQueryParams = {
    serviceVersionId,
    featureVersionId,
    keyId,
    'variation[]': variationToQueryParams(variation),
  };

  const { data: properties } = useGetServiceTypesServiceTypeIdVariationProperties(serviceTypeId!, {
    query: {
      staleTime: Infinity,
      enabled: !!serviceTypeId,
    },
  });

  const permissionQueryOptions = getMembershipPermissionsQueryOptions(membershipQueryParams);

  const permissionsQuery = useGetMembershipPermissions(membershipQueryParams);

  useEffect(() => {
    if (form.state.isDirty && permissionsQuery.data != null) {
      form.validate('change');
    }
  }, [permissionsQuery.data]);

  const deletePermissionMutation = useDeleteMembershipPermissionsPermissionId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(permissionQueryOptions);
      },
    },
  });

  const addPermissionMutation = usePostMembershipPermissions({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(permissionQueryOptions);
        form.reset();
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      member: undefined,
      permission: 'editor',
    } as FormValues,
    validators: {
      onChangeAsync: z.object({
        member: requiredMember('User or group is required').pipe(
          z.custom<MembershipMembershipObjectDto>().superRefine((value, ctx) => {
            if (permissionsQuery.data == null) {
              return;
            }

            if (value.type === 'user' && permissionsQuery.data.some((permission) => permission.userId === value.id)) {
              ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: 'Permission for this user already exists',
              });
            }

            if (value.type === 'group' && permissionsQuery.data.some((permission) => permission.groupId === value.id)) {
              ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: 'Permission for this group already exists',
              });
            }
          })
        ),
        permission: z.enum(['editor', 'admin']),
      }),
    },
    onSubmit: async ({ value }) => {
      addPermissionMutation.mutate({
        data: {
          serviceVersionId,
          featureVersionId,
          keyId,
          variation: variationToRequestParams(variation),
          userId: value.member?.type === 'user' ? value.member.id : undefined,
          groupId: value.member?.type === 'group' ? value.member.id : undefined,
          permission: value.permission as DbPermissionLevelEnum,
        },
      });
    },
  });

  return (
    <div className="flex flex-col gap-4">
      <MutationErrors mutations={[deletePermissionMutation]} />
      {properties != null && (
        <>
          <h2 className="text-lg font-semibold">Variation</h2>
          <p className="text-sm text-muted-foreground">
            Optinally define permissions for spcific variation. If all properties are "any" (default), then permission is applied to all
            values of the key.
          </p>
          <div className="flex flex-row gap-4">
            {properties.map((property) => (
              <div key={property.id} className="flex flex-col gap-2">
                <div>{property.name}</div>
                <VariationSelect
                  id={property.id.toString()}
                  values={property.values}
                  value={variation[property.id] || 'any'}
                  onValueChange={(value) => setVariation((prev) => ({ ...prev, [property.id]: value }))}
                />
              </div>
            ))}
          </div>
        </>
      )}
      <RenderQuery query={permissionsQuery} emptyMessage="No permissions defined">
        {(permissions) => (
          <List>
            {permissions.map((permission) => (
              <ListItem key={permission.id} variant="slim">
                <div className="flex flex-row gap-2 items-center justify-between">
                  {permission.userId && (
                    <div>
                      User{' '}
                      <Link className="link" to="/users/$userId" params={{ userId: permission.userId }}>
                        {permission.userName}
                      </Link>{' '}
                      has permission <strong>{permission.permission}</strong>
                    </div>
                  )}
                  {permission.groupId && (
                    <div>
                      Group{' '}
                      <Link className="link" to="/groups/$groupId" params={{ groupId: permission.groupId }}>
                        {permission.groupName}
                      </Link>{' '}
                      has permission <strong>{permission.permission}</strong>
                    </div>
                  )}
                  <Button variant="destructive" size="sm" onClick={() => deletePermissionMutation.mutate({ permission_id: permission.id })}>
                    Remove
                  </Button>
                </div>
              </ListItem>
            ))}
          </List>
        )}
      </RenderQuery>
      <form.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
        >
          <h2 className="text-lg font-semibold">Add Permission</h2>
          <MutationErrors mutations={[addPermissionMutation]} />

          <div className="flex flex-row gap-4">
            <form.AppField name="member">
              {(field) => (
                <MemberPicker
                  value={field.state.value}
                  onValueChange={(value) => field.handleChange(value)}
                  onBlur={() => field.handleBlur()}
                />
              )}
            </form.AppField>

            <form.AppField name="permission">
              {(field) => (
                <>
                  <field.FormControl>
                    <Select name={field.name} value={field.state.value} onValueChange={(value) => field.handleChange(value)}>
                      <SelectTrigger id={field.name}>
                        <SelectValue placeholder="Select a permission" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="editor">editor</SelectItem>
                        {!featureVersionId && <SelectItem value="admin">admin</SelectItem>}
                      </SelectContent>
                    </Select>
                  </field.FormControl>
                </>
              )}
            </form.AppField>
          </div>
          <form.Subscribe selector={(state) => [state.errors]} children={([errors]) => <ZodErrorMessage errors={errors} />} />
          <div>
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit || isSubmitting}>
                  Add
                </Button>
              )}
            />
          </div>
        </form>
      </form.AppForm>
    </div>
  );
}
