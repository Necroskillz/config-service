import { ServiceValueValidatorParameterType } from '~/gen';
import { Input } from '~/components/ui/input';
import { json } from '@codemirror/lang-json';
import CodeMirror from '@uiw/react-codemirror';
import { useTheme } from '~/ThemeProvider';
import { formatJsonSafe } from '~/utils/json';

export function ValidatorParameterEditor({
  parameterType,
  parameter,
  onChange,
  disabled,
  onBlur,
}: {
  parameterType: ServiceValueValidatorParameterType;
  parameter?: string;
  onChange?: (parameter: string) => void;
  disabled?: boolean;
  onBlur?: () => void;
}) {
  const { activeTheme } = useTheme();

  if (parameterType === 'none') {
    return null;
  }

  if (parameterType === 'json_schema') {
    return (
      <div className="json-editor-container">
        <CodeMirror
          value={formatJsonSafe(parameter)}
          onChange={onChange}
          onBlur={onBlur}
          theme={activeTheme === 'dark' ? 'dark' : 'light'}
          extensions={[json()]}
          readOnly={disabled}
        />
      </div>
    );
  } else {
    return <Input type="text" value={parameter} onChange={(e) => onChange?.(e.target.value)} onBlur={onBlur} disabled={disabled} />;
  }
}
