import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useParams } from 'react-router-dom';

const SettingsProvider = () => {
    const { providerId } = useParams<{ providerId: string }>();

    return (
        <Card>
            <CardHeader>
                <CardTitle>Provider Details</CardTitle>
                <CardDescription>Configure provider settings</CardDescription>
            </CardHeader>
            <CardContent>
                <p>Provider ID: {providerId}</p>
                <p>This section will be filled later.</p>
            </CardContent>
        </Card>
    );
};

export default SettingsProvider;
