import { useEffect, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';

import Logo from '@/components/icons/Logo';
import Markdown from '@/components/Markdown';
import { generateReport, generatePDFFromMarkdown, generateFileName } from '@/lib/report';
import { useFlowReportQuery } from '@/graphql/types';
import { Log } from '@/lib/log';

type ReportState = 'loading' | 'content' | 'generating' | 'error';

const Report = () => {
    const { flowId } = useParams<{ flowId: string }>();
    const [searchParams] = useSearchParams();
    const download = searchParams.has('download');
    const silent = searchParams.has('silent');

    const [state, setState] = useState<ReportState>('loading');
    const [error, setError] = useState<string | null>(null);
    const [reportContent, setReportContent] = useState<string>('');

    const { data, loading, error: queryError } = useFlowReportQuery({
        variables: { id: flowId! },
        skip: !flowId,
        errorPolicy: 'all'
    });

    // Reset state when component mounts or flowId changes
    useEffect(() => {
        setState('loading');
        setError(null);
        setReportContent('');
    }, [flowId]);

    useEffect(() => {
        if (loading) return;

        if (queryError || !data?.flow) {
            setError('Failed to load flow data');
            setState('error');
            return;
        }

        // Generate report content using flow and tasks from GraphQL response
        const content = generateReport(data.tasks || [], data.flow);
        setReportContent(content);

        if (download) {
            // Download mode - generate PDF and download it
            setState('generating');
            const fileName = `${generateFileName(data.flow)}.pdf`;
            
            generatePDFFromMarkdown(content, fileName)
                .then(() => {
                    if (silent) {
                        // Silent download - close window after successful download
                        setTimeout(() => window.close(), 1000);
                    } else {
                        // Normal download - show content after download
                        setState('content');
                    }
                })
                .catch((err) => {
                    Log.error('PDF generation failed:', err);
                    setError('Failed to generate PDF');
                    setState('error');
                });
        } else {
            setState('content');
        }
    }, [data, loading, queryError, download, silent]);

    // Loading state (for all modes during initial loading and PDF generation)
    if (state === 'loading' || state === 'generating') {
        return (
            <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
                <div className="flex flex-col items-center justify-center min-h-screen p-8">
                    <Logo className="mb-8 h-16 w-16 animate-logo-spin text-white" />
                    <div className="text-center space-y-4">
                        <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">
                            {state === 'loading' ? 'Loading Report...' : 'Generating PDF...'}
                        </h1>
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                        <p className="text-gray-600 dark:text-gray-400 max-w-md">
                            {state === 'loading'
                                ? 'Please wait while we prepare your penetration testing report.'
                                : 'Creating your PDF document. This may take a few moments.'}
                        </p>
                    </div>
                </div>
            </div>
        );
    }

    // Error state
    if (state === 'error') {
        return (
            <div className="min-h-screen bg-gradient-to-br from-red-50 via-white to-orange-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
                <div className="flex flex-col items-center justify-center min-h-screen p-8">
                    <Logo className="mb-8 h-16 w-16" />
                    <div className="text-center space-y-4">
                        <h1 className="text-2xl font-semibold text-red-600 dark:text-red-400">Error Loading Report</h1>
                        <p className="text-gray-600 dark:text-gray-400 max-w-md">
                            {error || 'An unexpected error occurred while loading the report.'}
                        </p>
                        <button
                            onClick={() => window.close()}
                            className="mt-4 px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors"
                        >
                            Close
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    // Content viewing state (normal mode without download)
    return (
        <div className="min-h-screen bg-white dark:bg-gray-900">
            <div className="w-full h-screen overflow-auto p-8">
                <div className="max-w-4xl mx-auto">
                    <div className="prose prose-slate dark:prose-invert max-w-none">
                        <Markdown>{reportContent}</Markdown>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Report;
