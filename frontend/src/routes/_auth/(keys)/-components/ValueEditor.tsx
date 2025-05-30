import { Input } from '~/components/ui/input';
import { Switch } from '~/components/ui/switch';
import { DbValueTypeKind } from '~/gen';
import { JsonEditor } from '~/components/JsonEditor';

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

  function inputOnChange(e: React.ChangeEvent<HTMLInputElement>) {
    onChange(e.target.value);
  }

  if (valueType === 'json') {
    return <JsonEditor value={props.value} onChange={onChange} id={props.id} disabled={props.disabled} onBlur={props.onBlur} />;
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
