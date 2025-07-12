import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useSettingsPromptsQuery } from '@/graphql/types';

const SettingsPrompts = () => {
    const { data, loading, error } = useSettingsPromptsQuery();

    if (loading) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle>Prompts</CardTitle>
                    <CardDescription>Manage prompt templates</CardDescription>
                </CardHeader>
                <CardContent>
                    <p>Loading prompts...</p>
                </CardContent>
            </Card>
        );
    }

    if (error) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle>Prompts</CardTitle>
                    <CardDescription>Manage prompt templates</CardDescription>
                </CardHeader>
                <CardContent>
                    <p className="text-red-500">Error loading prompts: {error.message}</p>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>Prompts</CardTitle>
                <CardDescription>Manage prompt templates</CardDescription>
            </CardHeader>
            <CardContent>
                <div className="space-y-4">
                    {/* User Defined Prompts */}
                    <div>
                        <h3 className="text-lg font-semibold mb-2">User Defined Prompts</h3>
                        {data?.settingsPrompts.userDefined &&
                        Array.isArray(data.settingsPrompts.userDefined) &&
                        data.settingsPrompts.userDefined.length > 0 ? (
                            <div className="space-y-2">
                                {data.settingsPrompts.userDefined.map((prompt) => (
                                    <div
                                        key={prompt.id}
                                        className="border rounded p-3"
                                    >
                                        <div className="flex justify-between items-start">
                                            <div>
                                                <h4 className="font-medium">{prompt.type}</h4>
                                                <p className="text-sm text-muted-foreground">
                                                    Created: {new Date(prompt.createdAt).toLocaleDateString()}
                                                </p>
                                                <p className="text-sm text-muted-foreground">
                                                    Updated: {new Date(prompt.updatedAt).toLocaleDateString()}
                                                </p>
                                            </div>
                                        </div>
                                        <div className="mt-2">
                                            <details>
                                                <summary className="cursor-pointer text-sm font-medium">
                                                    View Template
                                                </summary>
                                                <pre className="mt-2 text-xs bg-muted p-2 rounded overflow-auto max-h-32">
                                                    {prompt.template}
                                                </pre>
                                            </details>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <p className="text-muted-foreground">No user-defined prompts found.</p>
                        )}
                    </div>

                    {/* Default Prompts Summary */}
                    <div>
                        <h3 className="text-lg font-semibold mb-2">Default System Prompts</h3>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {/* Agent Prompts */}
                            <div className="border rounded p-3">
                                <h4 className="font-medium mb-2">Agent Prompts</h4>
                                <div className="text-sm space-y-1">
                                    <p>• Primary Agent</p>
                                    <p>• Assistant</p>
                                    <p>• Pentester</p>
                                    <p>• Coder</p>
                                    <p>• Installer</p>
                                    <p>• Searcher</p>
                                    <p>• Memorist</p>
                                    <p>• Adviser</p>
                                    <p>• Generator</p>
                                    <p>• Refiner</p>
                                    <p>• Reporter</p>
                                    <p>• Reflector</p>
                                    <p>• Enricher</p>
                                    <p>• Tool Call Fixer</p>
                                    <p>• Summarizer</p>
                                </div>
                            </div>

                            {/* Tool Prompts */}
                            <div className="border rounded p-3">
                                <h4 className="font-medium mb-2">Tool Prompts</h4>
                                <div className="text-sm space-y-1">
                                    <p>• Flow Description</p>
                                    <p>• Task Description</p>
                                    <p>• Execution Logs</p>
                                    <p>• Full Execution Context</p>
                                    <p>• Short Execution Context</p>
                                    <p>• Docker Image Chooser</p>
                                    <p>• Language Chooser</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
};

export default SettingsPrompts;
