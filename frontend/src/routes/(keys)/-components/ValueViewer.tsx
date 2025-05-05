import { useTheme } from '~/ThemeProvider';
import CodeMirror from '@uiw/react-codemirror';
import { json } from '@codemirror/lang-json';

export function ValueViewer({ valueType, value }: { valueType: string; value: string }) {
  const { activeTheme } = useTheme();

  if (valueType === 'json') {
    return (
      <div className="json-viewer-container">
        <CodeMirror
          suppressHydrationWarning
          value={value ? JSON.stringify(JSON.parse(value), null, 2) : undefined}
          extensions={[json()]}
          theme={activeTheme === 'dark' ? 'dark' : 'light'}
          readOnly
        />
      </div>
    );
  } else {
    return <pre>{value}</pre>;
  }
}
