import { createFileRoute } from '@tanstack/react-router';
import { RenderPagedQuery } from '~/components/RenderPagedQuery';
import { ChangesetChangeDescription } from '~/routes/_auth/(changesets)/-components/ChangesetChange';
import {
  ChangesetChangeHistoryItemDto,
  DbChangesetChangeKind,
  dbChangesetChangeKind,
  getChangeHistoryQueryOptions,
  getChangeHistoryServicesQueryOptions,
  useGetChangeHistory,
  useGetChangeHistoryFeatures,
  useGetChangeHistoryFeaturesFeatureIdVersions,
  useGetChangeHistoryKeys,
  useGetChangeHistoryServicesServiceIdVersions,
  useGetChangeHistoryServicesSuspense,
  useGetServiceTypesServiceTypeIdVariationProperties,
} from '~/gen';
import { zodValidator } from '@tanstack/zod-adapter';
import { z } from 'zod';
import { appTitle, seo } from '~/utils/seo';
import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from '@tanstack/react-table';
import { Link, useNavigate } from '@tanstack/react-router';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '~/components/ui/table';
import { Select, SelectValue, SelectTrigger, SelectContent, SelectItem } from '~/components/ui/select';
import { queryParamsToVariation, variationToQueryParams } from '~/lib/utils';
import { RenderQuery } from '~/components/RenderQuery';
import { useEffect, useReducer } from 'react';
import { Checkbox } from '~/components/ui/checkbox';
import { PageTitle } from '~/components/PageTitle';
import { Button } from '~/components/ui/button';
import { DateTimePicker } from '~/components/DateTimePicker';
import { SlimPage } from '~/components/SlimPage';
import { Badge } from '~/components/ui/badge';

const PAGE_SIZE = 20;

const columnHelper = createColumnHelper<ChangesetChangeHistoryItemDto>();

const columns = [
  columnHelper.accessor('appliedAt', {
    header: 'Timestamp',
    cell: (info) => new Date(info.getValue()).toLocaleString(),
  }),
  columnHelper.accessor('changesetId', {
    header: 'Changeset',
    cell: (info) => {
      return (
        <Link className="link" to="/changesets/$changesetId" params={{ changesetId: info.getValue() }}>
          #{info.getValue()}
        </Link>
      );
    },
  }),
  columnHelper.accessor((row) => ({ userName: row.userName, userId: row.userId }), {
    id: 'user',
    header: 'User',
    cell: (info) => {
      const { userName, userId } = info.getValue();
      return (
        <Link className="link" to="/users/$userId" params={{ userId }}>
          {userName}
        </Link>
      );
    },
  }),
  columnHelper.accessor((row) => row, {
    id: 'change',
    header: 'Change',
    cell: (info) => {
      return <ChangesetChangeDescription change={info.getValue()} />;
    },
  }),
];

const kindLabels: Record<DbChangesetChangeKind, string> = {
  service_version: 'Service version',
  feature_version: 'Feature version',
  feature_version_service_version: 'Feature version ↔ Service version',
  key: 'Key',
  variation_value: 'Value',
} as const;

const DAY_MS = 1000 * 60 * 60 * 24;

export const Route = createFileRoute('/_auth/(change-history)/change-history')({
  component: RouteComponent,
  validateSearch: zodValidator(
    z.object({
      page: z.number().min(1).default(1),
      serviceId: z.number().optional(),
      serviceVersionId: z.number().optional(),
      featureId: z.number().optional(),
      featureVersionId: z.number().optional(),
      keyName: z.string().optional(),
      applyVariation: z.boolean().optional(),
      from: z.coerce.date().optional(),
      to: z.coerce.date().optional(),
      'variation[]': z.array(z.string()).optional(),
      'kinds[]': z.array(z.nativeEnum(dbChangesetChangeKind)).optional(),
    })
  ),
  loaderDeps: ({ search }) => ({ ...search }),
  loader: async ({ context, deps }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getChangeHistoryServicesQueryOptions()),
      context.queryClient.ensureQueryData(
        getChangeHistoryQueryOptions({
          ...deps,
          pageSize: PAGE_SIZE,
        })
      ),
    ]);
  },
  head: () => ({
    meta: [...seo({ title: appTitle(['Change history']) })],
  }),
});

