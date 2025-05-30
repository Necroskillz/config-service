import { JsonEditor } from '~/components/JsonEditor';

export function ValueViewer({ valueType, value }: { valueType: string; value: string }) {
  if (valueType === 'json') {
    return <JsonEditor mode="viewer" value={value} />;
  } else {
    return <pre>{value}</pre>;
  }
}
