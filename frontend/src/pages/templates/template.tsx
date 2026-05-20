import type { ReactNode } from 'react';

import { zodResolver } from '@hookform/resolvers/zod';
import {
    ChevronDown,
    Ellipsis,
    FileSymlink,
    FileText,
    Loader2,
    PanelRightClose,
    PanelRightOpen,
    Pencil,
    Save,
    Trash,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { z } from 'zod';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import {
    DetailNavigationButtons,
    DetailNavigationSheet,
    DetailNavigationToolbar,
} from '@/components/shared/detail-navigation';
import { InlineEditInput, useInlineEdit } from '@/components/shared/inline-edit';
import { Badge } from '@/components/ui/badge';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupTextareaAutosize } from '@/components/ui/input-group';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Spinner } from '@/components/ui/spinner';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useTemplateDetailNavigation } from '@/features/templates/use-template-detail-navigation';
import { useFlowTemplateQuery } from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { cn } from '@/lib/utils';
import { type Template, useTemplates } from '@/providers/templates-provider';

const formSchema = z.object({
    text: z.string().trim().min(1, { message: 'Text is required' }),
    title: z.string().trim().min(1, { message: 'Title is required' }),
});

type FormValues = z.infer<typeof formSchema>;

const PRESET_TEMPLATES: { text: string; title: string }[] = [
    {
        text: `Perform comprehensive security assessment of web application: {{TARGET_URL}}

Action plan:
1. Application Exploration: Navigate all pages, test features, identify endpoints and input vectors
2. Vulnerability Testing per endpoint:
   - Path Traversal: attempt to read /etc/passwd, focus on file download/upload features
   - XSS: inject unique markers, scan responses, craft context-specific payloads
   - SQL Injection: run sqlmap on inputs, use tamper scripts for WAF bypass
   - Command Injection: use time-based detection, try commix utility
   - SSRF: use Interactsh for OOB, target file upload/PDF generation endpoints
   - XXE: test XML uploads and Office documents
   - Unsafe File Upload: test executable extensions, double extensions, null byte injection
   - CSRF: test token validation, POST to GET conversion
3. Authentication & Session: test for broken authentication, session fixation, weak password policies
4. Business Logic: identify privilege escalation, price manipulation, workflow bypass opportunities
5. Report: document all findings with reproduction steps and proof-of-concept exploits`,
        title: 'Web Application Security Assessment',
    },
    {
        text: `Perform network infrastructure reconnaissance of target: {{TARGET_NETWORK}}

Action plan:
1. Network Discovery: identify live hosts using nmap ping sweeps, map network topology
2. Port Scanning: comprehensive port scan (1-65535), identify all open services
3. Service Enumeration: fingerprint service versions, detect OS information
4. Vulnerability Scanning: run automated vulnerability scans against discovered services
5. SSL/TLS Analysis: check certificate validity, weak ciphers, protocol vulnerabilities
6. Banner Grabbing: collect detailed service information for exploit research
7. Network Diagram: create visual map of discovered infrastructure
8. Report: prioritized list of hosts, services, and potential attack vectors`,
        title: 'Network Infrastructure Discovery & Mapping',
    },
    {
        text: `Conduct Active Directory security assessment for domain: {{DOMAIN_NAME}}

Action plan:
1. Initial Access: test password spraying, check for AS-REP roasting, look for Kerberoastable accounts
2. Domain Enumeration: enumerate users, groups, computers, GPOs, trust relationships
3. Privilege Escalation: identify misconfigured ACLs, check for exploitable group memberships, find delegation issues
4. Credential Harvesting: search for credentials in SYSVOL, check for password in AD attributes, dump NTDS.dit if possible
5. Lateral Movement: test pass-the-hash, pass-the-ticket, overpass-the-hash techniques
6. Persistence: identify opportunities for golden ticket, silver ticket, DCSync rights
7. Domain Admin Path: map attack path from current privileges to Domain Admin
8. Report: document attack chain, compromised accounts, security gaps in AD configuration`,
        title: 'Active Directory Penetration Test',
    },
    {
        text: `Perform comprehensive API security assessment: {{API_BASE_URL}}

Action plan:
1. API Discovery: identify all endpoints, HTTP methods, parameters
2. Authentication Testing: test broken authentication, token manipulation, JWT vulnerabilities
3. Authorization Testing: test broken object-level authorization (BOLA/IDOR), function-level authorization bypass
4. Input Validation: test injection attacks (SQL, NoSQL, Command, XXE), mass assignment vulnerabilities
5. Rate Limiting: test for absence of rate limiting, brute force protection
6. Business Logic: test for excessive data exposure, lack of resource limiting, unsafe consumption of APIs
7. Security Misconfiguration: check CORS policy, security headers, verbose error messages
8. GraphQL Specific (if applicable): test introspection, query depth limits, batching attacks
9. Report: document API vulnerabilities with curl/Postman proof-of-concepts`,
        title: 'API Security Testing',
    },
    {
        text: `Perform security audit of AWS infrastructure: {{AWS_ACCOUNT_ID or DOMAIN}}

Action plan:
1. Reconnaissance: identify S3 buckets, EC2 instances, public endpoints, enumerate services via DNS
2. S3 Security: test bucket permissions, public access, ACL misconfigurations, bucket policies
3. IAM Assessment: review roles, policies, check for overly permissive permissions, find unused credentials
4. EC2 Security: scan for open security groups, test instance metadata service (169.254.169.254), check IMDSv2
5. Network Security: review VPC configurations, security groups, NACLs, public subnets
6. Database Exposure: check RDS public accessibility, security groups, encryption settings
7. Lambda Functions: test for function URL exposure, environment variable leaks, IAM role permissions
8. CloudTrail & Logging: verify logging is enabled, check for security monitoring gaps
9. Report: prioritized cloud security findings with AWS-specific remediation steps`,
        title: 'Cloud Infrastructure Security Audit (AWS)',
    },
    {
        text: `Conduct WordPress security assessment: {{WORDPRESS_URL}}

Action plan:
1. Version Detection: identify WordPress core version, theme, and active plugins
2. Plugin Vulnerabilities: enumerate installed plugins, check for known CVEs using WPScan and Sploitus
3. Theme Vulnerabilities: identify theme version, search for known exploits
4. User Enumeration: enumerate valid usernames via REST API, author archives, login responses
5. Authentication Testing: test weak passwords, brute force protection, 2FA bypass
6. File Upload: test media upload restrictions, arbitrary file upload vulnerabilities
7. XML-RPC: check if enabled, test pingback SSRF, brute force amplification
8. SQL Injection: test search functionality, custom query parameters, plugin-specific inputs
9. XSS Testing: test comments, search, contact forms, custom fields
10. Configuration Issues: check wp-config.php exposure, directory listing, sensitive file access
11. Report: document WordPress-specific vulnerabilities with exploit steps`,
        title: 'WordPress Security Assessment',
    },
    {
        text: `Perform external attack surface assessment for organization: {{ORGANIZATION_NAME or DOMAIN}}

Action plan:
1. Asset Discovery: enumerate all domains, subdomains (subfinder, amass), IP ranges, ASN information
2. Certificate Transparency: search crt.sh for subdomains, identify forgotten assets
3. Port Scanning: scan all discovered assets for open ports and services
4. Web Application Fingerprinting: identify technologies, CMS, frameworks, server versions
5. Email Security: test SPF, DKIM, DMARC records, email spoofing potential
6. Cloud Asset Discovery: search for exposed S3 buckets, Azure blobs, exposed cloud databases
7. Sensitive Data Exposure: search GitHub, GitLab, Pastebin for leaked credentials, API keys
8. Third-Party Integrations: identify SaaS applications, API endpoints, partner integrations
9. Vulnerability Prioritization: identify internet-facing critical vulnerabilities
10. Report: comprehensive external attack surface map with risk-prioritized findings`,
        title: 'External Attack Surface Assessment',
    },
    {
        text: `Conduct internal network penetration test from position: {{INITIAL_ACCESS_LEVEL}}

Action plan:
1. Network Reconnaissance: ARP scanning, identify network segments, map internal infrastructure
2. Service Discovery: comprehensive port scanning of internal hosts, identify critical servers
3. SMB/NetBIOS Enumeration: test null sessions, enumerate shares, check for anonymous access
4. Credential Attacks: LLMNR/NBT-NS poisoning (Responder), relay attacks, password spraying
5. Vulnerability Exploitation: exploit unpatched services, test default credentials, known CVEs
6. Privilege Escalation: exploit local vulnerabilities, misconfigured services, weak permissions
7. Lateral Movement: pass-the-hash, token impersonation, exploit trust relationships
8. Data Exfiltration: identify sensitive data locations, test data loss prevention controls
9. Persistence: establish persistent access mechanisms
10. Report: document internal security posture, attack path visualization, remediation priorities`,
        title: 'Internal Network Penetration Test',
    },
    {
        text: `Perform security testing of mobile application backend API: {{API_URL}}

Action plan:
1. Traffic Interception: analyze mobile app traffic, extract API endpoints and authentication
2. Authentication Mechanisms: test OAuth flows, JWT implementation, refresh token handling, certificate pinning bypass
3. API Endpoint Testing: test all discovered endpoints for BOLA/IDOR, broken function-level authorization
4. Data Validation: test for injection attacks in API parameters, test file upload endpoints
5. Business Logic: test premium feature bypass, subscription validation, in-app purchase verification
6. Session Management: test token expiration, concurrent session handling, session fixation
7. Sensitive Data: check for PII exposure, excessive data in responses, hardcoded secrets
8. Rate Limiting: test brute force protection on login, API rate limits, account lockout
9. Deep Linking: test for deep link hijacking, intent redirection (Android), URL scheme abuse (iOS)
10. Report: mobile-specific vulnerabilities with mitigation recommendations`,
        title: 'Mobile Application Security Testing (API Backend)',
    },
    {
        text: `Assess DevOps infrastructure and CI/CD pipeline security: {{ORGANIZATION}}

Action plan:
1. Repository Security: scan GitHub/GitLab for exposed secrets, API keys, credentials in commit history
2. CI/CD Configuration: review Jenkins/GitLab CI/GitHub Actions configurations, test for injection in pipeline definitions
3. Container Security: scan Docker images for vulnerabilities, test for container escape, check image sources
4. Secrets Management: test secret storage (HashiCorp Vault, AWS Secrets Manager), check for hardcoded secrets
5. Access Control: review permissions on repositories, pipeline access, deployment keys, service accounts
6. Artifact Security: scan build artifacts, test artifact repository access controls (Nexus, Artifactory)
7. Kubernetes Security: review pod security policies, RBAC, network policies, exposed dashboards
8. Infrastructure as Code: review Terraform/Ansible for misconfigurations, overly permissive IAM roles
9. Monitoring & Logging: verify security logging, test log tampering, check for security monitoring gaps
10. Report: DevOps security findings with secure pipeline recommendations`,
        title: 'DevOps & CI/CD Pipeline Security',
    },
    {
        text: `Conduct database security assessment: {{DATABASE_TYPE}} at {{HOST:PORT}}

Action plan:
1. Access Testing: test for default credentials, weak passwords, anonymous access
2. Network Exposure: verify database should not be internet-accessible, check firewall rules
3. Authentication: test authentication mechanisms, user enumeration, password policies
4. Authorization: review user permissions, test for privilege escalation, check for excessive grants
5. Injection Testing: SQL injection in application layer, test stored procedures for injection
6. Configuration Review: check for dangerous configuration options (xp_cmdshell, LOAD DATA, file_priv)
7. Encryption: verify data-at-rest encryption, SSL/TLS for connections, check for sensitive data in plaintext
8. Backup Security: test backup file access, check backup encryption, verify backup restoration procedures
9. Audit Logging: verify audit logs enabled, test log tampering, check retention policies
10. Report: database-specific security findings with hardening recommendations`,
        title: 'Database Security Assessment',
    },
];

