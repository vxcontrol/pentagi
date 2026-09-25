import { WifiOff } from 'lucide-react';
import { useEffect, useState } from 'react';

export function ConnectionStatus() {
    const [isDisconnected, setIsDisconnected] = useState(false);

    useEffect(() => {
        const markDown = () => setIsDisconnected(true);
        const markUp = () => setIsDisconnected(false);

        window.addEventListener('ws:disconnected', markDown);
        window.addEventListener('ws:connected', markUp);

        return () => {
            window.removeEventListener('ws:disconnected', markDown);
            window.removeEventListener('ws:connected', markUp);
        };
    }, []);

    if (!isDisconnected) {
        return null;
    }

    return (
        <div
            className="bg-background/90 text-muted-foreground fixed bottom-4 left-1/2 z-50 flex -translate-x-1/2 items-center gap-2 rounded-full border px-3 py-1.5 text-xs shadow-lg backdrop-blur"
            data-slot="connection-status"
            role="status"
        >
            <WifiOff className="size-3.5" />
            Not receiving updates — reconnecting
        </div>
    );
}
