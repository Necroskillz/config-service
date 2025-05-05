import { Input } from '~/components/ui/input';
import { Switch } from '~/components/ui/switch';
import { DbValueTypeKind } from '~/gen';
import { useTheme } from '~/ThemeProvider';
import CodeMirror from '@uiw/react-codemirror';
import { json } from '@codemirror/lang-json';
import { formatJsonSafe } from '~/utils/json';

type ValueEditorProps = {
  valueType: DbValueTypeKind;
  id: string;
  name: string;
  value: string;
  onChange: (value: string) => void;
  onBlur: () => void;
  disabled?: boolean;
};

export function ValueEditor({ valueType, ...props }: ValueEditorProps) {
  const { onChange, ...commonProps } = props;
  const { activeTheme } = useTheme();

  function inputOnChange(e: React.ChangeEvent<HTMLInputElement>) {
    onChange(e.target.value);
  }

  if (valueType === 'json') {
    return (
      <div className="json-editor-container">
        <CodeMirror
          value={formatJsonSafe(props.value)}
          onChange={onChange}
          theme={activeTheme === 'dark' ? 'dark' : 'light'}
          extensions={[json()]}
          readOnly={props.disabled}
          id={`${props.id}-editor`}
          onBlur={props.onBlur}
        />
      </div>
    );
  } else if (valueType === 'boolean') {
    return <Switch onCheckedChange={(checked) => onChange(checked ? 'TRUE' : 'FALSE')} checked={props.value === 'TRUE'} {...commonProps} />;
  } else {
    const inputModes = {
      string: 'text',
      integer: 'numeric',
      decimal: 'decimal',
    } as const;

    return <Input type="text" inputMode={inputModes[valueType]} onChange={inputOnChange} {...commonProps} />;
  }
}
