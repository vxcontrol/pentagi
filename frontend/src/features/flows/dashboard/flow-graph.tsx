import { Maximize, Minus, Plus } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';

import type { GraphData, GraphEdge, GraphNode as GraphNodeData } from './use-flow-dashboard';

// --- Types ---

interface FlowGraphProps {
    data: GraphData;
}

interface GraphLink {
    relationshipType: string;
    source: GraphNode | string;
    target: GraphNode | string;
}

interface GraphNode {
    description?: string;
    entityType: string;
    id: string;
    name: string;
    properties: Record<string, unknown>;
    summary?: string;
    x?: number;
    y?: number;
}

interface ViewTransform {
    scale: number;
    x: number;
    y: number;
}

// --- Color palette per entity type ---

const ENTITY_COLORS: Record<string, string> = {
    Account: '#a855f7', // purple
    Agent: '#8b5cf6', // violet
    Artifact: '#ec4899', // pink
    AttackTechnique: '#dc2626', // red-dark
    Attempt: '#f43f5e', // rose
    Capability: '#0d9488', // teal
    Credential: '#d946ef', // fuchsia
    Entity: '#64748b', // slate (generic)
    Episodic: '#78716c', // stone
    Host: '#3b82f6', // blue
    Misconfiguration: '#f59e0b', // amber
    Port: '#06b6d4', // cyan
    Run: '#6366f1', // indigo
    Service: '#22c55e', // green
    Subtask: '#0ea5e9', // sky
    Tool: '#6b7280', // gray
    ToolExecution: '#9ca3af', // gray-light
    ValidAccess: '#f97316', // orange
    Vulnerability: '#ef4444', // red
    WebApp: '#14b8a6', // teal-light
};

const DEFAULT_NODE_COLOR = '#64748b';

const NODE_WIDTH = 140;
const NODE_HEIGHT = 40;
const NODE_HALF_WIDTH = NODE_WIDTH / 2;
const NODE_HALF_HEIGHT = NODE_HEIGHT / 2;
const NODE_BORDER_RADIUS = 8;
const MIN_ZOOM = 0.15;
const MAX_ZOOM = 5;
const ZOOM_STEP = 1.15;

// --- Helpers ---

// Calculate a transform that centers and fits the graph within the viewport
function computeFitTransform(nodes: GraphNode[], viewportWidth: number, viewportHeight: number): ViewTransform {
    if (!nodes.length) {
        return { scale: 1, x: 0, y: 0 };
    }

    const padding = 60;
    let minX = Infinity;
    let minY = Infinity;
    let maxX = -Infinity;
    let maxY = -Infinity;

    for (const node of nodes) {
        if (node.x != null && node.y != null) {
            minX = Math.min(minX, node.x - NODE_HALF_WIDTH);
            minY = Math.min(minY, node.y - NODE_HALF_HEIGHT);
            maxX = Math.max(maxX, node.x + NODE_HALF_WIDTH);
            maxY = Math.max(maxY, node.y + NODE_HALF_HEIGHT);
        }
    }

    const graphWidth = maxX - minX + padding * 2;
    const graphHeight = maxY - minY + padding * 2;

    const scale = Math.min(Math.max(viewportWidth / graphWidth, MIN_ZOOM), Math.min(viewportHeight / graphHeight, 1));

    const centerX = (minX + maxX) / 2;
    const centerY = (minY + maxY) / 2;

    return {
        scale,
        x: viewportWidth / 2 - centerX * scale,
        y: viewportHeight / 2 - centerY * scale,
    };
}

