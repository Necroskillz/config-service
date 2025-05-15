import { ServiceValueValidatorParameterType } from '~/gen';
import { Input } from '~/components/ui/input';
import { JsonEditor } from '~/components/JsonEditor';

export function ValidatorParameterEditor({
  parameterType,
  parameter,
  onChange,
  disabled,
  onBlur,
}: {
  parameterType: ServiceValueValidatorParameterType;
  parameter: string;
  onChange?: (parameter: string) => void;
  disabled?: boolean;
  onBlur?: () => void;
}) {
  if (parameterType === 'none') {
    return null;
  }

  if (parameterType === 'json_schema') {
    return (
      <JsonEditor
        value={parameter}
        onChange={(val) => {
          onChange?.(val);
        }}
        onBlur={onBlur}
        disabled={disabled}
      />
    );
  } else {
    return <Input type="text" value={parameter} onChange={(e) => onChange?.(e.target.value)} onBlur={onBlur} disabled={disabled} />;
  }
}