const renderTemplateItem = (item: Template, isCurrent: boolean): ReactNode => (
    <span className={cn('min-w-0 flex-1 truncate', isCurrent && 'font-medium')}>{item.title}</span>
);

function Template() {
    const navigate = useNavigate();
    const { templateId } = useParams<{ templateId?: string }>();
    const { createTemplate, deleteTemplate, updateTemplate } = useTemplates();

    const { isMobile } = useBreakpoint();
    const isNew = templateId === 'new';

    // Pass `null` while creating a new template — there is no "current item"
    // to highlight, and the toolbar shouldn't render at all anyway (gated
    // below by `canShowActions`).
    const templateNav = useTemplateDetailNavigation(isNew ? null : templateId);

    const [isAsideOpen, setIsAsideOpen] = useState(false);
    const [expandedPresetIndex, setExpandedPresetIndex] = useState<null | number>(null);
    const [isReplaceConfirmOpen, setIsReplaceConfirmOpen] = useState(false);
    const [isSaving, setIsSaving] = useState(false);
    const [pendingPreset, setPendingPreset] = useState<null | { text: string; title: string }>(null);
    const [isRenaming, setIsRenaming] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

    const {
        handleDropdownCloseAutoFocus,
        inputRef: editingInputRef,
        isEditing: isEditingTitle,
        startEdit: handleTemplateRenameStart,
        stopEdit: handleTemplateRenameCancel,
    } = useInlineEdit({ resetKey: templateId });

    const { data: templateData, loading: isLoadingTemplate } = useFlowTemplateQuery({
        skip: isNew || !templateId,
        variables: templateId && !isNew ? { templateId } : undefined,
    });

    const form = useForm<FormValues>({
        defaultValues: { text: '', title: '' },
        mode: 'onChange',
        resolver: zodResolver(formSchema),
    });

    const { control, formState, getValues, handleSubmit: handleFormSubmit, reset, setValue } = form;

    useEffect(() => {
        if (isNew || !templateData?.flowTemplate) {
            return;
        }

        const { text, title } = templateData.flowTemplate;
        reset({ text, title }, { keepDefaultValues: false });
    }, [templateData, isNew, reset]);

    const hasUnsavedChanges = formState.isDirty;
    const templateName = templateData?.flowTemplate?.title ?? null;

    const handleTemplateRenameSave = useCallback(async () => {
        const newTitle = editingInputRef.current?.value.trim();
        const template = templateData?.flowTemplate;

        if (!templateId || !newTitle || !template) {
            return;
        }

        if (newTitle === template.title) {
            handleTemplateRenameCancel();

            return;
        }

        setIsRenaming(true);

        try {
            // Preserve the original `text` from the server so that an inline
            // rename never overwrites unsaved edits in the form below.
            await updateTemplate(templateId, { text: template.text, title: newTitle });
            toast.success('Template renamed successfully');
            handleTemplateRenameCancel();
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsRenaming(false);
        }
    }, [editingInputRef, handleTemplateRenameCancel, templateId, templateData?.flowTemplate, updateTemplate]);

    const handleTemplateDelete = useCallback(async () => {
        if (!templateId) {
            return;
        }

        setIsDeleting(true);

        try {
            await deleteTemplate(templateId);
            navigate('/templates', { replace: true });
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsDeleting(false);
        }
    }, [templateId, deleteTemplate, navigate]);

    const handleSubmit = async (values: FormValues) => {
        if (isSaving) {
            return;
        }

        setIsSaving(true);

        try {
            if (isNew) {
                await createTemplate(values.title, values.text);
                navigate('/templates');
            } else if (templateId) {
                await updateTemplate(templateId, { text: values.text, title: values.title });
                reset(values, { keepDefaultValues: false });
            }
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsSaving(false);
        }
    };

    const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
        const { ctrlKey, key, metaKey, shiftKey } = event;

        if (isSaving || key !== 'Enter' || shiftKey || ctrlKey || metaKey) {
            return;
        }

        event.preventDefault();
        handleFormSubmit(handleSubmit)();
    };

    const handleApplyPreset = useCallback(
        (preset: { text: string; title: string }) => {
            const current = getValues();
            const hasContent = (current.title?.trim().length ?? 0) > 0 || (current.text?.trim().length ?? 0) > 0;

            if (hasContent) {
                setPendingPreset(preset);
                setIsReplaceConfirmOpen(true);
            } else {
                setValue('title', preset.title, { shouldDirty: true, shouldValidate: true });
                setValue('text', preset.text, { shouldDirty: true, shouldValidate: true });
            }
        },
        [getValues, setValue],
    );

    const handleConfirmReplacePreset = useCallback(() => {
        if (pendingPreset) {
            setValue('title', pendingPreset.title, { shouldDirty: true, shouldValidate: true });
            setValue('text', pendingPreset.text, { shouldDirty: true, shouldValidate: true });
            setPendingPreset(null);
        }
    }, [pendingPreset, setValue]);

    const canShowActions = !isNew && !!templateData?.flowTemplate;

    const pageHeader = (
        <>
            <header className="bg-background sticky top-0 z-10 flex h-12 shrink-0 items-center gap-2 border-b px-4">
                <div className="flex min-w-0 flex-1 items-center gap-2">
                    <SidebarTrigger className="-ml-1 shrink-0" />
                    <Separator
                        className="mr-2 h-4 shrink-0"
                        orientation="vertical"
                    />
                    <Breadcrumb className="min-w-0 flex-1">
                        <BreadcrumbList className="min-w-0 flex-nowrap">
                            <BreadcrumbItem className="min-w-0 gap-2">
                                {isEditingTitle && canShowActions ? (
                                    <InlineEditInput
                                        busy={isRenaming}
                                        className="w-64 max-w-full min-w-0 flex-1"
                                        defaultValue={templateName ?? ''}
                                        inputRef={editingInputRef}
                                        onCancel={handleTemplateRenameCancel}
                                        onSave={handleTemplateRenameSave}
                                        placeholder="Template title"
                                    />
                                ) : canShowActions ? (
                                    <Tooltip>
                                        <TooltipTrigger asChild>
                                            <BreadcrumbPage
                                                className="max-w-64 min-w-0 cursor-text truncate select-none"
                                                onDoubleClick={handleTemplateRenameStart}
                                            >
                                                {templateName ?? 'Template'}
                                            </BreadcrumbPage>
                                        </TooltipTrigger>
                                        <TooltipContent>Double-click to rename</TooltipContent>
                                    </Tooltip>
                                ) : (
                                    <BreadcrumbPage className="min-w-0 truncate">
                                        {isNew ? 'New template' : (templateName ?? 'Template')}
                                    </BreadcrumbPage>
                                )}
                            </BreadcrumbItem>
                        </BreadcrumbList>
                    </Breadcrumb>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                    {canShowActions && !isMobile && (
                        <DetailNavigationToolbar<Template>
                            controller={templateNav}
                            renderItem={renderTemplateItem}
                            sheetIcon={<FileText className="size-4" />}
                            sheetTitle="Templates"
                        />
                    )}
                    <Button
                        onClick={() => setIsAsideOpen((open) => !open)}
                        size="icon"
                        variant="ghost"
                    >
                        {isAsideOpen ? <PanelRightClose /> : <PanelRightOpen />}
                    </Button>
                    {canShowActions && (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    aria-label="Template actions"
                                    className="size-8 p-0"
                                    variant="ghost"
                                >
                                    <Ellipsis />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-24"
                                onCloseAutoFocus={handleDropdownCloseAutoFocus}
                            >
                                {isMobile && templateNav.total > 0 && (
                                    <>
                                        <DropdownMenuItem
                                            className="cursor-default hover:bg-transparent focus:bg-transparent"
                                            onSelect={(event) => event.preventDefault()}
                                        >
                                            <FileText className="size-4" />
                                            Templates
                                            <div className="-my-1.5 -mr-2 ml-auto flex items-center">
                                                <DetailNavigationButtons<Template>
                                                    controller={templateNav}
                                                    sheetTitle="Templates"
                                                    size="sm"
                                                />
                                            </div>
                                        </DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                    </>
                                )}
                                <DropdownMenuItem onClick={handleTemplateRenameStart}>
                                    <Pencil className="size-3" />
                                    Rename
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    disabled={isDeleting}
                                    onClick={() => setIsDeleteDialogOpen(true)}
                                >
                                    {isDeleting ? (
                                        <>
                                            <Loader2 className="size-4 animate-spin" />
                                            Deleting...
                                        </>
                                    ) : (
                                        <>
                                            <Trash className="size-4" />
                                            Delete
                                        </>
                                    )}
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    )}
                </div>
            </header>
            {isMobile && canShowActions && (
                <DetailNavigationSheet<Template>
                    controller={templateNav}
                    renderItem={renderTemplateItem}
                    sheetIcon={<FileText className="size-4" />}
                    sheetTitle="Templates"
                />
            )}
        </>
    );

    const asideContent = useMemo(
        () => (
            <div className="flex w-full min-w-0 flex-col gap-2 p-2">
                {PRESET_TEMPLATES.map((preset, index) => (
                    <Collapsible
                        className="w-full min-w-0"
                        key={index}
                        onOpenChange={(open) => setExpandedPresetIndex(open ? index : null)}
                        open={expandedPresetIndex === index}
                    >
                        <Card className="w-full min-w-0">
                            <div className="flex w-full min-w-0">
                                <Button
                                    className={cn(
                                        'h-auto min-w-0 flex-1 justify-start rounded-none rounded-tl-[0.6875rem] px-3 py-2 text-left text-start',
                                        expandedPresetIndex !== index ? 'rounded-bl-[0.6875rem]' : 'whitespace-normal',
                                    )}
                                    onClick={() => handleApplyPreset(preset)}
                                    variant="ghost"
                                >
                                    <span className={cn('min-w-0', expandedPresetIndex !== index && 'truncate')}>
                                        {preset.title}
                                    </span>
                                </Button>
                                <CollapsibleTrigger asChild>
                                    <Button
                                        className={cn(
                                            'h-auto shrink-0 rounded-none rounded-tr-[0.6875rem] border-l px-2 py-2',
                                            expandedPresetIndex !== index && 'rounded-br-[0.6875rem]',
                                        )}
                                        variant="ghost"
                                    >
                                        <ChevronDown
                                            className={cn(
                                                'transition-transform',
                                                expandedPresetIndex === index && 'rotate-180',
                                            )}
                                        />
                                    </Button>
                                </CollapsibleTrigger>
                            </div>
                            <CollapsibleContent>
                                <CardContent className="border-t px-3 py-2">
                                    <p className="text-muted-foreground text-sm break-words whitespace-pre-wrap">
                                        {preset.text}
                                    </p>
                                </CardContent>
                            </CollapsibleContent>
                        </Card>
                    </Collapsible>
                ))}
            </div>
        ),
        [expandedPresetIndex, handleApplyPreset],
    );

    const aside = useMemo(
        () =>
            isMobile ? (
                <Sheet
                    onOpenChange={setIsAsideOpen}
                    open={isAsideOpen}
                >
                    <SheetContent
                        // The Sheet body is just a list of presets with no
                        // descriptive sub-text — opt out of the Radix
                        // Description warning explicitly.
                        aria-describedby={undefined}
                        className="flex w-full max-w-sm flex-col gap-0 p-0 sm:max-w-sm"
                        side="right"
                    >
                        <SheetHeader className="border-b p-4">
                            <SheetTitle className="flex items-center gap-2 pr-8 text-base">
                                <FileText className="size-4" />
                                <span>Preset templates</span>
                                <Badge
                                    className="ml-auto font-normal tabular-nums"
                                    variant="secondary"
                                >
                                    {PRESET_TEMPLATES.length}
                                </Badge>
                            </SheetTitle>
                        </SheetHeader>
                        {/* Plain overflow-y-auto instead of Radix ScrollArea —
                            ScrollArea's Viewport wraps children in a
                            `display: table` div whose width grows to fit
                            intrinsic content, defeating `w-full min-w-0` on
                            inner flex rows and pushing the chevron buttons
                            off-screen. */}
                        <div className="min-w-0 flex-1 overflow-y-auto">{asideContent}</div>
                    </SheetContent>
                </Sheet>
            ) : (
                <aside
                    className={cn(
                        'bg-background shrink-0 overflow-hidden transition-[width] duration-200',
                        isAsideOpen ? 'w-80 border-l sm:w-96' : 'w-0',
                    )}
                >
                    {isAsideOpen ? (
                        <div className="flex h-full w-80 min-w-0 flex-col sm:w-96">
                            <div className="border-b p-4">
                                <h3 className="flex items-center gap-2 text-base font-semibold">
                                    <FileText className="size-4" />
                                    <span>Preset templates</span>
                                    <Badge
                                        className="ml-auto font-normal tabular-nums"
                                        variant="secondary"
                                    >
                                        {PRESET_TEMPLATES.length}
                                    </Badge>
                                </h3>
                            </div>
                            <div className="min-w-0 flex-1 overflow-y-auto">{asideContent}</div>
                        </div>
                    ) : null}
                </aside>
            ),
        [isMobile, isAsideOpen, asideContent],
    );

    if (!isNew && isLoadingTemplate) {
        return (
            <>
                {pageHeader}
                <div className="flex min-h-[calc(100dvh-3rem)] items-center justify-center">
                    <Spinner variant="circle" />
                </div>
            </>
        );
    }

    if (!isNew && !isLoadingTemplate && !templateData?.flowTemplate) {
        return (
            <>
                {pageHeader}
                <div className="flex min-h-[calc(100dvh-3rem)] items-center justify-center p-4">
                    <Card className="w-full max-w-2xl">
                        <CardContent className="flex flex-col items-center gap-4 pt-6 text-center">
                            <h2 className="text-xl font-semibold">Template not found</h2>
                            <p className="text-muted-foreground">The template you are looking for does not exist.</p>
                            <Button onClick={() => navigate('/templates')}>Back to Templates</Button>
                        </CardContent>
                    </Card>
                </div>
            </>
        );
    }

    return (
        <>
            {pageHeader}
            <div className="flex min-h-[calc(100dvh-3rem)]">
                <div className="flex min-w-0 flex-1 items-center justify-center p-4">
                    <Card className="w-full max-w-2xl">
                        <CardContent className="flex flex-col gap-4 pt-6">
                            <div className="text-center">
                                <h1 className="text-2xl font-semibold">
                                    {isNew ? 'Create a new template' : 'Edit template'}
                                </h1>
                                <p className="text-muted-foreground mt-2">
                                    Add title and content for your template or use a
                                    <Button
                                        className="h-auto px-1.5 py-0 text-base"
                                        onClick={() => setIsAsideOpen((open) => !open)}
                                        variant="link"
                                    >
                                        Preset template
                                    </Button>
                                </p>
                            </div>
                            <Form {...form}>
                                <form
                                    className="flex flex-col gap-4"
                                    onSubmit={handleFormSubmit(handleSubmit)}
                                >
                                    <FormField
                                        control={control}
                                        name="title"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormControl>
                                                    <Input
                                                        autoFocus={isNew}
                                                        disabled={isSaving}
                                                        placeholder="Title"
                                                        {...field}
                                                    />
                                                </FormControl>
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={control}
                                        name="text"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormControl>
                                                    <InputGroup className="block">
                                                        <InputGroupTextareaAutosize
                                                            {...field}
                                                            className="min-h-0"
                                                            disabled={isSaving}
                                                            maxRows={9}
                                                            minRows={1}
                                                            onKeyDown={handleKeyDown}
                                                            placeholder="Content"
                                                        />
                                                        <InputGroupAddon align="block-end">
                                                            <InputGroupButton
                                                                aria-label={isNew ? 'Create template' : 'Save template'}
                                                                className="ml-auto"
                                                                disabled={
                                                                    isSaving ||
                                                                    !formState.isValid ||
                                                                    (!isNew && !hasUnsavedChanges)
                                                                }
                                                                size="icon-xs"
                                                                title={isNew ? 'Create template' : 'Save template'}
                                                                type="submit"
                                                                variant="default"
                                                            >
                                                                {isSaving ? (
                                                                    <Spinner variant="circle" />
                                                                ) : (
                                                                    <Save aria-hidden="true" />
                                                                )}
                                                            </InputGroupButton>
                                                        </InputGroupAddon>
                                                    </InputGroup>
                                                </FormControl>
                                            </FormItem>
                                        )}
                                    />
                                </form>
                            </Form>
                        </CardContent>
                    </Card>
                </div>
                {aside}
            </div>
            <ConfirmationDialog
                confirmIcon={<FileSymlink />}
                confirmText="Replace"
                confirmVariant="default"
                description="Current form has content. Replace with the selected preset?"
                handleConfirm={handleConfirmReplacePreset}
                handleOpenChange={(open) => {
                    if (!open) {
                        setPendingPreset(null);
                    }

                    setIsReplaceConfirmOpen(open);
                }}
                isOpen={isReplaceConfirmOpen}
                title="Replace content?"
            />
            <ConfirmationDialog
                cancelText="Cancel"
                confirmText="Delete"
                handleConfirm={handleTemplateDelete}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={templateName ?? undefined}
                itemType="template"
            />
        </>
    );
}

export default Template;
