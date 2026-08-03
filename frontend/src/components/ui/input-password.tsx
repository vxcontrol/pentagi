import { Eye, EyeOff } from 'lucide-react';
import * as React from 'react';

import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';

export type InputPasswordProps = Omit<React.ComponentProps<typeof InputGroupInput>, 'type'>;

function InputPassword(props: InputPasswordProps) {
    const [visible, setVisible] = React.useState(false);

    return (
        <InputGroup>
            <InputGroupInput
                {...props}
                type={visible ? 'text' : 'password'}
            />
            <InputGroupAddon align="inline-end">
                <InputGroupButton
                    aria-label={visible ? 'Hide password' : 'Show password'}
                    onClick={() => setVisible((prev) => !prev)}
                    size="icon-sm"
                    tabIndex={-1}
                >
                    {visible ? <EyeOff /> : <Eye />}
                </InputGroupButton>
            </InputGroupAddon>
        </InputGroup>
    );
}

export { InputPassword };