// Semantic left-to-right layout (fixed levels per entity type)
function computeLayout(rawNodes: GraphNode[], rawLinks: GraphLink[]): { links: GraphLink[]; nodes: GraphNode[] } {
    const nodes: GraphNode[] = rawNodes.map((node) => ({ ...node }));
    const links: GraphLink[] = rawLinks.map((link) => ({ ...link }));

    if (!nodes.length) {
        return { links, nodes };
    }

    const HORIZONTAL_SPACING = 240;
    const VERTICAL_SPACING = 70;

    // Build adjacency maps
    const nodeById = new Map<string, GraphNode>();
    const childrenIds = new Map<string, string[]>();
    const parentIds = new Map<string, string[]>();

    for (const node of nodes) {
        nodeById.set(node.id, node);
        childrenIds.set(node.id, []);
        parentIds.set(node.id, []);
    }

    for (const link of links) {
        const sourceId = typeof link.source === 'string' ? link.source : link.source.id;
        const targetId = typeof link.target === 'string' ? link.target : link.target.id;
        childrenIds.get(sourceId)?.push(targetId);
        parentIds.get(targetId)?.push(sourceId);
    }

    const SEMANTIC_LEVEL: Record<string, number> = {
        Host: 0,
        Port: 1,
        Service: 2,
        VHost: 2,
        WebApp: 3,
        Endpoint: 4,
        Vulnerability: 5,
        Misconfiguration: 5,
        Credential: 6,
        Artifact: 6,
        Attempt: 6,
        ValidAccess: 7,
        Account: 8,
        PrivChange: 8,
        Capability: 9,
        AttackTechnique: 9,
    };

    const DEFAULT_LEVEL = 5;

    // Assign level by entity type
    const depth = new Map<string, number>();

    for (const node of nodes) {
        depth.set(node.id, SEMANTIC_LEVEL[node.entityType] ?? DEFAULT_LEVEL);
    }

    // Group nodes by semantic level
    const maxDepth = Math.max(...Array.from(depth.values()));
    const allLevels: GraphNode[][] = Array.from({ length: maxDepth + 1 }, () => []);

    for (const node of nodes) {
        allLevels[depth.get(node.id) ?? 0]?.push(node);
    }

    // Compact: remove empty levels
    const levels = allLevels.filter((group) => group.length > 0);

    // Sort within each level by average neighbor position (barycenter heuristic)
    for (let levelIdx = 0; levelIdx < levels.length; levelIdx++) {
        const groupNodes = levels[levelIdx];

        if (!groupNodes || levelIdx === 0) {
            // First level: assign positions without sorting
            if (groupNodes) {
                const groupHeight = (groupNodes.length - 1) * VERTICAL_SPACING;

                for (const [index, node] of groupNodes.entries()) {
                    node.x = 0;
                    node.y = -groupHeight / 2 + index * VERTICAL_SPACING;
                }
            }

            continue;
        }

        // Collect y positions of all already-positioned nodes
        const positionedY = new Map<string, number>();

        for (let prev = 0; prev < levelIdx; prev++) {
            for (const node of levels[prev]!) {
                if (node.y != null) {
                    positionedY.set(node.id, node.y);
                }
            }
        }

        // Sort by average y of all connected neighbors in previous levels
        groupNodes.sort((a, b) => {
            const avgY = (nodeId: string): number => {
                const neighbors = [...(parentIds.get(nodeId) ?? []), ...(childrenIds.get(nodeId) ?? [])].filter(
                    (id) => positionedY.has(id),
                );

                if (!neighbors.length) {
                    return 0;
                }

                return neighbors.reduce((sum, id) => sum + (positionedY.get(id) ?? 0), 0) / neighbors.length;
            };

            return avgY(a.id) - avgY(b.id);
        });

        const groupHeight = (groupNodes.length - 1) * VERTICAL_SPACING;

        for (const [index, node] of groupNodes.entries()) {
            node.x = levelIdx * HORIZONTAL_SPACING;
            node.y = -groupHeight / 2 + index * VERTICAL_SPACING;
        }
    }

    // Resolve link references from string IDs to actual node objects
    for (const link of links) {
        const sourceId = typeof link.source === 'string' ? link.source : link.source.id;
        const targetId = typeof link.target === 'string' ? link.target : link.target.id;
        link.source = nodeById.get(sourceId)!;
        link.target = nodeById.get(targetId)!;
    }

    return { links, nodes };
}

function getEntityType(node: GraphNodeData): string {
    return node.labels.find((label) => label !== 'Entity') ?? node.labels[0] ?? 'Unknown';
}

