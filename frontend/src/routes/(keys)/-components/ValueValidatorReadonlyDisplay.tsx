import { Label } from '~/components/ui/label';
import { ServiceValidatorWithParameterTypeDto } from '~/gen';
import { ValidatorParameterEditor } from './ValidatorParameterEditor';
import { Input } from '~/components/ui/input';
import { Button } from '~/components/ui/button';

export function ValueValidatorReadonlyDisplay({ validator }: { validator: ServiceValidatorWithParameterTypeDto }) {
  return (
    <div key={validator.validatorType} className="flex flex-row gap-4 w-full">
      <div className="flex flex-3/12 flex-col gap-2">
        <Label>Validator</Label>
        <span className="text-lg">{validator.validatorType}</span>
      </div>
      <div className="flex flex-1/3 flex-col gap-2">
        <Label>Parameter</Label>
        <ValidatorParameterEditor parameterType={validator.parameterType} parameter={validator.parameter} disabled={true} />
      </div>
      <div className="flex flex-1/3 flex-col gap-2">
        <Label>Error Text</Label>
        <Input type="text" value={validator.errorText} disabled />
      </div>
      <div className="flex flex-1/12 flex-col gap-2">
        <Label>&nbsp;</Label>
        <Button variant="destructive" disabled>
          Remove
        </Button>
      </div>
    </div>
  );
}
