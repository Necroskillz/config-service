import { HandlerVariationValueSelectOption } from '~/gen/types/handler/VariationValueSelectOption';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { SelectProps } from '@radix-ui/react-select';

export function getIndent(indent: number, isLast: boolean): string {
  if (indent === 0) {
    return '';
  }

  let sb = '';

  for (let i = 0; i < indent; i++) {
    sb += ' ';
  }

  sb += isLast ? '╚' : '╠';
  sb += ' ';

  return sb;
}

export function VariationSelect({
  values,
  id,
  value,
  ...props
}: { values: HandlerVariationValueSelectOption[]; id: string; value: string } & SelectProps) {
  return (
    <Select value={value ?? 'any'} {...props}>
      <SelectTrigger id={id}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {values.map((value, index) => (
          <SelectItem
            key={value.value}
            value={value.value}
            prefix={getIndent(value.depth, index === values.length - 1 || values[index + 1].depth !== value.depth)}
          >
            {value.value}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