function getNodeColor(entityType: string): string {
    return ENTITY_COLORS[entityType] ?? DEFAULT_NODE_COLOR;
}

function getNodeId(node: GraphNodeData): string {
    const properties = (node.properties ?? {}) as Record<string, unknown>;
    const name = String(properties.name ?? properties.uuid ?? '');
    const type = getEntityType(node);

    return `${type}:${name}`;
}

function getNodeName(node: GraphNodeData): string {
    const properties = (node.properties ?? {}) as Record<string, unknown>;

    return String(properties.name ?? 'unnamed');
}

function getOptionalString(node: GraphNodeData, key: string): string | undefined {
    const properties = (node.properties ?? {}) as Record<string, unknown>;
    const value = properties[key];

    return typeof value === 'string' && value.trim() ? value.trim() : undefined;
}

// Parse flat graph edge records into graph nodes and links
function parseGraphEdges(data: GraphEdge[]): { links: GraphLink[]; nodes: GraphNode[] } {
    const nodesMap = new Map<string, GraphNode>();
    const linksSet = new Set<string>();
    const links: GraphLink[] = [];

    for (const record of data) {
        const sourceId = getNodeId(record.source);
        const targetId = getNodeId(record.target);
        const sourceProperties = (record.source.properties ?? {}) as Record<string, unknown>;
        const targetProperties = (record.target.properties ?? {}) as Record<string, unknown>;

        if (!nodesMap.has(sourceId)) {
            nodesMap.set(sourceId, {
                description: getOptionalString(record.source, 'description'),
                entityType: getEntityType(record.source),
                id: sourceId,
                name: getNodeName(record.source),
                properties: sourceProperties,
                summary: getOptionalString(record.source, 'summary'),
            });
        }

        if (!nodesMap.has(targetId)) {
            nodesMap.set(targetId, {
                description: getOptionalString(record.target, 'description'),
                entityType: getEntityType(record.target),
                id: targetId,
                name: getNodeName(record.target),
                properties: targetProperties,
                summary: getOptionalString(record.target, 'summary'),
            });
        }

        const linkKey = `${sourceId}|${targetId}`;

        if (!linksSet.has(linkKey)) {
            linksSet.add(linkKey);
            links.push({
                relationshipType: record.relationType,
                source: sourceId,
                target: targetId,
            });
        }
    }

    return { links, nodes: Array.from(nodesMap.values()) };
}

// Truncate long names for display
function truncateLabel(text: string, maxLength: number = 14): string {
    return text.length > maxLength ? `${text.slice(0, maxLength)}...` : text;
}

// --- Component ---

