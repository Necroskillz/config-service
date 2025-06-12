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
import { useEffect, useState } from 'react';
import { Checkbox } from '~/components/ui/checkbox';
import { PageTitle } from '~/components/PageTitle';
import { Button } from '~/components/ui/button';

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

  const [selectedServiceId, setSelectedServiceId] = useState(search.serviceId);
  const [selectedServiceVersionId, setSelectedServiceVersionId] = useState(search.serviceVersionId);
  const [selectedFeatureId, setSelectedFeatureId] = useState(search.featureId);
  const [selectedFeatureVersionId, setSelectedFeatureVersionId] = useState(search.featureVersionId);
  const [selectedKeyName, setSelectedKeyName] = useState(search.keyName);
  const [selectedApplyVariation, setSelectedApplyVariation] = useState(search.applyVariation);
  const [selectedVariation, setSelectedVariation] = useState(queryParamsToVariation(search['variation[]'] ?? []));
  const [selectedKinds, setSelectedKinds] = useState(parseKinds(search['kinds[]'] ?? []));
  const [selectedServiceType, setSelectedServiceType] = useState<number | undefined>(undefined);

  useEffect(() => {
    setSelectedServiceId(search.serviceId);
    setSelectedServiceVersionId(search.serviceVersionId);
    setSelectedFeatureId(search.featureId);
    setSelectedFeatureVersionId(search.featureVersionId);
    setSelectedKeyName(search.keyName);
    setSelectedApplyVariation(search.applyVariation);
    setSelectedVariation(queryParamsToVariation(search['variation[]'] ?? []));
    setSelectedKinds(parseKinds(search['kinds[]'] ?? []));
  }, [search]);

  useEffect(() => {
    if (selectedServiceId) {
      const serviceType = services.find((v) => v.id === selectedServiceId)?.serviceTypeId;
      if (serviceType) {
        setSelectedServiceType(serviceType);
      }
    }
  }, [selectedServiceId]);

  function filter() {
    const kindsArray: DbChangesetChangeKind[] = [];
    for (const [kind, selected] of Object.entries(selectedKinds)) {
      if (selected) {
        kindsArray.push(kind as DbChangesetChangeKind);
      }
    }

    navigate({
      to: '/change-history',
      search: {
        serviceId: selectedServiceId,
        serviceVersionId: selectedServiceVersionId,
        featureId: selectedFeatureId,
        featureVersionId: selectedFeatureVersionId,
        keyName: selectedKeyName,
        applyVariation: selectedApplyVariation,
        'variation[]': selectedApplyVariation ? variationToQueryParams(selectedVariation) : undefined,
        'kinds[]': kindsArray.length > 0 ? kindsArray : undefined,
      },
    });
  }

  const { data: services } = useGetChangeHistoryServicesSuspense();
  const serviceVersionsQuery = useGetChangeHistoryServicesServiceIdVersions(selectedServiceId!, {
    query: {
      enabled: !!selectedServiceId,
    },
  });

  const featuresQuery = useGetChangeHistoryFeatures(
    {
      serviceId: selectedServiceVersionId ? undefined : selectedServiceId,
      serviceVersionId: selectedServiceVersionId,
    },
    {
      query: {
        enabled: !!selectedServiceId || !!selectedServiceVersionId,
      },
    }
  );

  const featureVersionsQuery = useGetChangeHistoryFeaturesFeatureIdVersions(selectedFeatureId!, {
    query: {
      enabled: !!selectedFeatureId,
    },
  });

  const keysQuery = useGetChangeHistoryKeys(
    {
      featureId: selectedFeatureId,
      featureVersionId: selectedFeatureVersionId,
    },
    {
      query: {
        enabled: !!selectedFeatureId || !!selectedFeatureVersionId,
      },
    }
  );

  const variationPropertiesQuery = useGetServiceTypesServiceTypeIdVariationProperties(selectedServiceType!, {
    query: {
      enabled: !!selectedServiceType,
    },
  });

  useEffect(() => {
    if (serviceVersionsQuery.data && !serviceVersionsQuery.data.some((v) => v.id === selectedServiceVersionId)) {
      setSelectedServiceVersionId(undefined);
    }
  }, [serviceVersionsQuery.data]);

  useEffect(() => {
    if (featuresQuery.data && !featuresQuery.data.some((v) => v.id === selectedFeatureId)) {
      setSelectedFeatureId(undefined);
      setSelectedFeatureVersionId(undefined);
      setSelectedKeyName(undefined);
    }
  }, [featuresQuery.data]);

  useEffect(() => {
    if (featureVersionsQuery.data && !featureVersionsQuery.data.some((v) => v.id === selectedFeatureVersionId)) {
      setSelectedFeatureVersionId(undefined);
    }
  }, [featureVersionsQuery.data]);

  useEffect(() => {
    if (keysQuery.data && !keysQuery.data.some((v) => v.name === selectedKeyName)) {
      setSelectedKeyName(undefined);
    }
  }, [keysQuery.data]);

  useEffect(() => {
    if (variationPropertiesQuery.data) {
      setSelectedVariation((prev) => {
        const newVariation: Record<string, string> = {};
        for (const variationProperty of variationPropertiesQuery.data) {
          if (prev[variationProperty.name]) {
            newVariation[variationProperty.name] = prev[variationProperty.name];
          }
        }
        return newVariation;
      });
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
    <div className="p-4">
      <PageTitle>Change history</PageTitle>
      <div className="flex flex-col gap-4">
        <div className="flex flex-row gap-24">
          <div className="flex flex-col gap-4">
            <div className="flex flex-row gap-2">
              <div className="flex flex-col gap-1">
                <label>Service</label>
                <div className="flex flex-col gap-1">
                  <Select
                    value={selectedServiceId?.toString() ?? ''}
                    onValueChange={(v) => setSelectedServiceId(v === '' ? undefined : parseInt(v))}
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
                  value={selectedServiceVersionId?.toString() ?? 'all'}
                  onValueChange={(v) => setSelectedServiceVersionId(v === 'all' || !v ? undefined : parseInt(v))}
                  disabled={!selectedServiceId}
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
                <label>Feature</label>
                <div className="flex flex-col gap-1">
                  <Select
                    value={selectedFeatureId?.toString() ?? ''}
                    onValueChange={(v) => setSelectedFeatureId(parseInt(v))}
                    disabled={!selectedServiceId && !selectedServiceVersionId}
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
                    value={selectedFeatureVersionId?.toString() ?? 'all'}
                    onValueChange={(v) => setSelectedFeatureVersionId(v === 'all' || !v ? undefined : parseInt(v))}
                    disabled={!selectedFeatureId}
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
                <label>Key</label>
                <div className="flex flex-col gap-1">
                  <Select
                    value={selectedKeyName ?? ''}
                    onValueChange={(v) => setSelectedKeyName(v)}
                    disabled={!selectedFeatureId && !selectedFeatureVersionId}
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
                    checked={selectedKinds[kind]}
                    onCheckedChange={(v) => setSelectedKinds((prev) => ({ ...prev, [kind]: v === true }))}
                  />
                  <label htmlFor={kind}>{kindLabels[kind]}</label>
                </div>
              </div>
            ))}
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
                    checked={selectedApplyVariation}
                    onCheckedChange={(v) => setSelectedApplyVariation(v === true)}
                  />
                  <label htmlFor="applyVariation">Filter variation</label>
                </div>
                {variationProperties.map((variationProperty) => (
                  <div key={variationProperty.id}>
                    <div className="flex flex-col gap-1">
                      <label>{variationProperty.name}</label>
                      <Select
                        value={selectedVariation[variationProperty.name] ?? 'any'}
                        onValueChange={(v) => setSelectedVariation((prev) => ({ ...prev, [variationProperty.name]: v }))}
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
    </div>
  );
}
