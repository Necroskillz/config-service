import { json } from '@codemirror/lang-json';
import CodeMirror from '@uiw/react-codemirror';
import { useState } from 'react';
import { cn } from '~/lib/utils';
import { useTheme } from '~/ThemeProvider';
import { formatJsonSafe } from '~/utils/json';

export function JsonEditor({
  value,
  onChange,
  id,
  disabled,
  onBlur,
  mode = 'editor',
}: {
  value: string;
  onChange?: (value: string) => void;
  id?: string;
  disabled?: boolean;
  onBlur?: () => void;
  mode?: 'editor' | 'viewer';
}) {
  const { activeTheme } = useTheme();

  const [editorValue, setEditorValue] = useState(() => formatJsonSafe(value));

  return (
    <div className={cn({ 'json-editor-container': mode === 'editor', 'json-viewer-container': mode === 'viewer' })}>
      <CodeMirror
        suppressHydrationWarning
        value={editorValue}
        onChange={(val) => {
          setEditorValue(val);
          onChange?.(val);
        }}
        theme={activeTheme === 'dark' ? 'dark' : 'light'}
        extensions={[json()]}
        readOnly={disabled || mode === 'viewer'}
        id={id}
        onBlur={onBlur}
      />
    </div>
  );
}