const FlowGraph = ({ data }: FlowGraphProps) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const [dimensions, setDimensions] = useState({ height: 500, width: 800 });
    const [hoveredNode, setHoveredNode] = useState<GraphNode | null>(null);
    const [hoveredLink, setHoveredLink] = useState<GraphLink | null>(null);

    // Pan & zoom state: null = auto-fit, set = user override
    const [userTransform, setUserTransform] = useState<null | ViewTransform>(null);
    const [isPanning, setIsPanning] = useState(false);
    const panStartRef = useRef({ clientX: 0, clientY: 0, x: 0, y: 0 });

    // Observe container size for responsiveness
    useEffect(() => {
        const container = containerRef.current;

        if (!container) {
            return;
        }

        const observer = new ResizeObserver((entries) => {
            const entry = entries[0];

            if (entry) {
                const { height, width } = entry.contentRect;
                setDimensions({
                    height: Math.max(height, 400),
                    width: Math.max(width, 300),
                });
            }
        });

        observer.observe(container);

        return () => observer.disconnect();
    }, []);

    // Parse raw data
    const { links: rawLinks, nodes: rawNodes } = useMemo(() => parseGraphEdges(data.data), [data.data]);

    // Static layout
    const { links, nodes } = useMemo(() => computeLayout(rawNodes, rawLinks), [rawNodes, rawLinks]);

    // Auto-fit transform computed from layout bounds and current viewport
    const autoFitTransform = useMemo(
        () => computeFitTransform(nodes, dimensions.width, dimensions.height),
        [nodes, dimensions.width, dimensions.height],
    );

    // Active transform: user override or auto-fit
    const transform = userTransform ?? autoFitTransform;

    // Collect unique entity types for the legend toggles
    const entityTypes = useMemo(() => {
        const types = new Set(nodes.map((node) => node.entityType));

        return Array.from(types).sort();
    }, [nodes]);

    // Visible entity types filter (all visible by default)
    const [visibleTypes, setVisibleTypes] = useState<null | string[]>(null);

    // Use user selection if set, otherwise show all types
    const activeVisibleTypes = visibleTypes ?? entityTypes;

    // Derived: which nodes/links to actually render
    const visibleTypesSet = useMemo(() => new Set(activeVisibleTypes), [activeVisibleTypes]);

    const visibleNodes = useMemo(
        () => nodes.filter((node) => visibleTypesSet.has(node.entityType)),
        [nodes, visibleTypesSet],
    );

    const visibleLinks = useMemo(
        () =>
            links.filter((link) => {
                const source = link.source as GraphNode;
                const target = link.target as GraphNode;

                return visibleTypesSet.has(source.entityType) && visibleTypesSet.has(target.entityType);
            }),
        [links, visibleTypesSet],
    );

    // --- Pan & zoom handlers ---

    const clampScale = useCallback((scale: number) => Math.min(Math.max(scale, MIN_ZOOM), MAX_ZOOM), []);

    const zoomAtPoint = useCallback(
        (centerX: number, centerY: number, factor: number) => {
            const previous = userTransform ?? autoFitTransform;
            const newScale = clampScale(previous.scale * factor);
            const ratio = newScale / previous.scale;

            setUserTransform({
                scale: newScale,
                x: centerX - (centerX - previous.x) * ratio,
                y: centerY - (centerY - previous.y) * ratio,
            });
        },
        [clampScale, userTransform, autoFitTransform],
    );

    // Mouse wheel zoom — native listener to allow preventDefault on passive scroll
    useEffect(() => {
        const container = containerRef.current;

        if (!container) {
            return;
        }

        const handleWheel = (event: WheelEvent) => {
            event.preventDefault();
            event.stopPropagation();

            const rect = container.getBoundingClientRect();
            const svgX = event.clientX - rect.left;
            const svgY = event.clientY - rect.top;
            const factor = event.deltaY < 0 ? ZOOM_STEP : 1 / ZOOM_STEP;

            zoomAtPoint(svgX, svgY, factor);
        };

        container.addEventListener('wheel', handleWheel, { passive: false });

        return () => container.removeEventListener('wheel', handleWheel);
    }, [zoomAtPoint]);

    // Pan: pointer down — record start position, don't start panning yet
    const handlePointerDown = useCallback(
        (event: React.PointerEvent) => {
            if (event.button !== 0) {
                return;
            }

            panStartRef.current = {
                clientX: event.clientX,
                clientY: event.clientY,
                x: transform.x,
                y: transform.y,
            };

            setIsPanning(true);
        },
        [transform.x, transform.y],
    );

    // Pan: pointer move — only apply pan if moved beyond threshold
    const PAN_THRESHOLD = 4;

    const handlePointerMove = useCallback(
        (event: React.PointerEvent) => {
            if (!isPanning) {
                return;
            }

            const deltaX = event.clientX - panStartRef.current.clientX;
            const deltaY = event.clientY - panStartRef.current.clientY;

            // Only update transform if pointer moved enough to count as a drag
            if (Math.abs(deltaX) > PAN_THRESHOLD || Math.abs(deltaY) > PAN_THRESHOLD) {
                setUserTransform((previous) => ({
                    ...(previous ?? autoFitTransform),
                    x: panStartRef.current.x + deltaX,
                    y: panStartRef.current.y + deltaY,
                }));
            }
        },
        [isPanning, autoFitTransform],
    );

    // Pan: pointer up
    const handlePointerUp = useCallback(() => {
        setIsPanning(false);
    }, []);

    // Fit all nodes into view — reset to auto-fit
    const handleFitToView = useCallback(() => {
        setUserTransform(null);
    }, []);

    const handleZoomIn = useCallback(() => {
        zoomAtPoint(dimensions.width / 2, dimensions.height / 2, ZOOM_STEP);
    }, [dimensions.width, dimensions.height, zoomAtPoint]);

    const handleZoomOut = useCallback(() => {
        zoomAtPoint(dimensions.width / 2, dimensions.height / 2, 1 / ZOOM_STEP);
    }, [dimensions.width, dimensions.height, zoomAtPoint]);

    // Node / link hover handlers
    const handleNodeEnter = useCallback((node: GraphNode) => {
        setHoveredNode(node);
    }, []);

    const handleNodeLeave = useCallback(() => {
        setHoveredNode(null);
    }, []);

    const handleLinkEnter = useCallback((link: GraphLink) => {
        setHoveredLink(link);
    }, []);

    const handleLinkLeave = useCallback(() => {
        setHoveredLink(null);
    }, []);

    // Node detail drawer
    const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);

    const handleNodeClick = useCallback((node: GraphNode) => {
        setSelectedNode((previous) => (previous?.id === node.id ? null : node));
    }, []);

    const handleDrawerClose = useCallback(() => {
        setSelectedNode(null);
    }, []);

    // Find connections for selected node
    const selectedNodeConnections = useMemo(() => {
        if (!selectedNode) {
            return { incoming: [], outgoing: [] };
        }

        const outgoing = links
            .filter((link) => (link.source as GraphNode).id === selectedNode.id)
            .map((link) => ({
                node: link.target as GraphNode,
                relationship: link.relationshipType,
            }));

        const incoming = links
            .filter((link) => (link.target as GraphNode).id === selectedNode.id)
            .map((link) => ({
                node: link.source as GraphNode,
                relationship: link.relationshipType,
            }));

        return { incoming, outgoing };
    }, [selectedNode, links]);

    if (!data.data.length) {
        return null;
    }

    const svgTransform = `translate(${transform.x}, ${transform.y}) scale(${transform.scale})`;

    return (
        <div>
            {/* Entity type filter toggles */}
            <ToggleGroup
                className="mb-3 flex flex-wrap justify-start gap-1"
                onValueChange={setVisibleTypes}
                type="multiple"
                value={activeVisibleTypes}
            >
                {entityTypes.map((type) => (
                    <ToggleGroupItem
                        className="gap-1.5 px-2.5 text-xs"
                        key={type}
                        size="sm"
                        value={type}
                        variant="outline"
                    >
                        <span
                            className="inline-block size-2.5 rounded-full"
                            style={{ backgroundColor: getNodeColor(type) }}
                        />
                        {type}
                    </ToggleGroupItem>
                ))}
            </ToggleGroup>

            {/* Graph SVG */}
            <div
                className={`bg-background relative h-[500px] w-full overflow-hidden rounded-lg border ${isPanning ? 'cursor-grabbing' : 'cursor-grab'}`}
                onPointerDown={handlePointerDown}
                onPointerMove={handlePointerMove}
                onPointerUp={handlePointerUp}
                ref={containerRef}
            >
                <svg
                    className="size-full"
                    viewBox={`0 0 ${dimensions.width} ${dimensions.height}`}
                >
                    <defs>
                        <marker
                            id="arrowhead"
                            markerHeight="6"
                            markerWidth="8"
                            orient="auto"
                            refX="8"
                            refY="3"
                        >
                            <path
                                className="fill-muted-foreground/40"
                                d="M0,0 L0,6 L8,3 z"
                            />
                        </marker>
                        <marker
                            id="arrowhead-active"
                            markerHeight="6"
                            markerWidth="8"
                            orient="auto"
                            refX="8"
                            refY="3"
                        >
                            <path
                                className="fill-foreground"
                                d="M0,0 L0,6 L8,3 z"
                            />
                        </marker>
                    </defs>

                    {/* Transformed group: all graph content */}
                    <g transform={svgTransform}>
                        {/* Edges (bezier curves for hierarchical layout) */}
                        <g>
                            {visibleLinks.map((link, index) => {
                                const source = link.source as GraphNode;
                                const target = link.target as GraphNode;

                                if (source.x == null || source.y == null || target.x == null || target.y == null) {
                                    return null;
                                }

                                // Start/end at node edges, not centers
                                const startX = source.x + NODE_HALF_WIDTH;
                                const endX = target.x - NODE_HALF_WIDTH;
                                const midX = (startX + endX) / 2;
                                const midY = (source.y + target.y) / 2;

                                // Horizontal bezier control offset for smooth curves
                                const edgeDeltaX = Math.abs(endX - startX);
                                const curveOffset = Math.max(edgeDeltaX * 0.4, 30);

                                const pathData = `M ${startX} ${source.y} C ${startX + curveOffset} ${source.y}, ${endX - curveOffset} ${target.y}, ${endX} ${target.y}`;

                                const isHighlighted =
                                    hoveredLink === link ||
                                    hoveredNode?.id === source.id ||
                                    hoveredNode?.id === target.id;

                                return (
                                    <g
                                        key={index}
                                        onMouseEnter={() => handleLinkEnter(link)}
                                        onMouseLeave={handleLinkLeave}
                                    >
                                        <path
                                            d={pathData}
                                            fill="none"
                                            markerEnd={isHighlighted ? 'url(#arrowhead-active)' : 'url(#arrowhead)'}
                                            stroke={isHighlighted ? 'currentColor' : 'var(--color-muted-foreground)'}
                                            strokeOpacity={isHighlighted ? 0.8 : 0.25}
                                            strokeWidth={isHighlighted ? 2 : 1}
                                        />
                                        {/* Invisible wider path for easier hovering */}
                                        <path
                                            d={pathData}
                                            fill="none"
                                            stroke="transparent"
                                            strokeWidth={12}
                                        />
                                        {/* Edge label on hover */}
                                        {isHighlighted && (
                                            <text
                                                className="fill-foreground pointer-events-none text-[10px] select-none"
                                                textAnchor="middle"
                                                x={midX}
                                                y={midY - 6}
                                            >
                                                {link.relationshipType.replaceAll('_', ' ')}
                                            </text>
                                        )}
                                    </g>
                                );
                            })}
                        </g>

                        {/* Nodes */}
                        <g>
                            {visibleNodes.map((node) => {
                                if (node.x == null || node.y == null) {
                                    return null;
                                }

                                const color = getNodeColor(node.entityType);
                                const isHighlighted =
                                    hoveredNode?.id === node.id ||
                                    (hoveredLink &&
                                        ((hoveredLink.source as GraphNode).id === node.id ||
                                            (hoveredLink.target as GraphNode).id === node.id));

                                return (
                                    <g
                                        className="cursor-pointer"
                                        key={node.id}
                                        onClick={() => handleNodeClick(node)}
                                        onMouseEnter={() => handleNodeEnter(node)}
                                        onMouseLeave={handleNodeLeave}
                                        transform={`translate(${node.x}, ${node.y})`}
                                    >
                                        {/* Glow outline on hover */}
                                        {isHighlighted && (
                                            <rect
                                                fill="none"
                                                height={NODE_HEIGHT + 10}
                                                opacity={0.3}
                                                rx={NODE_BORDER_RADIUS + 2}
                                                ry={NODE_BORDER_RADIUS + 2}
                                                stroke={color}
                                                strokeWidth={3}
                                                width={NODE_WIDTH + 10}
                                                x={-(NODE_WIDTH + 10) / 2}
                                                y={-(NODE_HEIGHT + 10) / 2}
                                            />
                                        )}

                                        {/* Node rectangle */}
                                        <rect
                                            fill={color}
                                            fillOpacity={0.15}
                                            height={NODE_HEIGHT}
                                            rx={NODE_BORDER_RADIUS}
                                            ry={NODE_BORDER_RADIUS}
                                            stroke={color}
                                            strokeWidth={isHighlighted ? 2.5 : 1.5}
                                            width={NODE_WIDTH}
                                            x={-NODE_HALF_WIDTH}
                                            y={-NODE_HALF_HEIGHT}
                                        />

                                        {/* Entity type */}
                                        <text
                                            className="pointer-events-none text-[10px] font-bold uppercase select-none"
                                            dominantBaseline="central"
                                            fill={color}
                                            textAnchor="middle"
                                            y={-6}
                                        >
                                            {truncateLabel(node.name, 18)}
                                        </text>

                                        {/* Node name */}
                                        <text
                                            className="pointer-events-none text-[9px] select-none"
                                            dominantBaseline="central"
                                            fill={color}
                                            textAnchor="middle"
                                            y={8}
                                        >
                                            {node.entityType}
                                        </text>
                                    </g>
                                );
                            })}
                        </g>
                    </g>
                </svg>

                {/* Zoom controls — stop pointer events from triggering pan */}
                <div
                    className="absolute top-3 right-3 flex flex-col gap-1"
                    onPointerDown={(event) => event.stopPropagation()}
                >
                    <Button
                        className="size-8"
                        onClick={handleZoomIn}
                        size="icon"
                        title="Zoom in"
                        variant="outline"
                    >
                        <Plus className="size-4" />
                    </Button>
                    <Button
                        className="size-8"
                        onClick={handleZoomOut}
                        size="icon"
                        title="Zoom out"
                        variant="outline"
                    >
                        <Minus className="size-4" />
                    </Button>
                    <Button
                        className="size-8"
                        onClick={handleFitToView}
                        size="icon"
                        title="Fit to view"
                        variant="outline"
                    >
                        <Maximize className="size-4" />
                    </Button>
                </div>

                {/* Zoom level indicator */}
                <div className="text-muted-foreground pointer-events-none absolute right-3 bottom-3 text-[10px]">
                    {Math.round(transform.scale * 100)}%
                </div>

                {/* Tooltip */}
                {hoveredNode && (
                    <div className="bg-popover text-popover-foreground pointer-events-none absolute bottom-3 left-3 max-w-sm rounded-md border px-3 py-2 text-xs shadow-md">
                        <div className="flex items-center gap-1.5">
                            <span
                                className="inline-block size-2 shrink-0 rounded-full"
                                style={{ backgroundColor: getNodeColor(hoveredNode.entityType) }}
                            />
                            <span className="font-semibold">{hoveredNode.entityType}</span>
                            <span className="text-muted-foreground">·</span>
                            <span className="truncate font-mono">{hoveredNode.name}</span>
                        </div>
                        {hoveredNode.summary && (
                            <p className="text-muted-foreground mt-1 line-clamp-3">{hoveredNode.summary}</p>
                        )}
                        {hoveredNode.description && hoveredNode.description !== hoveredNode.summary && (
                            <p className="text-muted-foreground/70 mt-1 line-clamp-2 italic">
                                {hoveredNode.description}
                            </p>
                        )}
                    </div>
                )}
            </div>
            {/* Node detail sheet */}
            <Sheet
                modal={false}
                onOpenChange={(open) => {
                    if (!open) {
                        handleDrawerClose();
                    }
                }}
                open={!!selectedNode}
            >
                <SheetContent
                    className="flex w-[420px] flex-col gap-0 p-0 sm:max-w-[420px]"
                    onInteractOutside={(event) => event.preventDefault()}
                    overlay={false}
                >
                    {selectedNode && (
                        <>
                            <SheetHeader className="gap-1 p-4 pb-3">
                                <div className="flex items-center gap-2">
                                    <span
                                        className="inline-block size-3 shrink-0 rounded-full"
                                        style={{ backgroundColor: getNodeColor(selectedNode.entityType) }}
                                    />
                                    <Badge variant="outline">{selectedNode.entityType}</Badge>
                                </div>
                                <SheetTitle className="font-mono text-base">{selectedNode.name}</SheetTitle>
                                {selectedNode.summary && <SheetDescription>{selectedNode.summary}</SheetDescription>}
                            </SheetHeader>

                            <Separator />

                            <ScrollArea className="flex-1">
                                <div className="space-y-4 p-4">
                                    {/* Description */}
                                    {selectedNode.description && selectedNode.description !== selectedNode.summary && (
                                        <div>
                                            <h4 className="text-muted-foreground mb-1 text-xs font-semibold tracking-wide uppercase">
                                                Description
                                            </h4>
                                            <p className="text-sm">{selectedNode.description}</p>
                                        </div>
                                    )}

                                    {/* All properties */}
                                    <div>
                                        <h4 className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                                            Properties
                                        </h4>
                                        <div className="space-y-2">
                                            {Object.entries(selectedNode.properties)
                                                .filter(
                                                    ([key, value]) =>
                                                        value != null &&
                                                        !key.endsWith('_embedding') &&
                                                        key !== 'labels' &&
                                                        typeof value !== 'object',
                                                )
                                                .sort(([keyA], [keyB]) => keyA.localeCompare(keyB))
                                                .map(([key, value]) => (
                                                    <div
                                                        className="flex gap-2 text-sm"
                                                        key={key}
                                                    >
                                                        <span className="text-muted-foreground shrink-0 font-medium">
                                                            {key}
                                                        </span>
                                                        <span className="break-all">{String(value)}</span>
                                                    </div>
                                                ))}
                                        </div>
                                    </div>

                                    {/* Outgoing connections */}
                                    {!!selectedNodeConnections.outgoing.length && (
                                        <div>
                                            <h4 className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                                                Outgoing ({selectedNodeConnections.outgoing.length})
                                            </h4>
                                            <div className="space-y-1.5">
                                                {selectedNodeConnections.outgoing.map((connection, index) => (
                                                    <button
                                                        className="hover:bg-muted flex w-full items-center gap-2 rounded-md border px-2.5 py-1.5 text-left text-xs transition-colors"
                                                        key={index}
                                                        onClick={() => setSelectedNode(connection.node)}
                                                        type="button"
                                                    >
                                                        <span
                                                            className="inline-block size-2 shrink-0 rounded-full"
                                                            style={{
                                                                backgroundColor: getNodeColor(
                                                                    connection.node.entityType,
                                                                ),
                                                            }}
                                                        />
                                                        <span className="text-muted-foreground">
                                                            {connection.relationship.replaceAll('_', ' ')}
                                                        </span>
                                                        <span className="ml-auto truncate font-mono font-medium">
                                                            {connection.node.name}
                                                        </span>
                                                    </button>
                                                ))}
                                            </div>
                                        </div>
                                    )}

                                    {/* Incoming connections */}
                                    {!!selectedNodeConnections.incoming.length && (
                                        <div>
                                            <h4 className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                                                Incoming ({selectedNodeConnections.incoming.length})
                                            </h4>
                                            <div className="space-y-1.5">
                                                {selectedNodeConnections.incoming.map((connection, index) => (
                                                    <button
                                                        className="hover:bg-muted flex w-full items-center gap-2 rounded-md border px-2.5 py-1.5 text-left text-xs transition-colors"
                                                        key={index}
                                                        onClick={() => setSelectedNode(connection.node)}
                                                        type="button"
                                                    >
                                                        <span
                                                            className="inline-block size-2 shrink-0 rounded-full"
                                                            style={{
                                                                backgroundColor: getNodeColor(
                                                                    connection.node.entityType,
                                                                ),
                                                            }}
                                                        />
                                                        <span className="truncate font-mono font-medium">
                                                            {connection.node.name}
                                                        </span>
                                                        <span className="text-muted-foreground ml-auto">
                                                            {connection.relationship.replaceAll('_', ' ')}
                                                        </span>
                                                    </button>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            </ScrollArea>
                        </>
                    )}
                </SheetContent>
            </Sheet>
        </div>
    );
};

export default FlowGraph;