function RouteComponent() {
  const search = Route.useSearch();
  const navigate = useNavigate();

  function parseKinds(kinds: DbChangesetChangeKind[]) {
    const newKinds: Record<DbChangesetChangeKind, boolean> = {
      service_version: false,
      feature_version: false,
      feature_version_service_version: false,
      key: false,
      variation_value: false,
    };
    for (const kind of kinds) {
      newKinds[kind] = true;
    }
    return newKinds;
  }

  const { data: services } = useGetChangeHistoryServicesSuspense();

  function getServiceType(serviceId: number | undefined) {
    return serviceId ? services.find((v) => v.id === serviceId)?.serviceTypeId : undefined;
  }

  function searchToState() {
    return {
      selectedServiceId: search.serviceId,
      selectedServiceVersionId: search.serviceVersionId,
      selectedFeatureId: search.featureId,
      selectedFeatureVersionId: search.featureVersionId,
      selectedKeyName: search.keyName,
      selectedApplyVariation: search.applyVariation,
      selectedVariation: queryParamsToVariation(search['variation[]'] ?? []),
      selectedKinds: parseKinds(search['kinds[]'] ?? []),
      selectedFrom: search.from,
      selectedTo: search.to,
      selectedServiceType: getServiceType(search.serviceId),
    };
  }

  const [state, setState] = useReducer(
    (prev: ReturnType<typeof searchToState>, next: Partial<ReturnType<typeof searchToState>>) => ({ ...prev, ...next }),
    searchToState()
  );

  useEffect(() => {
    setState(searchToState());
  }, [search]);

  useEffect(() => {
    if (state.selectedServiceId) {
      setState({ selectedServiceType: getServiceType(state.selectedServiceId) });
    }
  }, [state.selectedServiceId]);

  useEffect(() => {
    if (state.selectedFrom && state.selectedTo) {
      const from = state.selectedFrom.getTime();
      const to = state.selectedTo.getTime();

      if (from > to) {
        setState({ selectedTo: new Date(from + DAY_MS) });
      }
    }
  }, [state.selectedFrom]);

  useEffect(() => {
    if (state.selectedFrom && state.selectedTo) {
      const from = state.selectedFrom.getTime();
      const to = state.selectedTo.getTime();

      if (to < from) {
        setState({ selectedFrom: new Date(to - DAY_MS) });
      }
    }
  }, [state.selectedTo]);

  function filter() {
    const kindsArray: DbChangesetChangeKind[] = [];
    for (const [kind, selected] of Object.entries(state.selectedKinds)) {
      if (selected) {
        kindsArray.push(kind as DbChangesetChangeKind);
      }
    }

    navigate({
      to: '/change-history',
      search: {
        serviceId: state.selectedServiceId,
        serviceVersionId: state.selectedServiceVersionId,
        featureId: state.selectedFeatureId,
        featureVersionId: state.selectedFeatureVersionId,
        keyName: state.selectedKeyName,
        applyVariation: state.selectedApplyVariation,
        from: state.selectedFrom,
        to: state.selectedTo,
        'variation[]': state.selectedApplyVariation ? variationToQueryParams(state.selectedVariation) : undefined,
        'kinds[]': kindsArray.length > 0 ? kindsArray : undefined,
      },
    });
  }

  const serviceVersionsQuery = useGetChangeHistoryServicesServiceIdVersions(state.selectedServiceId!, {
    query: {
      enabled: !!state.selectedServiceId,
    },
  });

  const featuresQuery = useGetChangeHistoryFeatures(
    {
      serviceId: state.selectedServiceVersionId ? undefined : state.selectedServiceId,
      serviceVersionId: state.selectedServiceVersionId,
    },
    {
      query: {
        enabled: !!state.selectedServiceId || !!state.selectedServiceVersionId,
      },
    }
  );

  const featureVersionsQuery = useGetChangeHistoryFeaturesFeatureIdVersions(state.selectedFeatureId!, {
    query: {
      enabled: !!state.selectedFeatureId,
    },
  });

  const keysQuery = useGetChangeHistoryKeys(
    {
      featureId: state.selectedFeatureId,
      featureVersionId: state.selectedFeatureVersionId,
    },
    {
      query: {
        enabled: !!state.selectedFeatureId || !!state.selectedFeatureVersionId,
      },
    }
  );

  const variationPropertiesQuery = useGetServiceTypesServiceTypeIdVariationProperties(state.selectedServiceType!, {
    query: {
      enabled: !!state.selectedServiceType,
    },
  });

  useEffect(() => {
    if (serviceVersionsQuery.data && !serviceVersionsQuery.data.some((v) => v.id === state.selectedServiceVersionId)) {
      setState({ selectedServiceVersionId: undefined });
    }
  }, [serviceVersionsQuery.data]);

  useEffect(() => {
    if (featuresQuery.data && !featuresQuery.data.some((v) => v.id === state.selectedFeatureId)) {
      setState({ selectedFeatureId: undefined, selectedFeatureVersionId: undefined, selectedKeyName: undefined });
    }
  }, [featuresQuery.data]);

  useEffect(() => {
    if (featureVersionsQuery.data && !featureVersionsQuery.data.some((v) => v.id === state.selectedFeatureVersionId)) {
      setState({ selectedFeatureVersionId: undefined });
    }
  }, [featureVersionsQuery.data]);

  useEffect(() => {
    if (keysQuery.data && !keysQuery.data.some((v) => v.name === state.selectedKeyName)) {
      setState({ selectedKeyName: undefined });
    }
  }, [keysQuery.data]);

  useEffect(() => {
    if (variationPropertiesQuery.data) {
      const newVariation: Record<string, string> = {};
      for (const variationProperty of variationPropertiesQuery.data) {
        if (state.selectedVariation[variationProperty.name]) {
          newVariation[variationProperty.name] = state.selectedVariation[variationProperty.name];
        }
      }

      setState({ selectedVariation: newVariation });
    }
  }, [variationPropertiesQuery.data]);

  const query = useGetChangeHistory({
    ...search,
    pageSize: PAGE_SIZE,
  });

  const table = useReactTable({
    data: query.data?.items ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <SlimPage size="lg">
      <PageTitle>Change history</PageTitle>
      <div className="flex flex-col gap-4">
        <div className="flex flex-row gap-24">
          <div className="flex flex-col gap-4">
            <div className="flex flex-row gap-2">
              <div className="flex flex-col gap-1">
                <WithClearButton
                  onClear={() =>
                    setState({
                      selectedServiceId: undefined,
                      selectedServiceVersionId: undefined,
                      selectedFeatureId: undefined,
                      selectedFeatureVersionId: undefined,
                      selectedKeyName: undefined,
                      selectedServiceType: undefined,
                    })
                  }
                >
                  <label>Service</label>
                </WithClearButton>
                <div className="flex flex-col gap-1">
                  <Select
                    value={state.selectedServiceId?.toString() ?? ''}
                    onValueChange={(v) => setState({ selectedServiceId: v === '' ? undefined : parseInt(v) })}
                  >
                    <SelectTrigger className="w-[400px]">
                      <SelectValue placeholder="Select a service" />
                    </SelectTrigger>
                    <SelectContent>
                      {services.map((service) => (
                        <SelectItem key={service.id} value={service.id.toString()}>
                          {service.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <label>Version</label>
                <Select
                  value={state.selectedServiceVersionId?.toString() ?? 'all'}
                  onValueChange={(v) => setState({ selectedServiceVersionId: v === 'all' || !v ? undefined : parseInt(v) })}
                  disabled={!state.selectedServiceId}
                >
                  <SelectTrigger className="w-[120px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <RenderQuery query={serviceVersionsQuery}>
                      {(serviceVersions) => (
                        <>
                          <SelectItem value="all">all versions</SelectItem>
                          {serviceVersions.map((serviceVersion) => (
                            <SelectItem key={serviceVersion.id} value={serviceVersion.id.toString()}>
                              v{serviceVersion.version}
                            </SelectItem>
                          ))}
                        </>
                      )}
                    </RenderQuery>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="flex flex-row gap-2">
              <div className="flex flex-col gap-1">
                <WithClearButton
                  onClear={() =>
                    setState({ selectedFeatureId: undefined, selectedFeatureVersionId: undefined, selectedKeyName: undefined })
                  }
                >
                  <label>Feature</label>
                </WithClearButton>
                <div className="flex flex-col gap-1">
                  <Select
                    value={state.selectedFeatureId?.toString() ?? ''}
                    onValueChange={(v) => setState({ selectedFeatureId: parseInt(v) })}
                    disabled={!state.selectedServiceId && !state.selectedServiceVersionId}
                  >
                    <SelectTrigger className="w-[400px]">
                      <SelectValue placeholder="Select a feature" />
                    </SelectTrigger>
                    <SelectContent>
                      <RenderQuery query={featuresQuery} emptyMessage="Service has no features">
                        {(features) =>
                          features.map((feature) => (
                            <SelectItem key={feature.id} value={feature.id.toString()}>
                              {feature.name}
                            </SelectItem>
                          ))
                        }
                      </RenderQuery>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <div className="flex flex-col gap-1">
                <label>Version</label>
                <div className="flex flex-col gap-1">
                  <Select
                    value={state.selectedFeatureVersionId?.toString() ?? 'all'}
                    onValueChange={(v) => setState({ selectedFeatureVersionId: v === 'all' || !v ? undefined : parseInt(v) })}
                    disabled={!state.selectedFeatureId}
                  >
                    <SelectTrigger className="w-[120px]">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <RenderQuery query={featureVersionsQuery}>
                        {(featureVersions) => (
                          <>
                            <SelectItem value="all">all versions</SelectItem>
                            {featureVersions.map((featureVersion) => (
                              <SelectItem key={featureVersion.id} value={featureVersion.id.toString()}>
                                v{featureVersion.version}
                              </SelectItem>
                            ))}
                          </>
                        )}
                      </RenderQuery>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>
            <div className="flex flex-row gap-2">
              <div className="flex flex-col gap-1">
                <WithClearButton onClear={() => setState({ selectedKeyName: undefined })}>
                  <label>Key</label>
                </WithClearButton>
                <div className="flex flex-col gap-1">
                  <Select
                    value={state.selectedKeyName ?? ''}
                    onValueChange={(v) => setState({ selectedKeyName: v })}
                    disabled={!state.selectedFeatureId && !state.selectedFeatureVersionId}
                  >
                    <SelectTrigger className="w-[528px]">
                      <SelectValue placeholder="Select a key" />
                    </SelectTrigger>
                    <SelectContent>
                      <RenderQuery query={keysQuery} emptyMessage="Feature has no keys">
                        {(keys) =>
                          keys.map((key) => (
                            <SelectItem key={key.name} value={key.name}>
                              {key.name}
                            </SelectItem>
                          ))
                        }
                      </RenderQuery>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>
          </div>
          <div className="flex flex-col gap-2">
            <h2 className="text-md font-semibold">Change types</h2>
            {(
              ['service_version', 'feature_version', 'feature_version_service_version', 'key', 'variation_value'] as DbChangesetChangeKind[]
            ).map((kind) => (
              <div key={kind}>
                <div className="flex flex-row gap-2 items-center">
                  <Checkbox
                    id={kind}
                    checked={state.selectedKinds[kind]}
                    onCheckedChange={(v) => setState({ selectedKinds: { ...state.selectedKinds, [kind]: v === true } })}
                  />
                  <label className="whitespace-nowrap" htmlFor={kind}>
                    {kindLabels[kind]}
                  </label>
                </div>
              </div>
            ))}
          </div>
          <div className="flex flex-col gap-2">
            <WithClearButton onClear={() => setState({ selectedFrom: undefined })}>
              <label>From</label>
            </WithClearButton>
            <DateTimePicker id="from" value={state.selectedFrom} onChange={(v) => setState({ selectedFrom: v })} className="w-[350px]" />
            <WithClearButton onClear={() => setState({ selectedTo: undefined })}>
              <label>To</label>
            </WithClearButton>
            <DateTimePicker id="to" value={state.selectedTo} onChange={(v) => setState({ selectedTo: v })} className="w-[350px]" />
          </div>
        </div>
        <div className="flex flex-row gap-2">
          <RenderQuery
            query={variationPropertiesQuery}
            emptyMessage="Service type has no variation properties"
            disabledMessage="Select service to enable filtering by variation"
          >
            {(variationProperties) => (
              <div className="flex flex-row gap-4 items-end">
                <div className="flex flex-row gap-2 items-center pb-2">
                  <Checkbox
                    id="applyVariation"
                    checked={state.selectedApplyVariation}
                    onCheckedChange={(v) => setState({ selectedApplyVariation: v === true })}
                  />
                  <label htmlFor="applyVariation">Filter variation</label>
                </div>
                {variationProperties.map((variationProperty) => (
                  <div key={variationProperty.id}>
                    <div className="flex flex-col gap-1">
                      <label>{variationProperty.name}</label>
                      <Select
                        value={state.selectedVariation[variationProperty.name] ?? 'any'}
                        onValueChange={(v) => setState({ selectedVariation: { ...state.selectedVariation, [variationProperty.name]: v } })}
                      >
                        <SelectTrigger className="w-[150px]">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {variationProperty.values.map((value) => (
                            <SelectItem key={value.id} value={value.value}>
                              {value.value}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </RenderQuery>
        </div>
        <div>
          <Button onClick={filter}>Filter</Button>
        </div>
      </div>
      <div className="mt-4">
        <RenderPagedQuery
          query={query}
          page={search.page}
          pageSize={PAGE_SIZE}
          linkTo="/change-history"
          linkSearch={search}
          emptyMessage="No change history"
          pageKey="page"
        >
          {() => (
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
                  <TableRow key={row.id}>
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </RenderPagedQuery>
      </div>
    </SlimPage>
  );
}

function WithClearButton({ children, onClear }: { children: React.ReactNode; onClear: () => void }) {
  return (
    <div className="flex flex-row gap-2 items-center justify-between">
      {children}
      <Badge variant="outline" className="cursor-pointer" onClick={onClear}>
        Clear
      </Badge>
    </div>
  );
}
